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
	UserStructureID   int64
	UserIslandID      int64
	StructureID       int64
	X                 int
	Y                 int
	IsComplete        int
	IsUpgrading       int
	UpgradeTo         int64 // target structure_id while upgrading; 0 when idle
	Flip              int
	Muted             int
	Scale             float64
	DateCreated       int64
	BuildingCompleted int64
	LastCollection    int64
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
		PutLong("date_created", s.DateCreated).
		PutLong("last_collection", s.LastCollection).
		PutDouble("diamonds_collected", 0)

	if s.IsComplete == 0 || s.IsUpgrading != 0 {
		obj.PutLong("building_completed", s.BuildingCompleted)
	}
	inv := data.MakeGFSArray()
	inv.AddSFSObject(data.MakeGFSObject().PutInt("m", 0))
	obj.PutGFSArray("inv", inv)
	obj.PutUtfString("req", `[{"m":68}]`)
	return obj
}

type Monster struct {
	UserMonsterID  int64
	UserIslandID   int64
	MonsterID      int64
	X              int
	Y              int
	Flip           int
	Level          int
	Happiness      int
	CollectedCoins int
	TimesFed       int
	Volume         float64
	Muted          int
	Name           string
	DateCreated    int64
	LastCollection int64
	MegaPerma      bool
	MegaCurrent    int
	MegaStart      int64
	MegaFinish     int64
}

type GoldMonster struct {
	UserMonsterID   int64   `json:"user_monster_id"`
	ParentMonsterID int64   `json:"parent_monster_id"`
	MonsterID       int64   `json:"monster_id"`
	X               int     `json:"pos_x"`
	Y               int     `json:"pos_y"`
	Flip            int     `json:"flip"`
	Level           int     `json:"level"`
	Name            string  `json:"name"`
	Volume          float64 `json:"volume"`
	Muted           int     `json:"muted"`
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

func (gm *GoldMonster) GetSFSObject(islandID int64) *data.GFSObject {
	return data.MakeGFSObject().
		PutLong("user_monster_id", gm.UserMonsterID).
		PutLong("user_island_id", islandID).
		PutLong("island", islandID).
		PutInt("monster", int(gm.MonsterID)).
		PutInt("pos_x", gm.X).
		PutInt("pos_y", gm.Y).
		PutInt("flip", gm.Flip).
		PutInt("level", gm.Level).
		PutUtfString("name", gm.Name).
		PutDouble("volume", gm.Volume).
		PutInt("muted", gm.Muted).
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

type Island struct {
	UserIslandID int64
	IslandID     int64
	BBBID        int64
	Likes        int
	Dislikes     int
	WarpSpeed    float64
	Structures   []*Structure
	GoldMonsters []*GoldMonster `json:"gold_monsters,omitempty"` // Don't break older saves
	Monsters     []*Monster
	Eggs         []*Egg
	Breedings    []*Breeding
}

func (i *Island) FindStructure(userStructureID int64) *Structure {
	for _, s := range i.Structures {
		if s.UserStructureID == userStructureID {
			return s
		}
	}
	return nil
}

func (i *Island) FindStructureByType(structureID int64) *Structure {
	for _, s := range i.Structures {
		if s.StructureID == structureID {
			return s
		}
	}
	return nil
}

func (i *Island) FindMonster(userMonsterID int64) *Monster {
	for _, m := range i.Monsters {
		if m.UserMonsterID == userMonsterID {
			return m
		}
	}
	return nil
}

func (i *Island) FindGoldMonster(userMonsterID int64) *GoldMonster {
	for _, m := range i.GoldMonsters {
		if m.UserMonsterID == userMonsterID {
			return m
		}
	}
	return nil
}

func (i *Island) FindEgg(userEggID int64) *Egg {
	for _, e := range i.Eggs {
		if e.UserEggID == userEggID {
			return e
		}
	}
	return nil
}

func (i *Island) FindBreeding(userBreedingID int64) *Breeding {
	for _, b := range i.Breedings {
		if b.UserBreedingID == userBreedingID {
			return b
		}
	}
	return nil
}

func (i *Island) AddMonster(m *Monster) {
	i.Monsters = append(i.Monsters, m)
}

func (i *Island) RemoveEgg(userEggID int64) {
	out := i.Eggs[:0]
	for _, e := range i.Eggs {
		if e.UserEggID != userEggID {
			out = append(out, e)
		}
	}
	i.Eggs = out
}

func (i *Island) RemoveStructure(userStructureID int64) {
	out := i.Structures[:0]
	for _, s := range i.Structures {
		if s.UserStructureID != userStructureID {
			out = append(out, s)
		}
	}
	i.Structures = out
}

func (i *Island) RemoveMonster(userMonsterID int64) {
	out := i.Monsters[:0]
	for _, m := range i.Monsters {
		if m.UserMonsterID != userMonsterID {
			out = append(out, m)
		}
	}
	i.Monsters = out
}

func (i *Island) RemoveBreeding(userBreedingID int64) {
	out := i.Breedings[:0]
	for _, b := range i.Breedings {
		if b.UserBreedingID != userBreedingID {
			out = append(out, b)
		}
	}
	i.Breedings = out
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
			monsters.AddSFSObject(gm.GetSFSObject(i.UserIslandID))
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

	island.PutGFSArray("structures", structures).
		PutGFSArray("monsters", monsters).
		PutGFSArray("breeding", breeding).
		PutGFSArray("torches", data.MakeGFSArray()).
		PutGFSArray("eggs", eggs).
		PutGFSArray("baking", data.MakeGFSArray()).
		PutGFSArray("gi_mappings", giMappings)
	return island
}

func (i *Island) IsGold() bool {
	return i.IslandID == 6
}

func (i *Island) IsEthereal() bool {
	return i.IslandID == 7
}
