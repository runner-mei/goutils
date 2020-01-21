package util

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/runner-mei/goutils/as"
)

type SafeInterfaceMap struct {
	dataMutex sync.Mutex
	data      map[string]interface{}
}

func (sess *SafeInterfaceMap) ForEach(cb func(key string, value interface{})) {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()

	for key, value := range sess.data {
		cb(key, value)
	}
}

// Set put a value to SafeInterfaceMap
func (sess *SafeInterfaceMap) Set(key string, value interface{}) error {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		sess.data = map[string]interface{}{}
	}

	sess.data[key] = value
	return nil
}

// GetWithDefault return a value with the key, if it isn't exists then return default value.
func (sess *SafeInterfaceMap) GetWithDefault(key string, defValue interface{}) interface{} {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return defValue
	}

	value, ok := sess.data[key]
	if !ok {
		return defValue
	}
	if value == nil {
		return defValue
	}
	return value
}

// Get return a value with the key, if it isn't exists then return null.
func (sess *SafeInterfaceMap) Get(key string) interface{} {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return nil
	}
	return sess.data[key]
}

// StringWithDefault return a string with the key, if it isn't exists then return default value.
func (sess *SafeInterfaceMap) StringWithDefault(key, defValue string) string {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return defValue
	}
	return GetStringWithDefault(sess.data, key, defValue)
}

// StringWith return a string with the key, if it isn't exists then return error.
func (sess *SafeInterfaceMap) StringWith(key string) (string, error) {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return "", ErrValueNotFound
	}
	return GetString(sess.data, key)
}

// IntWithDefault return a int with the key, if it isn't exists then return default value.
func (sess *SafeInterfaceMap) IntWithDefault(key string, defValue int) int {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return defValue
	}
	return GetIntWithDefault(sess.data, key, defValue)
}

// IntWith return a int with the key, if it isn't exists then return error.
func (sess *SafeInterfaceMap) IntWith(key string) (int, error) {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return 0, ErrValueNotFound
	}
	return GetInt(sess.data, key)
}

// Int64WithDefault return a int64 with the key, if it isn't exists then return default value.
func (sess *SafeInterfaceMap) Int64WithDefault(key string, defValue int64) int64 {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return defValue
	}
	return GetInt64WithDefault(sess.data, key, defValue)
}

// Int64With return a int64 with the key, if it isn't exists then return error.
func (sess *SafeInterfaceMap) Int64With(key string) (int64, error) {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return 0, ErrValueNotFound
	}
	return GetInt64(sess.data, key)
}

// BoolWithDefault return a bool with the key, if it isn't exists then return default value.
func (sess *SafeInterfaceMap) BoolWithDefault(key string, defValue bool) bool {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return defValue
	}

	return GetBoolWithDefault(sess.data, key, defValue)
}

// BoolWith return a bool with the key, if it isn't exists then return error.
func (sess *SafeInterfaceMap) BoolWith(key string) (bool, error) {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return false, ErrValueNotFound
	}

	return GetBool(sess.data, key)
}

// DurationWithDefault return a Duration with the key, if it isn't exists then return default value.
func (sess *SafeInterfaceMap) DurationWithDefault(key string, defValue time.Duration) time.Duration {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return defValue
	}
	return GetDurationWithDefault(sess.data, key, defValue)
}

// DurationWith return a Duration with the key, if it isn't exists then return error.
func (sess *SafeInterfaceMap) DurationWith(key string) (time.Duration, error) {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return 0, ErrValueNotFound
	}
	return GetDuration(sess.data, key)
}

// TimeWithDefault return a Time with the key, if it isn't exists then return default value.
func (sess *SafeInterfaceMap) TimeWithDefault(key string, defValue time.Time) time.Time {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return defValue
	}
	return GetTimeWithDefault(sess.data, key, defValue)
}

// TimeWith return a Time with the key, if it isn't exists then return error.
func (sess *SafeInterfaceMap) TimeWith(key string) (time.Time, error) {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return time.Time{}, ErrValueNotFound
	}
	return GetTime(sess.data, key)
}

type SafeStringMap struct {
	dataMutex sync.Mutex
	data      map[string]string
}

func (sess *SafeStringMap) ForEach(cb func(key, value string)) {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	for key, value := range sess.data {
		cb(key, value)
	}
}

// Set put a value to SafeStringMap
func (sess *SafeStringMap) Set(key string, value interface{}) error {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		sess.data = map[string]string{}
	}

	switch v := value.(type) {
	case string:
		sess.data[key] = v
	case int:
		sess.data[key] = strconv.FormatInt(int64(v), 10)
	case int16:
		sess.data[key] = strconv.FormatInt(int64(v), 10)
	case int32:
		sess.data[key] = strconv.FormatInt(int64(v), 10)
	case int64:
		sess.data[key] = strconv.FormatInt(v, 10)
	case uint:
		sess.data[key] = strconv.FormatUint(uint64(v), 10)
	case uint16:
		sess.data[key] = strconv.FormatUint(uint64(v), 10)
	case uint32:
		sess.data[key] = strconv.FormatUint(uint64(v), 10)
	case uint64:
		sess.data[key] = strconv.FormatUint(v, 10)
	case time.Duration:
		sess.data[key] = v.String()
	case time.Time:
		sess.data[key] = v.Format(time.RFC3339Nano)
	default:
		return fmt.Errorf("unknow type(%T - %#v)", value, value)
	}
	return nil
}

