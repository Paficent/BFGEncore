/* Project Encore: BFG - Localized Private Game Restoration Server
 * Copyright (C) 2026 Paficent <paficent@tutamail.com> & Contributors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"paficent/bfg/commands"
	"paficent/bfg/db"
	"paficent/bfg/game"
	"paficent/bfg/save"
)

func main() {
	configPath := flag.String("config", "./config.json", "path to config.json")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	serverAddr := cfg.GameAddr
	authAddr := cfg.AuthAddr
	dbDir := cfg.DBPath
	saveFile := cfg.SavePath
	dlcDir := cfg.DLCPath
	usersFile := cfg.UsersPath
	logFilePath := cfg.LogPath
	debugOn := cfg.Debug

	openFlags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if cfg.RefreshLog {
		openFlags |= os.O_TRUNC
	}
	logFile, err := os.OpenFile(logFilePath, openFlags, 0o644)
	if err != nil {
		log.Fatalf("open log file: %v", err)
	}
	defer logFile.Close()

	logs := newLogStream()
	log.SetOutput(io.MultiWriter(logFile, logs))
	log.Print("Project Encore: BFG — Copyright (C) 2026 Paficent & Contributors")

	database, err := db.Open(dbDir)
	if err != nil {
		fatal(logFile, "open game data: %v", err)
	}
	static := db.LoadStatic(database)
	constantsPath := cfg.ConstantsPath
	if constantsPath == "" {
		constantsPath = "./constants.json"
	}
	consts, err := db.LoadConstants(constantsPath)
	if err != nil {
		fatal(logFile, "load constants: %v", err)
	}
	static.Constants = consts
	log.Printf("static data: %d genes, %d islands, %d monsters, %d structures, %d levels, %d store items",
		static.Genes.Size(), static.Islands.Size(), static.Monsters.Size(),
		static.Structures.Size(), static.Levels.Size(), static.StoreItems.Size())

	store, err := save.Open(saveFile)
	if err != nil {
		fatal(logFile, "open player store: %v", err)
	}
	defer store.Close()

	mgr := game.New(static, store, debugOn)
	mgr.StartAutosave(60 * time.Second)
	cmds := commands.New(mgr)

	auth := newAuthServer(cfg, usersFile, dlcDir)
	go func() {
		log.Printf("auth/content listening on %s, dlc=%s users=%s", authAddr, dlcDir, usersFile)
		if err := http.ListenAndServe(authAddr, auth.handler()); err != nil {
			log.Printf("auth server stopped: %v", err)
		}
	}()

	go func() {
		log.Printf("SFS2X game server starting on %s", serverAddr)
		if err := mgr.Run(serverAddr); err != nil {
			log.Printf("game server stopped: %v", err)
		}
	}()

	if debugOn {
		log.Printf("debug packet logging is ON")
	}

	p := tea.NewProgram(
		newServerTUI(logs, mgr, cmds, serverAddr, authAddr),
		tea.WithAltScreen(),
	)

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		p.Quit()
	}()

	if _, err := p.Run(); err != nil {
		fatal(logFile, "console: %v", err)
	}

	log.SetOutput(io.MultiWriter(os.Stderr, logFile))
	log.Print("shutting down, saving players...")
	if _, err := mgr.SaveAll(); err != nil {
		log.Printf("save on shutdown failed: %v", err)
	}
}

func fatal(logFile *os.File, format string, args ...any) {
	log.SetOutput(io.MultiWriter(os.Stderr, logFile))
	log.Fatalf(format, args...)
}
