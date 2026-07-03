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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type DB struct {
	Genes            []Gene
	Levels           []Level
	ScratchOffs      []ScratchOff
	Torches          []Torch
	GameSettings     []GameSetting
	Islands          []Island
	IslandMonsters   []IslandMonster
	IslandStructures []IslandStructure
	Entities         []Entity
	Structures       []Structure
	Monsters         []Monster
	MonsterLevels    []MonsterLevel
	Breeding         []BreedingCombo
	Quests           []Quest
	StoreGroups      []StoreGroup
	StoreCurrencies  []StoreCurrency
	Store            []StoreItem
	Teleports        []Teleport

	entityIndex map[int]Entity
}

func Open(dir string) (*DB, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory of table files", dir)
	}

	l := &loader{dir: dir}
	db := &DB{
		Genes:            read[Gene](l, "genes"),
		Levels:           read[Level](l, "level_xp"),
		ScratchOffs:      read[ScratchOff](l, "scratch_offs"),
		Torches:          read[Torch](l, "island_torches"),
		GameSettings:     read[GameSetting](l, "game_settings"),
		Islands:          read[Island](l, "islands"),
		IslandMonsters:   read[IslandMonster](l, "island_monsters"),
		IslandStructures: read[IslandStructure](l, "island_structures"),
		Entities:         read[Entity](l, "entities"),
		Structures:       read[Structure](l, "structures"),
		Monsters:         read[Monster](l, "monsters"),
		MonsterLevels:    read[MonsterLevel](l, "monster_levels"),
		Breeding:         read[BreedingCombo](l, "breeding_combinations"),
		Quests:           read[Quest](l, "quests"),
		StoreGroups:      read[StoreGroup](l, "store_groups"),
		StoreCurrencies:  read[StoreCurrency](l, "store_currency"),
		Store:            read[StoreItem](l, "store_data"),
		Teleports:        read[Teleport](l, "monster_island_2_island_map"),
	}
	if l.err != nil {
		return nil, l.err
	}

	db.entityIndex = indexBy(db.Entities, func(e Entity) int { return e.ID })
	return db, nil
}

func (db *DB) entity(id int) (Entity, bool) {
	e, ok := db.entityIndex[id]
	return e, ok
}

type loader struct {
	dir string
	err error
}

func read[T any](l *loader, name string) []T {
	if l.err != nil {
		return nil
	}
	rows, err := loadTable[T](l.dir, name)
	if err != nil {
		l.err = err
	}
	return rows
}

func loadTable[T any](dir, name string) ([]T, error) {
	b, err := os.ReadFile(filepath.Join(dir, name+".json"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(b)) == 0 {
		return nil, nil
	}
	var rows []T
	if err := json.Unmarshal(b, &rows); err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}
	return rows, nil
}

func groupBy[T any, K comparable](rows []T, key func(T) K) map[K][]T {
	out := map[K][]T{}
	for _, r := range rows {
		k := key(r)
		out[k] = append(out[k], r)
	}
	return out
}

func indexBy[T any, K comparable](rows []T, key func(T) K) map[K]T {
	out := make(map[K]T, len(rows))
	for _, r := range rows {
		out[key(r)] = r
	}
	return out
}
