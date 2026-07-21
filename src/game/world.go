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

import "github.com/Paficent/GoFox2X/data"

type Structure struct {
	UserStructureID    int64
	UserIslandID       int64
	StructureID        int64
	X                  int
	Y                  int
	IsComplete         int
	IsObstacleComplete int
	IsUpgrading        int
	UpgradeTo          int64 // target structure_id while upgrading; 0 when idle
	Flip               int
	Muted              int
	Scale              float64
	DateCreated        int64
	BuildingCompleted  int64
	LastCollection     int64
}

func (s *Structure) GetSFSObject() *data.GFSObject {
	obj := data.MakeGFSObject().
		PutLong("user_structure_id", s.UserStructureID).
		PutLong("user_island_id", s.UserIslandID).
		PutLong("island", s.UserIslandID).
		PutLong("structure", s.StructureID).
		PutFloat("scale", float32(s.Scale)).
		PutDouble("size", s.Scale).
		PutInt("pos_x", s.X).
		PutInt("pos_y", s.Y).
		PutInt("flip", boolInt(s.Flip != 0)).
		PutInt("muted", boolInt(s.Muted != 0)).
		PutInt("is_complete", s.IsComplete).
		PutInt("is_upgrading", s.IsUpgrading).
		PutInt("in_warehouse", 0).

		PutLong("last_collection", s.LastCollection).
		PutDouble("diamonds_collected", 0)

	if s.IsComplete == 0 || s.IsObstacleComplete == 0 || s.IsUpgrading != 0 {
		obj.PutLong("building_completed", s.BuildingCompleted)
		obj.PutLong("date_created", s.DateCreated)
	}
	inv := data.MakeGFSArray()
	inv.AddSFSObject(data.MakeGFSObject().PutInt("m", 0))
	obj.PutGFSArray("inv", inv)
	obj.PutUtfString("req", `[{"m":68}]`)
	return obj
}

type Monster struct {
	UserMonsterID   int64
	UserIslandID    int64
	MonsterID       int64
	ParentMonsterID int64
	X               int
	Y               int
	Flip            int
	Level           int
	Happiness       int
	CollectedCoins  int
	TimesFed        int
	Volume          float64
	Muted           int
	Name            string
	DateCreated     int64
	LastCollection  int64
	MegaPerma       bool
	MegaCurrent     int
	MegaStart       int64
	MegaFinish      int64
}

func (m *Monster) GetSFSObject() *data.GFSObject {
	obj := data.MakeGFSObject().
		PutLong("user_monster_id", m.UserMonsterID).
		PutLong("user_island_id", m.UserIslandID).
		PutLong("island", m.UserIslandID).
		PutInt("monster", int(m.MonsterID)).
		PutInt("pos_x", m.X).
		PutInt("pos_y", m.Y).
		PutInt("flip", m.Flip).
		PutInt("level", m.Level).
		PutInt("happiness", m.Happiness).
		PutInt("collected_coins", m.CollectedCoins).
		PutInt("collected_ethereal", 0).
		PutInt("collected_diamonds", 0).
		PutInt("collected_food", 0).
		PutInt("times_fed", m.TimesFed).
		PutDouble("volume", m.Volume).
		PutInt("muted", m.Muted).
		PutInt("in_hotel", 0).
		PutBool("limited", false).
		PutLong("last_feeding", m.DateCreated).
		PutLong("date_created", m.DateCreated).
		PutLong("last_collection", m.LastCollection).
		PutUtfString("name", m.Name)
	if m.hasMega() {
		obj.PutGFSObject("megamonster", m.megaObject())
	}
	return obj
}

func (m *Monster) goldSFSObject(islandID int64) *data.GFSObject {
	return data.MakeGFSObject().
		PutLong("user_monster_id", m.UserMonsterID).
		PutLong("user_island_id", islandID).
		PutLong("island", islandID).
		PutInt("monster", int(m.MonsterID)).
		PutInt("pos_x", m.X).
		PutInt("pos_y", m.Y).
		PutInt("flip", m.Flip).
		PutInt("level", m.Level).
		PutUtfString("name", m.Name).
		PutDouble("volume", m.Volume).
		PutInt("muted", m.Muted).
		PutInt("times_fed", 0).
		PutInt("happiness", 0).
		PutInt("collected_coins", 0).
		PutInt("collected_ethereal", 0).
		PutInt("in_hotel", 0).
		PutBool("limited", false)
}

