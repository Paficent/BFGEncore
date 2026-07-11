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

package db

import "math/rand"

type breedCombo struct {
	Result      int
	Probability int
}

func loadBreedingCombos(db *DB) map[[2]int][]breedCombo {
	out := map[[2]int][]breedCombo{}
	for _, e := range db.Breeding {
		a, b := e.Monsters[0], e.Monsters[1]
		if a > b {
			a, b = b, a
		}
		key := [2]int{a, b}
		for _, r := range e.Results {
			out[key] = append(out[key], breedCombo{Result: r[0], Probability: r[1]})
		}
	}
	for key, combos := range out {
		for i := 1; i < len(combos); i++ {
			for j := i; j > 0 && combos[j].Probability > combos[j-1].Probability; j-- {
				combos[j], combos[j-1] = combos[j-1], combos[j]
			}
		}
		out[key] = combos
	}
	return out
}

func (sd *StaticData) BreedingResult(monster1, monster2, level1, level2, playerLevel int) int {
	if monster1 > monster2 {
		monster1, monster2 = monster2, monster1
	}
	if combos := sd.breedingCombos[[2]int{monster1, monster2}]; len(combos) > 0 {
		return combos[0].Result
	}
	total := level1 + level2
	if total <= 0 {
		if rand.Intn(2) == 0 {
			return monster1
		}
		return monster2
	}
	firstProb := int(float64(level1) / float64(total) * 100)
	if rand.Intn(100)+1 <= firstProb {
		return monster1
	}
	return monster2
}

type LevelInfo struct {
	Food      int
	Coins     int
	Shards    int
	MaxCoins  int
	MaxShards int
}

func loadMonsterLevels(db *DB) map[[2]int]LevelInfo {
	out := map[[2]int]LevelInfo{}
	for _, m := range db.MonsterLevels {
		for _, r := range m.Levels {
			out[[2]int{m.Monster, r.Level}] = LevelInfo{
				Food:      r.Food,
				Coins:     r.Coins,
				Shards:    r.Eth,
				MaxCoins:  r.MaxCoins,
				MaxShards: r.MaxEth,
			}
		}
	}
	return out
}

func (sd *StaticData) MonsterLevel(monsterID, level int) (LevelInfo, bool) {
	li, ok := sd.monsterLevels[[2]int{monsterID, level}]
	return li, ok
}

func loadLevelXP(db *DB) map[int]int {
	out := map[int]int{}
	for _, r := range db.Levels {
		out[r.Level] = r.XP
	}
	return out
}

type TeleportDest struct {
	DestIsland  int
	DestMonster int
}

func loadTeleportInfo(db *DB) map[[2]int]TeleportDest {
	out := map[[2]int]TeleportDest{}
	for _, e := range db.Teleports {
		out[[2]int{e.SrcIsland, e.SrcMonster}] = TeleportDest{e.DestIsland, e.DestMonster}
	}
	return out
}

func (sd *StaticData) Teleport(island, monster int) (TeleportDest, bool) {
	d, ok := sd.teleports[[2]int{island, monster}]
	return d, ok
}

func loadScratchOffs(db *DB) map[string][]ScratchOff {
	out := map[string][]ScratchOff{}
	for _, s := range db.ScratchOffs {
		out[s.Type] = append(out[s.Type], s)
	}
	return out
}

func (sd *StaticData) RollScratchOff(kind string) (ScratchOff, bool) {
	pool := sd.scratchByType[kind]
	total := 0
	for _, s := range pool {
		if s.Probability > 0 {
			total += s.Probability
		}
	}
	if total == 0 {
		return ScratchOff{}, false
	}
	n := rand.Intn(total)
	for _, s := range pool {
		if s.Probability <= 0 {
			continue
		}
		if n < s.Probability {
			return s, true
		}
		n -= s.Probability
	}
	return ScratchOff{}, false
}
