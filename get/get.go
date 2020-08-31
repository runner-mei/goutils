package get

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/runner-mei/goutils/as"
)

var ErrValueNotFound = as.ErrValueNotFound
var ErrValueNull = as.ErrValueNull

func BoolWithDefault(attributes map[string]interface{}, key string, defaultValue bool) bool {
	res, e := Bool(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func Bool(attributes map[string]interface{}, key string) (bool, error) {
	v, ok := attributes[key]
	if !ok {
		return false, ErrValueNotFound
	}
	if nil == v {
		return false, ErrValueNull
	}
	return as.Bool(v)
}

func IntWithDefault(attributes map[string]interface{}, key string, defaultValue int) int {
	res, e := Int(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func Int(attributes map[string]interface{}, key string) (int, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return as.Int(v)
}

func UintWithDefault(attributes map[string]interface{}, key string, defaultValue uint) uint {
	res, e := Uint(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func Uint(attributes map[string]interface{}, key string) (uint, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return as.Uint(v)
}

func FloatWithDefault(attributes map[string]interface{}, key string, defaultValue float64) float64 {
	res, e := Float(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func Float(attributes map[string]interface{}, key string) (float64, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return as.Float64(v)
}

func Int32WithDefault(attributes map[string]interface{}, key string, defaultValue int32) int32 {
	res, e := Int32(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func Int32(attributes map[string]interface{}, key string) (int32, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return as.Int32(v)
}

func Int64WithDefault(attributes map[string]interface{}, key string, defaultValue int64) int64 {
	res, e := Int64(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func Int64(attributes map[string]interface{}, key string) (int64, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return as.Int64(v)
}

func Uint32WithDefault(attributes map[string]interface{}, key string, defaultValue uint32) uint32 {
	res, e := Uint32(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func Uint32(attributes map[string]interface{}, key string) (uint32, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return as.Uint32(v)
}

func Uint64WithDefault(attributes map[string]interface{}, key string, defaultValue uint64) uint64 {
	res, e := Uint64(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}
func Uint64(attributes map[string]interface{}, key string) (uint64, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return as.Uint64(v)
}
func StringWithDefault(attributes map[string]interface{}, key string, defaultValue string) string {
	res, e := String(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func String(attributes map[string]interface{}, key string) (string, error) {
	v, ok := attributes[key]
	if !ok {
		return "", ErrValueNotFound
	}
	if nil == v {
		return "", ErrValueNull
	}
	return as.String(v)
}
func TimeWithDefault(attributes map[string]interface{}, key string, defaultValue time.Time) time.Time {
	res, e := Time(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func Time(attributes map[string]interface{}, key string) (time.Time, error) {
	v, ok := attributes[key]
	if !ok {
		return time.Time{}, ErrValueNotFound
	}
	if nil == v {
		return time.Time{}, ErrValueNull
	}
	return as.Time(v)
}

func DurationWithDefault(attributes map[string]interface{}, key string, defaultValue time.Duration) time.Duration {
	res, e := Duration(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func Duration(attributes map[string]interface{}, key string) (time.Duration, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return as.Duration(v)
}

func StringArray(attributes map[string]interface{}, key string) ([]string, error) {
	v, ok := attributes[key]
	if !ok {
		return nil, ErrValueNotFound
	}

	if nil == v {
		return nil, ErrValueNull
	}

	if s, ok := v.(string); ok {
		if strings.HasPrefix(s, "[") {
			var ss []string
			if e := json.Unmarshal([]byte(s), &ss); nil != e {
				return nil, e
			}
			return ss, nil
		}
		return []string{s}, nil
	}

	res, ok := v.([]interface{})
	if !ok {
		ss, ok := v.([]string)
		if !ok {
			return nil, as.CreateTypeError(v, "array")
		}
		return ss, nil
	}
	if 0 == len(res) {
		return nil, nil
	}

	ss := make([]string, 0, len(res))
	for _, s := range res {
		if nil != s {
			ss = append(ss, fmt.Sprint(s))
		}
	}

	return ss, nil
}

func StringArrayWithDefault(attributes map[string]interface{}, key string, defaultValue []string) []string {
	v, ok := attributes[key]
	if !ok {
		return defaultValue
	}

	if nil == v {
		return defaultValue
	}
	if s, ok := v.(string); ok {
		if strings.HasPrefix(s, "[") {
			var ss []string
			if e := json.Unmarshal([]byte(s), &ss); nil != e {
				return defaultValue
			}
			return ss
		}

		return []string{s}
	}

	res, ok := v.([]interface{})
	if !ok {
		ss, ok := v.([]string)
		if !ok {
			return defaultValue
		}
		return ss
	}

	if 0 == len(res) {
		return defaultValue
	}

	ss := make([]string, 0, len(res))
	for _, s := range res {
		if nil != s {
			ss = append(ss, fmt.Sprint(s))
		}
	}

	return ss
}

func IntArray(attributes map[string]interface{}, key string) ([]int, error) {
	v, ok := attributes[key]
	if !ok {
		return nil, ErrValueNotFound
	}

	return as.Ints(v)
}

func IntArrayWithDefault(attributes map[string]interface{}, key string, defValue []int) []int {
	v, ok := attributes[key]
	if !ok {
		return defValue
	}

	ints, e := as.Ints(v)
	if e != nil {
		return defValue
	}
	return ints
}

func Int64Array(attributes map[string]interface{}, key string) ([]int64, error) {
	v, ok := attributes[key]
	if !ok {
		return nil, ErrValueNotFound
	}

	return as.Int64s(v)
}

func Int64ArrayWithDefault(attributes map[string]interface{}, key string, defValue []int64) []int64 {
	v, ok := attributes[key]
	if !ok {
		return defValue
	}

	ints, e := as.Int64s(v)
	if e != nil {
		return defValue
	}
	return ints
}

func Array(attributes map[string]interface{}, key string) ([]interface{}, error) {
	v, ok := attributes[key]
	if !ok {
		return nil, ErrValueNotFound
	}

	if nil == v {
		return nil, ErrValueNull
	}

	res, ok := v.([]interface{})
	if !ok {
		return nil, as.CreateTypeError(v, "array")
	}
	return res, nil
}

func ArrayWithDefault(attributes map[string]interface{}, key string, defaultValue []interface{}) []interface{} {
	v, ok := attributes[key]
	if !ok {
		return defaultValue
	}

	if nil == v {
		return defaultValue
	}

	res, ok := v.([]interface{})
	if !ok {
		return defaultValue
	}
	return res
}

func ObjectWithDefault(attributes map[string]interface{}, key string, defaultValue map[string]interface{}) map[string]interface{} {
	v, ok := attributes[key]
	if !ok {
		return defaultValue
	}

	if nil == v {
		return defaultValue
	}

	res, ok := v.(map[string]interface{})
	if !ok {
		return defaultValue
	}
	return res
}

func Object(attributes map[string]interface{}, key string) (map[string]interface{}, error) {
	v, ok := attributes[key]
	if !ok {
		return nil, ErrValueNotFound
	}

	if nil == v {
		return nil, ErrValueNull
	}

	res, ok := v.(map[string]interface{})
	if !ok {
		return nil, as.CreateTypeError(v, "map")
	}
	return res, nil
}

func ObjectsWithDefault(attributes map[string]interface{}, key string, defaultValue []map[string]interface{}) []map[string]interface{} {
	v, ok := attributes[key]
	if !ok {
		return defaultValue
	}

	if nil == v {
		return defaultValue
	}

	results, e := as.Objects(v)
	if nil != e {
		return defaultValue
	}
	return results
}

func Objects(attributes map[string]interface{}, key string) ([]map[string]interface{}, error) {
	v, ok := attributes[key]
	if !ok {
		return nil, ErrValueNotFound
	}

	if nil == v {
		return nil, ErrValueNull
	}

	return as.Objects(v)
}

func IntList(params map[string]string, key string) ([]int, error) {
	v, ok := params[key]
	if !ok {
		return nil, errors.New("'" + key + "' is not exists in the params.")
	}

	ss := strings.Split(v, ",")
	results := make([]int, 0, len(ss))
	for _, s := range ss {
		i, e := strconv.ParseInt(s, 10, 32)
		if nil != e {
			return nil, errors.New("'" + key + "' contains nonnumber - " + v + ".")
		}
		results = append(results, int(i))
	}
	return results, nil
}
