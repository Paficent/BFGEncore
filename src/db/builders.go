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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Paficent/GoFox2X/data"
)

func jstr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func parseLastChanged(s string) int64 {
	if strings.TrimSpace(s) == "" {
		return nowMS()
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t.Unix() * 1000
	}
	return nowMS()
}

func getGenes(db *DB) *data.GFSArray {
	now := nowMS()
	return buildArray(db.Table("genes"), func(r Row) *data.GFSObject {
		return data.MakeGFSObject().
			PutUtfString("gene_letter", r.Str("gene_letter")).
			PutUtfString("gene_graphic", r.Str("gene_graphic")).
			PutUtfString("min_server_version", r.Str("min_server_version")).
			PutInt("gene_id", r.Int("gene_id")).
			PutLong("last_changed", now)
	})
}

func getLevels(db *DB) *data.GFSArray {
	return buildArray(db.Table("level_xp"), func(r Row) *data.GFSObject {
		return data.MakeGFSObject().
			PutInt("level", r.Int("level")).
			PutInt("xp", r.Int("xp")).
			PutInt("max_bakeries", r.Int("max_bakeries"))
	})
}

func getScratchOffs(db *DB) *data.GFSArray {
	return buildArray(db.Table("scratch_offs"), func(r Row) *data.GFSObject {
		return data.MakeGFSObject().
			PutInt("id", r.Int("id")).
			PutInt("scratch_id", r.Int("id")).
			PutUtfString("type", r.Str("type")).
			PutUtfString("prize", r.Str("prize")).
			PutInt("amount", r.Int("amount")).
			PutInt("probability", r.Int("probability")).
			PutInt("is_top_prize", r.Int("is_top_prize")).
			PutUtfString("min_server_version", r.Str("min_server_version"))
	})
}

func getTorchData(db *DB) *data.GFSArray {
	return buildArray(db.Table("island_torches"), func(r Row) *data.GFSObject {
		return data.MakeGFSObject().
			PutInt("island_id", r.Int("island_id")).
			PutUtfString("torch_graphic", r.Str("torch_graphic")).
			PutLong("last_changed", parseLastChanged(r.Str("last_changed")))
	})
}

var gameSettingDefaults = [][2]string{
	{"USER_SELLING_PERCENTAGE", "0.75"},
	{"USER_MAX_NUM_TORCHES_PER_ISLAND", "10"},
	{"USER_DIAMOND_COST_PER_LIT_TORCH", "2"},
	{"USER_DIAMOND_COST_PER_PERMALIT_TORCH", "100"},
	{"USER_DIAMOND_COST_PER_DAILY_MEGAFY", "50"},
	{"USER_DIAMOND_COST_PER_PERMALIT_MEGAMONSTER", "20"},
	{"USER_COIN_COST_PER_DAILY_MEGAMONSTER", "25000"},
	{"USER_COIN_COST_PER_PERMALIT_MEGAMONSTER", "250000"},
	{"USER_ETHEREAL_ISLAND_HATCH_XP_MODIFIER", "0.027"},
	{"MEMORY_DIAMOND_PRICE", "2"},
	{"MEMORY_COIN_PRICE", "0"},
	{"USER_SCRATCHOFF_PRICE", "2"},
	{"USER_MONSTER_SCRATCHOFF_PRICE", "10"},
	{"USER_MORE_GAMES_IOS", "playhaven"},
	{"USER_MORE_GAMES_ANDROID", "playhaven"},
	{"USER_MORE_GAMES_AMAZON", "chartboost"},
	{"USER_FB_ACHIEVEMENTS_URL", "http://www.bbbarcade.com/mysingingmonsters/msm_facebook/admin/post_achievement.php"},
	{"USER_FB_MONSTERS_URL", "http://www.bbbarcade.com/mysingingmonsters/msm_facebook/content/monsters/jpg/"},
	{"USER_FB_CUSTOM_EVENTS_URL", "http://www.mysingingmonsters.com/facebook/actions/"},
	{"USER_FB_PLATFORM_REDIRECT_URL", "http://www.bbbarcade.com/mysingingmonsters/msm_facebook/platform_redirect.php"},
	{"USER_FB_POST_REWARD_REFRESH", "24"},
	{"USER_COIN_ETH_EXCHANGE_RATE", "500000,50"},
	{"USER_DIAMOND_ETH_EXCHANGE_RATE", "50,100"},
	{"USER_ETH_DIAMOND_EXCHANGE_RATE", "30000,1"},
	{"USER_NEWS_DATA", "0"},
}

func getGameSettings(db *DB) *data.GFSArray {
	arr := data.MakeGFSArray()
	existing := map[string]bool{}
	for _, r := range db.Table("game_settings") {
		setting := r.Str("setting")
		existing[setting] = true
		arr.AddSFSObject(data.MakeGFSObject().
			PutUtfString("key", setting).
			PutUtfString("value", r.Str("value")))
	}
	for _, kv := range gameSettingDefaults { //potentially not necessary
		if !existing[kv[0]] {
			arr.AddSFSObject(data.MakeGFSObject().
				PutUtfString("key", kv[0]).
				PutUtfString("value", kv[1]))
		}
	}
	return arr
}

