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

type MonsterInfo struct {
	Entity          int
	CostCoins       int
	CostDiamonds    int
	CostEth         int
	BuildTime       int
	XP              int
	BoxRequirements string
	Names           []string
}

type MineInfo struct {
	Time     int
	Diamonds int
}

type FoodOption struct {
	Food  int
	Cost  int
	Time  int
	Xp    int
	Label string
}

type StructureInfo struct {
	Entity       int
	Type         string
	UpgradesTo   int
	CostCoins    int
	CostDiamonds int
	CostEth      int
	BuildTime    int
	XP           int
	Mine         *MineInfo
	Food         []FoodOption
}

type IslandInfo struct {
	CostCoins    int
	CostDiamonds int
	Castle       int
	Obstacle     int
}

func loadMonsters(db *DB) map[int]MonsterInfo {
	out := make(map[int]MonsterInfo, len(db.Monsters))
	for _, m := range db.Monsters {
		e, _ := db.entity(m.Entity)
		out[m.ID] = MonsterInfo{
			Entity:          m.Entity,
			CostCoins:       e.CostCoins,
			CostDiamonds:    e.CostDiamonds,
			CostEth:         e.CostEth,
			BuildTime:       e.BuildTime,
			XP:              e.XP,
			Names:           m.Names.V,
			BoxRequirements: m.BoxRequirements,
		}
	}
	return out
}

func loadStructures(db *DB) map[int]StructureInfo {
	out := make(map[int]StructureInfo, len(db.Structures))
	for _, s := range db.Structures {
		e, _ := db.entity(s.Entity)
		info := StructureInfo{
			Entity:       s.Entity,
			Type:         s.Type,
			UpgradesTo:   s.UpgradesTo,
			CostCoins:    e.CostCoins,
			CostDiamonds: e.CostDiamonds,
			CostEth:      e.CostEth,
			BuildTime:    e.BuildTime,
			XP:           e.XP,
		}
		if s.Type == "mine" {
			info.Mine = &MineInfo{
				Time:     toInt(s.Extra.V["time"]),
				Diamonds: toInt(s.Extra.V["diamonds"]),
			}
		}
		if opts, ok := s.Extra.V["food_options"].([]any); ok {
			for _, o := range opts {
				m, ok := o.(map[string]any)
				if !ok {
					continue
				}
				info.Food = append(info.Food, FoodOption{
					Food:  toInt(m["food"]),
					Cost:  toInt(m["cost"]),
					Time:  toInt(m["time"]),
					Xp:    toInt(m["xp"]),
					Label: jstr(m["label"]),
				})
			}
		}
		out[s.ID] = info
	}
	return out
}

func loadIslands(db *DB) map[int]IslandInfo {
	out := make(map[int]IslandInfo, len(db.Islands))
	for _, r := range db.Islands {
		out[r.ID] = IslandInfo{
			CostCoins:    r.CostCoins,
			CostDiamonds: r.CostDiamonds,
			Castle:       r.Castle,
			Obstacle:     r.Obstacle,
		}
	}
	return out
}

func (sd *StaticData) Monster(id int) (MonsterInfo, bool) {
	mi, ok := sd.monsters[id]
	return mi, ok
}

func (sd *StaticData) Structure(id int) (StructureInfo, bool) {
	si, ok := sd.structures[id]
	return si, ok
}

func (sd *StaticData) Island(id int) (IslandInfo, bool) {
	ii, ok := sd.islands[id]
	return ii, ok
}
