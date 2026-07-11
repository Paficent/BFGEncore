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

	"paficent/bfg/utils"

	"github.com/Paficent/GoFox2X/data"
	"github.com/Paficent/GoFox2X/protocol"
	"github.com/Paficent/GoFox2X/server"
	"github.com/Paficent/GoFox2X/transport"
)

func (m *Manager) Run(addr string) error {
	srv := &server.Server{
		Handler:      server.HandlerFunc(m.handleMessage),
		OnConnect:    m.onConnect,
		OnDisconnect: m.onDisconnect,
	}
	log.Printf("SFS2X server listening on %s", addr)
	return srv.ListenAndServe(addr)
}

func (m *Manager) handleMessage(c *transport.Conn, msg *protocol.Message) {
	if m.debug {
		utils.Dump("recv", msg)
	}
	switch {
	case msg.Controller == protocol.System && msg.Action == protocol.ActionHandshake:
		log.Printf("recv sys  handshake")
		m.handleHandshake(c)
	case msg.Controller == protocol.System && msg.Action == protocol.ActionLogin:
		m.handleLogin(c, msg)
	case msg.Controller == protocol.System:
		log.Printf("recv sys  action=%d (unhandled)", msg.Action)
	default:
		if cmd, _, ok := protocol.ParseExtensionRequest(msg); ok {
			log.Printf("recv ext  %s", cmd)
		} else {
			log.Printf("recv ???  controller=%s action=%d", msg.Controller, msg.Action)
		}
		if !m.router.Dispatch(c, msg) {
			m.handleUnknownExtension(c, msg)
		}
	}
}

func (m *Manager) onConnect(c *transport.Conn) {
	log.Printf("connect    %s", c.RemoteAddr())
}

func (m *Manager) onDisconnect(c *transport.Conn, err error) {
	log.Printf("disconnect %s (%v)", c.RemoteAddr(), err)
	if p := playerFromConn(c); p != nil {
		m.untrackConn(p.BBBID)
		if serr := m.savePlayer(p); serr != nil {
			log.Printf("save on disconnect failed: %v", serr)
		}
	}
}

func (m *Manager) trackConn(bbbID int64, c *transport.Conn) {
	m.mu.Lock()
	m.conns[bbbID] = c
	m.mu.Unlock()
}

func (m *Manager) untrackConn(bbbID int64) {
	m.mu.Lock()
	delete(m.conns, bbbID)
	m.mu.Unlock()
}

func (m *Manager) connFor(bbbID int64) *transport.Conn {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.conns[bbbID]
}

func (m *Manager) Push(bbbID int64, command string, payload *data.GFSObject) {
	if c := m.connFor(bbbID); c != nil {
		m.reply(c, command, payload)
	}
}

func (m *Manager) PushProperties(p *Player) {
	if p == nil {
		return
	}
	m.Push(p.BBBID, "gs_update_properties", data.MakeGFSObject().PutGFSArray("properties", p.GetProperties()))
}

func (m *Manager) handleHandshake(c *transport.Conn) {
	sessionInfo := data.MakeGFSObject().
		PutInt("ct", 1_000_000).
		PutInt("ms", 8_000_000).
		PutUtfString("tk", "0123456789abcdef0123456789abcdef")
	m.sendSystem(c, protocol.ActionHandshake, "handshake", sessionInfo)
}

func (m *Manager) handleUnknownExtension(c *transport.Conn, msg *protocol.Message) {
	cmd, _, ok := protocol.ParseExtensionRequest(msg)
	if !ok {
		return
	}
	log.Printf("  NOT IMPLEMENTED: %s", cmd)
	notice := data.MakeGFSObject().
		PutBool("force_logout", false).
		PutUtfString("msg", cmd+" is not implemented yet")
	m.reply(c, "gs_display_generic_message", notice)
	m.reply(c, cmd, data.MakeGFSObject())
}
