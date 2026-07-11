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

// abstraction layer for handling game logic.
package game

import (
	"log"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	"paficent/bfg/db"
	"paficent/bfg/save"
	"paficent/bfg/utils"

	"github.com/Paficent/GoFox2X/data"
	"github.com/Paficent/GoFox2X/protocol"
	"github.com/Paficent/GoFox2X/server"
	"github.com/Paficent/GoFox2X/transport"
)

type HandlerFunc func(*Context)

type Context struct {
	Conn   *transport.Conn
	Params *data.GFSObject
	mgr    *Manager
}

func (c *Context) Player() *Player { return playerFromConn(c.Conn) }
func (c *Context) Island() *Island {
	p := c.Player()
	if p == nil {
		return nil
	}
	return p.GetActiveIsland()
}

func (c *Context) Int(key string) int       { return paramInt(c.Params, key) }
func (c *Context) Int64(key string) int64   { return int64(paramInt(c.Params, key)) }
func (c *Context) Float(key string) float64 { return paramFloat(c.Params, key) }
func (c *Context) Str(key string) string    { s, _ := c.Params.GetUtfString(key); return s }

func (c *Context) Reply(command string, payload *data.GFSObject) {
	c.mgr.reply(c.Conn, command, payload)
}
func (c *Context) Fail(command, message string) {
	c.Reply(command, data.MakeGFSObject().PutBool("success", false).
		PutUtfString("error", message).PutUtfString("message", message))
}

func (c *Context) Manager() *Manager      { return c.mgr }
func (c *Context) Static() *db.StaticData { return c.mgr.Static }

type Manager struct {
	Static *db.StaticData

	store  *save.Store
	router *server.ExtensionRouter
	debug  bool

	mu      sync.Mutex
	players map[int64]*Player
	conns   map[int64]*transport.Conn

	timers *timerService
	sendMu sync.Mutex // serialises writes: handler + timer goroutines can both send

	pendingScratch  map[scratchKey]db.ScratchOff
	memoryHighScore int
}

type scratchKey struct {
	id  int64
	typ string
}

func New(static *db.StaticData, store *save.Store, debug bool) *Manager {
	m := &Manager{
		Static:  static,
		store:   store,
		router:  server.NewExtensionRouter(),
		debug:   debug,
		players: map[int64]*Player{},
		conns:   map[int64]*transport.Conn{},
		timers:  newTimerService(),

		pendingScratch: map[scratchKey]db.ScratchOff{},
	}
	m.loadPlayers()

	registerLoginHandlers(m)
	registerMonsterHandlers(m)
	registerStructureHandlers(m)
	registerIslandHandlers(m)
	registerEconomyHandlers(m)
	registerQuestHandlers(m)
	registerBakingHandlers(m)

	m.rearmUpgradeTimers()
	return m
}

func (m *Manager) setPendingScratch(id int64, typ string, s db.ScratchOff) {
	m.mu.Lock()
	m.pendingScratch[scratchKey{id, typ}] = s
	m.mu.Unlock()
}

func (m *Manager) takePendingScratch(id int64, typ string) (db.ScratchOff, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := scratchKey{id, typ}
	s, ok := m.pendingScratch[k]
	if ok {
		delete(m.pendingScratch, k)
	}
	return s, ok
}

func (m *Manager) peekPendingScratch(id int64, typ string) (db.ScratchOff, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.pendingScratch[scratchKey{id, typ}]
	return s, ok
}

func (m *Manager) recordMemoryScore(score int) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	if score > m.memoryHighScore {
		m.memoryHighScore = score
	}
	return m.memoryHighScore
}

func (m *Manager) scratchReveal(prize db.ScratchOff, props *data.GFSArray) *data.GFSObject {
	if prize.Prize == "monster" {
		return data.MakeGFSObject().
			PutUtfString("type", "M").
			PutLong("success", 0).
			PutLong("has_egg", 1).
			PutGFSArray("properties", props)
	}
	matches := 3 + rand.Intn(4)
	scaled := data.MakeGFSObject().
		PutInt("amount", prize.Amount)
	ticket := data.MakeGFSObject().
		PutInt("amount", prize.Amount).
		PutInt("matches", matches).
		PutGFSObject("scaled_prizes", scaled).
		PutUtfString("type", prize.Type).
		PutUtfString("prize", prize.Prize)
	return data.MakeGFSObject().
		PutLong("success", 1).
		PutUtfString("type", prize.Type).
		PutUtfString("prize", prize.Prize).
		PutInt("amount", prize.Amount).
		PutInt("matches", matches).
		PutGFSObject("ticket", ticket).
		PutGFSObject("scaled_prizes", scaled).
		PutGFSArray("properties", props)
}

func (m *Manager) Handle(command string, fn HandlerFunc) {
	m.router.On(command, func(conn *transport.Conn, params *data.GFSObject) {
		fn(&Context{Conn: conn, Params: params, mgr: m})
	})
}