func getIslands(db *DB) *data.GFSArray {
	monstersByIsland := db.Group("island_monsters", "island")
	structuresByIsland := db.Group("island_structures", "island")
	now := nowMS()

	return buildArray(db.Table("islands"), func(r Row) *data.GFSObject {
		islandID := r.Int("island_id")
		island := data.MakeGFSObject().
			PutInt("id", islandID).
			PutInt("island_id", islandID).
			PutInt("island_type", islandID).
			PutUtfString("name", r.Str("name")).
			PutUtfString("description", r.Str("description")).
			PutUtfString("genes", r.Str("genes")).
			PutUtfString("midi", r.Str("midi")).
			PutUtfString("min_server_version", r.Str("min_server_version")).
			PutLong("last_changed", now).
			PutUtfString("fb_object_id", "").
			PutInt("enabled", 1).
			PutInt("level", r.Int("level")).
			PutInt("cost_coins", r.Int("cost_coins")).
			PutInt("cost_diamonds", r.Int("cost_diamonds")).
			PutInt("castle_structure_id", r.Int("castle_structure_id")).
			PutUtfString("remix_url", bbsURL).
			PutUtfString("remix_url_2", bbsURL)

		g := r.JSON("graphic")
		island.PutGFSObject("graphic", data.MakeGFSObject().
			PutUtfString("file", jstr(g["file"])).
			PutUtfString("tileset", jstr(g["tileset"])).
			PutUtfString("grid", "main_grid.bin").
			PutUtfString("bg", jstr(g["bg"])))
		island.PutUtfString("grid", "main_grid.bin")

		monsters := data.MakeGFSArray()
		for _, m := range monstersByIsland[islandID] {
			if skipMonsterIDs[m.Int("monster")] {
				continue
			}
			monsters.AddSFSObject(data.MakeGFSObject().
				PutInt("monster", m.Int("monster")).
				PutUtfString("instrument", m.Str("instrument")))
		}
		island.PutGFSArray("monsters", monsters)

		structures := data.MakeGFSArray()
		for _, s := range structuresByIsland[islandID] {
			structures.AddSFSObject(data.MakeGFSObject().
				PutInt("structure", s.Int("structure")).
				PutUtfString("instrument", s.Str("instrument")))
		}
		island.PutGFSArray("structures", structures)
		return island
	})
}

func getStructures(db *DB) *data.GFSArray {
	now := nowMS()
	return buildArray(db.Table("structures"), func(r Row) *data.GFSObject {
		structureID := r.Int("structure_id")
		if skipStructureIDs[structureID] {
			return nil
		}
		structure := data.MakeGFSObject().
			PutInt("structure_id", structureID).
			PutInt("id", structureID).
			PutInt("entity_id", r.Int("entity")).
			PutUtfString("structure_type", r.Str("structure_type")).
			PutInt("upgrades_to", r.Int("upgrades_to")).
			PutUtfString("sound", r.Str("sound")).
			PutLong("last_changed", now).
			PutInt("limit_to_island", r.Int("limit_to_island"))

		extra := data.MakeGFSObject()
		putValues(extra, r.JSON("extra"))
		structure.PutGFSObject("extra", extra)

		addEntityData(db, structure, r.Int("entity"))
		return structure
	})
}

func getMonsters(db *DB, levels map[[2]int]LevelInfo) *data.GFSArray {
	now := nowMS()

	levelsByMonster := map[int][]int{}
	for key := range levels {
		monsterID, lvl := key[0], key[1]
		levelsByMonster[monsterID] = append(levelsByMonster[monsterID], lvl)
	}
	for _, lvls := range levelsByMonster {
		sort.Ints(lvls)
	}

	return buildArray(db.Table("monsters"), func(r Row) *data.GFSObject {
		monsterID := r.Int("monster_id")
		if skipMonsterIDs[monsterID] {
			return nil
		}
		genes := r.Str("genes")
		levelupIsland := r.Str("levelup_island")

		monster := data.MakeGFSObject().
			PutInt("monster_id", monsterID).
			PutInt("id", monsterID).
			PutInt("entity_id", r.Int("entity")).
			PutUtfString("genes", genes).
			PutUtfString("common_name", "Monster").
			PutUtfString("spore_graphic", "spore_"+genes).
			PutBool("limited", true).
			PutLong("last_changed", now).
			PutInt("beds", r.Int("beds")).
			PutInt("hide_friends", 0)

		happiness := data.MakeGFSArray()
		for _, h := range r.JSONArray("happiness") {
			if hm, ok := h.(map[string]any); ok {
				happiness.AddSFSObject(data.MakeGFSObject().
					PutInt("entity", numToInt(hm["entity"])).
					PutInt("value", numToInt(hm["value"])))
			}
		}
		monster.PutGFSArray("happiness", happiness)
		monster.PutGFSArray("likes", happiness)
		monster.PutGFSArray("dislikes", data.MakeGFSArray())

		names := data.MakeGFSArray()
		for _, n := range r.JSONArray("names") {
			if s, ok := n.(string); ok {
				names.AddUtfString(s)
			}
		}
		monster.PutGFSArray("names", names)
		monster.PutInt("level_up_xp", r.Int("level_up_xp"))
		monster.PutUtfString("levelup_island", levelupIsland)

		binsID := monsterBinIDs[monsterID]
		if strings.EqualFold(levelupIsland, "ethereal") {
			binsID = etherealBinIDs[monsterID]
		}
		if binsID != "" {
			monster.PutUtfString("bins_id", binsID)
			monster.PutUtfString("bin_id", binsID)
		}

		monster.PutUtfString("link_title", bbsTitle)
		monster.PutUtfString("link_address", bbsURL)

		addEntityData(db, monster, r.Int("entity"))

		levelsArr := data.MakeGFSArray()
		for _, lvl := range levelsByMonster[monsterID] {
			li := levels[[2]int{monsterID, lvl}]
			levelsArr.AddSFSObject(data.MakeGFSObject().
				PutInt("max_coins", li.MaxCoins).
				PutInt("coins", li.Coins).
				PutInt("level", lvl).
				PutInt("food", li.Food).
				PutInt("ethereal_currency", li.Shards).
				PutInt("max_ethereal", li.MaxShards))
		}
		monster.PutGFSArray("levels", levelsArr)
		return monster
	})
}
