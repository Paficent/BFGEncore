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
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"paficent/bfg/db"
)

type config struct {
	Key           string `json:"key"`
	IV            string `json:"iv"`
	MaxPlayers    int    `json:"max_players"`
	ServerIP      string `json:"server_ip"`
	GameAddr      string `json:"game_addr"`
	AuthAddr      string `json:"auth_addr"`
	DBPath        string `json:"db_path"`
	SavePath      string `json:"save_path"`
	DLCPath       string `json:"dlc_path"`
	UsersPath     string `json:"users_path"`
	LogPath       string `json:"log_path"`
	ConstantsPath string `json:"constants_path"`
	RefreshLog    bool   `json:"refresh_log"`
	Debug         bool   `json:"debug"`
}

func defaults() config {
	return config{
		MaxPlayers:    200,
		ServerIP:      "127.0.0.1",
		GameAddr:      "0.0.0.0:9933",
		AuthAddr:      "127.0.0.1:900",
		DBPath:        "./db",
		SavePath:      "./players.db",
		DLCPath:       "./dlc",
		UsersPath:     "./auth_users.json",
		LogPath:       "./server.log",
		ConstantsPath: "./constants.json",
		RefreshLog:    true,
	}
}

const secretAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

func genSecret(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	for i := range b {
		b[i] = secretAlphabet[b[i]&0x3f]
	}
	return string(b)
}

func writeConfig(path string, c config) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o600); err != nil {
		return err
	}
	return ensureConstants(path)
}

func ensureConstants(configPath string) error {
	path := filepath.Join(filepath.Dir(configPath), "constants.json")
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	b, err := json.MarshalIndent(db.DefaultConstants(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o600)
}

func atoiOr(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}

func truthy(s string) bool {
	switch s {
	case "1", "t", "true", "y", "yes", "on":
		return true
	}
	return false
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
