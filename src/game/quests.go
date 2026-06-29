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

package game

import (
	"log"
	"paficent/bfg/db"

	"github.com/Paficent/GoFox2X/data"
)

const (
	questActive    = "active"
	questComplete  = "complete"
	questCollected = "collected"
)

func (m *Manager) ensureQuests(p *Player) {
	if p.QuestStatus == nil {
		p.QuestStatus = map[int]string{}
	}
	if p.QuestCounters == nil {
		p.QuestCounters = map[string]int{}
	}
	for _, def := range m.Static.QuestDefs {
		if def.Initial {
			if _, ok := p.QuestStatus[def.ID]; !ok {
				p.QuestStatus[def.ID] = questActive
			}
		}
	}
}

func (p *Player) bumpCounter(key string) {
	if p.QuestCounters == nil {
		p.QuestCounters = map[string]int{}
	}
	p.QuestCounters[key]++
}

func (m *Manager) evaluateQuests(p *Player) []*data.GFSObject {
	m.ensureQuests(p)
	var changed []*data.GFSObject
	for {
		progressed := false
		for _, def := range m.Static.QuestDefs {
			if p.QuestStatus[def.ID] != questActive {
				continue
			}
			if !m.questGoalsMet(p, def) {
				continue
			}
			p.QuestStatus[def.ID] = questComplete
			changed = append(changed, questStatusEntry(def.ID))
			progressed = true
			if !def.Visible {
				m.collectQuest(p, def.ID)
			}
		}
		if !progressed {
			break
		}
	}
	return changed
}

func (m *Manager) collectQuest(p *Player, questID int) bool {
	m.ensureQuests(p)
	def := m.Static.QuestByID[questID]
	if def == nil || p.QuestStatus[questID] != questComplete {
		return false
	}
	p.AddProperties(
		int64(def.Rewards.Coins),
		int64(def.Rewards.Diamonds),
		int64(def.Rewards.Food),
		int64(def.Rewards.XP),
		int64(def.Rewards.Eth),
	)
	p.QuestStatus[questID] = questCollected
	for _, name := range def.Next {
		if nd := m.Static.QuestByName[name]; nd != nil {
			if _, ok := p.QuestStatus[nd.ID]; !ok {
				p.QuestStatus[nd.ID] = questActive
			}
		}
	}
	return true
}

func registerQuestHandlers(m *Manager) {
	m.HandlePlayer("gs_quest_collect", func(ctx *Context, p *Player) {
		questID := ctx.Int("quest_id")
		if !m.collectQuest(p, questID) {
			ctx.Reply("gs_quest_collect", data.MakeGFSObject())
			return
		}
		ctx.Reply("gs_quest_collect", data.MakeGFSObject().
			PutGFSArray("properties", p.GetProperties()))
		result := data.MakeGFSArray()
		result.AddSFSObject(data.MakeGFSObject().PutLong("collect", int64(questID)))
		ctx.Reply("gs_quest", contentResponse("result", result))
	})
	m.HandlePlayer("gs_quest_read", func(ctx *Context, p *Player) {
		questID := ctx.Int("quest_id")

		log.Print(questID)
	})
}

func (m *Manager) questGoalsMet(p *Player, def *db.QuestDef) bool {
	if len(def.Goals) == 0 {
		return false
	}
	for _, g := range def.Goals {
		if !m.evalGoal(p, g) {
			return false
		}
	}
	return true
}

func (m *Manager) evalGoal(p *Player, g db.QuestGoal) bool {
	switch g.Key {
	case "monster":
		return m.countMonsters(p, g.Int) >= g.Num
	case "object":
		return m.countObjects(p, g.Int) >= g.Num
	case "structure_type":
		return m.countStructureType(p, g.Str) >= g.Num
	case "monster_level":
		return cmpEval(maxMonsterLevel(p), g.Int, g.Eval)
	case "castle_level":
		return false
	case "coins":
		return cmpEval(int(p.Coins), g.Int, g.Eval)
	case "food":
		return cmpEval(int(p.Food), g.Int, g.Eval)
	case "diamonds":
		return cmpEval(int(p.Diamonds), g.Int, g.Eval)
	case "xp":
		return cmpEval(int(p.XP), g.Int, g.Eval)
	case "ethereal_currency":
		return cmpEval(int(p.Shards), g.Int, g.Eval)
	case "obstacle_removed", "rename_monster":
		return p.QuestCounters[g.Key] >= g.Int
	case "collect":
		return p.QuestStatus[g.Int] == questCollected
	case "neighbor":
		return false
	default:
		return false
	}
}

