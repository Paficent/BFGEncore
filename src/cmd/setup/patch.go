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
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const officialAuthURL = "https://bf-auth.bbbgame.net/"

func clientAuthURL(cfg config) string {
	port := "900"
	if _, p, err := net.SplitHostPort(cfg.AuthAddr); err == nil && p != "" {
		port = p
	}
	return "http://" + net.JoinHostPort(cfg.ServerIP, port) + "/"
}

func backupPath(bin string) string {
	ext := filepath.Ext(bin)
	return strings.TrimSuffix(bin, ext) + "_BACKUP" + ext
}

func patchClient(bin, clientURL string) error {
	if bin == "" {
		return nil
	}
	info, err := os.Stat(bin)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory, expected the game executable", bin)
	}

	find := []byte(officialAuthURL)
	if len(clientURL) > len(find) {
		return fmt.Errorf("server URL %q is %d bytes, but only %d fit in place; use a shorter address (an IP with a short port)", clientURL, len(clientURL), len(find))
	}
	repl := make([]byte, len(find))
	copy(repl, clientURL) // string terminator + padding

	backup := backupPath(bin)
	data, err := os.ReadFile(bin)
	if err != nil {
		return err
	}

	switch {
	case bytes.Contains(data, find):
	case bytes.Contains(data, repl):
		return nil
	default:
		restored, rerr := os.ReadFile(backup)
		if rerr != nil || !bytes.Contains(restored, find) {
			return fmt.Errorf("couldn't find %s in %s — is this the Big Fish MSM v1.2.9 executable? if it was patched before, restore %s and try again", officialAuthURL, bin, filepath.Base(backup))
		}
		data = restored
	}

	if _, statErr := os.Stat(backup); os.IsNotExist(statErr) {
		if err := os.WriteFile(backup, data, info.Mode()); err != nil {
			return fmt.Errorf("write backup %s: %w", backup, err)
		}
	}

	patched := bytes.ReplaceAll(data, find, repl)
	return os.WriteFile(bin, patched, info.Mode())
}