// for commands that change a players state (autosaves immediately)
func (m *Manager) HandleWrite(command string, fn HandlerFunc) {
	m.router.On(command, func(conn *transport.Conn, params *data.GFSObject) {
		ctx := &Context{Conn: conn, Params: params, mgr: m}
		fn(ctx)
		if p := ctx.Player(); p != nil {
			m.runQuestEval(ctx, p)
			if err := m.savePlayer(p); err != nil {
				log.Printf("save after %s failed: %v", command, err)
			}
		}
	})
}

// returned object sent back immediately under the same name
func (m *Manager) HandleReply(command string, build func(*Context) *data.GFSObject) {
	m.Handle(command, func(ctx *Context) {
		ctx.Reply(command, build(ctx))
	})
}

// commands that need an empty reply (i'm looking at you keep-alive)
func (m *Manager) HandleAck(commands ...string) {
	for _, command := range commands {
		command := command
		m.Handle(command, func(ctx *Context) {
			ctx.Reply(command, data.MakeGFSObject())
		})
	}
}

type PlayerHandlerFunc func(*Context, *Player)

// checks for player, writes save
func (m *Manager) HandlePlayer(command string, fn PlayerHandlerFunc) {
	m.HandleWrite(command, func(ctx *Context) {
		if p := ctx.Player(); p != nil {
			fn(ctx, p)
		}
	})
}

// checks for player, does not write
func (m *Manager) HandlePlayerRead(command string, fn PlayerHandlerFunc) {
	m.Handle(command, func(ctx *Context) {
		if p := ctx.Player(); p != nil {
			fn(ctx, p)
		}
	})
}

func stamp(o *data.GFSObject) *data.GFSObject {
	now := nowMS()
	return o.PutLong("server_time", now).PutLong("last_updated", now)
}

func (m *Manager) GetOrCreatePlayer(bbbID int64) *Player {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.players[bbbID]; ok {
		return p
	}
	p := newPlayer(bbbID, bbbID, m.Static)
	m.players[bbbID] = p
	return p
}

func (m *Manager) Player(bbbID int64) *Player {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.players[bbbID]
}

func (m *Manager) reply(conn *transport.Conn, command string, payload *data.GFSObject) {
	m.send(conn, protocol.ExtensionResponse(command, payload), "ext  "+command)
}

func (m *Manager) sendSystem(conn *transport.Conn, action protocol.Action, label string, payload *data.GFSObject) {
	m.send(conn, protocol.NewMessage(protocol.System, action, payload), "sys  "+label)
}

func (m *Manager) send(conn *transport.Conn, msg *protocol.Message, label string) {
	m.sendMu.Lock()
	defer m.sendMu.Unlock()
	body, err := msg.MarshalBinary()
	if err != nil {
		log.Printf("  send %-30s MARSHAL ERROR: %v", label, err)
		return
	}
	if m.debug {
		utils.Dump("send", msg)
	}
	if err := conn.Send(msg); err != nil {
		log.Printf("  send %-30s SEND ERROR (%d bytes): %v", label, len(body), err)
		return
	}
	flag := ""
	if len(body) > 0xFFFF {
		flag = "  [BIG FRAME >64KB]"
	}
	log.Printf("  send %-30s %6d bytes%s", label, len(body), flag)
}

// utils:
// TODO: potentially abstract this to the utils/ folder

func nowMS() int64 {
	return time.Now().Unix() * 1000
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func playerFromConn(c *transport.Conn) *Player {
	if p, ok := c.Session.(*Player); ok {
		return p
	}
	return nil
}

func paramInt(o *data.GFSObject, key string) int {
	if v, ok := o.GetInt(key); ok {
		return v
	}
	if v, ok := o.GetShort(key); ok {
		return int(v)
	}
	if v, ok := o.GetLong(key); ok {
		return int(v)
	}
	if v, ok := o.GetByte(key); ok {
		return int(v)
	}
	return 0
}

func paramFloat(o *data.GFSObject, key string) float64 {
	if v, ok := o.GetDouble(key); ok {
		return v
	}
	if v, ok := o.GetFloat(key); ok {
		return float64(v)
	}
	if v, ok := o.GetInt(key); ok {
		return float64(v)
	}
	return 1.0
}

type Stats struct {
	Online        int
	LoadedPlayers int
}

func (m *Manager) Stats() Stats {
	m.mu.Lock()
	defer m.mu.Unlock()
	return Stats{Online: len(m.conns), LoadedPlayers: len(m.players)}
}

func (m *Manager) OnlinePlayers() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	names := make([]string, 0, len(m.conns))
	for id := range m.conns {
		label := strconv.FormatInt(id, 10)
		if p := m.players[id]; p != nil && p.DisplayName != "" {
			label = p.DisplayName + "  " + label
		}
		names = append(names, label)
	}
	sort.Strings(names)
	return names
}
