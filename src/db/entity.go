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

func addEntityData(db *DB, obj *data.GFSObject, entityID int) {
	e, ok := db.entity(entityID)
	if !ok {
		return
	}

	obj.PutInt("entity_id", entityID)
	obj.PutUtfString("name", e.Name)
	obj.PutUtfString("description", e.Description)
	obj.PutUtfString("entity_type", e.Type)

	if g := e.Graphic.V; g != nil {
		obj.PutGFSObject("graphic", dynamicObject(g))
	}

	obj.PutInt("size_x", e.SizeX)
	obj.PutInt("size_y", e.SizeY)
	obj.PutInt("level", e.Level)
	obj.PutInt("buildTime", e.BuildTime*1000)
	obj.PutInt("build_time", e.BuildTime*1000)
	obj.PutInt("cost_coins", e.CostCoins)
	obj.PutInt("cost_eth_currency", e.CostEth)
	obj.PutInt("cost_diamonds", e.CostDiamonds)
	obj.PutInt("cost_sale", e.CostSale)

	obj.PutUtfString("keywords", e.Keywords)
	obj.PutUtfString("min_server_version", orDefault(e.MinServer, "0.0"))

	obj.PutBool("movable", bool(e.Movable))
	obj.PutBool("view_in_market", bool(e.ViewInMarket))
	obj.PutBool("premium", bool(e.Premium))

	reqs := data.MakeGFSArray()
	for _, req := range e.Requirements.V {
		reqs.AddSFSObject(data.MakeGFSObject().PutInt("entity", req.Entity))
	}
	obj.PutGFSArray("requirements", reqs)

	obj.PutLong("last_changed", nowMS())
	obj.PutInt("xp", e.XP)
	obj.PutInt("y_offset", e.YOffset)
	obj.PutInt("sticker_offset", e.YOffset)
	obj.PutUtfString("fb_object_id", "")
	obj.PutInt("tier", 1)
}
