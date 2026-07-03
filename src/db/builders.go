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
	return buildArray(db.Genes, func(g Gene) *data.GFSObject {
		return data.MakeGFSObject().
			PutUtfString("gene_letter", g.Letter).
			PutUtfString("gene_graphic", g.Graphic).
			PutUtfString("min_server_version", g.MinServer).
			PutInt("gene_id", g.ID).
			PutLong("last_changed", now)
	})
}

func getLevels(db *DB) *data.GFSArray {
	return buildArray(db.Levels, func(l Level) *data.GFSObject {
		return data.MakeGFSObject().
			PutInt("level", l.Level).
			PutInt("xp", l.XP).
			PutInt("max_bakeries", l.MaxBakeries)
	})
}

func getScratchOffs(db *DB) *data.GFSArray {
	return buildArray(db.ScratchOffs, func(s ScratchOff) *data.GFSObject {
		return data.MakeGFSObject().
			PutInt("id", s.ID).
			PutInt("scratch_id", s.ID).
			PutUtfString("type", s.Type).
			PutUtfString("prize", s.Prize).
			PutInt("amount", s.Amount).
			PutInt("probability", s.Probability).
			PutInt("is_top_prize", s.IsTopPrize).
			PutUtfString("min_server_version", s.MinServer)
	})
}

func getTorchData(db *DB) *data.GFSArray {
	return buildArray(db.Torches, func(t Torch) *data.GFSObject {
		return data.MakeGFSObject().
			PutInt("island_id", t.IslandID).
			PutUtfString("torch_graphic", t.Graphic).
			PutLong("last_changed", parseLastChanged(t.LastChanged))
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
	for _, s := range db.GameSettings {
		existing[s.Setting] = true
		arr.AddSFSObject(data.MakeGFSObject().
			PutUtfString("key", s.Setting).
			PutUtfString("value", s.Value))
	}
	for _, kv := range gameSettingDefaults {
		if !existing[kv[0]] {
			arr.AddSFSObject(data.MakeGFSObject().
				PutUtfString("key", kv[0]).
				PutUtfString("value", kv[1]))
		}
	}
	return arr
}

func getIslands(db *DB) *data.GFSArray {
	monstersByIsland := groupBy(db.IslandMonsters, func(m IslandMonster) int { return m.Island })
	structuresByIsland := groupBy(db.IslandStructures, func(s IslandStructure) int { return s.Island })
	now := nowMS()

	return buildArray(db.Islands, func(r Island) *data.GFSObject {
		island := data.MakeGFSObject().
			PutInt("id", r.ID).
			PutInt("island_id", r.ID).
			PutInt("island_type", r.ID).
			PutUtfString("name", r.Name).
			PutUtfString("description", r.Description).
			PutUtfString("genes", r.Genes).
			PutUtfString("midi", r.Midi).
			PutUtfString("min_server_version", r.MinServer).
			PutLong("last_changed", now).
			PutUtfString("fb_object_id", "").
			PutInt("enabled", 1).
			PutInt("level", r.Level).
			PutInt("cost_coins", r.CostCoins).
			PutInt("cost_diamonds", r.CostDiamonds).
			PutInt("castle_structure_id", r.Castle).
			PutUtfString("remix_url", bbsURL).
			PutUtfString("remix_url_2", bbsURL)

		g := r.Graphic.V
		island.PutGFSObject("graphic", data.MakeGFSObject().
			PutUtfString("file", jstr(g["file"])).
			PutUtfString("tileset", jstr(g["tileset"])).
			PutUtfString("grid", "main_grid.bin").
			PutUtfString("bg", jstr(g["bg"])))
		island.PutUtfString("grid", "main_grid.bin")

		monsters := data.MakeGFSArray()
		for _, m := range monstersByIsland[r.ID] {
			if skipMonsterIDs[m.Monster] {
				continue
			}
			monsters.AddSFSObject(data.MakeGFSObject().
				PutInt("monster", m.Monster).
				PutUtfString("instrument", m.Instrument))
		}
		island.PutGFSArray("monsters", monsters)

		structures := data.MakeGFSArray()
		for _, s := range structuresByIsland[r.ID] {
			structures.AddSFSObject(data.MakeGFSObject().
				PutInt("structure", s.Structure).
				PutUtfString("instrument", s.Instrument))
		}
		island.PutGFSArray("structures", structures)
		return island
	})
}

func getStructures(db *DB) *data.GFSArray {
	now := nowMS()
	return buildArray(db.Structures, func(r Structure) *data.GFSObject {
		if skipStructureIDs[r.ID] {
			return nil
		}
		structure := data.MakeGFSObject().
			PutInt("structure_id", r.ID).
			PutInt("id", r.ID).
			PutInt("entity_id", r.Entity).
			PutUtfString("structure_type", r.Type).
			PutInt("upgrades_to", r.UpgradesTo).
			PutUtfString("sound", r.Sound).
			PutLong("last_changed", now).
			PutInt("limit_to_island", r.LimitToIsland)

		structure.PutGFSObject("extra", dynamicObject(r.Extra.V))
		addEntityData(db, structure, r.Entity)
		return structure
	})
}

func getMonsters(db *DB, levels map[[2]int]LevelInfo) *data.GFSArray {
	now := nowMS()

	levelsByMonster := map[int][]int{}
	for key := range levels {
		levelsByMonster[key[0]] = append(levelsByMonster[key[0]], key[1])
	}
	for _, lvls := range levelsByMonster {
		sort.Ints(lvls)
	}

	return buildArray(db.Monsters, func(r Monster) *data.GFSObject {
		if skipMonsterIDs[r.ID] {
			return nil
		}
		monster := data.MakeGFSObject().
			PutInt("monster_id", r.ID).
			PutInt("id", r.ID).
			PutInt("entity_id", r.Entity).
			PutUtfString("genes", r.Genes).
			PutUtfString("common_name", "Monster").
			PutUtfString("spore_graphic", "spore_"+r.Genes).
			PutBool("limited", true).
			PutLong("last_changed", now).
			PutInt("beds", r.Beds).
			PutInt("hide_friends", 0)

		happiness := data.MakeGFSArray()
		for _, h := range r.Happiness.V {
			happiness.AddSFSObject(data.MakeGFSObject().
				PutInt("entity", h.Entity).
				PutInt("value", h.Value))
		}
		monster.PutGFSArray("happiness", happiness)
		monster.PutGFSArray("likes", happiness)
		monster.PutGFSArray("dislikes", data.MakeGFSArray())

		names := data.MakeGFSArray()
		for _, n := range r.Names.V {
			names.AddUtfString(n)
		}
		monster.PutGFSArray("names", names)
		monster.PutInt("level_up_xp", r.LevelUpXP)
		monster.PutUtfString("levelup_island", r.LevelupIsland)

		binsID := monsterBinIDs[r.ID]
		if strings.EqualFold(r.LevelupIsland, "ethereal") {
			binsID = etherealBinIDs[r.ID]
		}
		if binsID != "" {
			monster.PutUtfString("bins_id", binsID)
			monster.PutUtfString("bin_id", binsID)
		}

		monster.PutUtfString("link_title", bbsTitle)
		monster.PutUtfString("link_address", bbsURL)

		addEntityData(db, monster, r.Entity)

		levelsArr := data.MakeGFSArray()
		for _, lvl := range levelsByMonster[r.ID] {
			li := levels[[2]int{r.ID, lvl}]
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
