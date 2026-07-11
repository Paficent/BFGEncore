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
	"encoding/json"
	"log"
	"time"

	"paficent/bfg/db"
)

type playerSave struct {
	BBBID        int64     `json:"bbb_id"`
	UserID       int64     `json:"user_id"`
	DisplayName  string    `json:"display_name"`
	Coins        int64     `json:"coins"`
	Diamonds     int64     `json:"diamonds"`
	Food         int64     `json:"food"`
	XP           int64     `json:"xp"`
	Shards       int64     `json:"shards"`
	Level        int       `json:"level"`
	ActiveIsland int64     `json:"active_island"`
	EggSeq       int64     `json:"egg_seq"`
	BreedingSeq  int64     `json:"breeding_seq"`
	BakingSeq    int64     `json:"baking_seq"`
	MonsterSeq   int64     `json:"monster_seq"`
	StructureSeq int64     `json:"structure_seq"`
	IslandSeq    int64     `json:"island_seq"`
	Islands      []*Island `json:"islands"`

	QuestStatus   map[int]string `json:"quest_status"`
	QuestCounters map[string]int `json:"quest_counters"`
}

func (p *Player) toSave() *playerSave {
	return &playerSave{
		BBBID:        p.BBBID,
		UserID:       p.UserID,
		DisplayName:  p.DisplayName,
		Coins:        p.Coins,
		Diamonds:     p.Diamonds,
		Food:         p.Food,
		XP:           p.XP,
		Shards:       p.Shards,
		Level:        p.Level,
		ActiveIsland: p.ActiveIsland,
		EggSeq:       p.eggSeq,
		BreedingSeq:  p.breedingSeq,
		BakingSeq:    p.bakingSeq,
		MonsterSeq:   p.monsterSeq,
		StructureSeq: p.structureSeq,
		IslandSeq:    p.islandSeq,
		Islands:      p.Islands,

		QuestStatus:   p.QuestStatus,
		QuestCounters: p.QuestCounters,
	}
}

func (ps *playerSave) toPlayer(static *db.StaticData) *Player {
	return &Player{
		BBBID:        ps.BBBID,
		UserID:       ps.UserID,
		DisplayName:  ps.DisplayName,
		Coins:        ps.Coins,
		Diamonds:     ps.Diamonds,
		Food:         ps.Food,
		XP:           ps.XP,
		Shards:       ps.Shards,
		Level:        ps.Level,
		ActiveIsland: ps.ActiveIsland,
		Islands:      ps.Islands,
		levelXP:      static.LevelXP,
		eggSeq:       ps.EggSeq,
		breedingSeq:  ps.BreedingSeq,
		bakingSeq:    ps.BakingSeq,
		monsterSeq:   ps.MonsterSeq,
		structureSeq: ps.StructureSeq,
		islandSeq:    ps.IslandSeq,

		QuestStatus:   ps.QuestStatus,
		QuestCounters: ps.QuestCounters,
	}
}

func (m *Manager) loadPlayers() {
	if m.store == nil {
		return
	}
	records, err := m.store.Load()
	if err != nil {
		log.Printf("players: load failed: %v", err)
		return
	}
	m.mu.Lock()
	for _, rec := range records {
		var ps playerSave
		if err := json.Unmarshal(rec.Data, &ps); err != nil {
			log.Printf("players: skipping bad save for %d: %v", rec.BBBID, err)
			continue
		}
		if ps.BBBID <= 0 {
			ps.BBBID = rec.BBBID
		}
		m.players[ps.BBBID] = ps.toPlayer(m.Static)
	}
	n := len(m.players)
	m.mu.Unlock()
	log.Printf("loaded %d player(s) from save store", n)
}

func (m *Manager) savePlayer(p *Player) error {
	if m.store == nil || p == nil {
		return nil
	}
	blob, err := json.Marshal(p.toSave())
	if err != nil {
		return err
	}
	return m.store.Save(p.BBBID, blob, nowMS())
}

func (m *Manager) SavePlayer(p *Player) error { return m.savePlayer(p) }

func (m *Manager) SaveAll() (int, error) {
	if m.store == nil {
		return 0, nil
	}
	m.mu.Lock()
	players := make([]*Player, 0, len(m.players))
	for _, p := range m.players {
		players = append(players, p)
	}
	m.mu.Unlock()

	n := 0
	for _, p := range players {
		if err := m.savePlayer(p); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// TODO: maybe useless with Manager.HandleWrite()
// also could potentially cause db overwrite issues...
func (m *Manager) StartAutosave(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			n, err := m.SaveAll()
			if err != nil {
				log.Printf("autosave failed: %v", err)
			} else if n > 0 {
				log.Printf("autosaved %d player(s)", n)
			}
		}
	}()
}
