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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Paficent/GoFox2X/data"
)

// TODO: rewrite store_data.json to not just used strings
type Row map[string]any

func (r Row) Str(key string) string {
	switch v := r[key].(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprint(v)
	}
}

func (r Row) Int(key string) int {
	switch v := r[key].(type) {
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return 0
		}
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return int(i)
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return int(f)
		}
	case float64:
		return int(v)
	case bool:
		if v {
			return 1
		}
	}
	return 0
}

func (r Row) Float(key string) float64 {
	switch v := r[key].(type) {
	case json.Number:
		f, _ := v.Float64()
		return f
	case float64:
		return v
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return f
	}
	return 0
}

func (r Row) Bool(key string) bool { return r.Int(key) != 0 }

func (r Row) Has(key string) bool {
	v, ok := r[key]
	return ok && v != nil
}

func (r Row) RawArray(key string) []any {
	a, _ := r[key].([]any)
	return a
}

func (r Row) JSON(key string) map[string]any { return decodeJSONMap(r.Str(key)) }

func (r Row) JSONArray(key string) []any { return decodeJSONArray(r.Str(key)) }

type DB struct {
	tables map[string][]Row
	entIdx map[int]Row
}

func (db *DB) Table(name string) []Row { return db.tables[name] }

func (db *DB) Group(table, key string) map[int][]Row {
	out := map[int][]Row{}
	for _, r := range db.tables[table] {
		out[r.Int(key)] = append(out[r.Int(key)], r)
	}
	return out
}

func (db *DB) entityByID(id int) (Row, bool) {
	if db.entIdx == nil {
		db.entIdx = map[int]Row{}
		for _, r := range db.tables["entities"] {
			db.entIdx[r.Int("entity_id")] = r
		}
	}
	r, ok := db.entIdx[id]
	return r, ok
}

func Open(dir string) (*DB, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory of table JSON files", dir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	db := &DB{tables: map[string][]Row{}}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		rows, err := decodeRows(b)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		db.tables[strings.TrimSuffix(e.Name(), ".json")] = rows
	}
	return db, nil
}

func decodeRows(b []byte) ([]Row, error) {
	if len(bytes.TrimSpace(b)) == 0 {
		return nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var rows []Row
	if err := dec.Decode(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// builder helpers:
func nowMS() int64 { return time.Now().Unix() * 1000 }

func buildArray(rows []Row, fn func(Row) *data.GFSObject) *data.GFSArray {
	arr := data.MakeGFSArray()
	for _, r := range rows {
		if obj := fn(r); obj != nil {
			arr.AddSFSObject(obj)
		}
	}
	return arr
}

func putValue(obj *data.GFSObject, key string, v any) {
	switch val := v.(type) {
	case bool:
		obj.PutBool(key, val)
	case json.Number:
		if isIntNumber(val) {
			i, _ := val.Int64()
			obj.PutInt(key, int(i))
		} else {
			f, _ := val.Float64()
			obj.PutDouble(key, f)
		}
	case string:
		obj.PutUtfString(key, val)
	case map[string]any:
		sub := data.MakeGFSObject()
		putValues(sub, val)
		obj.PutGFSObject(key, sub)
	case []any:
		arr := data.MakeGFSArray()
		for _, item := range val {
			addValue(arr, item)
		}
		obj.PutGFSArray(key, arr)
	default:
		obj.PutUtfString(key, fmt.Sprint(val))
	}
}

func putValues(obj *data.GFSObject, m map[string]any) {
	for k, v := range m {
		putValue(obj, k, v)
	}
}

func isIntNumber(n json.Number) bool { return !strings.ContainsAny(string(n), ".eE") }

func numToInt(v any) int {
	if n, ok := v.(json.Number); ok {
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}

func decodeJSONMap(s string) map[string]any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var m map[string]any
	if dec.Decode(&m) != nil {
		return nil
	}
	return m
}

func decodeJSONArray(s string) []any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var a []any
	if dec.Decode(&a) != nil {
		return nil
	}
	return a
}
