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
	dlcOptions   = []string{"Skip", "Local folder or .ZIP", "Download from a URL"}
	patchOptions = []string{"Skip", "Patch binary"}
)

type field struct {
	name   string
	input  textinput.Model
	secret bool
	apply  func(*config, string)
}

type appliedMsg struct{ notes []string }

type model struct {
	step   step
	out    string
	width  int
	height int
	sp     spinner.Model
	fields []field
	focus  int
	cfg    config

	dlcChooser   chooser
	patchChooser chooser
	dlc          dlcSource
	patchBin     string

	notes []string
}

func newModel(out string) model {
	d := defaults()
	mk := func(name, val string, secret bool, apply func(*config, string)) field {
		ti := textinput.New()
		ti.Prompt = ""
		ti.CharLimit = 256
		ti.Width = 44
		ti.SetValue(val)
		return field{name: name, input: ti, secret: secret, apply: apply}
	}
	fields := []field{
		mk("Encryption key", genSecret(16), true, func(c *config, v string) { c.Key = v }),
		mk("Encryption IV", genSecret(16), true, func(c *config, v string) { c.IV = v }),
		mk("Max players", strconv.Itoa(d.MaxPlayers), false, func(c *config, v string) { c.MaxPlayers = atoiOr(v, c.MaxPlayers) }),
		mk("Server IP", d.ServerIP, false, func(c *config, v string) { c.ServerIP = v }),
		mk("Game address", d.GameAddr, false, func(c *config, v string) { c.GameAddr = v }),
		mk("Auth address", d.AuthAddr, false, func(c *config, v string) { c.AuthAddr = v }),
		mk("DB path", d.DBPath, false, func(c *config, v string) { c.DBPath = v }),
		mk("Save path", d.SavePath, false, func(c *config, v string) { c.SavePath = v }),
		mk("DLC path", d.DLCPath, false, func(c *config, v string) { c.DLCPath = v }),
		mk("Users path", d.UsersPath, false, func(c *config, v string) { c.UsersPath = v }),
		mk("Log path", d.LogPath, false, func(c *config, v string) { c.LogPath = v }),
		mk("Refresh log on start", boolStr(d.RefreshLog), false, func(c *config, v string) { c.RefreshLog = truthy(v) }),
		mk("Debug logging", boolStr(d.Debug), false, func(c *config, v string) { c.Debug = truthy(v) }),
	}
	fields[0].input.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = spinStyle

	return model{
		step:   stepWelcome,
		out:    out,
		sp:     sp,
		fields: fields,
		dlcChooser: newChooser(dlcOptions, map[int]string{
			1: "path to a .zip or a directory",
			2: "https://example.com/dlc.zip",
		}),
		patchChooser: newChooser(patchOptions, map[int]string{
			1: "path to MySingingMonsters_SDK.exe",
		}),
	}
}

func (m model) Init() tea.Cmd { return textinput.Blink }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
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

func (m *model) resize(w, h int) {
	m.width, m.height = w, h
	iw := w - 16
	if iw < 20 {
		iw = 20
	}
	if iw > 60 {
		iw = 60
	}
	for i := range m.fields {
		m.fields[i].input.Width = iw
	}
	m.dlcChooser.input.Width = iw
	m.patchChooser.input.Width = iw
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
		f.apply(&c, strings.TrimSpace(f.input.Value()))
	}
	m.cfg = c
}

func (m model) updateDLC(msg tea.Msg) (tea.Model, tea.Cmd) {
	res, cmd := m.dlcChooser.handle(msg)
	switch res {
	case chooserQuit:
		return m, tea.Quit
	case chooserPicked:
		if m.dlcChooser.cursor == 0 {
			m.dlc = dlcSource{}
		} else if ref := m.dlcChooser.value(); ref != "" {
			kind := "url"
			if m.dlcChooser.cursor == 1 {
				kind = "local"
			}
			m.dlc = dlcSource{kind: kind, ref: ref}
		}
		m.step = stepPatch
		return m, nil
	}
	return m, cmd
}

func (m model) updatePatch(msg tea.Msg) (tea.Model, tea.Cmd) {
	res, cmd := m.patchChooser.handle(msg)
	switch res {
	case chooserQuit:
		return m, tea.Quit
	case chooserPicked:
		if m.patchChooser.cursor == 0 {
			m.patchBin = ""
		} else {
			m.patchBin = m.patchChooser.value()
		}
		m.step = stepReview
		return m, nil
	}
	return m, cmd
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
