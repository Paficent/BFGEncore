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
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Paficent/GoFox2X/data"
)

func getStoreGroups(db *DB) *data.GFSArray {
	now := nowMS()
	return buildArray(db.StoreGroups, func(g StoreGroup) *data.GFSObject {
		return data.MakeGFSObject().
			PutInt("id", g.ID).
			PutInt("storegroup_id", g.ID).
			PutUtfString("name", g.Name).
			PutInt("currency", g.Currency).
			PutUtfString("group_title", g.Title).
			PutLong("last_changed", now).
			PutUtfString("min_server_version", g.MinServer)
	})
}

func getStoreCurrencies(db *DB) *data.GFSArray {
	now := nowMS()
	return buildArray(db.StoreCurrencies, func(c StoreCurrency) *data.GFSObject {
		return data.MakeGFSObject().
			PutInt("storecur_id", c.ID).
			PutInt("id", c.ID).
			PutUtfString("currency_name", c.Name).
			PutInt("starting_amount", c.StartingAmount).
			PutLong("last_changed", now).
			PutUtfString("min_server_version", c.MinServer)
	})
}

func getStoreItems(db *DB) *data.GFSArray {
	items := buildArray(db.Store, func(s StoreItem) *data.GFSObject {
		switch s.Currency {
		case "coins", "diamonds", "food", "ethereal":
		default:
			return nil
		}
		if strings.Contains(strings.ToLower(s.Title), "warm") {
			return nil
		}
		if s.MinServer != "0.0" {
			return nil
		}
		maxVal := -1
		if m := strings.TrimSpace(s.Max); m != "" {
			maxVal, _ = strconv.Atoi(m)
		}
		return data.MakeGFSObject().
			PutInt("id", int(s.ID)).
			PutInt("item_id", int(s.ID)).
			PutUtfString("item_name", s.Name).
			PutUtfString("item_title", s.Title).
			PutUtfString("item_desc", s.Desc).
			PutInt("price", int(s.Price)).
			PutInt("consumable", int(s.Consumable)).
			PutInt("amount", int(s.Amount)).
			PutInt("max", maxVal).
			PutInt("group_id", int(s.GroupID)).
			PutInt("sale_amount", 0).
			PutInt("currency_id", int(s.CurrencyID)).
			PutUtfString("sheet_id", s.SheetID)
	})

	now := nowMS()

	var castles []Structure
	for _, s := range db.Structures {
		if s.Type == "castle" {
			castles = append(castles, s)
		}
	}
	sort.Slice(castles, func(i, j int) bool { return castles[i].ID < castles[j].ID })
	for _, s := range castles {
		addPermanentStructureItem(db, items, s, now)
	}

	for _, s := range db.Structures {
		if s.ID == 2 {
			addPermanentStructureItem(db, items, s, now)
			break
		}
	}

	addPermanentMonsterItem(db, items, 82, "001_E_rare.bin", now, 1)
	return items
}

func addPermanentStructureItem(db *DB, items *data.GFSArray, s Structure, now int64) {
	const offset = 300000
	e, ok := db.entity(s.Entity)
	if !ok {
		return
	}
	storeID := offset + s.ID
	itemName := orDefault(e.Name, fmt.Sprintf("STRUCTURE_%d", s.ID))

	item := data.MakeGFSObject().
		PutInt("id", storeID).
		PutInt("item_id", storeID).
		PutUtfString("item_name", itemName).
		PutUtfString("item_title", e.Name).
		PutUtfString("item_desc", e.Description)

	switch {
	case e.CostCoins > 0:
		item.PutInt("price", e.CostCoins).PutUtfString("currency", "coins")
	case e.CostDiamonds > 0:
		item.PutInt("price", e.CostDiamonds).PutUtfString("currency", "diamonds")
	case e.CostEth > 0:
		item.PutInt("price", e.CostEth).PutUtfString("currency", "ethereal")
	default:
		item.PutInt("price", 0).PutUtfString("currency", "coins")
	}

	item.PutInt("consumable", 0).
		PutInt("amount", 1).
		PutInt("max", 1).
		PutInt("group_id", 1).
		PutInt("sale_amount", 0).
		PutInt("currency_id", 1).
		PutUtfString("sheet_id", "").
		PutUtfString("image_id", "").
		PutUtfString("ios_platform_id", "").
		PutUtfString("android_platform_id", "").
		PutUtfString("amazon_platform_id", "").
		PutLong("last_changed", now).
		PutInt("enabled", 1).
		PutUtfString("min_server_version", orDefault(e.MinServer, "0.0"))

	addEntityData(db, item, s.Entity)
	item.PutUtfString("structure_type", s.Type).PutInt("upgrades_to", s.UpgradesTo)
	items.AddSFSObject(item)
}

func addPermanentMonsterItem(db *DB, items *data.GFSArray, monsterID int, binsID string, now int64, maxLimit int) {
	const offset = 100000

	var monster *Monster
	for i := range db.Monsters {
		if db.Monsters[i].ID == monsterID {
			monster = &db.Monsters[i]
			break
		}
	}
	if monster == nil {
		return
	}
	e, ok := db.entity(monster.Entity)
	if !ok {
		return
	}
	storeID := offset + monsterID
	itemName := orDefault(e.Name, fmt.Sprintf("Monster_%d", monsterID))

	item := data.MakeGFSObject().
		PutInt("id", storeID).
		PutInt("item_id", storeID).
		PutUtfString("item_name", itemName).
		PutUtfString("item_title", e.Name).
		PutUtfString("item_desc", e.Description)

	switch {
	case e.CostCoins > 0:
		item.PutInt("price", e.CostCoins).PutInt("currency_id", 1)
	case e.CostDiamonds > 0:
		item.PutInt("price", e.CostDiamonds).PutInt("currency_id", 2)
	default:
		item.PutInt("price", 0).PutInt("currency_id", 1)
	}

	item.PutInt("consumable", 0).
		PutInt("amount", 1).
		PutInt("max", maxLimit).
		PutInt("group_id", 1).
		PutInt("sale_amount", 0).
		PutUtfString("sheet_id", "").
		PutUtfString("image_id", "").
		PutUtfString("ios_platform_id", "").
		PutUtfString("android_platform_id", "").
		PutUtfString("amazon_platform_id", "").
		PutLong("last_changed", now).
		PutInt("enabled", 1).
		PutUtfString("min_server_version", orDefault(e.MinServer, "0.0"))

	if binsID != "" {
		item.PutUtfString("bins_id", binsID).PutUtfString("bin_id", binsID)
	}

	addEntityData(db, item, monster.Entity)
	items.AddSFSObject(item)
}
