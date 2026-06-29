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
	"github.com/Paficent/GoFox2X/data"
)

func registerStructureHandlers(m *Manager) {
	static := m.Static

	m.HandlePlayer("gs_buy_structure", func(ctx *Context, p *Player) {
		island := ctx.Island()
		structureID := ctx.Int64("structure_id")
		info, ok := static.StructureBuy[int(structureID)]
		if !ok {
			ctx.Fail("gs_buy_structure", "Unknown structure")
			return
		}
		if !p.BuyInIsland(int64(info.CostCoins), int64(info.CostDiamonds), int64(info.CostEth), island) {
			return
		}
		now := nowMS()
		s := &Structure{
			UserStructureID:   p.NextStructureID(),
			UserIslandID:      island.UserIslandID,
			StructureID:       structureID,
			X:                 ctx.Int("pos_x"),
			Y:                 ctx.Int("pos_y"),
			IsComplete:        1,
			IsUpgrading:       0,
			Flip:              ctx.Int("flip"),
			Muted:             0,
			Scale:             ctx.Float("scale"),
			BuildingCompleted: now,
			DateCreated:       now,
			LastCollection:    now,
		}
		if s.Scale == 0 {
			s.Scale = 1.0
		}
		island.Structures = append(island.Structures, s)

		structures := data.MakeGFSArray()
		structures.AddSFSObject(s.GetSFSObject())

		ctx.Reply("gs_buy_structure", data.MakeGFSObject().
			PutLong("success", 1).
			PutGFSArray("properties", p.GetProperties()).
			PutGFSObject("user_structure", s.GetSFSObject()))

		// ctx.Reply("gs_update_structure", data.MakeGFSObject().PutGFSArray("structures", structures))
	})

	m.HandlePlayer("gs_sell_structure", func(ctx *Context, p *Player) {
		island := ctx.Island()
		usid := ctx.Int64("user_structure_id")
		s := island.FindStructure(usid)
		if s == nil {
			ctx.Fail("gs_sell_structure", "Invalid structure ID")
			return
		}
		if info, ok := static.StructureBuy[int(s.StructureID)]; ok {
			p.AddProperties(int64(info.CostCoins*3/4), 0, 0, 0, 0)
		}
		island.RemoveStructure(usid)
		ctx.Reply("gs_sell_structure", data.MakeGFSObject().
			PutLong("success", 1).
			PutLong("user_structure_id", usid).
			PutGFSArray("properties", p.GetProperties()))
	})

	m.HandlePlayer("gs_flip_structure", func(ctx *Context, p *Player) {
		island := ctx.Island()
		usid := ctx.Int64("user_structure_id")
		s := island.FindStructure(usid)
		if s == nil {
			ctx.Fail("gs_flip_structure", "Invalid structure ID")
			return
		}
		s.Flip = boolInt(s.Flip == 0)
		props := data.MakeGFSArray()
		props.AddSFSObject(data.MakeGFSObject().PutInt("flip", s.Flip))
		ctx.Reply("gs_update_structure", data.MakeGFSObject().
			PutLong("user_structure_id", usid).
			PutGFSArray("properties", props))
	})

	m.HandlePlayer("gs_mute_structure", func(ctx *Context, p *Player) {
		island := ctx.Island()
		usid := ctx.Int64("user_structure_id")
		s := island.FindStructure(usid)
		if s == nil {
			ctx.Fail("gs_mute_structure", "Invalid structure ID")
			return
		}
		s.Muted = boolInt(s.Muted == 0)
		props := data.MakeGFSArray()
		props.AddSFSObject(data.MakeGFSObject().PutInt("muted", s.Muted))
		ctx.Reply("gs_update_structure", data.MakeGFSObject().
			PutLong("user_structure_id", usid).
			PutGFSArray("properties", props))
	})

	m.HandlePlayer("gs_move_structure", func(ctx *Context, p *Player) {
		userStructureID := ctx.Int64("user_structure_id")
		newX := ctx.Int("pos_x")
		newY := ctx.Int("pos_y")
		scale := ctx.Float("scale")

		island := ctx.Island()
		if island == nil {
			ctx.Reply("gs_move_structure", data.MakeGFSObject().PutBool("success", false))
			return
		}
		s := island.FindStructure(userStructureID)
		if s == nil {
			ctx.Reply("gs_move_structure", data.MakeGFSObject().PutBool("success", false))
			return
		}
		s.X = newX
		s.Y = newY
		s.Scale = scale

		properties := p.GetProperties()
		properties.AddSFSObject(data.MakeGFSObject().PutInt("pos_x", newX))
		properties.AddSFSObject(data.MakeGFSObject().PutInt("pos_y", newY))
		properties.AddSFSObject(data.MakeGFSObject().PutDouble("scale", scale))

		ctx.Reply("gs_move_structure", data.MakeGFSObject().
			PutGFSArray("properties", properties).
			PutLong("user_structure_id", userStructureID).
			PutGFSObject("user_structure", s.GetSFSObject()).
			PutBool("success", true))

		ctx.Reply("gs_update_structure", data.MakeGFSObject().
			PutLong("user_structure_id", userStructureID).
			PutGFSArray("properties", properties))
	})

	m.HandlePlayer("gs_start_upgrade_structure", func(ctx *Context, p *Player) {
		island := ctx.Island()
		usid := ctx.Int64("user_structure_id")
		s := island.FindStructure(usid)
		if s == nil {
			ctx.Fail("gs_start_upgrade_structure", "Structure not found")
			return
		}
		if s.IsUpgrading == 1 {
			ctx.Fail("gs_start_upgrade_structure", "Already upgrading")
			return
		}
		nextID := static.StructureUpgradesTo[int(s.StructureID)]
		if nextID == 0 {
			ctx.Fail("gs_start_upgrade_structure", "No upgrade available")
			return
		}
		cost, ok := static.StructureBuy[nextID]
		if !ok {
			ctx.Fail("gs_start_upgrade_structure", "Upgrade data missing")
			return
		}
		if !p.Buy(int64(cost.CostCoins), int64(cost.CostDiamonds), int64(cost.CostEth)) {
			ctx.Fail("gs_start_upgrade_structure", "Not enough currency")
			return
		}
		now := nowMS()

		s.UpgradeTo = int64(nextID)
		s.IsUpgrading = 1
		s.IsComplete = 0
		s.BuildingCompleted = now + int64(cost.BuildTime)*1000

		bbbID := p.BBBID
		m.ScheduleAt(upgradeKey(bbbID, usid), s.BuildingCompleted, func() {
			m.completeUpgrade(bbbID, usid)
		})

		props := p.GetProperties()
		props.AddSFSObject(data.MakeGFSObject().PutInt("is_complete", 0))
		props.AddSFSObject(data.MakeGFSObject().PutInt("is_upgrading", 1))
		props.AddSFSObject(data.MakeGFSObject().PutLong("building_completed", s.BuildingCompleted))
		props.AddSFSObject(data.MakeGFSObject().PutLong("date_created", s.DateCreated))
		ctx.Reply("gs_update_structure", data.MakeGFSObject().
			PutLong("user_structure_id", s.UserStructureID).
			PutGFSArray("properties", props))
	})

	m.HandlePlayer("gs_finish_upgrade_structure", func(ctx *Context, p *Player) {
		island := ctx.Island()
		usid := ctx.Int64("user_structure_id")
		s := island.FindStructure(usid)
		if s == nil {
			ctx.Fail("gs_finish_upgrade_structure", "Structure not found")
			return
		}
		m.finishUpgradeNow(p, s)
		ctx.Reply("gs_finish_upgrade_structure", upgradedStructurePayload(p, s))
	})

	m.HandlePlayer("gs_speed_up_structure", func(ctx *Context, p *Player) {
		island := ctx.Island()
		usid := ctx.Int64("user_structure_id")
		s := island.FindStructure(usid)
		if s == nil {
			ctx.Fail("gs_speed_up_structure", "Structure not found")
			return
		}
		wasUpgrade := s.UpgradeTo != 0
		m.finishUpgradeNow(p, s)

		ctx.Reply("gs_speed_up_structure", data.MakeGFSObject().
			PutLong("success", 1).
			PutLong("user_structure_id", usid).
			PutGFSArray("properties", p.GetProperties()))
		if wasUpgrade {
			ctx.Reply("gs_finish_upgrade_structure", upgradedStructurePayload(p, s))
		}
	})

	m.HandlePlayer("gs_finish_structure", func(ctx *Context, p *Player) {
		island := ctx.Island()
		usid := ctx.Int64("user_structure_id")
		s := island.FindStructure(usid)
		if s == nil {
			ctx.Fail("gs_finish_structure", "Structure not found")
			return
		}

		wasUpgrade := s.UpgradeTo != 0
		m.finishUpgradeNow(p, s)

		ctx.Reply("gs_finish_structure", data.MakeGFSObject().
			PutLong("success", 1).
			PutLong("user_structure_id", usid).
			PutGFSObject("user_structure", s.GetSFSObject()).
			PutGFSArray("properties", p.GetProperties()))
		if wasUpgrade {
			ctx.Reply("gs_finish_upgrade_structure", upgradedStructurePayload(p, s))
		}
	})

	m.HandlePlayer("gs_clear_obstacle", func(ctx *Context, p *Player) {
		island := ctx.Island()
		usid := ctx.Int64("user_structure_id")
		s := island.FindStructure(usid)
		if s == nil {
			ctx.Fail("gs_clear_obstacle", "Invalid structure ID")
			return
		}

		if info, ok := static.StructureBuy[int(s.StructureID)]; ok && s.IsComplete != 0 {
			if !p.Buy(int64(info.CostCoins), int64(info.CostDiamonds), int64(info.CostEth)) {
				ctx.Fail("gs_clear_obstacle", "Not enough resources to clear obstacle")
				return
			}
		}
		m.CancelTimer(clearObstacleKey(p.BBBID, usid))
		m.removeObstacle(p, usid)
		ctx.Reply("gs_clear_obstacle", obstacleClearedPayload(p, usid))
	})

	m.HandlePlayer("gs_start_obstacle", func(ctx *Context, p *Player) {
		island := ctx.Island()
		usid := ctx.Int64("user_structure_id")
		s := island.FindStructure(usid)
		if s == nil {
			ctx.Fail("gs_start_obstacle", "Invalid structure ID")
			return
		}
		info, ok := static.StructureBuy[int(s.StructureID)]
		if ok {
			if !p.Buy(int64(info.CostCoins), int64(info.CostDiamonds), int64(info.CostEth)) {
				ctx.Fail("gs_start_obstacle", "Not enough resources to clear obstacle")
				return
			}
		}
		clearMS := int64(0)
		if ok {
			clearMS = int64(info.BuildTime) * 1000
		}
		if clearMS <= 0 {
			m.removeObstacle(p, usid)
			ctx.Reply("gs_start_obstacle", data.MakeGFSObject().
				PutLong("success", 1).
				PutLong("user_structure_id", usid).
				PutGFSObject("user_structure", s.GetSFSObject()).
				PutGFSArray("properties", p.GetProperties()))
			m.Push(p.BBBID, "gs_clear_obstacle", obstacleClearedPayload(p, usid))
			return
		}
		s.IsComplete = 0
		s.BuildingCompleted = nowMS() + clearMS
		bbbID := p.BBBID
		m.ScheduleAt(clearObstacleKey(bbbID, usid), s.BuildingCompleted, func() {
			m.completeClearObstacle(bbbID, usid)
		})
		ctx.Reply("gs_start_obstacle", data.MakeGFSObject().
			PutLong("success", 1).
			PutLong("user_structure_id", usid).
			PutGFSObject("user_structure", s.GetSFSObject()).
			PutGFSArray("properties", p.GetProperties()))
	})

	m.HandlePlayer("gs_clear_obstacle_speed_up", func(ctx *Context, p *Player) {
		island := ctx.Island()
		usid := ctx.Int64("user_structure_id")
		s := island.FindStructure(usid)
		if s == nil {
			ctx.Fail("gs_clear_obstacle_speed_up", "Invalid structure ID")
			return
		}
		m.CancelTimer(clearObstacleKey(p.BBBID, usid))
		m.removeObstacle(p, usid)

		ctx.Reply("gs_clear_obstacle_speed_up", data.MakeGFSObject().PutBool("success", true))
		m.Push(p.BBBID, "gs_clear_obstacle", obstacleClearedPayload(p, usid))
	})

	m.HandlePlayerRead("gs_collect_from_mine", func(ctx *Context, p *Player) {
		island := ctx.Island()
		var mineStructure *Structure
		for _, s := range island.Structures {
			if _, ok := static.MineInfo[int(s.StructureID)]; ok {
				mineStructure = s
				break
			}
		}
		if mineStructure == nil {
			ctx.Reply("gs_collect_from_mine", data.MakeGFSObject().PutLong("success", 0))
			return
		}
		// Timer does not reset for some reason, but at least the game does not crash
		now := nowMS()
		mine := static.MineInfo[int(mineStructure.StructureID)]
		p.AddProperties(0, int64(mine.Diamonds), 0, 0, 0)
		ctx.Reply("gs_update_properties", data.MakeGFSObject().PutGFSArray("properties", p.GetProperties()))
		ctx.Reply("gs_collect_from_mine", data.MakeGFSObject().PutLong("success", 1))
		ctx.Reply("gs_update_structure", data.MakeGFSObject().
			PutLong("user_structure_id", mineStructure.UserStructureID).
			PutLong("last_collection", now).
			PutGFSArray("properties", p.GetProperties()))
	})
}
