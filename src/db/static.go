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

import "github.com/Paficent/GoFox2X/data"

type StaticData struct {
	GameSettings    *data.GFSArray
	Genes           *data.GFSArray
	Islands         *data.GFSArray
	Torches         *data.GFSArray
	Monsters        *data.GFSArray
	Structures      *data.GFSArray
	Levels          *data.GFSArray
	ScratchOffs     *data.GFSArray
	TimedEvents     *data.GFSArray
	StoreItems      *data.GFSArray
	StoreGroups     *data.GFSArray
	StoreCurrencies *data.GFSArray

	LevelXP map[int]int

	QuestDefs   []*QuestDef
	QuestByName map[string]*QuestDef
	QuestByID   map[int]*QuestDef
	QuestStatic map[int]*data.GFSObject
	QuestOrder  []int

	Constants Constants

	monsters       map[int]MonsterInfo
	structures     map[int]StructureInfo
	islands        map[int]IslandInfo
	monsterLevels  map[[2]int]LevelInfo
	teleports      map[[2]int]TeleportDest
	breedingCombos map[[2]int][]breedCombo
	scratchByType  map[string][]ScratchOff
}

func LoadStatic(db *DB) *StaticData {
	questDefs, questByName, questByID := loadQuestDefs(db)
	questStatic, questOrder := loadQuestStatics(db)
	monLevels := loadMonsterLevels(db)
	return &StaticData{
		GameSettings:    getGameSettings(db),
		Genes:           getGenes(db),
		Islands:         getIslands(db),
		Torches:         getTorchData(db),
		Monsters:        getMonsters(db, monLevels),
		Structures:      getStructures(db),
		Levels:          getLevels(db),
		ScratchOffs:     getScratchOffs(db),
		TimedEvents:     getTimedEvents(db),
		StoreItems:      getStoreItems(db),
		StoreGroups:     getStoreGroups(db),
		StoreCurrencies: getStoreCurrencies(db),

		LevelXP: loadLevelXP(db),

		QuestDefs:   questDefs,
		QuestByName: questByName,
		QuestByID:   questByID,
		QuestStatic: questStatic,
		QuestOrder:  questOrder,

		monsters:       loadMonsters(db),
		structures:     loadStructures(db),
		islands:        loadIslands(db),
		monsterLevels:  loadMonsterLevels(db),
		teleports:      loadTeleportInfo(db),
		breedingCombos: loadBreedingCombos(db),
		scratchByType:  loadScratchOffs(db),
	}
}

const (
	bbsURL   = "https://127.0.0.1:9933"
	bbsTitle = "placeholder"
)

var (
	skipMonsterIDs   = map[int]bool{30: true, 79: true, 80: true}
	skipStructureIDs = map[int]bool{232: true, 233: true, 234: true, 235: true, 236: true}

	monsterBinIDs = map[int]string{
		32: "S01", 33: "CR", 34: "V", 49: "W", 52: "X", 50: "L", 55: "G",
		56: "M", 57: "KM", 59: "GM", 75: "G", 76: "M", 77: "L", 78: "LM",
		82: "001_E_rare.bin",
	}
	etherealBinIDs = map[int]string{
		50: "G", 54: "J", 56: "M", 57: "L", 58: "LM", 76: "GM",
	}
)