func (m *Monster) hasMega() bool {
	return m.MegaPerma == true || m.MegaFinish > 0
}

func (m *Monster) megaObject() *data.GFSObject {
	return data.MakeGFSObject().
		PutBool("permamega", m.MegaPerma).
		PutLong("user_monster_id", m.UserMonsterID).
		PutLong("currently_mega", int64(m.MegaCurrent)).
		PutLong("started_at", m.MegaStart).
		PutLong("finishes_at", m.MegaFinish)
}

type Egg struct {
	IslandID        int64
	LaidOn          int64
	HatchesOn       int64
	MonsterID       int64
	UserEggID       int64
	UserStructureID int64
}

func (e *Egg) GetSFSObject() *data.GFSObject {
	return data.MakeGFSObject().
		PutLong("island", e.IslandID).
		PutInt("structure", int(e.UserStructureID)).
		PutInt("monster", int(e.MonsterID)).
		PutLong("user_egg_id", e.UserEggID).
		PutLong("hatches_on", e.HatchesOn).
		PutLong("laid_on", e.LaidOn)
}

type Breeding struct {
	IslandID       int64
	UserBreedingID int64
	StructureID    int64
	Monster1       int
	Monster2       int
	NewMonster     int
	StartedOn      int64
	CompleteOn     int64
}

func (b *Breeding) GetSFSObject() *data.GFSObject {
	return data.MakeGFSObject().
		PutLong("island", b.IslandID).
		PutLong("user_breeding_id", b.UserBreedingID).
		PutLong("structure", b.StructureID).
		PutInt("monster_1", b.Monster1).
		PutInt("monster_2", b.Monster2).
		PutInt("new_monster", b.NewMonster).
		PutLong("started_on", b.StartedOn).
		PutLong("complete_on", b.CompleteOn)
}

type Baking struct {
	IslandID        int64
	UserBakingID    int64
	UserStructureID int64
	FoodIndex       int
	Food            int
	Xp              int
	StartedOn       int64
	CompleteOn      int64
}

func (b *Baking) GetSFSObject() *data.GFSObject {
	return data.MakeGFSObject().
		PutLong("island", b.IslandID).
		PutLong("user_baking_id", b.UserBakingID).
		PutLong("user_structure", b.UserStructureID).
		PutInt("food_count", b.Food).
		PutInt("food_index", b.FoodIndex).
		PutLong("started_at", b.StartedOn).
		PutLong("finished_at", b.CompleteOn)
}

type Island struct {
	UserIslandID int64
	IslandID     int64
	BBBID        int64
	Likes        int
	Dislikes     int
	WarpSpeed    float64
	Structures   []*Structure
	GoldMonsters []*Monster `json:"gold_monsters,omitempty"` // Don't break older saves
	Monsters     []*Monster
	Eggs         []*Egg
	Breedings    []*Breeding
	Bakings      []*Baking
}

func find[T any](items []*T, match func(*T) bool) *T {
	for _, it := range items {
		if match(it) {
			return it
		}
	}
	return nil
}

func remove[T any](items []*T, match func(*T) bool) []*T {
	out := items[:0]
	for _, it := range items {
		if !match(it) {
			out = append(out, it)
		}
	}
	return out
}

func (i *Island) FindStructure(id int64) *Structure {
	return find(i.Structures, func(s *Structure) bool { return s.UserStructureID == id })
}

func (i *Island) FindStructureByType(structureID int64) *Structure {
	return find(i.Structures, func(s *Structure) bool { return s.StructureID == structureID })
}

func (i *Island) FindMonster(id int64) *Monster {
	return find(i.Monsters, func(m *Monster) bool { return m.UserMonsterID == id })
}

func (i *Island) FindGoldMonster(id int64) *Monster {
	return find(i.GoldMonsters, func(m *Monster) bool { return m.UserMonsterID == id })
}

