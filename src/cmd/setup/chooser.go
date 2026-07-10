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

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type chooserResult int

const (
	chooserNone chooserResult = iota
	chooserPicked
	chooserQuit
)

type chooser struct {
	options      []string
	placeholders map[int]string
	cursor       int
	typing       bool
	input        textinput.Model
}

func newChooser(options []string, placeholders map[int]string) chooser {
	in := textinput.New()
	in.Prompt = "› "
	in.CharLimit = 512
	in.Width = 48
	return chooser{options: options, placeholders: placeholders, input: in}
}

func (c *chooser) handle(msg tea.Msg) (chooserResult, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if c.typing {
		if ok {
			switch k.String() {
			case "esc":
				c.typing = false
				c.input.Blur()
				return chooserNone, nil
			case "enter":
				c.typing = false
				c.input.Blur()
				return chooserPicked, nil
			}
		}
		var cmd tea.Cmd
		c.input, cmd = c.input.Update(msg)
		return chooserNone, cmd
	}
	if !ok {
		return chooserNone, nil
	}
	switch k.String() {
	case "esc":
		return chooserQuit, nil
	case "up", "k":
		if c.cursor > 0 {
			c.cursor--
		}
	case "down", "j":
		if c.cursor < len(c.options)-1 {
			c.cursor++
		}
	case "enter":
		if ph, needsText := c.placeholders[c.cursor]; needsText {
			c.typing = true
			c.input.SetValue("")
			c.input.Placeholder = ph
			c.input.Focus()
			return chooserNone, textinput.Blink
		}
		return chooserPicked, nil
	}
	return chooserNone, nil
}

func (c chooser) value() string { return strings.TrimSpace(c.input.Value()) }
