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
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/Paficent/GoFox2X/data"
)

type timerService struct {
	mu     sync.Mutex
	timers map[string]*time.Timer
}

func newTimerService() *timerService {
	return &timerService{timers: map[string]*time.Timer{}}
}

func (m *Manager) Schedule(key string, delay time.Duration, fn func()) {
	ts := m.timers
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if old, ok := ts.timers[key]; ok {
		old.Stop()
		delete(ts.timers, key)
	}
	if delay <= 0 {
		go fn()
		return
	}
	ts.timers[key] = time.AfterFunc(delay, func() {
		ts.mu.Lock()
		delete(ts.timers, key)
		ts.mu.Unlock()
		fn()
	})
}

// absolute epoch-millis deadline.
func (m *Manager) ScheduleAt(key string, atMS int64, fn func()) {
	m.Schedule(key, time.Duration(atMS-nowMS())*time.Millisecond, fn)
}

func (m *Manager) CancelTimer(key string) {
	ts := m.timers
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if t, ok := ts.timers[key]; ok {
		t.Stop()
		delete(ts.timers, key)
	}
}

func upgradeKey(bbbID, userStructureID int64) string {
	return fmt.Sprintf("upgrade:%d:%d", bbbID, userStructureID)
}

func upgradedStructurePayload(p *Player, s *Structure) *data.GFSObject {
	return data.MakeGFSObject().
		PutLong("success", 1).
		PutLong("user_structure_id", s.UserStructureID).
		PutGFSObject("user_structure", s.GetSFSObject()).
		PutGFSArray("properties", p.GetProperties())
}

func (m *Manager) finishUpgradeNow(p *Player, s *Structure) {
	m.CancelTimer(upgradeKey(p.BBBID, s.UserStructureID))
	now := nowMS()
	if s.UpgradeTo != 0 {
		s.StructureID = s.UpgradeTo
	}
	s.UpgradeTo = 0
	s.IsUpgrading = 0
	s.IsComplete = 1
	s.DateCreated = now
	s.LastCollection = now
}

func (m *Manager) completeUpgrade(bbbID, userStructureID int64) {
	p := m.Player(bbbID)
	if p == nil {
		return
	}
	s := p.findStructure(userStructureID)
	if s == nil || s.IsUpgrading == 0 {
		return
	}
	m.finishUpgradeNow(p, s)

	if err := m.savePlayer(p); err != nil {
		log.Printf("save after upgrade complete failed: %v", err)
	}
	m.Push(bbbID, "gs_finish_upgrade_structure", upgradedStructurePayload(p, s))
}

func clearObstacleKey(bbbID, userStructureID int64) string {
	return fmt.Sprintf("clearobs:%d:%d", bbbID, userStructureID)
}

func obstacleClearedPayload(p *Player, userStructureID int64) *data.GFSObject {
	return data.MakeGFSObject().
		PutLong("success", 1).
		PutLong("user_structure_id", userStructureID).
		PutGFSArray("properties", p.GetProperties())
}

func (m *Manager) removeObstacle(p *Player, userStructureID int64) {
	for _, isl := range p.Islands {
		s := isl.FindStructure(userStructureID)
		if s == nil {
			continue
		}
		isl.RemoveStructure(userStructureID)
		p.bumpCounter("obstacle_removed")
		if info, ok := m.Static.Structure(int(s.StructureID)); ok && info.XP > 0 {
			p.AddProperties(0, 0, 0, int64(info.XP), 0)
		}
		return
	}
}

func (m *Manager) completeClearObstacle(bbbID, userStructureID int64) {
	p := m.Player(bbbID)
	if p == nil {
		return
	}
	m.removeObstacle(p, userStructureID)
	if err := m.savePlayer(p); err != nil {
		log.Printf("save after obstacle clear failed: %v", err)
	}
	m.Push(bbbID, "gs_clear_obstacle", obstacleClearedPayload(p, userStructureID))
}

func (m *Manager) rearmUpgradeTimers() {
	m.mu.Lock()
	players := make([]*Player, 0, len(m.players))
	for _, p := range m.players {
		players = append(players, p)
	}
	m.mu.Unlock()

	now := nowMS()
	armed, finished := 0, 0
	for _, p := range players {
		for _, isl := range p.Islands {
			for _, s := range isl.Structures {
				if s.IsUpgrading == 0 {
					continue
				}
				bbbID, usid, at := p.BBBID, s.UserStructureID, s.BuildingCompleted
				if at <= now {
					m.completeUpgrade(bbbID, usid)
					finished++
				} else {
					m.ScheduleAt(upgradeKey(bbbID, usid), at, func() {
						m.completeUpgrade(bbbID, usid)
					})
					armed++
				}
			}
		}
	}
	if armed > 0 || finished > 0 {
		log.Printf("upgrades on load: %d re-armed, %d finished immediately", armed, finished)
	}
}

func (m *Manager) calculateSpeedupCost(p *Player, id int64, kind string) int64 {
	minutes := 0

	switch kind {
	case "structure":
		s := p.findStructure(id)
		minutes = int(((s.BuildingCompleted - s.DateCreated) / 1000) / 60)
	case "baking":
		b := p.GetActiveIsland().FindBaking(id)
		minutes = int(((b.CompleteOn - b.StartedOn) / 1000) / 60)
	case "egg":
		e := p.GetActiveIsland().FindEgg(id)
		minutes = int(((e.HatchesOn - e.LaidOn) / 1000) / 60)
	case "breeding":
		b := p.GetActiveIsland().FindBreeding(id)
		minutes = int(((b.CompleteOn - b.StartedOn) / 1000) / 60)
	}
	return int64(math.Ceil(float64(minutes) / 60.0))
}
