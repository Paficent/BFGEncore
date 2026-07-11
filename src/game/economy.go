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

const (
	scratchOffPrice        = 2
	monsterScratchOffPrice = 10

	monsterScratchCooldownMS = 7 * 24 * 60 * 60 * 1000
)

func registerEconomyHandlers(m *Manager) {
	m.HandleReply("gs_player_has_scratch_off", func(ctx *Context) *data.GFSObject {
		kind := ctx.Str("type")
		success := int64(0)
		if kind == "M" {
			if p := ctx.Player(); p != nil && p.freeMonsterScratch(nowMS()) {
				success = 1
			}
		}
		return data.MakeGFSObject().
			PutUtfString("type", kind).
			PutLong("success", success)
	})

	m.HandlePlayer("gs_purchase_scratch_off", func(ctx *Context, p *Player) {
		kind := ctx.Str("type")
		switch kind {
		case "C":
			p.Buy(0, scratchOffPrice, 0)
		case "M":
			p.Buy(0, monsterScratchOffPrice, 0)
		default:
			return
		}
		prize, ok := m.Static.RollScratchOff(kind)
		if !ok {
			return
		}
		m.setPendingScratch(p.BBBID, kind, prize)
		ctx.Reply("gs_play_scratch_off", m.scratchReveal(prize, p.GetProperties()))
	})

	m.HandlePlayer("gs_play_scratch_off", func(ctx *Context, p *Player) {
		kind := ctx.Str("type")
		if prize, ok := m.peekPendingScratch(p.BBBID, kind); ok {
			ctx.Reply("gs_play_scratch_off", m.scratchReveal(prize, p.GetProperties()))
			return
		}
		if kind == "M" && p.freeMonsterScratch(nowMS()) {
			if prize, ok := m.Static.RollScratchOff("M"); ok {
				p.MonsterScratchTime = nowMS()
				m.setPendingScratch(p.BBBID, "M", prize)
				ctx.Reply("gs_play_scratch_off", m.scratchReveal(prize, p.GetProperties()))
				return
			}
		}
		ctx.Reply("gs_play_scratch_off", data.MakeGFSObject().
			PutLong("success", 0).
			PutGFSArray("properties", p.GetProperties()))
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
		local := 0
		if p := ctx.Player(); p != nil {
			local = p.MemoryHighScore
		}
		ctx.Reply("gs_get_memory_game_numbers", data.MakeGFSObject().
			// difficulty?
			PutInt("tier1ResponseLevel", 5).
			PutInt("tier2ResponseLevel", 10).
			PutInt("tier3ResponseLevel", 20).
			PutInt("tier4ResponseLevel", 50).
			// double tap
			PutFloat("doubleTapChance", 0.5).
			PutInt("doubleTapBeginStep", 10).
			// swapping
			PutFloat("swapAnimationSpeed", 5000.0).
			PutInt("stepDurationOfSwap", 1).
			PutFloat("monsterSwapChance", 0.5).
			PutInt("swapBeginStep", -1).
			// timing
			PutFloat("failPauseDuration", 1.0).
			PutFloat("postSwapPauseDuration", 0.5).
			PutFloat("postNotePauseDuration", 0.0).
			PutFloat("startSeqPauseDuration", 0.0).
			PutFloat("startGamePauseDuration", 2.0).
			// audio
			PutFloat("toneDuration", 2.0).
			PutInt("fixedToneDuration", 0).
			PutInt("memoryGameAudioSampleNumber", 100).
			// scoring
			PutInt("topscore", m.recordMemoryScore(0)).
			PutInt("prev_highscore", local).
			PutBool("success", true))
	})

	// TODO: Assuming that daily minigame if available would return 0?
	m.HandleReply("gs_memory_minigame_current_cost", func(ctx *Context) *data.GFSObject {
		//PutInt("diamond_cost", 0).
		return data.MakeGFSObject().PutInt("diamond_cost", 2).PutInt("coin_cost", 0).PutBool("success", true)
	})

	// TODO: Assuming that daily minigame if available would return 0?
	m.HandleReply("gs_purchase_memory_mini_game", func(ctx *Context) *data.GFSObject {
		//PutInt("diamond_cost", 0).
		return data.MakeGFSObject().PutInt("diamond_cost", 2).PutInt("coin_cost", 0).PutBool("success", true)
	})

	m.HandlePlayer("gs_collect_memory_mini_game", func(ctx *Context, p *Player) {
		score := ctx.Int("score")
		prev := p.MemoryHighScore
		newHigh := int64(0)
		if score > p.MemoryHighScore {
			p.MemoryHighScore = score
			newHigh = 1
		}
		globalHigh := m.recordMemoryScore(score)

		coinReward := score * 100
		foodReward := score * 50
		diamondReward := 0
		if newHigh != 0 {
			diamondReward = 2
		}
		p.AddProperties(int64(coinReward), int64(diamondReward), int64(foodReward), 0, 0)

		ctx.Reply("gs_collect_memory_mini_game", data.MakeGFSObject().
			PutLong("new_high_score", newHigh).
			PutInt("diamond_reward", diamondReward).
			PutInt("coin_reward", coinReward).
			PutInt("food_reward", foodReward).
			PutGFSArray("properties", p.GetProperties()).
			PutInt("topscore", globalHigh).
			PutInt("prev_highscore", prev).
			PutInt("diamond_replay_cost", 2).
			PutInt("coin_replay_cost", 0).
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

		giMonster := &Monster{
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
			PutGFSObject("monster", giMonster.goldSFSObject(goldIsland.UserIslandID)))
	})

	// currency conversions
	m.HandlePlayer("gs_currency_conversion", func(ctx *Context, p *Player) {
		if !p.AddProperties(0, -50, 0, 0, 0) {
			return
		}
		p.AddProperties(1_000_000, 0, 0, 0, 0)
		ctx.Reply("gs_update_properties", data.MakeGFSObject().PutGFSArray("properties", p.GetProperties()))
	})

	m.HandlePlayer("gs_currency_diamonds2eth_conversion", func(ctx *Context, p *Player) {
		if !p.AddProperties(0, -50, 0, 0, 0) {
			return
		}
		p.AddProperties(0, 0, 0, 0, 100)
		ctx.Reply("gs_update_properties", data.MakeGFSObject().PutGFSArray("properties", p.GetProperties()))
	})

	m.HandlePlayer("gs_currency_coins2eth_conversion", func(ctx *Context, p *Player) {
		if !p.AddProperties(-500_000, 0, 0, 0, 0) {
			return
		}
		p.AddProperties(0, 0, 0, 0, 50)
		ctx.Reply("gs_update_properties", data.MakeGFSObject().PutGFSArray("properties", p.GetProperties()))
	})

	// success + the player's current properties (for some reason)
	withProperties := []string{
		"gs_collect_daily_reward",
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

	m.HandlePlayer("gs_collect_scratch_off", func(ctx *Context, p *Player) {
		prize, ok := m.takePendingScratch(p.BBBID, ctx.Str("type"))
		if !ok {
			ctx.Reply("gs_collect_scratch_off", data.MakeGFSObject().PutLong("success", 0))
			return
		}
		if prize.Prize == "monster" {
			m.awardEgg(ctx, p, ctx.Island(), prize.Amount)
			ctx.Reply("gs_collect_scratch_off", data.MakeGFSObject().
				PutLong("success", 0).
				PutLong("has_egg", 0))
			return
		}
		switch prize.Prize {
		case "coins":
			p.AddProperties(int64(prize.Amount), 0, 0, 0, 0)
		case "diamonds":
			p.AddProperties(0, int64(prize.Amount), 0, 0, 0)
		case "food":
			p.AddProperties(0, 0, int64(prize.Amount), 0, 0)
		}
		ctx.Reply("gs_collect_scratch_off", data.MakeGFSObject().
			PutLong("success", 1).
			PutGFSArray("properties", p.GetProperties()))
	})

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
