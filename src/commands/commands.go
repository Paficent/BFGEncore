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

// TODO: Abstract this into multiple files when it becomes too big to manage
package commands

import (
	"fmt"
	"sort"
	"strconv"
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
					fmt.Fprintf(&b, "%-14s %s\n", c.Name, c.Help)
				}
				return strings.TrimRight(b.String(), "\n"), nil
			},
		},
		{
			Name:  "save",
			Usage: "save",
			Help:  "save all loaded players now",
			Run: func(r *Registry, _ []string) (string, error) {
				n, err := r.mgr.SaveAll()
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("saved %d player(s)", n), nil
			},
		},
		{
			Name:  "set_level",
			Usage: "set_level <bbb_id> <level>",
			Help:  "set a player's level (1-30)",
			Run: func(r *Registry, args []string) (string, error) {
				if len(args) != 2 {
					return "", fmt.Errorf("usage: set_level <bbb_id> <level>")
				}
				p, err := r.player(args[0])
				if err != nil {
					return "", err
				}
				level, err := strconv.Atoi(args[1])
				if err != nil || level < 1 || level > 30 {
					return "", fmt.Errorf("level must be a number from 1 to 30")
				}
				p.Level = level
				return r.persist(p, fmt.Sprintf("set %s to level %d", label(p), level))
			},
		},
		giveCommand("give_coins", "coins", func(p *game.Player, n int64) bool { return p.AddProperties(n, 0, 0, 0, 0) }),
		giveCommand("give_diamonds", "diamonds", func(p *game.Player, n int64) bool { return p.AddProperties(0, n, 0, 0, 0) }),
		giveCommand("give_food", "food", func(p *game.Player, n int64) bool { return p.AddProperties(0, 0, n, 0, 0) }),
		giveCommand("give_xp", "xp", func(p *game.Player, n int64) bool { return p.AddProperties(0, 0, 0, n, 0) }),
		giveCommand("give_shards", "shards", func(p *game.Player, n int64) bool { return p.AddProperties(0, 0, 0, 0, n) }),
	}
}

func giveCommand(name, currency string, add func(*game.Player, int64) bool) *Command {
	return &Command{
		Name:  name,
		Usage: name + " <bbb_id> <amount>",
		Help:  "give " + currency + " to a player (negative to remove)",
		Run: func(r *Registry, args []string) (string, error) {
			if len(args) != 2 {
				return "", fmt.Errorf("usage: %s <bbb_id> <amount>", name)
			}
			p, err := r.player(args[0])
			if err != nil {
				return "", err
			}
			amount, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return "", fmt.Errorf("invalid amount %q", args[1])
			}
			if !add(p, amount) {
				return "", fmt.Errorf("%s would drop %s below zero", label(p), currency)
			}
			return r.persist(p, fmt.Sprintf("gave %s %d %s", label(p), amount, currency))
		},
	}
}

func (r *Registry) player(arg string) (*game.Player, error) {
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid bbb_id %q", arg)
	}
	p := r.mgr.Player(id)
	if p == nil {
		return nil, fmt.Errorf("no loaded player with bbb_id %d", id)
	}
	return p, nil
}

func (r *Registry) persist(p *game.Player, msg string) (string, error) {
	if err := r.mgr.SavePlayer(p); err != nil {
		return "", fmt.Errorf("%s (but saving failed: %v)", msg, err)
	}
	r.mgr.PushProperties(p)
	return msg, nil
}

func label(p *game.Player) string {
	if p.DisplayName != "" {
		return fmt.Sprintf("%s (%d)", p.DisplayName, p.BBBID)
	}
	return strconv.FormatInt(p.BBBID, 10)
}
