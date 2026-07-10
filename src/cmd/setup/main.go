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
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"paficent/bfg/cmd/tui"
)

var (
	accent = tui.Accent
	subtle = tui.Subtle
	good   = tui.Good
	bad    = tui.Bad

	titleStyle = lipgloss.NewStyle().Foreground(accent).Bold(true).MarginBottom(1)
	boxStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(1, 3)
	label      = lipgloss.NewStyle().Foreground(subtle)
	labelOn    = lipgloss.NewStyle().Foreground(accent).Bold(true)
	pick       = lipgloss.NewStyle().Foreground(accent).Bold(true)
	help       = lipgloss.NewStyle().Foreground(subtle).MarginTop(1)
	note       = lipgloss.NewStyle().Foreground(subtle).Italic(true)
	okText     = lipgloss.NewStyle().Foreground(good).Bold(true)
	badText    = lipgloss.NewStyle().Foreground(bad)
	spinStyle  = lipgloss.NewStyle().Foreground(accent)
)

func main() {
	out := flag.String("o", "config.json", "where to write the generated config")
	y := flag.Bool("y", false, "headless: accepts all defaults")
	flag.Parse()

	if *y {
		c := defaults()
		c.Key, c.IV = genSecret(16), genSecret(16)
		if err := writeConfig(*out, c); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Println("wrote", *out)
		return
	}

	if _, err := tea.NewProgram(newModel(*out), tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
