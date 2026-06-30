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

func registerMonsterHandlers(m *Manager) {
	static := m.Static

	m.HandlePlayer("gs_buy_egg", func(ctx *Context, p *Player) {
		monsterID := ctx.Int("monster_id")
		info, ok := static.MonsterBuy[monsterID]
		if !ok {
			ctx.Fail("gs_buy_egg", "Unknown monster")
			return
		}
		island := ctx.Island()
		if island == nil {
			ctx.Fail("gs_buy_egg", "Error")
			return
		}
		if !p.BuyInIsland(int64(info.CostCoins), int64(info.CostDiamonds), int64(info.CostEth), island) {
			return
		}
		nursery := island.FindStructureByType(1)
		if nursery == nil {
			ctx.Fail("gs_buy_egg", "Error")
			return
		}

		now := nowMS()
		egg := &Egg{
			IslandID:        island.UserIslandID,
			LaidOn:          now,
			HatchesOn:       now + int64(info.BuildTime)*1000,
			MonsterID:       int64(monsterID),
			UserEggID:       p.NextEggID(),
			UserStructureID: nursery.UserStructureID,
		}
		island.Eggs = append(island.Eggs, egg)

		eggs := data.MakeGFSArray()
		eggs.AddSFSObject(egg.GetSFSObject())
		ctx.Reply("gs_update_eggs", data.MakeGFSObject().PutGFSArray("eggs", eggs))

		ctx.Reply("gs_buy_egg", data.MakeGFSObject().
			PutGFSObject("user_egg", egg.GetSFSObject()).
			PutBool("success", true).
			PutBool("remove_buyback", false).
			PutGFSArray("properties", p.GetProperties()))
	})

	m.HandlePlayer("gs_hatch_egg", func(ctx *Context, p *Player) {
		island := ctx.Island()
		userEggID := ctx.Int64("user_egg_id")
		egg := island.FindEgg(userEggID)
		if egg == nil {
			ctx.Fail("gs_hatch_egg", "Error")
			return
		}

		x, y := ctx.Int("pos_x"), ctx.Int("pos_y")
		if x == 0 {
			x = 1
		}
		if y == 0 {
			y = 1
		}
		flip := ctx.Int("flip")
		island.RemoveEgg(userEggID)

		now := nowMS()
		monster := &Monster{
			UserMonsterID:  p.NextMonsterID(),
			UserIslandID:   island.UserIslandID,
			MonsterID:      egg.MonsterID,
			X:              x,
			Y:              y,
			Flip:           flip,
			Level:          1,
			Happiness:      0,
			Volume:         0.5,
			Name:           "Monster",
			DateCreated:    now,
			LastCollection: now,
		}
		island.AddMonster(monster)

		if p.Level < 4 {
			p.AddProperties(0, 0, 0, 150, 0)
		}

		ctx.Reply("gs_player", data.MakeGFSObject().
			PutGFSObject("player_object", p.GetSFSObject()).
			PutLong("server_time", now))

		ctx.Reply("gs_hatch_egg", data.MakeGFSObject().
			PutGFSArray("properties", p.GetProperties()).
			PutLong("user_egg_id", userEggID).
			PutLong("island", island.UserIslandID).
			PutGFSObject("monster", monster.GetSFSObject()).
			PutBool("success", true).
			PutBool("directPlace", false).
			PutBool("remove_buyback", false).
			PutLong("user_structure_id", egg.UserStructureID))

		ctx.Reply("gs_update_monster", data.MakeGFSObject().
			PutBool("success", true).
			PutLong("user_monster_id", monster.UserMonsterID).
			PutGFSObject("monster", monster.GetSFSObject()))
	})

	m.HandlePlayer("gs_speed_up_hatching", func(ctx *Context, p *Player) {
		island := ctx.Island()
		userEggID := ctx.Int64("user_egg_id")
		egg := island.FindEgg(userEggID)
		if egg == nil {
			ctx.Fail("gs_speed_up_hatching", "Error")
			return
		}
		now := nowMS()
		minutesRemaining := (egg.HatchesOn - now) / 1000 / 60
		diamonds := (minutesRemaining + 59) / 60
		if !p.Buy(0, int64(diamonds), 0) {
			return
		}
		egg.HatchesOn = now
		ctx.Reply("gs_speed_up_hatching", data.MakeGFSObject().
			PutLong("success", 1).
			PutLong("user_egg_id", userEggID).
			PutLong("hatches_on", now).
			PutLong("laid_on", egg.LaidOn).
			PutGFSArray("properties", p.GetProperties()))
	})

	m.HandlePlayer("gs_sell_egg", func(ctx *Context, p *Player) {
		island := ctx.Island()
		userEggID := ctx.Int64("user_egg_id")
		egg := island.FindEgg(userEggID)
		if egg == nil {
			ctx.Fail("gs_sell_egg", "Egg not found")
			return
		}
		if info, ok := static.MonsterBuy[int(egg.MonsterID)]; ok {
			if island.IsEthereal() {
				p.AddProperties(0, 0, 0, 0, int64(info.CostEth*3/4))
			} else {
				p.AddProperties(int64(info.CostCoins*3/4), 0, 0, 0, 0)
			}
		}
		island.RemoveEgg(userEggID)
		ctx.Reply("gs_sell_egg", data.MakeGFSObject().
			PutBool("success", true).
			PutLong("user_egg_id", userEggID).
			PutGFSArray("properties", p.GetProperties()))
	})

	m.HandlePlayer("gs_breed_monsters", func(ctx *Context, p *Player) {
		userMonsterID1 := ctx.Int64("user_monster_id_1")
		userMonsterID2 := ctx.Int64("user_monster_id_2")

		island := ctx.Island()
		if island == nil {
			ctx.Fail("gs_breed_monsters", "Structure ID is required")
			return
		}
		breedingStruct := island.FindStructureByType(2)
		if breedingStruct == nil {
			ctx.Fail("gs_breed_monsters", "Structure ID is required")
			return
		}

		m1 := island.FindMonster(userMonsterID1)
		m2 := island.FindMonster(userMonsterID2)
		if m1 == nil || m2 == nil {
			ctx.Fail("gs_breed_monsters", "Invalid monster IDs")
			return
		}

		result := static.BreedingResult(int(m1.MonsterID), int(m2.MonsterID), m1.Level, m2.Level, p.Level)
		info, ok := static.MonsterBuy[result]
		if !ok {
			ctx.Fail("gs_breed_monsters", "Breeding result not found")
			return
		}

		now := nowMS()
		breeding := &Breeding{
			IslandID:       island.UserIslandID,
			UserBreedingID: p.NextBreedingID(),
			StructureID:    breedingStruct.UserStructureID,
			Monster1:       int(m1.MonsterID),
			Monster2:       int(m2.MonsterID),
			NewMonster:     result,
			StartedOn:      now,
			CompleteOn:     now + int64(info.BuildTime)*1000,
		}
		island.Breedings = append(island.Breedings, breeding)

		ctx.Reply("gs_breed_monsters", data.MakeGFSObject().
			PutLong("success", 1).
			PutLong("last_bred_monster_1", userMonsterID1).
			PutLong("last_bred_monster_2", userMonsterID2).
			PutGFSObject("user_breeding", breeding.GetSFSObject()))
	})

	m.HandlePlayer("gs_finish_breeding", func(ctx *Context, p *Player) {
		island := ctx.Island()
		userBreedingID := ctx.Int64("user_breeding_id")
		breeding := island.FindBreeding(userBreedingID)
		if breeding == nil {
			return
		}
		info, ok := static.MonsterBuy[breeding.NewMonster]
		if !ok {
			return
		}
		if !p.Buy(int64(info.CostCoins), int64(info.CostDiamonds), int64(info.CostEth)) {
			return
		}
		now := nowMS()
		endTime := now + int64(info.BuildTime)*1000
		island.RemoveBreeding(userBreedingID)

		structureID := breeding.StructureID
		if nursery := island.FindStructureByType(1); nursery != nil {
			structureID = nursery.UserStructureID
		}
		egg := &Egg{
			IslandID:        island.UserIslandID,
			LaidOn:          now,
			HatchesOn:       endTime,
			MonsterID:       int64(breeding.NewMonster),
			UserEggID:       p.NextEggID(),
			UserStructureID: structureID,
		}
		island.Eggs = append(island.Eggs, egg)

		ctx.Reply("gs_finish_breeding", data.MakeGFSObject().
			PutGFSObject("user_egg", egg.GetSFSObject()).
			PutLong("success", 1).
			PutLong("user_breeding_id", userBreedingID))
	})

	m.HandlePlayer("gs_speed_up_breeding", func(ctx *Context, p *Player) {
		island := ctx.Island()
		userBreedingID := ctx.Int64("user_breeding_id")
		breeding := island.FindBreeding(userBreedingID)
		if breeding == nil {
			ctx.Fail("gs_speed_up_breeding", "Error")
			return
		}
		now := nowMS()
		breeding.CompleteOn = now
		ctx.Reply("gs_speed_up_breeding", data.MakeGFSObject().
			PutLong("success", 1).
			PutLong("userBreedingId", userBreedingID).
			PutLong("complete_on", now).
			PutLong("started_on", breeding.StartedOn))
	})

	m.HandlePlayer("gs_move_monster", func(ctx *Context, p *Player) {
		island := ctx.Island()
		umid := ctx.Int64("user_monster_id")
		mon := island.FindMonster(umid)
		if mon == nil {
			ctx.Fail("gs_move_monster", "Invalid monster ID")
			return
		}
		mon.X = ctx.Int("pos_x")
		mon.Y = ctx.Int("pos_y")
		mon.Volume = ctx.Float("volume")

		ctx.Reply("gs_move_monster", data.MakeGFSObject().PutBool("success", true))

		ctx.Reply("gs_update_monster", data.MakeGFSObject().
			PutLong("user_monster_id", umid).
			PutInt("pos_x", mon.X).
			PutInt("pos_y", mon.Y).
			PutDouble("volume", mon.Volume).
			PutGFSArray("properties", p.GetProperties()))
	})

	m.HandlePlayer("gs_flip_monster", func(ctx *Context, p *Player) {
		island := ctx.Island()
		umid := ctx.Int64("user_monster_id")

		if island.IsGold() {
			gm := island.FindGoldMonster(umid)
			if gm == nil {
				ctx.Fail("gs_mute_monster", "Invalid monster ID")
				return
			}
			gm.Flip = boolInt(gm.Flip == 0)
			ctx.Reply("gs_flip_monster", data.MakeGFSObject().PutBool("success", true))
			ctx.Reply("gs_update_monster", data.MakeGFSObject().
				PutBool("success", true).
				PutLong("user_monster_id", umid).
				PutGFSObject("monster", gm.GetSFSObject(island.UserIslandID)).
				PutInt("flip", gm.Flip))
		}

		mon := island.FindMonster(umid)
		if mon == nil {
			ctx.Fail("gs_flip_monster", "Invalid monster ID")
			return
		}
		mon.Flip = boolInt(mon.Flip == 0)
		ctx.Reply("gs_flip_monster", data.MakeGFSObject().PutBool("success", true))
		ctx.Reply("gs_update_monster", data.MakeGFSObject().
			PutBool("success", true).
			PutLong("user_monster_id", umid).
			PutInt("flip", mon.Flip).
			PutGFSObject("monster", mon.GetSFSObject()))
	})

	m.HandlePlayer("gs_mute_monster", func(ctx *Context, p *Player) {
		island := ctx.Island()
		umid := ctx.Int64("user_monster_id")

		if island.IsGold() {
			gm := island.FindGoldMonster(umid)
			if gm == nil {
				ctx.Fail("gs_mute_monster", "Invalid monster ID")
				return
			}
			gm.Muted = boolInt(gm.Muted == 0)
			ctx.Reply("gs_mute_monster", data.MakeGFSObject().PutBool("success", true))
			ctx.Reply("gs_update_monster", data.MakeGFSObject().
				PutBool("success", true).
				PutLong("user_monster_id", umid).
				PutGFSObject("monster", gm.GetSFSObject(island.UserIslandID)).
				PutInt("muted", gm.Muted))
		}

		mon := island.FindMonster(umid)
		if mon == nil {
			ctx.Fail("gs_mute_monster", "Invalid monster ID")
			return
		}
		mon.Muted = boolInt(mon.Muted == 0)
		ctx.Reply("gs_mute_monster", data.MakeGFSObject().PutBool("success", true))
		ctx.Reply("gs_update_monster", data.MakeGFSObject().
			PutBool("success", true).
			PutLong("user_monster_id", umid).
			PutGFSObject("monster", mon.GetSFSObject()).
			PutInt("muted", mon.Muted))
	})

	m.HandlePlayer("gs_name_monster", func(ctx *Context, p *Player) {
		island := ctx.Island()
		umid := ctx.Int64("user_monster_id")
		mon := island.FindMonster(umid)
		if mon == nil {
			ctx.Fail("gs_name_monster", "Invalid monster ID")
			return
		}
		name, _ := ctx.Params.GetUtfString("name")
		if name == "" {
			name, _ = ctx.Params.GetUtfString("monster_name")
		}
		if name == "" {
			name, _ = ctx.Params.GetUtfString("newName")
		}
		name = strings.TrimSpace(sanitizeName(name))
		if reason := invalidName(name); reason != "" {
			ctx.Fail("gs_name_monster", reason)
			return
		}
		mon.Name = name
		p.bumpCounter("rename_monster")
		ctx.Reply("gs_name_monster", data.MakeGFSObject().
			PutBool("success", true).
			PutLong("user_monster_id", umid).
			PutUtfString("name", mon.Name).
			PutGFSObject("monster", mon.GetSFSObject()))
	})

	m.HandlePlayer("gs_feed_monster", func(ctx *Context, p *Player) {
		island := ctx.Island()
		umid := ctx.Int64("user_monster_id")
		mon := island.FindMonster(umid)
		if mon == nil {
			ctx.Fail("gs_feed_monster", "Invalid monster ID")
			return
		}
		level, ok := static.MonsterLevel(int(mon.MonsterID), mon.Level)
		if !ok {
			ctx.Fail("gs_feed_monster", "Monster level data not found")
			return
		}
		if !p.AddProperties(0, 0, -int64(level.Food), 0, 0) {
			ctx.Fail("gs_feed_monster", "Not enough food")
			return
		}
		mon.TimesFed++
		leveledUp := false
		if mon.TimesFed >= 4 {
			if _, hasNext := static.MonsterLevel(int(mon.MonsterID), mon.Level+1); hasNext {
				mon.TimesFed = 0
				mon.Level++
				leveledUp = true
			} else {
				mon.TimesFed = 3
			}
		}

		ctx.Reply("gs_feed_monster", data.MakeGFSObject().
			PutBool("success", true).
			PutInt("times_fed", mon.TimesFed).
			PutLong("user_monster_id", umid))

		update := data.MakeGFSObject().
			PutLong("user_monster_id", umid).
			PutInt("times_fed", mon.TimesFed)
		if leveledUp {
			update.PutInt("level", mon.Level).
				PutLong("last_collection", mon.LastCollection).
				PutInt("collected_coins", mon.CollectedCoins).
				PutInt("collected_ethereal", 0).
				PutGFSObject("monster", mon.GetSFSObject())
		}
		update.PutGFSArray("properties", p.GetProperties())
		ctx.Reply("gs_update_monster", update)
	})

	m.HandlePlayer("gs_collect_monster", func(ctx *Context, p *Player) {
		island := ctx.Island()
		umid := ctx.Int64("user_monster_id")
		mon := island.FindMonster(umid)
		if mon == nil {
			ctx.Fail("gs_collect_monster", "Invalid monster ID")
			return
		}
		level, ok := static.MonsterLevel(int(mon.MonsterID), mon.Level)
		if !ok {
			ctx.Fail("gs_collect_monster", "Monster level data not found")
			return
		}
		now := nowMS()
		lastCollection := mon.LastCollection
		if lastCollection == 0 {
			lastCollection = now
		}
		deltaSeconds := (now - lastCollection) / 1000
		mainCurrency := int64(level.Coins)
		maxCurrency := int64(level.MaxCoins)
		useShards := false
		timeDivider := 60 // Normal monsters (coin givers) use per minute
		if mainCurrency == 0 || maxCurrency == 0 {
			mainCurrency = int64(level.Shards)
			maxCurrency = int64(level.MaxShards)
			useShards = true
			timeDivider = 3600 // Ethereal monsters use per hour
		}
		reward := (mainCurrency * deltaSeconds) / int64(timeDivider)
		total := reward
		if total > maxCurrency {
			total = maxCurrency
		}
		if !useShards {
			p.AddProperties(total, 0, 0, 0, 0)
		} else {
			p.AddProperties(0, 0, 0, 0, total)
		}
		mon.CollectedCoins = 0
		mon.LastCollection = now

		collectResp := data.MakeGFSObject().PutLong("user_monster_id", umid)
		if total > 0 {
			collectResp.PutLong("success", 1)
			if !useShards {
				collectResp.PutInt("coins", int(total))
			} else {
				collectResp.PutInt("ethereal_currency", int(total))
			}
		} else {
			collectResp.PutLong("success", 0).PutUtfString("message", "nothing to collect")
		}
		ctx.Reply("gs_collect_monster", collectResp)

		ctx.Reply("gs_update_monster", data.MakeGFSObject().
			PutLong("user_monster_id", umid).
			PutGFSArray("properties", p.GetProperties()).
			PutLong("last_collection", now).
			PutInt("collected_coins", 0))

		ctx.Reply("gs_update_properties", data.MakeGFSObject().
			PutGFSArray("properties", p.GetProperties()))
	})

	m.HandlePlayer("gs_sell_monster", func(ctx *Context, p *Player) {
		island := ctx.Island()
		umid := ctx.Int64("user_monster_id")

		if island.IsGold() {
			for i, gm := range island.GoldMonsters {
				if gm.UserMonsterID == umid {
					island.GoldMonsters = append(island.GoldMonsters[:i], island.GoldMonsters[i+1:]...)
					ctx.Reply("gs_sell_monster", data.MakeGFSObject().
						PutBool("success", true).
						PutLong("user_monster_id", umid).
						PutGFSArray("properties", p.GetProperties()))
					return
				}
			}
			ctx.Fail("gs_sell_monster", "Invalid monster ID")
			return
		}

		mon := island.FindMonster(umid)
		if mon == nil {
			ctx.Fail("gs_sell_monster", "Invalid monster ID")
			return
		}
		if info, ok := static.MonsterBuy[int(mon.MonsterID)]; ok {
			p.AddProperties(int64(info.CostCoins*3/4), 0, 0, 0, 0)
		}
		island.RemoveMonster(umid)
		ctx.Reply("gs_sell_monster", data.MakeGFSObject().
			PutBool("success", true).
			PutLong("user_monster_id", umid).
			PutGFSArray("properties", p.GetProperties()))
	})

	// currently a weird behavior where only permanent mega monsters work
	// TODO
	m.HandlePlayer("gs_mega_monster_message", func(ctx *Context, p *Player) {
		island := ctx.Island()
		umid := ctx.Int64("user_monster_id")
		mon := island.FindMonster(umid)
		if mon == nil {
			ctx.Fail("gs_mega_monster_message", "Invalid monster ID")
			return
		}

		now := nowMS()

		permanent, ok := ctx.Params.GetBool("permanent")
		if !ok {
			permanent = false
		}

		enable, ok := ctx.Params.GetBool("mega_enable")
		if ok {
			mon.MegaCurrent = boolInt(enable)
		} else {
			cost := int64(2)
			if permanent {
				cost = 20
			}
			if !p.AddProperties(0, -cost, 0, 0, 0) {
				return
			}
			if permanent {
				mon.MegaPerma, mon.MegaCurrent, mon.MegaStart, mon.MegaFinish = true, 1, 0, 0
			} else {
				mon.MegaPerma, mon.MegaCurrent, mon.MegaStart, mon.MegaFinish = false, 1, now, now+int64(60*60*24)*1000
			}
		}

		// expire
		if mon.MegaPerma == false && mon.MegaFinish > 0 && mon.MegaFinish < now {
			mon.MegaFinish, mon.MegaStart, mon.MegaCurrent = 0, 0, 0
		}

		ctx.Reply("gs_mega_monster_message", data.MakeGFSObject().
			PutBool("success", true).
			PutLong("user_monster_id", umid))
		ctx.Reply("gs_update_monster", data.MakeGFSObject().
			PutLong("user_monster_id", umid).
			PutGFSArray("properties", p.GetProperties()).
			PutGFSObject("megamonster", mon.megaObject()))
	})

	m.HandlePlayer("gs_send_monster_home", func(ctx *Context, p *Player) {
		umid := ctx.Int64("user_monster_id")
		srcIsland := ctx.Island()
		mon := srcIsland.FindMonster(umid)
		if mon == nil {
			ctx.Reply("gs_send_monster_home", data.MakeGFSObject().
				PutLong("success", 0).
				PutLong("has_egg", 0).
				PutUtfString("message", "Monster not found"))
			return
		}
		dest, ok := static.TeleportInfo[[2]int{int(srcIsland.IslandID), int(mon.MonsterID)}]
		if !ok {
			ctx.Reply("gs_send_monster_home", data.MakeGFSObject().
				PutLong("success", 0).
				PutLong("has_egg", 0).
				PutUtfString("message", "No teleport destination for this monster"))
			return
		}
		var dstIsland *Island
		for _, isl := range p.Islands {
			if isl.IslandID == int64(dest.DestIsland) {
				dstIsland = isl
				break
			}
		}
		if dstIsland == nil {
			ctx.Reply("gs_send_monster_home", data.MakeGFSObject().
				PutLong("success", 0).
				PutLong("has_egg", 0).
				PutUtfString("message", "Player doesn't own destination island"))
			return
		}
		nursery := dstIsland.FindStructureByType(1)
		if nursery == nil {
			ctx.Reply("gs_send_monster_home", data.MakeGFSObject().
				PutLong("success", 0).
				PutLong("has_egg", 0).
				PutUtfString("message", "No nursery on destination island"))
			return
		}
		for _, e := range dstIsland.Eggs {
			if e.UserStructureID == nursery.UserStructureID {
				ctx.Reply("gs_send_monster_home", data.MakeGFSObject().
					PutLong("success", 0).
					PutLong("has_egg", 1).
					PutUtfString("message", "Nursery is occupied"))
				return
			}
		}
		info, ok := static.MonsterBuy[dest.DestMonster]
		if !ok {
			ctx.Reply("gs_send_monster_home", data.MakeGFSObject().
				PutLong("success", 0).
				PutLong("has_egg", 0).
				PutUtfString("message", "Monster buy info not found"))
			return
		}
		srcIsland.RemoveMonster(umid)
		now := nowMS()
		egg := &Egg{
			IslandID:        dstIsland.UserIslandID,
			LaidOn:          now,
			HatchesOn:       now + int64(info.BuildTime)*1000,
			MonsterID:       int64(dest.DestMonster),
			UserEggID:       p.NextEggID(),
			UserStructureID: nursery.UserStructureID,
		}
		dstIsland.Eggs = append(dstIsland.Eggs, egg)
		ctx.Reply("gs_send_monster_home", data.MakeGFSObject().PutLong("success", 1).PutInt("sent_to_island", int(dstIsland.IslandID)).PutLong("user_monster_id", umid))

		eggs := data.MakeGFSArray()
		eggs.AddSFSObject(egg.GetSFSObject())
		ctx.Reply("gs_update_eggs", data.MakeGFSObject().PutGFSArray("eggs", eggs))

		ctx.Reply("gs_buy_egg", data.MakeGFSObject().
			PutGFSObject("user_egg", egg.GetSFSObject()).
			PutBool("success", true).
			PutBool("remove_buyback", false).
			PutGFSArray("properties", p.GetProperties()))
	})
}
