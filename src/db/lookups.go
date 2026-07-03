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

type BuyInfo struct {
	Entity       int
	CostCoins    int
	CostDiamonds int
	CostEth      int
	BuildTime    int
	Xp           int
}

func buyInfoFromEntity(db *DB, entity int) (BuyInfo, bool) {
	e, ok := db.entity(entity)
	if !ok {
		return BuyInfo{}, false
	}
	return BuyInfo{
		Entity:       entity,
		CostCoins:    e.CostCoins,
		CostDiamonds: e.CostDiamonds,
		CostEth:      e.CostEth,
		BuildTime:    e.BuildTime,
		Xp:           e.XP,
	}, true
}

func loadMonsterBuy(db *DB) map[int]BuyInfo {
	out := map[int]BuyInfo{}
	for _, m := range db.Monsters {
		if info, ok := buyInfoFromEntity(db, m.Entity); ok {
			out[m.ID] = info
		}
	}
	return out
}

func loadStructureBuy(db *DB) map[int]BuyInfo {
	out := map[int]BuyInfo{}
	for _, s := range db.Structures {
		if info, ok := buyInfoFromEntity(db, s.Entity); ok {
			out[s.ID] = info
		}
	}
	return out
}

type breedCombo struct {
	Result      int
	Probability int
}

func loadBreedingCombos(db *DB) map[[2]int][]breedCombo {
	out := map[[2]int][]breedCombo{}
	for _, r := range db.Breeding {
		a, b := r.Monster1, r.Monster2
		if a > b {
			a, b = b, a
		}
		key := [2]int{a, b}
		out[key] = append(out[key], breedCombo{Result: r.Result, Probability: r.Probability})
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

type IslandBuyInfo struct {
	CostCoins    int
	CostDiamonds int
	Castle       int
}

func loadIslandBuy(db *DB) map[int]IslandBuyInfo {
	out := map[int]IslandBuyInfo{}
	for _, r := range db.Islands {
		out[r.ID] = IslandBuyInfo{
			CostCoins:    r.CostCoins,
			CostDiamonds: r.CostDiamonds,
			Castle:       r.Castle,
		}
	}
	return out
}

func loadLevelXP(db *DB) map[int]int {
	out := map[int]int{}
	for _, r := range db.Levels {
		out[r.Level] = r.XP
	}
	return out
}

func loadStructureUpgrades(db *DB) map[int]int {
	out := map[int]int{}
	for _, r := range db.Structures {
		if r.UpgradesTo != 0 {
			out[r.ID] = r.UpgradesTo
		}
	}
	return out
}

type MineInfo struct {
	Time     int
	Diamonds int
}

func loadMineInfo(db *DB) map[int]MineInfo {
	out := map[int]MineInfo{}
	for _, r := range db.Structures {
		if r.Type != "mine" {
			continue
		}
		out[r.ID] = MineInfo{
			Time:     toInt(r.Extra.V["time"]),
			Diamonds: toInt(r.Extra.V["diamonds"]),
		}
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
