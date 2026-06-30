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

import (
	"encoding/json"
	"math/rand"
)

type BuyInfo struct {
	Entity       int
	CostCoins    int
	CostDiamonds int
	CostEth      int
	BuildTime    int
	Xp           int
}

func buyInfoFromEntity(db *DB, entity int) (BuyInfo, bool) {
	e, ok := db.entityByID(entity)
	if !ok {
		return BuyInfo{}, false
	}
	return BuyInfo{
		Entity:       entity,
		CostCoins:    e.Int("cost_coins"),
		CostDiamonds: e.Int("cost_diamonds"),
		CostEth:      e.Int("cost_eth_currency"),
		BuildTime:    e.Int("build_time"),
		Xp:           e.Int("xp"),
	}, true
}

func loadMonsterBuy(db *DB) map[int]BuyInfo {
	out := map[int]BuyInfo{}
	for _, m := range db.Table("monsters") {
		if info, ok := buyInfoFromEntity(db, m.Int("entity")); ok {
			out[m.Int("monster_id")] = info
		}
	}
	return out
}

func loadStructureBuy(db *DB) map[int]BuyInfo {
	out := map[int]BuyInfo{}
	for _, s := range db.Table("structures") {
		if info, ok := buyInfoFromEntity(db, s.Int("entity")); ok {
			out[s.Int("structure_id")] = info
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
	for _, r := range db.Table("breeding_combinations") {
		a, b := r.Int("monster_1"), r.Int("monster_2")
		if a > b {
			a, b = b, a
		}
		key := [2]int{a, b}
		out[key] = append(out[key], breedCombo{Result: r.Int("result"), Probability: r.Int("probability")})
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
	for _, r := range db.Table("monster_levels") {
		out[[2]int{r.Int("monster"), r.Int("level")}] = LevelInfo{
			Food:      r.Int("food"),
			Coins:     r.Int("coins"),
			Shards:    r.Int("ethereal_currency"),
			MaxCoins:  r.Int("max_coins"),
			MaxShards: r.Int("max_ethereal"),
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
	for _, r := range db.Table("islands") {
		out[r.Int("island_id")] = IslandBuyInfo{
			CostCoins:    r.Int("cost_coins"),
			CostDiamonds: r.Int("cost_diamonds"),
			Castle:       r.Int("castle_structure_id"),
		}
	}
	return out
}

func loadLevelXP(db *DB) map[int]int {
	out := map[int]int{}
	for _, r := range db.Table("level_xp") {
		out[r.Int("level")] = r.Int("xp")
	}
	return out
}

func loadStructureUpgrades(db *DB) map[int]int {
	out := map[int]int{}
	for _, r := range db.Table("structures") {
		if up := r.Int("upgrades_to"); up != 0 {
			out[r.Int("structure_id")] = up
		}
	}
	return out
}

type MineInfo struct {
	Time     int // Time in minutes for the mine to finish its job
	Diamonds int
}

func loadMineInfo(db *DB) map[int]MineInfo {
	out := map[int]MineInfo{}
	for _, r := range db.Table("structures") {
		if r.Str("structure_type") != "mine" {
			continue
		}
		extra := r.Str("extra")
		var m MineInfo
		if err := json.Unmarshal([]byte(extra), &m); err != nil {
			continue
		}
		out[r.Int("structure_id")] = m
	}
	return out
}

type TeleportDest struct {
	DestIsland  int
	DestMonster int
}

func loadTeleportInfo(db *DB) map[[2]int]TeleportDest {
	out := map[[2]int]TeleportDest{}
	for _, e := range db.Table("monster_island_2_island_map") {
		out[[2]int{e.Int("source_island"), e.Int("source_monster")}] = TeleportDest{e.Int("dest_island"), e.Int("dest_monster")}
	}
	return out
}