func (m *Manager) countObjects(p *Player, object int) int {
	n := 0
	for _, isl := range p.Islands {
		for _, s := range isl.Structures {
			sid := int(s.StructureID)
			if sid == object || m.Static.StructureEntity[sid] == object {
				n++
			}
		}
		for _, mon := range isl.Monsters {
			mid := int(mon.MonsterID)
			if mid == object || m.Static.MonsterEntity[mid] == object {
				n++
			}
		}
	}
	return n
}

func (m *Manager) countMonsters(p *Player, monsterID int) int {
	n := 0
	for _, isl := range p.Islands {
		for _, mon := range isl.Monsters {
			if int(mon.MonsterID) == monsterID {
				n++
			}
		}
		for _, egg := range isl.Eggs {
			if int(egg.MonsterID) == monsterID {
				n++
			}
		}
	}
	return n
}

func (m *Manager) countStructureType(p *Player, structureType string) int {
	n := 0
	for _, isl := range p.Islands {
		for _, s := range isl.Structures {
			if m.Static.StructureType[int(s.StructureID)] == structureType {
				n++
			}
		}
	}
	return n
}

func maxMonsterLevel(p *Player) int {
	best := 0
	for _, isl := range p.Islands {
		for _, mon := range isl.Monsters {
			if mon.Level > best {
				best = mon.Level
			}
		}
	}
	return best
}

func activeIslandLevel(p *Player) int {
	if p.GetActiveIsland() != nil {
		return 30
	}
	return 0
}

var _ = activeIslandLevel

func cmpEval(actual, target int, eval string) bool {
	switch eval {
	case ">":
		return actual > target
	case ">=":
		return actual >= target
	case "<":
		return actual < target
	case "<=":
		return actual <= target
	default:
		return actual == target
	}
}

func questStatusEntry(questID int) *data.GFSObject {
	return data.MakeGFSObject().
		PutLong("quest_id", int64(questID)).
		PutUtfString("status", "true")
}

func (m *Manager) runQuestEval(ctx *Context, p *Player) {
	changed := m.evaluateQuests(p)
	if len(changed) == 0 {
		return
	}
	result := data.MakeGFSArray()
	for _, e := range changed {
		result.AddSFSObject(e)
	}
	ctx.Reply("gs_quest", contentResponse("result", result))
}

func (m *Manager) questListFor(p *Player) *data.GFSArray {
	m.ensureQuests(p)
	out := data.MakeGFSArray()
	for _, id := range m.Static.QuestOrder {
		status := p.QuestStatus[id]
		if status == questCollected {
			continue
		}
		statusStr, collected, isNew := "false", 0, 0
		switch status {
		case questComplete:
			statusStr = "true"
		case questCollected:
			statusStr, collected = "true", 1
		case questActive:
			if def := m.Static.QuestByID[id]; def != nil && def.Initial {
				isNew = 1
			}
		}
		log := data.MakeGFSObject().
			PutInt("id", id).
			PutInt("quest_id", id).
			PutInt("user", 0).
			PutUtfString("status", statusStr).
			PutInt("collected", collected).
			PutInt("new", isNew)
		entry := data.MakeGFSArray()
		entry.AddSFSObject(log)
		entry.AddSFSObject(m.Static.QuestStatic[id])
		out.AddSFSObject(data.MakeGFSObject().
			PutGFSArray("new", entry).
			PutLong("id", int64(id)))
	}
	return out
}
