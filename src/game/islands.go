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
	"log"

	"github.com/Paficent/GoFox2X/data"
)

func registerIslandHandlers(m *Manager) {
	static := m.Static

	m.HandlePlayer("gs_buy_island", func(ctx *Context, p *Player) {
		islandID := ctx.Int64("island_id")
		if islandID == 0 {
			ctx.Fail("gs_buy_island", "Error")
			return
		}
		for _, isl := range p.Islands {
			if isl.IslandID == islandID {
				ctx.Fail("gs_buy_island", "Error")
				return
			}
		}
		info, ok := static.Island(int(islandID))
		if !ok {
			ctx.Fail("gs_buy_island", "Error")
			return
		}
		if !p.AddProperties(-int64(info.CostCoins), -int64(info.CostDiamonds), 0, 0, 0) {
			ctx.Fail("gs_buy_island", "Not enough resources")
			return
		}
		island := p.NewIsland(islandID, int64(info.Castle), info.Obstacle)
		p.Islands = append(p.Islands, island)
		ctx.Reply("gs_buy_island", data.MakeGFSObject().
			PutLong("success", 1).
			PutGFSArray("properties", p.GetProperties()).
			PutGFSObject("user_island", island.GetSFSObject()))
	})

	m.HandlePlayer("gs_change_island", func(ctx *Context, p *Player) {
		islandID := ctx.Int64("user_island_id")
		owned := false
		var newIsland *Island

		for _, isl := range p.Islands {
			if isl.UserIslandID == islandID {
				owned = true
				newIsland = isl
				break
			}
		}

		resp := data.MakeGFSObject()
		if owned {
			p.ActiveIsland = islandID
			resp.PutBool("success", true).PutLong("user_island_id", islandID)
		} else {
			resp.PutBool("success", false).PutUtfString("error", "You don't have this island")
		}
		
		resp.PutGFSObject("hidden_objects", data.MakeGFSObject().PutGFSArray("objects", data.MakeGFSArray()))
		ctx.Reply("gs_change_island", resp)
		
		if owned {
			for _, id := range newIsland.KillMonsters {
				ctx.Reply("gs_sell_monster", data.MakeGFSObject().
				
					PutLong("pure_destroy", 1).
					PutLong("user_monster_id", int64(*id)))
			}
			
			newIsland.KillMonsters = []*int{}

		}
	})

	m.HandlePlayer("gs_light_torch", func(ctx *Context, p *Player) {
		permanent := ctx.Int("permanent") != 0 || ctx.Int("is_permanent") != 0 || ctx.Int("permalit") != 0
		cost := int64(2)
		if permanent {
			cost = 100
		}
		if !p.AddProperties(0, -cost, 0, 0, 0) {
			ctx.Fail("gs_light_torch", "Not enough diamonds")
			return
		}
		islandID := ctx.Int64("user_island_id")
		if islandID == 0 {
			islandID = p.ActiveIsland
		}
		animations := data.MakeGFSArray()
		animations.AddSFSObject(data.MakeGFSObject().
			PutUtfString("animation", "light_torch").
			PutUtfString("animation_alias", "torch_lighting").
			PutLong("user_island_id", islandID))
		ctx.Reply("gs_light_torch", data.MakeGFSObject().
			PutBool("success", true).
			PutLong("user_island_id", islandID).
			PutUtfString("animation", "light_torch").
			PutGFSArray("properties", p.GetProperties()).
			PutGFSArray("animations", animations))
	})

	m.Handle("gs_get_torchgifts", func(ctx *Context) {
		props := data.MakeGFSArray()
		props.AddSFSObject(data.MakeGFSObject().PutGFSArray("can_gift_torch_times", data.MakeGFSArray()))
		ctx.Reply("gs_get_torchgifts", data.MakeGFSObject().
			PutBool("success", true).
			PutGFSArray("torch_gifts", data.MakeGFSArray()).
			PutGFSArray("properties", props))
	})

	m.HandlePlayerRead("gs_save_island_warp_speed", func(ctx *Context, p *Player) {
		// (gfs_object):
		//       (long) user_island_id: 2
		//       (double) warp_speed: 1.927374243736267

		// islandID := ctx.Int64("user_island_id")
		warpSpeed := ctx.Float("warp_speed")
		log.Print(warpSpeed)

		// assumes p.ActiveIsland is correct
		activeIsland := p.GetActiveIsland()
		activeIsland.WarpSpeed = warpSpeed
	})
}
