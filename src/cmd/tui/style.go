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

package tui

import "github.com/charmbracelet/lipgloss"

var (
	Accent = lipgloss.AdaptiveColor{Light: "205", Dark: "212"}
	Subtle = lipgloss.AdaptiveColor{Light: "240", Dark: "244"}
	Good   = lipgloss.AdaptiveColor{Light: "29", Dark: "42"}
	Bad    = lipgloss.AdaptiveColor{Light: "160", Dark: "203"}
	Bright = lipgloss.AdaptiveColor{Light: "231", Dark: "231"}
	BarBG  = lipgloss.AdaptiveColor{Light: "252", Dark: "237"}
)

func Center(width, height int, content string) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
