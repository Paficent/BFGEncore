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
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"paficent/bfg/cmd/tui"
	"paficent/bfg/commands"
	"paficent/bfg/game"
)

const maxLogLines = 2000

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(tui.Bright).Background(tui.Accent)
	panelBox   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(tui.Subtle)
	inputBox   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(tui.Accent)
	panelTitle = lipgloss.NewStyle().Foreground(tui.Accent).Bold(true).Align(lipgloss.Center)

	promptStyle = lipgloss.NewStyle().Foreground(tui.Accent).Bold(true)
	cmdStyle    = lipgloss.NewStyle().Foreground(tui.Accent)
	errStyle    = lipgloss.NewStyle().Foreground(tui.Bad)
	respStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "238", Dark: "250"})
	logStyle    = lipgloss.NewStyle().Foreground(tui.Subtle)
	dimStyle    = lipgloss.NewStyle().Foreground(tui.Subtle)
)

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

type layout struct {
	leftW, rightW   int
	cmdsH, playersH int
	outH, inputH    int
}

func computeLayout(w, h int, leftRatio, topRatio float64) layout {
	contentH := h - 1
	if contentH < 6 {
		contentH = 6
	}
	leftW := int(float64(w) * leftRatio)
	if leftW > w-14 {
		leftW = w - 14
	}
	if leftW < 14 {
		leftW = 14
	}
	rightW := w - leftW
	if rightW < 14 {
		rightW = 14
	}
	cmdsH := int(float64(contentH) * topRatio)
	if cmdsH < 4 {
		cmdsH = 4
	}
	playersH := contentH - cmdsH
	if playersH < 4 {
		playersH = 4
		cmdsH = contentH - playersH
	}
	inputH := 3
	outH := contentH - inputH
	if outH < 3 {
		outH = 3
	}
	return layout{leftW: leftW, rightW: rightW, cmdsH: cmdsH, playersH: playersH, outH: outH, inputH: inputH}
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

type serverTUI struct {
	logs        *logStream
	mgr         *game.Manager
	cmds        *commands.Registry
	commandList []string
	stats       game.Stats
	online      []string
	buf         []string
	vp          viewport.Model
	input       textinput.Model
	w, h        int
	ready       bool
	leftRatio   float64
	topRatio    float64
	gameAddr    string
	authAddr    string
	start       time.Time
}

func newServerTUI(logs *logStream, mgr *game.Manager, cmds *commands.Registry, gameAddr, authAddr string) serverTUI {
	in := textinput.New()
	in.Placeholder = "type a command (try \"help\") ..."
	in.Prompt = promptStyle.Render("> ")
	in.Focus()

	list := make([]string, 0)
	for _, c := range cmds.Commands() {
		list = append(list, c.Usage)
	}

	return serverTUI{
		logs:        logs,
		mgr:         mgr,
		cmds:        cmds,
		commandList: list,
		input:       in,
		leftRatio:   0.32,
		topRatio:    0.65,
		gameAddr:    gameAddr,
		authAddr:    authAddr,
		start:       time.Now(),
	}
}

func (m serverTUI) Init() tea.Cmd {
	return tea.Batch(waitForLogs(m.logs.ch), tick(), textinput.Blink)
}

func (m serverTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		return m.onKey(msg)
	case logLinesMsg:
		m.pushLog(logStyle, msg...)
		return m, waitForLogs(m.logs.ch)
	case tickMsg:
		m.stats = m.mgr.Stats()
		m.online = m.mgr.OnlinePlayers()
		return m, tick()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m serverTUI) onKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "enter":
		line := strings.TrimSpace(m.input.Value())
		m.input.Reset()
		if line != "" {
			m.runCommand(line)
		}
		return m, nil
	case "pgup", "pgdown", "home", "end", "up", "down":
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(k)
		return m, cmd
	case "ctrl+left":
		m.leftRatio = clampF(m.leftRatio-0.03, 0.15, 0.6)
		m.applyLayout()
		return m, nil
	case "ctrl+right":
		m.leftRatio = clampF(m.leftRatio+0.03, 0.15, 0.6)
		m.applyLayout()
		return m, nil
	case "ctrl+up":
		m.topRatio = clampF(m.topRatio-0.05, 0.25, 0.8)
		m.applyLayout()
		return m, nil
	case "ctrl+down":
		m.topRatio = clampF(m.topRatio+0.05, 0.25, 0.8)
		m.applyLayout()
		return m, nil
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(k)
		return m, cmd
	}
}

