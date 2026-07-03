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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Paficent/GoFox2X/data"
)

func nowMS() int64 { return time.Now().Unix() * 1000 }

func buildArray[T any](rows []T, fn func(T) *data.GFSObject) *data.GFSArray {
	arr := data.MakeGFSArray()
	for _, r := range rows {
		if obj := fn(r); obj != nil {
			arr.AddSFSObject(obj)
		}
	}
	return arr
}

func dynamicObject(m map[string]any) *data.GFSObject {
	obj := data.MakeGFSObject()
	for k, v := range m {
		putDynamic(obj, k, v)
	}
	return obj
}

func dynamicArray(items []any) *data.GFSArray {
	arr := data.MakeGFSArray()
	for _, v := range items {
		addDynamic(arr, v)
	}
	return arr
}

func putDynamic(obj *data.GFSObject, key string, v any) {
	switch x := v.(type) {
	case nil:
	case bool:
		obj.PutBool(key, x)
	case json.Number:
		if isInt(x) {
			i, _ := x.Int64()
			obj.PutInt(key, int(i))
		} else {
			f, _ := x.Float64()
			obj.PutDouble(key, f)
		}
	case string:
		obj.PutUtfString(key, x)
	case map[string]any:
		obj.PutGFSObject(key, dynamicObject(x))
	case []any:
		obj.PutGFSArray(key, dynamicArray(x))
	default:
		obj.PutUtfString(key, fmt.Sprint(x))
	}
}

func addDynamic(arr *data.GFSArray, v any) {
	switch x := v.(type) {
	case nil:
	case bool:
		arr.AddBool(x)
	case json.Number:
		if isInt(x) {
			i, _ := x.Int64()
			arr.AddInt(int(i))
		} else {
			f, _ := x.Float64()
			arr.AddDouble(f)
		}
	case string:
		arr.AddUtfString(x)
	case map[string]any:
		arr.AddSFSObject(dynamicObject(x))
	case []any:
		arr.AddSFSArray(dynamicArray(x))
	default:
		arr.AddUtfString(fmt.Sprint(x))
	}
}

func isInt(n json.Number) bool { return !strings.ContainsAny(string(n), ".eE") }

func toInt(v any) int {
	switch x := v.(type) {
	case json.Number:
		i, _ := x.Int64()
		return int(i)
	case float64:
		return int(x)
	case int:
		return x
	}
	return 0
}

func truthy(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return x != 0
	case int:
		return x != 0
	case json.Number:
		if isInt(x) {
			i, _ := x.Int64()
			return i != 0
		}
		f, _ := x.Float64()
		return f != 0
	case string:
		s := strings.TrimSpace(strings.ToLower(x))
		return s == "true" || s == "1"
	}
	return false
}

func jstr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
