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

func registerBakingHandlers(m *Manager) {
	static := m.Static

	m.HandlePlayer("gs_start_baking", func(ctx *Context, p *Player) {
		island := ctx.Island()
		usid := ctx.Int64("user_structure_id")
		foodIndex := ctx.Int("food_index")
		s := island.FindStructure(usid)
		if s == nil {
			ctx.Fail("gs_start_baking", "Invalid structure ID")
			return
		}
		opts := static.FoodOptions[int(s.StructureID)]
		if foodIndex < 0 || foodIndex >= len(opts) {
			ctx.Fail("gs_start_baking", "Invalid food option")
			return
		}
		opt := opts[foodIndex]
		if !p.Buy(int64(opt.Cost), 0, 0) {
			ctx.Fail("gs_start_baking", "Not enough coins to bake")
			return
		}
		now := nowMS()
		baking := &Baking{
			IslandID:        island.UserIslandID,
			UserBakingID:    p.NextBakingID(),
			UserStructureID: usid,
			FoodIndex:       foodIndex,
			Food:            opt.Food,
			Xp:              opt.Xp,
			StartedOn:       now,
			CompleteOn:      now + int64(opt.Time)*1000,
		}
		island.Bakings = append(island.Bakings, baking)

		ctx.Reply("gs_start_baking", data.MakeGFSObject().
			PutBool("success", true).
			PutLong("user_structure_id", usid).
			PutGFSObject("user_baking", baking.GetSFSObject()).
			PutGFSArray("properties", p.GetProperties()))
	})

	m.HandlePlayer("gs_finish_baking", func(ctx *Context, p *Player) {
		island := ctx.Island()
		bakingID := ctx.Int64("user_baking_id")
		baking := island.FindBaking(bakingID)
		if baking == nil {
			ctx.Fail("gs_finish_baking", "Invalid baking ID")
			return
		}

		island.RemoveBaking(bakingID)
		p.AddProperties(0, 0, int64(baking.Food), int64(baking.Xp), 0)

		ctx.Reply("gs_finish_baking", data.MakeGFSObject().
			PutBool("success", true).
			PutLong("user_baking_id", bakingID).
			PutGFSArray("properties", p.GetProperties()))
	})

	m.HandlePlayer("gs_speed_up_baking", func(ctx *Context, p *Player) {
		island := ctx.Island()
		bakingID := ctx.Int64("user_baking_id")
		baking := island.FindBaking(bakingID)
		if baking == nil {
			ctx.Fail("gs_speed_up_baking", "Invalid baking ID")
			return
		}

		if !p.Buy(0, m.calculateSpeedupCost(p, bakingID, "baking"), 0) {
			ctx.Fail("gs_speed_up_baking", "Not enough diamonds!")
			return
		}

		now := nowMS()
		baking.CompleteOn = now

		p.AddProperties(0, 0, 0, 0, 0) //TODO

		ctx.Reply("gs_speed_up_baking", data.MakeGFSObject().
			PutBool("success", true).
			PutLong("user_baking_id", bakingID).
			PutLong("finished_at", now).
			PutGFSArray("properties", p.GetProperties()))
	})
}
