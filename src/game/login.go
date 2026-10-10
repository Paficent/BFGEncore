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
	"strconv"

	"github.com/Paficent/GoFox2X/data"
	"github.com/Paficent/GoFox2X/protocol"
	"github.com/Paficent/GoFox2X/transport"
)

func contentResponse(key string, value *data.GFSArray) *data.GFSObject {
	return stamp(data.MakeGFSObject().PutGFSArray(key, value))
}

func limboRoom() protocol.Room {
	return protocol.Room{ID: 0, Name: "Limbo", Type: "default", UserCount: 1, MaxPlayers: 200}
}

func (m *Manager) handleLogin(c *transport.Conn, msg *protocol.Message) {
	un, _ := msg.Payload.GetUtfString("un")
	bbbID, _ := strconv.ParseInt(un, 10, 64)
	log.Printf("recv sys  login as %q", un)

	login := data.MakeGFSObject().
		PutShort("rs", 0).
		PutUtfString("zn", "MySingingMonsters").
		PutUtfString("un", un).
		PutShort("pi", 1).
		PutInt("id", 1).
		PutGFSObject("p", data.MakeGFSObject()).
		PutGFSArray("rl", protocol.RoomList(limboRoom()))

	m.sendSystem(c, protocol.ActionLogin, "login", login)

	c.Session = m.GetOrCreatePlayer(bbbID)
	m.trackConn(bbbID, c)
	m.sendLoginSequence(c, bbbID)
}

func (m *Manager) sendLoginSequence(c *transport.Conn, bbbID int64) {
	settings := data.MakeGFSObject().PutGFSArray("user_game_settings", m.Static.GameSettings)
	constants := m.Static.Constants
	
	m.reply(c, "game_settings", settings)

	m.reply(c, "gs_initialized", data.MakeGFSObject().PutLong("bbb_id", bbbID))
	if constants.CustomLogon == "" {
		return
	}
	welcome := data.MakeGFSObject().
		PutBool("force_logout", false).
		PutUtfString("msg", constants.CustomLogon)
	m.reply(c, "gs_display_generic_message", welcome)
}

func registerLoginHandlers(m *Manager) {
	static := m.Static

	content := func(cmd, key string, value *data.GFSArray) {
		m.Handle(cmd, func(ctx *Context) {
			ctx.Reply(cmd, contentResponse(key, value))
		})
	}

	content("db_gene", "genes_data", static.Genes)
	content("db_island", "islands_data", static.Islands)
	content("db_island_torches", "island_torch_data", static.Torches)
	content("db_monster", "monsters_data", static.Monsters)
	content("db_structure", "structures_data", static.Structures)
	content("db_level", "level_data", static.Levels)
	content("db_scratch_offs", "scratch_offs", static.ScratchOffs)

	content("gs_timed_events", "timed_event_list", static.TimedEvents)
	m.Handle("gs_quest", func(ctx *Context) {
		p := ctx.Player()
		if p == nil {
			p = &Player{}
		}
		ctx.Reply("gs_quest", contentResponse("result", m.questListFor(p)))
	})

	m.Handle("db_store", func(ctx *Context) {
		ctx.Reply("db_store", stamp(data.MakeGFSObject().
			PutGFSArray("store_item_data", static.StoreItems).
			PutGFSArray("store_group_data", static.StoreGroups).
			PutGFSArray("store_currency_data", static.StoreCurrencies)))
	})

	m.Handle("gs_promos", func(ctx *Context) {
		ctx.Reply("gs_promos", data.MakeGFSObject())
	})

	m.Handle("gs_player", func(ctx *Context) {
		player := data.MakeGFSObject()
		if p := ctx.Player(); p != nil {
			player = p.GetSFSObject()
		}
		ctx.Reply("gs_player", data.MakeGFSObject().PutGFSObject("player_object", player))
	})

	m.HandleAck(
		"keep_alive",
		"gs_multi_neighbors",
		"gs_get_messages",
		"gs_handle_facebook_help_instances",
		"gs_process_unclaimed_purchases",
	)
}
