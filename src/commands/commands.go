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

package commands

import (
	"fmt"
	"sort"
	"strings"

	"paficent/bfg/game"
)

type Command struct {
	Name  string
	Usage string
	Help  string
	Run   func(r *Registry, args []string) (string, error)
}

type Registry struct {
	mgr  *game.Manager
	cmds map[string]*Command
}

func New(mgr *game.Manager) *Registry {
	r := &Registry{mgr: mgr, cmds: map[string]*Command{}}
	r.Register(builtins()...)
	return r
}

func (r *Registry) Register(cmds ...*Command) {
	for _, c := range cmds {
		r.cmds[c.Name] = c
	}
}

func (r *Registry) Manager() *game.Manager { return r.mgr }

func (r *Registry) Commands() []*Command {
	out := make([]*Command, 0, len(r.cmds))
	for _, c := range r.cmds {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (r *Registry) Run(line string) (string, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", nil
	}
	name, args := fields[0], fields[1:]
	cmd, ok := r.cmds[name]
	if !ok {
		return "", fmt.Errorf("unknown command %q (try \"help\")", name)
	}
	return cmd.Run(r, args)
}

func builtins() []*Command {
	return []*Command{
		{
			Name:  "help",
			Usage: "help",
			Help:  "list available commands",
			Run: func(r *Registry, _ []string) (string, error) {
				var b strings.Builder
				for _, c := range r.Commands() {
					fmt.Fprintf(&b, "%-10s %s\n", c.Name, c.Help)
				}
				return strings.TrimRight(b.String(), "\n"), nil
			},
		},
		todo("give", "give <bbb_id> <coins|diamonds|food|shards|xp> <amount>", "grant currency to a player"),
		todo("setlevel", "setlevel <bbb_id> <level>", "set a player's level"),
		todo("save", "save", "force an immediate save of all loaded players"),
	}
}

// lowkey everything rn
func todo(name, usage, help string) *Command {
	return &Command{
		Name:  name,
		Usage: usage,
		Help:  help + " (not implemented yet)",
		Run: func(r *Registry, _ []string) (string, error) {
			return "", fmt.Errorf("%q is not implemented yet — usage: %s", name, usage)
		},
	}
}
