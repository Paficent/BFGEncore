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
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type logStream struct{ ch chan string }

func newLogStream() *logStream { return &logStream{ch: make(chan string, 1024)} }

func (l *logStream) Write(p []byte) (int, error) {
	for _, line := range strings.Split(strings.TrimRight(string(p), "\n"), "\n") {
		select {
		case l.ch <- line:
		default:
		}
	}
	return len(p), nil
}

type logLinesMsg []string

func waitForLogs(ch chan string) tea.Cmd {
	return func() tea.Msg {
		lines := []string{<-ch}
		for {
			select {
			case l := <-ch:
				lines = append(lines, l)
			default:
				return logLinesMsg(lines)
			}
		}
	}
}