// GetWithDefault return a value with the key, if it isn't exists then return default value.
func (sess *SafeStringMap) GetWithDefault(key string, defValue interface{}) interface{} {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return defValue
	}

	value, ok := sess.data[key]
	if !ok {
		return defValue
	}
	return value
}

// Get return a value with the key, if it isn't exists then return null.
func (sess *SafeStringMap) Get(key string) interface{} {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return nil
	}

	return sess.data[key]
}

// StringWithDefault return a string with the key, if it isn't exists then return default value.
func (sess SafeStringMap) StringWithDefault(key, defValue string) string {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return defValue
	}

	s, ok := sess.data[key]
	if !ok {
		return defValue
	}
	return s
}

// StringWith return a string with the key, if it isn't exists then return error.
func (sess *SafeStringMap) StringWith(key string) (string, error) {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return "", ErrValueNotFound
	}

	s, ok := sess.data[key]
	if !ok {
		return "", ErrValueNotFound
	}
	return s, nil
}

// IntWithDefault return a int with the key, if it isn't exists then return default value.
func (sess SafeStringMap) IntWithDefault(key string, defValue int) int {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return defValue
	}

	s, ok := sess.data[key]
	if !ok {
		return defValue
	}
	i, e := strconv.ParseInt(s, 10, 0)
	if nil != e {
		return defValue
	}
	return int(i)
}

// IntWith return a int with the key, if it isn't exists then return error.
func (sess *SafeStringMap) IntWith(key string) (int, error) {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	s, ok := sess.data[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	i, e := strconv.ParseInt(s, 10, 0)
	if nil != e {
		return 0, as.CreateTypeError(s, "int")
	}
	return int(i), nil
}

// Int64WithDefault return a int64 with the key, if it isn't exists then return default value.
func (sess *SafeStringMap) Int64WithDefault(key string, defValue int64) int64 {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return defValue
	}

	s, ok := sess.data[key]
	if !ok {
		return defValue
	}
	i, e := strconv.ParseInt(s, 10, 64)
	if nil != e {
		return defValue
	}
	return i
}

// Int64With return a int64 with the key, if it isn't exists then return error.
func (sess *SafeStringMap) Int64With(key string) (int64, error) {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return 0, ErrValueNotFound
	}

	s, ok := sess.data[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	i, e := strconv.ParseInt(s, 10, 64)
	if nil != e {
		return 0, as.CreateTypeError(s, "int64")
	}
	return i, nil
}

// BoolWithDefault return a bool with the key, if it isn't exists then return default value.
func (sess *SafeStringMap) BoolWithDefault(key string, defValue bool) bool {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return defValue
	}

	s, ok := sess.data[key]
	if !ok {
		return defValue
	}
	switch s {
	case "true", "True", "TRUE", "1":
		return true
	case "false", "False", "FALSE", "0":
		return false
	default:
		return defValue
	}
}

// BoolWith return a bool with the key, if it isn't exists then return error.
func (sess *SafeStringMap) BoolWith(key string) (bool, error) {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return false, ErrValueNotFound
	}

	s, ok := sess.data[key]
	if !ok {
		return false, ErrValueNotFound
	}
	switch s {
	case "true", "True", "TRUE", "1":
		return true, nil
	case "false", "False", "FALSE", "0":
		return false, nil
	default:
		return false, as.CreateTypeError(s, "boolean")
	}
}

// DurationWithDefault return a Duration with the key, if it isn't exists then return default value.
func (sess *SafeStringMap) DurationWithDefault(key string, defValue time.Duration) time.Duration {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return defValue
	}

	s, ok := sess.data[key]
	if !ok {
		return defValue
	}
	duration, e := time.ParseDuration(s)
	if nil != e {
		return defValue
	}
	return duration
}

// DurationWith return a Duration with the key, if it isn't exists then return error.
func (sess *SafeStringMap) DurationWith(key string) (time.Duration, error) {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return 0, ErrValueNotFound
	}

	s, ok := sess.data[key]
	if !ok {
		return 0, ErrValueNotFound
	}
	duration, e := time.ParseDuration(s)
	if nil != e {
		return 0, as.CreateTypeError(s, "duration")
	}
	return duration, nil
}

// TimeWithDefault return a Time with the key, if it isn't exists then return default value.
func (sess *SafeStringMap) TimeWithDefault(key string, defValue time.Time) time.Time {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return defValue
	}

	s, ok := sess.data[key]
	if !ok {
		return defValue
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		t, e := time.Parse(layout, s)
		if e == nil {
			return t
		}
	}
	return defValue
}

// TimeWith return a Time with the key, if it isn't exists then return error.
func (sess *SafeStringMap) TimeWith(key string) (time.Time, error) {
	sess.dataMutex.Lock()
	defer sess.dataMutex.Unlock()
	if sess.data == nil {
		return time.Time{}, ErrValueNotFound
	}

	s, ok := sess.data[key]
	if !ok {
		return time.Time{}, ErrValueNotFound
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		t, e := time.Parse(layout, s)
		if e == nil {
			return t, nil
		}
	}
	return time.Time{}, as.CreateTypeError(s, "datetime")
}
