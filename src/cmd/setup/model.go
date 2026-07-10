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
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type step int

const (
	stepWelcome step = iota
	stepConfig
	stepDLC
	stepPatch
	stepReview
	stepApplying
	stepDone
)

var (
	dlcOptions   = []string{"Local folder or .ZIP", "Download from a URL", "Skip"}
	patchOptions = []string{"Patch binary", "Skip"}
)

type field struct {
	key    string
	name   string
	input  textinput.Model
	secret bool
}

type appliedMsg struct{ notes []string }

type model struct {
	step   step
	out    string
	width  int
	sp     spinner.Model
	fields []field
	focus  int
	cfg    config

	dlcCursor int
	dlcInput  textinput.Model
	dlcTyping bool
	dlc       dlcSource

	patchCursor int
	patchInput  textinput.Model
	patchTyping bool
	patchBin    string

	notes []string
}

func newModel(out string) model {
	d := defaults()
	mk := func(key, name, val string, secret bool) field {
		ti := textinput.New()
		ti.Prompt = ""
		ti.CharLimit = 256
		ti.Width = 44
		ti.SetValue(val)
		return field{key: key, name: name, input: ti, secret: secret}
	}
	fields := []field{
		mk("key", "Encryption key", genSecret(16), true),
		mk("iv", "Encryption IV", genSecret(16), true),
		mk("max_players", "Max players", strconv.Itoa(d.MaxPlayers), false),
		mk("server_ip", "Server IP", d.ServerIP, false),
		mk("game_addr", "Game address", d.GameAddr, false),
		mk("auth_addr", "Auth address", d.AuthAddr, false),
		mk("db_path", "DB path", d.DBPath, false),
		mk("save_path", "Save path", d.SavePath, false),
		mk("dlc_path", "DLC path", d.DLCPath, false),
		mk("users_path", "Users path", d.UsersPath, false),
		mk("log_path", "Log path", d.LogPath, false),
		mk("refresh_log", "Refresh log on start", boolStr(d.RefreshLog), false),
		mk("debug", "Debug logging", boolStr(d.Debug), false),
	}
	fields[0].input.Focus()

	di := textinput.New()
	di.Prompt = "› "
	di.CharLimit = 512
	di.Width = 48
	pi := textinput.New()
	pi.Prompt = "› "
	pi.CharLimit = 512
	pi.Width = 48

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = spinStyle

	return model{step: stepWelcome, out: out, sp: sp, fields: fields, dlcInput: di, patchInput: pi}
}

func (m model) Init() tea.Cmd { return textinput.Blink }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case appliedMsg:
		m.notes = msg.notes
		m.step = stepDone
		return m, nil
	}
	switch m.step {
	case stepWelcome:
		return m.updateWelcome(msg)
	case stepConfig:
		return m.updateConfig(msg)
	case stepDLC:
		return m.updateDLC(msg)
	case stepPatch:
		return m.updatePatch(msg)
	case stepReview:
		return m.updateReview(msg)
	case stepApplying:
		var cmd tea.Cmd
		m.sp, cmd = m.sp.Update(msg)
		return m, cmd
	case stepDone:
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) updateWelcome(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "q", "esc":
			return m, tea.Quit
		case "enter":
			m.step = stepConfig
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m model) updateConfig(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "esc":
			return m, tea.Quit
		case "tab", "down":
			m.focusField(m.focus + 1)
			return m, textinput.Blink
		case "shift+tab", "up":
			m.focusField(m.focus - 1)
			return m, textinput.Blink
		case "ctrl+r":
			if m.fields[m.focus].secret {
				m.fields[m.focus].input.SetValue(genSecret(16))
			}
			return m, nil
		case "enter":
			if m.focus == len(m.fields)-1 {
				m.collect()
				m.step = stepDLC
				return m, nil
			}
			m.focusField(m.focus + 1)
			return m, textinput.Blink
		}
	}
	var cmd tea.Cmd
	m.fields[m.focus].input, cmd = m.fields[m.focus].input.Update(msg)
	return m, cmd
}

func (m *model) focusField(i int) {
	if i < 0 {
		i = len(m.fields) - 1
	}
	if i >= len(m.fields) {
		i = 0
	}
	m.fields[m.focus].input.Blur()
	m.focus = i
	m.fields[m.focus].input.Focus()
}

