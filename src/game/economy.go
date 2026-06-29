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

import (
	"strings"

	"github.com/Paficent/GoFox2X/data"
)

func registerEconomyHandlers(m *Manager) {
	m.HandleWrite("gs_currency_conversion", func(ctx *Context) {
		p := ctx.Player()
		if p == nil {
			return
		}
		if !p.AddProperties(0, -50, 0, 0, 0) {
			return
		}
		p.AddProperties(1_000_000, 0, 0, 0, 0)
		ctx.Reply("gs_update_properties", data.MakeGFSObject().PutGFSArray("properties", p.GetProperties()))
	})

	m.HandleReply("gs_player_has_scratch_off", func(ctx *Context) *data.GFSObject {
		return data.MakeGFSObject().PutBool("success", false)
	})

	m.HandleReply("gs_purchase_scratch_off", func(ctx *Context) *data.GFSObject {
		ticket := data.MakeGFSObject().
			PutInt("id", 9).
			PutUtfString("type", "C").
			PutInt("amount", 1000).
			PutUtfString("prize", "diamonds")
		scaled := data.MakeGFSObject().
			PutInt("tier1", 50).
			PutInt("tier2", 100).
			PutInt("tier3", 200)
		resp := data.MakeGFSObject().
			PutBool("success", false).
			PutGFSObject("ticket", ticket).
			PutGFSObject("scaled_prizes", scaled)
		if p := ctx.Player(); p != nil {
			resp.PutGFSArray("properties", p.GetProperties())
		}
		return resp
	})

	// scratch offs are currently unimplemented
	// not replying with false (until implemented) causes UI bugs
	m.HandleReply("gs_play_scratch_off", func(ctx *Context) *data.GFSObject {
		resp := data.MakeGFSObject().PutBool("success", false)
		if p := ctx.Player(); p != nil {
			resp.PutGFSArray("properties", p.GetProperties())
		}
		return resp
	})

	m.HandleReply("gs_get_island_rank", func(ctx *Context) *data.GFSObject {
		return data.MakeGFSObject().PutBool("success", false)
	})

	m.HandleReply("gs_get_random_visit_data", func(ctx *Context) *data.GFSObject {
		return data.MakeGFSObject().PutBool("success", false)
	})

	m.HandleReply("gs_get_friend_visit_data", func(ctx *Context) *data.GFSObject {
		return data.MakeGFSObject().PutBool("success", false)
	})

	m.Handle("gs_get_ranked_island_data", func(ctx *Context) {
		notice := data.MakeGFSObject().
			PutBool("force_logout", false).
			PutUtfString("msg", "No ranked island found")
		ctx.Reply("gs_display_generic_message", notice)
	})

	m.HandleWrite("gs_set_displayname", func(ctx *Context) {
		p := ctx.Player()
		if p == nil {
			return
		}
		name, _ := ctx.Params.GetUtfString("display_name")
		if name == "" {
			name, _ = ctx.Params.GetUtfString("name")
		}
		name = strings.TrimSpace(sanitizeName(name))
		if reason := invalidName(name); reason != "" {
			ctx.Fail("gs_set_displayname", reason)
			return
		}
		p.DisplayName = name
		ctx.Reply("gs_set_displayname", data.MakeGFSObject().
			PutBool("success", true).
			PutUtfString("display_name", p.DisplayName))
	})

	m.Handle("gs_get_memory_game_numbers", func(ctx *Context) {
		ctx.Reply("gs_get_memory_game_numbers", data.MakeGFSObject().
			PutInt("memoryGameAudioSampleNumber", 100).
			PutFloat("toneDuration", 2.0).
			PutFloat("startGamePauseDuration", 2.0).
			PutFloat("startSeqPauseDuration", 0.0).
			PutFloat("postNotePauseDuration", 0.0).
			PutFloat("postSwapPauseDuration", 0.5).
			PutFloat("failPauseDuration", 1.0).
			PutInt("swapBeginStep", -1).
			PutFloat("monsterSwapChance", 0.5).
			PutInt("stepDurationOfSwap", 1).
			PutFloat("swapAnimationSpeed", 5000.0).
			PutInt("doubleTapBeginStep", 10).
			PutFloat("doubleTapChance", 0.5).
			PutInt("tier1ResponseLevel", 5).
			PutInt("tier2ResponseLevel", 10).
			PutInt("tier3ResponseLevel", 20).
			PutInt("tier4ResponseLevel", 50).
			PutInt("fixedToneDuration", 0).
			PutInt("diamondPrice", 2).
			PutInt("coinPrice", 0).
			PutInt("diamondReward", 1).
			PutInt("coinReward", 25).
			PutInt("foodReward", 50).
			PutInt("coinRewardFreq", 1).
			PutInt("foodRewardFreq", 5))
	})

	m.Handle("gs_memory_minigame_current_cost", func(ctx *Context) {
		ctx.Reply("gs_memory_minigame_current_cost", data.MakeGFSObject().
			PutInt("diamond_cost", 2).
			PutInt("coin_cost", 0).
			PutBool("success", true))
	})

	m.HandlePlayer("gs_place_on_gold_island", func(ctx *Context, p *Player) {
		umid := ctx.Int64("user_monster_id")
		parentIslandID := ctx.Int64("user_parent_island_id")
		x, y := ctx.Int("pos_x"), ctx.Int("pos_y")
		flip := ctx.Int("flip")

		var parentIsland *Island
		for _, isl := range p.Islands {
			if isl.UserIslandID == parentIslandID {
				parentIsland = isl
				break
			}
		}
		if parentIsland == nil {
			ctx.Reply("gs_place_on_gold_island", data.MakeGFSObject().
				PutLong("success", 0).
				PutLong("parent_id", parentIslandID))
			return
		}

		mon := parentIsland.FindMonster(umid)
		if mon == nil {
			ctx.Reply("gs_place_on_gold_island", data.MakeGFSObject().
				PutLong("success", 0).
				PutLong("parent_id", parentIslandID))
			return
		}

		var goldIsland *Island
		for _, isl := range p.Islands {
			if isl.IsGold() {
				goldIsland = isl
				break
			}
		}
		if goldIsland == nil {
			ctx.Reply("gs_place_on_gold_island", data.MakeGFSObject().
				PutLong("success", 0).
				PutLong("parent_id", parentIslandID))
			return
		}

		giMonster := &GoldMonster{
			UserMonsterID:   p.NextMonsterID(),
			ParentMonsterID: umid,
			MonsterID:       mon.MonsterID,
			X:               x,
			Y:               y,
			Flip:            flip,
			Level:           mon.Level,
			Name:            mon.Name,
			Volume:          mon.Volume,
			Muted:           mon.Muted,
		}
		goldIsland.GoldMonsters = append(goldIsland.GoldMonsters, giMonster)

		ctx.Reply("gs_place_on_gold_island", data.MakeGFSObject().
			PutLong("success", 1).
			PutLong("user_monster_id", umid).
			PutLong("user_parent_island_id", parentIslandID).
			PutGFSObject("monster", giMonster.GetSFSObject(goldIsland.UserIslandID)))
	})

	// success + the player's current properties (for some reason)
	withProperties := []string{
		"gs_collect_daily_reward",
		"gs_collect_scratch_off",
	}
	for _, cmd := range withProperties {
		m.HandleReply(cmd, func(ctx *Context) *data.GFSObject {
			resp := data.MakeGFSObject().PutBool("success", true)
			if p := ctx.Player(); p != nil {
				resp.PutGFSArray("properties", p.GetProperties())
			}
			return resp
		})
	}

	m.HandleReply("gs_rate_island", func(ctx *Context) *data.GFSObject {
		return data.MakeGFSObject().PutBool("success", true)
	})
	m.HandleReply("gs_referral_request", func(ctx *Context) *data.GFSObject {
		return data.MakeGFSObject().PutBool("success", true)
	})
	m.HandleReply("gs_collect_monster_from_hotel", func(ctx *Context) *data.GFSObject {
		return data.MakeGFSObject().PutBool("success", true)
	})
}
