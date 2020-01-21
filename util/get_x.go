package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/runner-mei/goutils/as"
)

func GetBoolWithDefault(attributes map[string]interface{}, key string, defaultValue bool) bool {
	res, e := GetBool(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetBool(attributes map[string]interface{}, key string) (bool, error) {
	v, ok := attributes[key]
	if !ok {
		return false, ErrValueNotFound
	}
	if nil == v {
		return false, ErrValueNull
	}
	return AsBool(v)
}

func GetIntWithDefault(attributes map[string]interface{}, key string, defaultValue int) int {
	res, e := GetInt(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetInt(attributes map[string]interface{}, key string) (int, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return AsInt(v)
}

func GetUintWithDefault(attributes map[string]interface{}, key string, defaultValue uint) uint {
	res, e := GetUint(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetUint(attributes map[string]interface{}, key string) (uint, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return AsUint(v)
}

func GetFloatWithDefault(attributes map[string]interface{}, key string, defaultValue float64) float64 {
	res, e := GetFloat(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetFloat(attributes map[string]interface{}, key string) (float64, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return AsFloat64(v)
}

func GetInt32WithDefault(attributes map[string]interface{}, key string, defaultValue int32) int32 {
	res, e := GetInt32(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetInt32(attributes map[string]interface{}, key string) (int32, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return AsInt32(v)
}

func GetInt64WithDefault(attributes map[string]interface{}, key string, defaultValue int64) int64 {
	res, e := GetInt64(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetInt64(attributes map[string]interface{}, key string) (int64, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return AsInt64(v)
}

func GetUint32WithDefault(attributes map[string]interface{}, key string, defaultValue uint32) uint32 {
	res, e := GetUint32(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetUint32(attributes map[string]interface{}, key string) (uint32, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return AsUint32(v)
}

func GetUint64WithDefault(attributes map[string]interface{}, key string, defaultValue uint64) uint64 {
	res, e := GetUint64(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}
func GetUint64(attributes map[string]interface{}, key string) (uint64, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return AsUint64(v)
}
func GetStringWithDefault(attributes map[string]interface{}, key string, defaultValue string) string {
	res, e := GetString(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetString(attributes map[string]interface{}, key string) (string, error) {
	v, ok := attributes[key]
	if !ok {
		return "", ErrValueNotFound
	}
	if nil == v {
		return "", ErrValueNull
	}
	return AsString(v)
}
func GetTimeWithDefault(attributes map[string]interface{}, key string, defaultValue time.Time) time.Time {
	res, e := GetTime(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetTime(attributes map[string]interface{}, key string) (time.Time, error) {
	v, ok := attributes[key]
	if !ok {
		return time.Time{}, ErrValueNotFound
	}
	if nil == v {
		return time.Time{}, ErrValueNull
	}
	return AsTime(v)
}

func GetDurationWithDefault(attributes map[string]interface{}, key string, defaultValue time.Duration) time.Duration {
	res, e := GetDuration(attributes, key)
	if nil != e {
		return defaultValue
	}
	return res
}

func GetDuration(attributes map[string]interface{}, key string) (time.Duration, error) {
	v, ok := attributes[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	if nil == v {
		return 0, ErrValueNull
	}
	return AsDuration(v)
}

func GetStringArray(attributes map[string]interface{}, key string) ([]string, error) {
	v, ok := attributes[key]
	if !ok {
		return nil, ErrValueNotFound
	}

	if nil == v {
		return nil, ErrValueNull
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

func GetStringArrayWithDefault(attributes map[string]interface{}, key string, defaultValue []string) []string {
	v, ok := attributes[key]
	if !ok {
		return defaultValue
	}

	if nil == v {
		return defaultValue
	}
	if s, ok := v.(string); ok && strings.HasPrefix(s, "[") {
		var ss []string
		if e := json.Unmarshal([]byte(s), &ss); nil != e {
			return defaultValue
		}
		return ss
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

func GetIntArray(attributes map[string]interface{}, key string) ([]int, error) {
	v, ok := attributes[key]
	if !ok {
		return nil, ErrValueNotFound
	}

	return AsInts(v)
}

func GetIntArrayWithDefault(attributes map[string]interface{}, key string, defValue []int) []int {
	v, ok := attributes[key]
	if !ok {
		return defValue
	}

	ints, e := AsInts(v)
	if e != nil {
		return defValue
	}
	return ints
}

func GetInt64Array(attributes map[string]interface{}, key string) ([]int64, error) {
	v, ok := attributes[key]
	if !ok {
		return nil, ErrValueNotFound
	}

	return AsInt64s(v)
}

func GetInt64ArrayWithDefault(attributes map[string]interface{}, key string, defValue []int64) []int64 {
	v, ok := attributes[key]
	if !ok {
		return defValue
	}

	ints, e := AsInt64s(v)
	if e != nil {
		return defValue
	}
	return ints
}

func GetArray(attributes map[string]interface{}, key string) ([]interface{}, error) {
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

func GetArrayWithDefault(attributes map[string]interface{}, key string, defaultValue []interface{}) []interface{} {
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

func GetObjectWithDefault(attributes map[string]interface{}, key string, defaultValue map[string]interface{}) map[string]interface{} {
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

func GetObject(attributes map[string]interface{}, key string) (map[string]interface{}, error) {
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

func GetObjectsWithDefault(attributes map[string]interface{}, key string, defaultValue []map[string]interface{}) []map[string]interface{} {
	v, ok := attributes[key]
	if !ok {
		return defaultValue
	}

	if nil == v {
		return defaultValue
	}

	results, e := AsObjects(v)
	if nil != e {
		return defaultValue
	}
	return results
}

func GetObjects(attributes map[string]interface{}, key string) ([]map[string]interface{}, error) {
	v, ok := attributes[key]
	if !ok {
		return nil, ErrValueNotFound
	}

	if nil == v {
		return nil, ErrValueNull
	}

	return AsObjects(v)
}

func GetIntList(params map[string]string, key string) ([]int, error) {
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
