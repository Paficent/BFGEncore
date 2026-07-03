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

func getQuests(db *DB) *data.GFSArray {
	return buildArray(db.Quests, func(r Quest) *data.GFSObject {
		log := questLogObject(r.ID, "false", 0, r.Initial)
		return questEntryWrap(r.ID, log, questStaticObject(r))
	})
}

func loadQuestStatics(db *DB) (map[int]*data.GFSObject, []int) {
	statics := map[int]*data.GFSObject{}
	var order []int
	for _, r := range db.Quests {
		statics[r.ID] = questStaticObject(r)
		order = append(order, r.ID)
	}
	return statics, order
}

func questLogObject(id int, status string, collected, isNew int) *data.GFSObject {
	return data.MakeGFSObject().
		PutInt("id", id).
		PutInt("quest_id", id).
		PutInt("user", 0).
		PutUtfString("status", status).
		PutInt("collected", collected).
		PutInt("new", isNew)
}

func questEntryWrap(id int, log, static *data.GFSObject) *data.GFSObject {
	entry := data.MakeGFSArray()
	entry.AddSFSObject(log)
	entry.AddSFSObject(static)

	return data.MakeGFSObject().
		PutGFSArray("new", entry).
		PutLong("id", int64(id))
}

func questStaticObject(r Quest) *data.GFSObject {
	static := data.MakeGFSObject().
		PutInt("id", r.ID).
		PutUtfString("name", r.Name).
		PutUtfString("description", r.Description).
		PutUtfString("type", r.Type)

	goals := data.MakeGFSArray()
	for _, g := range r.Goals.V {
		goal := data.MakeGFSObject()
		for k, v := range g {
			putDynamic(goal, k, v)
		}
		goals.AddSFSObject(goal)
	}
	static.PutGFSArray("goals", goals)

	next := data.MakeGFSArray()
	for _, name := range r.Next.V {
		next.AddSFSObject(data.MakeGFSObject().PutUtfString("quest", name))
	}
	static.PutGFSArray("next", next)

	rewards := data.MakeGFSObject()
	if len(r.Rewards.V) > 0 {
		for k, v := range r.Rewards.V[0] {
			putDynamic(rewards, k, v)
		}
	}
	static.PutGFSObject("rewards", rewards)

	visible := 0
	if r.Visible != nil {
		visible = *r.Visible
	}
	static.PutUtfString("sheet", r.Sheet).
		PutUtfString("image", r.Image).
		PutInt("visible", visible).
		PutUtfString("min_server_version", r.MinServer)
	if r.Comment != "" {
		static.PutUtfString("comment", r.Comment)
	}
	return static
}

func getTimedEvents(db *DB) *data.GFSArray {
	now := nowMS()
	const oneYearMS = int64(60*60*24*365) * 1000
	endDate := now + oneYearMS

	return buildArray(db.Entities, func(e Entity) *data.GFSObject {
		if bool(e.ViewInMarket) {
			return nil
		}
		eventData := data.MakeGFSArray()
		eventData.AddSFSObject(data.MakeGFSObject().PutInt("entity", e.ID))

		return data.MakeGFSObject().
			PutLong("end_date", endDate).
			PutLong("last_updated", now).
			PutUtfString("event_type", "EntityStoreAvailability").
			PutInt("event_id", 3).
			PutGFSArray("data", eventData).
			PutLong("id", int64(200000+e.ID)).
			PutLong("start_date", now)
	})
}