func (m *serverTUI) runCommand(line string) {
	m.pushLog(cmdStyle, "> "+line)
	switch out, err := m.cmds.Run(line); {
	case err != nil:
		m.pushLog(errStyle, "! "+err.Error())
	case out != "":
		m.pushLog(respStyle, out)
	}
}

func (m *serverTUI) resize(w, h int) {
	m.w, m.h = w, h
	m.applyLayout()
}

func (m *serverTUI) applyLayout() {
	l := computeLayout(m.w, m.h, m.leftRatio, m.topRatio)
	vpW := l.rightW - 2
	vpH := l.outH - 3
	if vpW < 1 {
		vpW = 1
	}
	if vpH < 1 {
		vpH = 1
	}
	if !m.ready {
		m.vp = viewport.New(vpW, vpH)
		m.ready = true
	} else {
		m.vp.Width, m.vp.Height = vpW, vpH
	}
	m.input.Width = l.rightW - 2 - lipgloss.Width(m.input.Prompt) - 1
	if m.input.Width < 1 {
		m.input.Width = 1
	}
	m.vp.SetContent(m.renderLog())
	m.vp.GotoBottom()
}

func (m *serverTUI) pushLog(style lipgloss.Style, lines ...string) {
	for _, l := range lines {
		for _, ln := range strings.Split(l, "\n") {
			m.buf = append(m.buf, style.Render(ln))
		}
	}
	if len(m.buf) > maxLogLines {
		m.buf = m.buf[len(m.buf)-maxLogLines:]
	}
	if m.ready {
		m.vp.SetContent(m.renderLog())
		m.vp.GotoBottom()
	}
}

func (m serverTUI) renderLog() string {
	return strings.Join(m.buf, "\n")
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= w {
		return s
	}
	return string(r[:w])
}

func listBody(items []string, w, h int) string {
	if h < 1 {
		h = 1
	}
	rows := make([]string, 0, h)
	for _, it := range items {
		if len(rows) == h {
			break
		}
		rows = append(rows, truncate(it, w))
	}
	if len(items) > h && h > 0 {
		rows[h-1] = truncate("… +"+strconv.Itoa(len(items)-h+1)+" more", w)
	}
	return strings.Join(rows, "\n")
}

func (m serverTUI) playerList() []string {
	if len(m.online) == 0 {
		return []string{dimStyle.Render("(no players online)")}
	}
	return m.online
}

func panel(title string, ow, oh int, body string) string {
	iw, ih := ow-2, oh-2
	if iw < 1 {
		iw = 1
	}
	if ih < 1 {
		ih = 1
	}
	head := panelTitle.Width(iw).Render(title)
	return panelBox.Width(iw).Height(ih).Render(head + "\n" + body)
}

func (m serverTUI) titleBar() string {
	left := " Project Encore: BFG"
	up := time.Since(m.start).Round(time.Second).String()
	for _, right := range []string{
		"game " + m.gameAddr + " · up " + up + " · ctrl+arrows resize · ctrl+c quit ",
		"up " + up + " · ctrl+arrows resize · ctrl+c quit ",
		"ctrl+arrows resize · ctrl+c quit ",
		"ctrl+c quit ",
	} {
		gap := m.w - lipgloss.Width(left) - lipgloss.Width(right)
		if gap >= 1 {
			return titleStyle.Width(m.w).Render(left + strings.Repeat(" ", gap) + right)
		}
	}
	return titleStyle.Width(m.w).Render(left)
}

func (m serverTUI) View() string {
	if !m.ready {
		return "starting ..."
	}
	l := computeLayout(m.w, m.h, m.leftRatio, m.topRatio)

	cmdsPanel := panel("Available Commands", l.leftW, l.cmdsH, listBody(m.commandList, l.leftW-2, l.cmdsH-3))
	players := panel("Online Players", l.leftW, l.playersH, listBody(m.playerList(), l.leftW-2, l.playersH-3))
	output := panel("Server Output", l.rightW, l.outH, m.vp.View())
	entry := inputBox.Width(l.rightW - 2).Height(1).Render(m.input.View())

	left := lipgloss.JoinVertical(lipgloss.Left, cmdsPanel, players)
	right := lipgloss.JoinVertical(lipgloss.Left, output, entry)
	return lipgloss.JoinVertical(lipgloss.Left, m.titleBar(), lipgloss.JoinHorizontal(lipgloss.Top, left, right))
}
