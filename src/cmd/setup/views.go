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

	"paficent/bfg/cmd/tui"
)

func (m model) View() string {
	var body string
	switch m.step {
	case stepWelcome:
		body = m.viewWelcome()
	case stepConfig:
		body = m.viewConfig()
	case stepDLC:
		body = m.viewDLC()
	case stepPatch:
		body = m.viewPatch()
	case stepApplying:
		body = m.viewApplying()
	case stepReview:
		body = m.viewReview()
	case stepDone:
		body = m.viewDone()
	}
	card := titleStyle.Render("Project Encore: BFG — Setup Wizard") + "\n" + boxStyle.Render(body)
	if m.width > 0 && m.height > 0 {
		return tui.Center(m.width, m.height, card)
	}
	return card + "\n"
}

func (m model) viewWelcome() string {
	return strings.Join([]string{
		"Simple setup to get started running a private server for the",
		"Big Fish version of MSM\n",
		"This script does not download any game files, you'll have",
		"to provide those yourself",
		help.Render("enter=start | q=quit"),
	}, "\n")
}

func (m model) viewConfig() string {
	var b strings.Builder
	for i, f := range m.fields {
		if i == m.focus {
			b.WriteString(labelOn.Render("› "+f.name) + "\n")
		} else {
			b.WriteString(label.Render("  "+f.name) + "\n")
		}
		b.WriteString("    " + f.input.View() + "\n")
	}
	b.WriteString(help.Render("tab/↑↓=move | ctrl+r=reroll secret | enter=next | esc=quit"))
	return b.String()
}

func (m model) viewDLC() string {
	if m.dlcChooser.typing {
		return strings.Join([]string{
			"Point at your own copy:",
			"",
			"  " + m.dlcChooser.input.View(),
			help.Render("enter=confirm | esc=back"),
		}, "\n")
	}
	return menu("DLC assets (optional)", m.dlcChooser.options, m.dlcChooser.cursor, "")
}

func (m model) viewPatch() string {
	if m.patchChooser.typing {
		return strings.Join([]string{
			"Binary to patch:",
			"",
			"  " + m.patchChooser.input.View(),
			help.Render("enter=confirm | esc=back"),
		}, "\n")
	}
	return menu("Client patch (optional)", m.patchChooser.options, m.patchChooser.cursor, "")
}

func (m model) viewReview() string {
	dlc := "skipped"
	if m.dlc.kind != "" {
		dlc = m.dlc.kind + "  " + m.dlc.ref
	}
	patch := "skipped"
	if m.patchBin != "" {
		patch = m.patchBin
	}
	return strings.Join([]string{
		"Ready to apply:",
		"",
		label.Render("  config   ") + m.out,
		label.Render("  dlc      ") + dlc,
		label.Render("  patch    ") + patch,
		"",
		help.Render("enter=apply | esc=quit"),
	}, "\n")
}

func (m model) viewApplying() string {
	return m.sp.View() + " working…"
}

func (m model) viewDone() string {
	var b strings.Builder
	b.WriteString(okText.Render("Done.") + "\n\n")
	for _, n := range m.notes {
		mark := okText.Render("ok")
		if strings.Contains(n, "not wired up") || strings.Contains(n, "error") || strings.Contains(n, "no such") {
			mark = badText.Render("··")
		}
		b.WriteString("  " + mark + "  " + n + "\n")
	}
	b.WriteString("\n" + help.Render("press any key to exit"))
	return b.String()
}

func menu(title string, opts []string, cursor int, footer string) string {
	var b strings.Builder
	b.WriteString(title + "\n\n")
	for i, opt := range opts {
		if i == cursor {
			b.WriteString(pick.Render("› "+opt) + "\n")
		} else {
			b.WriteString(label.Render("  "+opt) + "\n")
		}
	}
	b.WriteString("\n" + note.Render(footer) + "\n")
	b.WriteString(help.Render("↑↓=move | enter=select | esc=quit"))
	return b.String()
}
