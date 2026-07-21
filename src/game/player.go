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
	"paficent/bfg/db"

	"github.com/Paficent/GoFox2X/data"
)

const maxDisplayResource = 999_999_999

func clampDisplay(v int64) int {
	if v > maxDisplayResource {
		return maxDisplayResource
	}
	return int(v)
}

type Player struct {
	BBBID        int64
	UserID       int64
	DisplayName  string
	Coins        int64
	Diamonds     int64
	Food         int64
	XP           int64
	Shards       int64
	Level        int
	ActiveIsland int64
	Islands      []*Island

	MonsterScratchTime int64
	MemoryHighScore    int

	QuestStatus   map[int]string // TODO: NEED TO CONFIRM THIS WITH THE ACTUAL CLIENT
	QuestCounters map[string]int //

	levelXP      map[int]int
	eggSeq       int64
	breedingSeq  int64
	bakingSeq    int64
	monsterSeq   int64
	structureSeq int64
	islandSeq    int64
}

func newPlayer(bbbID, userID int64, static *db.StaticData) *Player {
	c := static.Constants
	p := &Player{
		BBBID:        bbbID,
		UserID:       userID,
		DisplayName:  "New Player",
		Coins:        c.StartingCoins,
		Diamonds:     c.StartingDiamonds,
		Food:         c.StartingFood,
		XP:           c.StartingXP,
		Shards:       c.StartingShards,
		Level:        c.StartingLevel,
		ActiveIsland: 1,
		levelXP:      static.LevelXP,
	}
	castle := int64(7)
	if info, ok := static.Island(1); ok && info.Castle != 0 {
		castle = int64(info.Castle)
	}
	p.structureSeq = 0
	p.islandSeq = 1
	p.Islands = append(p.Islands, p.buildIsland(1, 1, castle))
	return p
}

func (p *Player) handleLevelUp() {
	for p.Level < 30 {
		nextXP, ok := p.levelXP[p.Level+1]
		if !ok {
			break
		}
		if p.XP >= int64(nextXP) {
			p.Level++
			p.XP -= int64(nextXP)
		} else {
			break
		}
	}
}

func (p *Player) GetActiveIsland() *Island {
	for _, island := range p.Islands {
		if island.UserIslandID == p.ActiveIsland {
			return island
		}
	}
	return nil
}

func (p *Player) findStructure(userStructureID int64) *Structure {
	for _, island := range p.Islands {
		if s := island.FindStructure(userStructureID); s != nil {
			return s
		}
	}
	return nil
}

func (p *Player) AddProperties(coins, diamonds, food, xp, shards int64) bool {
	if p.Coins+coins < 0 || p.Diamonds+diamonds < 0 || p.Food+food < 0 || p.XP+xp < 0 || p.Shards+shards < 0 {
		return false
	}
	p.Coins += coins
	p.Diamonds += diamonds
	p.Food += food
	p.XP += xp
	p.Shards += shards

	const maxResource = 10_000_000_000
	clamp := func(v int64) int64 {
		if v > maxResource {
			return maxResource
		}
		return v
	}
	p.Coins = clamp(p.Coins)
	p.Diamonds = clamp(p.Diamonds)
	p.Food = clamp(p.Food)
	p.XP = clamp(p.XP)
	p.Shards = clamp(p.Shards)

	if xp > 0 {
		p.handleLevelUp()
	}
	return true
}

func (p *Player) Buy(coins, diamonds, eth int64) bool {
	return p.AddProperties(-coins, -diamonds, 0, 0, -eth)
}

func (p *Player) BuyInIsland(coins, diamonds, eth int64, island *Island) bool {
	if diamonds > 0 {
		if !p.Buy(0, int64(diamonds), 0) {
			return false
		}
	} else {
		if island.IsEthereal() {
			if !p.Buy(0, 0, int64(eth)) {
				return false
			}
		} else {
			if !p.Buy(int64(coins), 0, 0) {
				return false
			}
		}
	}
	return true
}

func (p *Player) NextEggID() int64 {
	p.eggSeq++
	return p.eggSeq
}

func (p *Player) NextBreedingID() int64 {
	p.breedingSeq++
	return p.breedingSeq
}

func (p *Player) NextBakingID() int64 {
	p.bakingSeq++
	return p.bakingSeq
}

func (p *Player) NextMonsterID() int64 {
	p.monsterSeq++
	return p.monsterSeq
}

func (p *Player) NextStructureID() int64 {
	p.structureSeq++
	return p.structureSeq
}

func (p *Player) GetProperties() *data.GFSArray {
	props := data.MakeGFSArray()
	props.AddSFSObject(data.MakeGFSObject().PutInt("coins", clampDisplay(p.Coins)))
	props.AddSFSObject(data.MakeGFSObject().PutInt("diamonds", clampDisplay(p.Diamonds)))
	props.AddSFSObject(data.MakeGFSObject().PutInt("food", clampDisplay(p.Food)))
	props.AddSFSObject(data.MakeGFSObject().PutInt("xp", clampDisplay(p.XP)))
	props.AddSFSObject(data.MakeGFSObject().PutInt("ethereal_currency", clampDisplay(p.Shards)))
	props.AddSFSObject(data.MakeGFSObject().PutInt("level", p.Level))
	props.AddSFSObject(data.MakeGFSObject().PutLong("monsterScratchTime", p.MonsterScratchTime))
	props.AddSFSObject(data.MakeGFSObject().PutBool("has_scratch_off_m", p.freeMonsterScratch(nowMS())))
	return props
}

func (p *Player) freeMonsterScratch(now int64) bool {
	return now-p.MonsterScratchTime >= monsterScratchCooldownMS
}

func (p *Player) GetSFSObject() *data.GFSObject {
	now := nowMS()
	obj := data.MakeGFSObject().
		PutInt("coins", clampDisplay(p.Coins)).
		PutInt("diamonds", clampDisplay(p.Diamonds)).
		PutInt("food", clampDisplay(p.Food)).
		PutInt("ethereal_currency", clampDisplay(p.Shards)).
		PutInt("premium", 1).
		PutLong("last_login", now).
		PutInt("xp", clampDisplay(p.XP)).
		PutInt("level", p.Level).
		PutInt("max_level", 30).
		PutInt("bbb_id", int(p.BBBID)).
		PutInt("user_id", int(p.UserID)).
		PutInt("referral", 0).
		PutLong("active_island", p.ActiveIsland).
		PutInt("fb_invite_reward", 1).
		PutInt("twitter_invite_reward", 1).
		PutInt("email_invite_reward", 1).
		PutLong("last_fb_post_reward", now).
		PutGFSArray("achievements", data.MakeGFSArray()).
		PutGFSArray("viewable_ads", data.MakeGFSArray()).
		PutUtfString("extra_ad_params", "").
		PutBool("third_party_ads", false).
		PutBool("third_party_video_ads", false).
		PutUtfString("display_name", p.DisplayName)

	islands := data.MakeGFSArray()
	for _, island := range p.Islands {
		islands.AddSFSObject(island.GetSFSObject())
	}
	obj.PutGFSArray("islands", islands)
	return obj
}