func (m *model) collect() {
	c := defaults()
	for _, f := range m.fields {
		v := strings.TrimSpace(f.input.Value())
		switch f.key {
		case "key":
			c.Key = v
		case "iv":
			c.IV = v
		case "max_players":
			c.MaxPlayers = atoiOr(v, c.MaxPlayers)
		case "server_ip":
			c.ServerIP = v
		case "game_addr":
			c.GameAddr = v
		case "auth_addr":
			c.AuthAddr = v
		case "db_path":
			c.DBPath = v
		case "save_path":
			c.SavePath = v
		case "dlc_path":
			c.DLCPath = v
		case "users_path":
			c.UsersPath = v
		case "log_path":
			c.LogPath = v
		case "refresh_log":
			c.RefreshLog = truthy(v)
		case "debug":
			c.Debug = truthy(v)
		}
	}
	m.cfg = c
}

func (m model) updateDLC(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if m.dlcTyping {
		if ok {
			switch k.String() {
			case "esc":
				m.dlcTyping = false
				m.dlcInput.Blur()
				return m, nil
			case "enter":
				ref := strings.TrimSpace(m.dlcInput.Value())
				if ref != "" {
					kind := "url"
					if m.dlcCursor == 1 {
						kind = "local"
					}
					m.dlc = dlcSource{kind: kind, ref: ref}
				}
				m.dlcTyping = false
				m.dlcInput.Blur()
				m.step = stepPatch
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.dlcInput, cmd = m.dlcInput.Update(msg)
		return m, cmd
	}
	if ok {
		switch k.String() {
		case "esc":
			return m, tea.Quit
		case "up", "k":
			if m.dlcCursor > 0 {
				m.dlcCursor--
			}
		case "down", "j":
			if m.dlcCursor < len(dlcOptions)-1 {
				m.dlcCursor++
			}
		case "enter":
			if m.dlcCursor == 0 {
				m.dlc = dlcSource{}
				m.step = stepPatch
				return m, nil
			}
			m.dlcTyping = true
			m.dlcInput.SetValue("")
			if m.dlcCursor == 1 {
				m.dlcInput.Placeholder = "path to a .zip or a directory"
			} else {
				m.dlcInput.Placeholder = "https://example.com/dlc.zip"
			}
			m.dlcInput.Focus()
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m model) updatePatch(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if m.patchTyping {
		if ok {
			switch k.String() {
			case "esc":
				m.patchTyping = false
				m.patchInput.Blur()
				return m, nil
			case "enter":
				m.patchBin = strings.TrimSpace(m.patchInput.Value())
				m.patchTyping = false
				m.patchInput.Blur()
				m.step = stepReview
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.patchInput, cmd = m.patchInput.Update(msg)
		return m, cmd
	}
	if ok {
		switch k.String() {
		case "esc":
			return m, tea.Quit
		case "up", "k":
			if m.patchCursor > 0 {
				m.patchCursor--
			}
		case "down", "j":
			if m.patchCursor < len(patchOptions)-1 {
				m.patchCursor++
			}
		case "enter":
			if m.patchCursor == 0 {
				m.patchBin = ""
				m.step = stepReview
				return m, nil
			}
			m.patchTyping = true
			m.patchInput.SetValue("")
			m.patchInput.Placeholder = "path to MySingingMonsters_SDK.exe"
			m.patchInput.Focus()
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m model) updateReview(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "esc":
			return m, tea.Quit
		case "enter":
			m.step = stepApplying
			return m, tea.Batch(m.sp.Tick, applyCmd(m.out, m.cfg, m.dlc, m.patchBin))
		}
	}
	return m, nil
}

func applyCmd(out string, cfg config, dlc dlcSource, patchBin string) tea.Cmd {
	return func() tea.Msg {
		var notes []string
		if err := writeConfig(out, cfg); err != nil {
			notes = append(notes, "config: "+err.Error())
		} else {
			notes = append(notes, "config: wrote "+out)
		}

		dir, err := binDir()
		switch {
		case dlc.kind == "":
			notes = append(notes, "dlc: skipped")
		case err != nil:
			notes = append(notes, "dlc: "+err.Error())
		default:
			if err := importDLC(dlc, dir); err != nil {
				notes = append(notes, "dlc: "+err.Error())
			} else {
				notes = append(notes, "dlc: imported into "+dir)
			}
		}

		if patchBin == "" {
			notes = append(notes, "patch: skipped")
		} else {
			url := clientAuthURL(cfg)
			if err := patchClient(patchBin, url); err != nil {
				notes = append(notes, "patch: "+err.Error())
			} else {
				notes = append(notes, "patch: pointed client at "+url)
			}
		}
		return appliedMsg{notes: notes}
	}
}
