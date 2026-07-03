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
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

type Gene struct {
	ID        int    `json:"gene_id"`
	Letter    string `json:"gene_letter"`
	Graphic   string `json:"gene_graphic"`
	MinServer string `json:"min_server_version"`
}

type Level struct {
	Level       int `json:"level"`
	XP          int `json:"xp"`
	MaxBakeries int `json:"max_bakeries"`
}

type ScratchOff struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	Prize       string `json:"prize"`
	Amount      int    `json:"amount"`
	Probability int    `json:"probability"`
	IsTopPrize  int    `json:"is_top_prize"`
	MinServer   string `json:"min_server_version"`
}

type Torch struct {
	IslandID    int    `json:"island_id"`
	Graphic     string `json:"torch_graphic"`
	LastChanged string `json:"last_changed"`
}

type GameSetting struct {
	Setting string `json:"setting"`
	Value   string `json:"value"`
}

type Island struct {
	ID           int             `json:"island_id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Graphic      quoted[blobMap] `json:"graphic"`
	Genes        string          `json:"genes"`
	Midi         string          `json:"midi"`
	Level        int             `json:"level"`
	CostCoins    int             `json:"cost_coins"`
	CostDiamonds int             `json:"cost_diamonds"`
	Castle       int             `json:"castle_structure_id"`
	MinServer    string          `json:"min_server_version"`
}

type IslandMonster struct {
	Island     int    `json:"island"`
	Monster    int    `json:"monster"`
	Instrument string `json:"instrument"`
}

type IslandStructure struct {
	Island     int    `json:"island"`
	Structure  int    `json:"structure"`
	Instrument string `json:"instrument"`
}

type Entity struct {
	ID           int                   `json:"entity_id"`
	Type         string                `json:"entity_type"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	Keywords     string                `json:"keywords"`
	Graphic      quoted[blobMap]       `json:"graphic"`
	SizeX        int                   `json:"size_x"`
	SizeY        int                   `json:"size_y"`
	CostCoins    int                   `json:"cost_coins"`
	CostEth      int                   `json:"cost_eth_currency"`
	CostDiamonds int                   `json:"cost_diamonds"`
	CostSale     int                   `json:"cost_sale"`
	BuildTime    int                   `json:"build_time"`
	Level        int                   `json:"level"`
	Requirements quoted[[]Requirement] `json:"requirements"`
	Movable      intbool               `json:"movable"`
	XP           int                   `json:"xp"`
	YOffset      int                   `json:"y_offset"`
	ViewInMarket intbool               `json:"view_in_market"`
	Premium      intbool               `json:"premium"`
	MinServer    string                `json:"min_server_version"`
}

type Structure struct {
	ID            int             `json:"structure_id"`
	Entity        int             `json:"entity"`
	Type          string          `json:"structure_type"`
	UpgradesTo    int             `json:"upgrades_to"`
	Sound         string          `json:"sound"`
	LimitToIsland int             `json:"limit_to_island"`
	Extra         quoted[blobMap] `json:"extra"`
}

type Monster struct {
	ID            int                 `json:"monster_id"`
	Entity        int                 `json:"entity"`
	Genes         string              `json:"genes"`
	Beds          int                 `json:"beds"`
	Happiness     quoted[[]Happiness] `json:"happiness"`
	Names         quoted[[]string]    `json:"names"`
	LevelUpXP     int                 `json:"level_up_xp"`
	LevelupIsland string              `json:"levelup_island"`
}

type MonsterLevel struct {
	Monster int               `json:"monster"`
	Levels  []MonsterLevelRow `json:"levels"`
}

type MonsterLevelRow struct {
	Level    int `json:"level"`
	Food     int `json:"food"`
	Coins    int `json:"coins"`
	MaxCoins int `json:"max_coins"`
	Eth      int `json:"ethereal_currency"`
	MaxEth   int `json:"max_ethereal"`
}

type BreedingCombo struct {
	Monster1    int `json:"monster_1"`
	Monster2    int `json:"monster_2"`
	Result      int `json:"result"`
	Probability int `json:"probability"`
}

type Quest struct {
	ID          int                      `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Type        string                   `json:"type"`
	Goals       quoted[[]map[string]any] `json:"goals"`
	Next        quoted[[]string]         `json:"next"`
	Rewards     quoted[[]map[string]any] `json:"rewards"`
	Sheet       string                   `json:"sheet"`
	Image       string                   `json:"image"`
	Initial     int                      `json:"initial"`
	Visible     *int                     `json:"visible"`
	Comment     string                   `json:"comment"`
	MinServer   string                   `json:"min_server_version"`
}

type StoreGroup struct {
	ID        int    `json:"storegroup_id"`
	Name      string `json:"group_name"`
	Currency  int    `json:"currency"`
	Title     string `json:"group_title"`
	MinServer string `json:"min_server_version"`
}

type StoreCurrency struct {
	ID             int    `json:"storecur_id"`
	Name           string `json:"currency_name"`
	StartingAmount int    `json:"starting_amount"`
	MinServer      string `json:"min_server_version"`
}

type StoreItem struct {
	ID         intish `json:"storeitem_id"`
	Name       string `json:"item_name"`
	Title      string `json:"item_title"`
	Desc       string `json:"item_desc"`
	Price      intish `json:"price"`
	Consumable intish `json:"consumable"`
	Amount     intish `json:"amount"`
	Max        string `json:"max"`
	GroupID    intish `json:"group_id"`
	CurrencyID intish `json:"currency_id"`
	Currency   string `json:"currency"`
	SheetID    string `json:"sheet_id"`
	MinServer  string `json:"min_server_version"`
}

type Teleport struct {
	SrcIsland   int `json:"source_island"`
	SrcMonster  int `json:"source_monster"`
	DestIsland  int `json:"dest_island"`
	DestMonster int `json:"dest_monster"`
}

type Happiness struct {
	Entity int `json:"entity"`
	Value  int `json:"value"`
}

type Requirement struct {
	Entity int
}

func (r *Requirement) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) > 0 && b[0] == '{' {
		var obj struct {
			Entity int `json:"entity"`
		}
		if err := json.Unmarshal(b, &obj); err != nil {
			return err
		}
		r.Entity = obj.Entity
		return nil
	}
	return json.Unmarshal(b, &r.Entity)
}

type blobMap = map[string]any

type quoted[T any] struct {
	V T
}

func (q *quoted[T]) UnmarshalJSON(b []byte) error {
	return decodeEmbedded(b, &q.V)
}

func decodeEmbedded(b []byte, v any) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		if strings.TrimSpace(s) == "" {
			return nil
		}
		b = []byte(s)
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	return dec.Decode(v)
}

type intbool bool

func (v *intbool) UnmarshalJSON(b []byte) error {
	var raw any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	*v = intbool(truthy(raw))
	return nil
}

type intish int

func (n *intish) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		b = []byte(strings.TrimSpace(s))
	}
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	if i, err := strconv.ParseInt(string(b), 10, 64); err == nil {
		*n = intish(i)
		return nil
	}
	f, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return err
	}
	*n = intish(f)
	return nil
}