func (i *Island) monster(id int64) *Monster {
	if i.IsGold() {
		return i.FindGoldMonster(id)
	}
	return i.FindMonster(id)
}

func (i *Island) monsterSFS(m *Monster) *data.GFSObject {
	if i.IsGold() {
		return m.goldSFSObject(i.UserIslandID)
	}
	return m.GetSFSObject()
}

func (i *Island) FindEgg(id int64) *Egg {
	return find(i.Eggs, func(e *Egg) bool { return e.UserEggID == id })
}

func (i *Island) FindBreeding(id int64) *Breeding {
	return find(i.Breedings, func(b *Breeding) bool { return b.UserBreedingID == id })
}

func (i *Island) FindBaking(id int64) *Baking {
	return find(i.Bakings, func(b *Baking) bool { return b.UserBakingID == id })
}

func (i *Island) AddMonster(m *Monster) {
	i.Monsters = append(i.Monsters, m)
}

func (i *Island) RemoveEgg(id int64) {
	i.Eggs = remove(i.Eggs, func(e *Egg) bool { return e.UserEggID == id })
}

func (i *Island) RemoveStructure(id int64) {
	i.Structures = remove(i.Structures, func(s *Structure) bool { return s.UserStructureID == id })
}

func (i *Island) RemoveMonster(id int64) {
	i.Monsters = remove(i.Monsters, func(m *Monster) bool { return m.UserMonsterID == id })
}

func (i *Island) RemoveBreeding(id int64) {
	i.Breedings = remove(i.Breedings, func(b *Breeding) bool { return b.UserBreedingID == id })
}

func (i *Island) RemoveBaking(id int64) {
	i.Bakings = remove(i.Bakings, func(b *Baking) bool { return b.UserBakingID == id })
}

func (i *Island) GetSFSObject() *data.GFSObject {
	island := data.MakeGFSObject().
		PutLong("user_island_id", i.UserIslandID).
		PutLong("user", i.BBBID).
		PutLong("upgrading_until", 0).
		PutLong("upgrade_started", 0).
		PutLong("date_created", 0).
		PutUtfString("name", "Island").
		PutInt("last_player_level", 30).
		PutInt("likes", i.Likes).
		PutInt("dislikes", i.Dislikes).
		PutInt("level", 30).
		PutInt("type", int(i.UserIslandID)).
		PutInt("island", int(i.IslandID)).
		PutDouble("warp_speed", i.WarpSpeed)

	structures := data.MakeGFSArray()
	for _, s := range i.Structures {
		structures.AddSFSObject(s.GetSFSObject())
	}
	monsters := data.MakeGFSArray()
	giMappings := data.MakeGFSArray()
	if i.IsGold() {
		for _, gm := range i.GoldMonsters {
			monsters.AddSFSObject(gm.goldSFSObject(i.UserIslandID))
			giMappings.AddSFSObject(data.MakeGFSObject().
				PutLong("user_monster", gm.ParentMonsterID).
				PutLong("user_gi_monster", gm.UserMonsterID))
		}
	} else {
		for _, m := range i.Monsters {
			monsters.AddSFSObject(m.GetSFSObject())
		}
	}
	eggs := data.MakeGFSArray()
	for _, e := range i.Eggs {
		eggs.AddSFSObject(e.GetSFSObject())
	}
	breeding := data.MakeGFSArray()
	for _, b := range i.Breedings {
		breeding.AddSFSObject(b.GetSFSObject())
	}

	baking := data.MakeGFSArray()
	for _, b := range i.Bakings {
		baking.AddSFSObject(b.GetSFSObject())
	}

	island.PutGFSArray("structures", structures).
		PutGFSArray("monsters", monsters).
		PutGFSArray("breeding", breeding).
		PutGFSArray("torches", data.MakeGFSArray()).
		PutGFSArray("eggs", eggs).
		PutGFSArray("baking", baking).
		PutGFSArray("gi_mappings", giMappings)
	return island
}

func (i *Island) IsGold() bool {
	return i.IslandID == 6
}

func (i *Island) IsEthereal() bool {
	return i.IslandID == 7
}
