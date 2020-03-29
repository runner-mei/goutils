package as

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/runner-mei/errors"
)

var ErrValueNotFound = errors.ErrValueNotFound
var ErrValueNull = errors.ErrValueNull
var NewTypeError = errors.NewTypeError

func CreateTypeError(value interface{}, exceptedType string, err ...error) error {
	return errors.NewTypeError(fmt.Sprintf("value is not a %s, actual is %T '%#v'", exceptedType, value, value))
}

func errOverflow(value interface{}, exceptedType string) error {
	return errors.NewTypeError(fmt.Sprintf("value is overflow for %s, actual is '%#v'", exceptedType, value))
}

func ConvertToInt64List(value, sep string) ([]int64, error) {
	ss := strings.Split(value, sep)
	results := make([]int64, 0, len(ss))
	for _, s := range ss {
		i, e := strconv.ParseInt(s, 10, 64)
		if nil != e {
			if "" == s {
				continue
			}
			return nil, errors.New("'" + value + "' contains nonnumber.")
		}
		results = append(results, i)
	}
	return results, nil
}

func ConvertToIntList(value, sep string) ([]int, error) {
	ss := strings.Split(value, sep)
	results := make([]int, 0, len(ss))
	for _, s := range ss {
		i, e := strconv.ParseInt(s, 10, 64)
		if nil != e {
			if "" == s {
				continue
			}
			return nil, errors.New("'" + value + "' contains nonnumber.")
		}
		results = append(results, int(i))
	}
	return results, nil
}

// Map type AsSerts to `map`
func Map(value interface{}) (map[string]interface{}, error) {
	if nil == value {
		return nil, ErrValueNull
	}
	if m, ok := value.(map[string]interface{}); ok {
		return m, nil
	}
	if m, ok := value.(map[string]string); ok {
		result := make(map[string]interface{}, len(m))
		for key, value := range m {
			result[key] = value
		}
		return result, nil
	}
	return nil, CreateTypeError(value, "map")
}

func Object(value interface{}) (map[string]interface{}, error) {
	return Map(value)
}

func Objects(v interface{}) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0, 10)
	switch value := v.(type) {
	case []interface{}:
		for i, o := range value {
			r, ok := o.(map[string]interface{})
			if !ok {
				return nil, CreateTypeError(o, "object["+strconv.FormatInt(int64(i), 10)+"]")
			}
			results = append(results, r)
		}
	case map[string]interface{}:
		for k, o := range value {
			r, ok := o.(map[string]interface{})
			if !ok {
				return nil, CreateTypeError(o, "object["+k+"]")
			}
			results = append(results, r)
		}
	case []map[string]interface{}:
		return value, nil
	default:
		if nil == value {
			return nil, ErrValueNull
		}
		return nil, CreateTypeError(value, "map or array")
	}
	return results, nil
}

func IsEmptyArray(value interface{}) bool {
	if a, ok := value.([]interface{}); ok {
		return len(a) == 0
	}
	if a, ok := value.([]map[string]interface{}); ok {
		return len(a) == 0
	}
	if a, ok := value.([]map[string]string); ok {
		return len(a) == 0
	}
	return true
}

func Int64Array(value interface{}) ([]int64, error) {
	switch a := value.(type) {
	case []interface{}:
		ints := make([]int64, len(a))
		for i := range a {
			iv, err := Int64(a[i])
			if err != nil {
				return nil, CreateTypeError(value, "int64Array")
			}
			ints[i] = iv
		}
		return ints, nil
	case []string:
		ints := make([]int64, len(a))
		for i := range a {
			iv, err := strconv.ParseInt(a[i], 10, 64)
			if err != nil {
				return nil, CreateTypeError(value, "int64Array")
			}
			ints[i] = iv
		}
		return ints, nil
	case []int64:
		return a, nil
	case []int:
		ints := make([]int64, len(a))
		for i := range a {
			ints[i] = int64(a[i])
		}
		return ints, nil
	case []int32:
		ints := make([]int64, len(a))
		for i := range a {
			ints[i] = int64(a[i])
		}
		return ints, nil
	case []uint64:
		ints := make([]int64, len(a))
		for i := range a {
			if a[i] > math.MaxInt64 {
				return nil, CreateTypeError(value, "int64Array")
			}
			ints[i] = int64(a[i])
		}
		return ints, nil
	case []uint:
		ints := make([]int64, len(a))
		for i := range a {
			ints[i] = int64(a[i])
		}
		return ints, nil
	case []uint32:
		ints := make([]int64, len(a))
		for i := range a {
			ints[i] = int64(a[i])
		}
		return ints, nil
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}

		if rv.Kind() == reflect.Slice {
			aLen := rv.Len()
			ints := make([]int64, aLen)
			for i := 0; i < aLen; i++ {
				iv, err := Int64(rv.Index(i).Interface())
				if err != nil {
					return nil, CreateTypeError(value, "int64Array")
				}
				ints[i] = iv
			}
			return ints, nil
		}
	}
	return nil, CreateTypeError(value, "int64Array")
}

func Int64ArrayWithDefault(value interface{}, defValue []int64) []int64 {
	if value == nil {
		return defValue
	}
	ints, err := Int64Array(value)
	if err != nil {
		return defValue
	}
	if ints == nil {
		return defValue
	}
	return ints
}

func Uint64Array(value interface{}) ([]uint64, error) {
	switch a := value.(type) {
	case []interface{}:
		uints := make([]uint64, len(a))
		for i := range a {
			iv, err := Uint64(a[i])
			if err != nil {
				return nil, CreateTypeError(value, "uint64Array")
			}
			uints[i] = iv
		}
		return uints, nil
	case []string:
		uints := make([]uint64, len(a))
		for i := range a {
			iv, err := strconv.ParseUint(a[i], 10, 64)
			if err != nil {
				return nil, CreateTypeError(value, "uint64Array")
			}
			uints[i] = iv
		}
		return uints, nil
	case []int64:
		uints := make([]uint64, len(a))
		for i := range a {
			if a[i] < 0 {
				return nil, CreateTypeError(value, "uint64Array")
			}
			uints[i] = uint64(a[i])
		}
		return uints, nil
	case []int:
		uints := make([]uint64, len(a))
		for i := range a {
			if a[i] < 0 {
				return nil, CreateTypeError(value, "uint64Array")
			}
			uints[i] = uint64(a[i])
		}
		return uints, nil
	case []int32:
		uints := make([]uint64, len(a))
		for i := range a {
			if a[i] < 0 {
				return nil, CreateTypeError(value, "uint64Array")
			}
			uints[i] = uint64(a[i])
		}
		return uints, nil
	case []uint64:
		return a, nil
	case []uint:
		uints := make([]uint64, len(a))
		for i := range a {
			uints[i] = uint64(a[i])
		}
		return uints, nil
	case []uint32:
		uints := make([]uint64, len(a))
		for i := range a {
			uints[i] = uint64(a[i])
		}
		return uints, nil
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}

		if rv.Kind() == reflect.Slice {
			aLen := rv.Len()
			uints := make([]uint64, aLen)
			for i := 0; i < aLen; i++ {
				iv, err := Uint64(rv.Index(i).Interface())
				if err != nil {
					return nil, CreateTypeError(value, "uint64Array")
				}
				uints[i] = iv
			}
			return uints, nil
		}
	}
	return nil, CreateTypeError(value, "uint64Array")
}

func Uint64ArrayWithDefault(value interface{}, defValue []uint64) []uint64 {
	if value == nil {
		return defValue
	}
	uints, err := Uint64Array(value)
	if err != nil {
		return defValue
	}
	if uints == nil {
		return defValue
	}
	return uints
}

func Array(value interface{}) ([]interface{}, error) {
	if a, ok := value.([]interface{}); ok {
		return a, nil
	}
	if a, ok := value.([]map[string]interface{}); ok {
		res := make([]interface{}, len(a))
		for idx, v := range a {
			res[idx] = v
		}
		return res, nil
	}
	if a, ok := value.([]map[string]string); ok {
		res := make([]interface{}, len(a))
		for idx, v := range a {
			res[idx] = v
		}
		return res, nil
	}
	return nil, CreateTypeError(value, "array")
}

func IntsWithDefault(value interface{}, defValue []int) []int {
	if value == nil {
		return defValue
	}
	switch vv := value.(type) {
	case []int:
		return vv
	default:
		ints, err := Int64Array(value)
		if err != nil {
			return defValue
		}
		if ints == nil {
			return defValue
		}
		var values = make([]int, 0, len(ints))
		for _, v := range ints {
			values = append(values, int(v))
		}
		return values
	}
}

// InplaceReader a inplace reader for bufio.Scanner
type InplaceReader int

func (p *InplaceReader) Read([]byte) (int, error) {
	if *p == 0 {
		return 0, io.EOF
	}
	ret := int(*p)
	*p = 0
	return ret, io.EOF
}

func ToStrings(o interface{}) []string {
	if ss, ok := o.([]string); ok {
		return ss
	}

	if ss, ok := o.([]interface{}); ok {
		var ipList []string
		for _, i := range ss {
			ipList = append(ipList, fmt.Sprint(i))
		}
		return ipList
	}

	s, ok := o.(string)
	if ok {

		ss, err := SplitStrings([]byte(s))
		if err != nil {
			panic(err)
		}
		return ss
	}

	bs, ok := o.([]byte)
	if !ok {
		panic(fmt.Errorf("o is unsupport type - %T %s", o, o))
	}
	ss, err := SplitStrings(bs)
	if err != nil {
		panic(err)
	}
	return ss
}

func SplitStrings(bs []byte) ([]string, error) {
	if len(bs) == 0 {
		return nil, nil
	}

	var ipList []string
	if err := json.Unmarshal(bs, &ipList); err == nil {
		return ipList, nil
	}

	r := InplaceReader(len(bs))
	scanner := bufio.NewScanner(&r)
	scanner.Buffer(bs, len(bs))

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		for _, field := range bytes.Split(line, []byte(",")) {
			if len(field) == 0 {
				continue
			}
			field = bytes.TrimSpace(field)
			if len(field) == 0 {
				continue
			}
			ipList = append(ipList, string(field))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ipList, nil
}

func Strings(value interface{}) ([]string, error) {
	if value == nil {
		return nil, ErrValueNull
	}
	switch vv := value.(type) {
	case []string:
		return vv, nil
	case []interface{}:
		results := make([]string, 0, len(vv))
		for _, v := range vv {
			s, e := String(v)
			if e != nil {
				return nil, e
			}
			results = append(results, s)
		}
		return results, nil
	}

	return nil, CreateTypeError(value, "string array")
}

func StringsWithDefault(value interface{}, defValue []string) []string {
	if value == nil {
		return defValue
	}
	switch vv := value.(type) {
	case []string:
		return vv
	case []interface{}:
		results := make([]string, 0, len(vv))
		for _, v := range vv {
			s, e := String(v)
			if e != nil {
				return defValue
			}
			results = append(results, s)
		}
		return results
	}

	return defValue
}

func Ints(value interface{}) ([]int, error) {
	if a, ok := value.([]interface{}); ok {
		ints := make([]int, 0, len(a))
		for _, v := range a {
			i, e := Int(v)
			if nil != e {
				return nil, e
			}
			ints = append(ints, i)
		}
		return ints, nil
	}
	return nil, CreateTypeError(value, "int array")
}

func Int64s(value interface{}) ([]int64, error) {
	if a, ok := value.([]interface{}); ok {
		ints := make([]int64, 0, len(a))
		for _, v := range a {
			i, e := Int64(v)
			if nil != e {
				return nil, e
			}
			ints = append(ints, i)
		}
		return ints, nil
	}
	return nil, CreateTypeError(value, "int64 array")
}

func ArrayWithDefault(value interface{}, defValue []interface{}) []interface{} {
	arr, err := Array(value)
	if nil != err {
		return defValue
	}
	return arr
}

// Bool type AsSerts to `bool`
func Bool(value interface{}) (bool, error) {
	if s, ok := value.(bool); ok {
		return s, nil
	}
	if s, ok := value.(string); ok {
		switch s {
		case "TRUE", "True", "true", "YES", "Yes", "yes", "on", "enabled":
			return true, nil
		case "FALSE", "False", "false", "NO", "No", "no":
			return false, nil
		}
	}
	if nil == value {
		return false, ErrValueNull
	}
	return false, CreateTypeError(value, "bool")
}

// Bool type AsSerts to `bool`
func BoolWithDefaultValue(value interface{}, defaultValue bool) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	if s, ok := value.(string); ok {
		switch s {
		case "TRUE", "True", "true", "YES", "Yes", "yes":
			return true
		case "FALSE", "False", "false", "NO", "No", "no":
			return false
		}
	}
	return defaultValue
}

func Int(value interface{}) (int, error) {
	a, err := Int32(value)
	return int(a), err
}

func Uint(value interface{}) (uint, error) {
	a, err := Uint32(value)
	return uint(a), err
}

// Int type AsSerts to `float64` then converts to `int`
func Int64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		if 9223372036854775807 >= int64(v) {
			return int64(v), nil
		}
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		if 9223372036854775807 >= v {
			return int64(v), nil
		}
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case []byte:
		i64, err := strconv.ParseInt(string(v), 10, 64)
		if nil == err {
			return i64, nil
		}
	case string:
		i64, err := strconv.ParseInt(v, 10, 64)
		if nil == err {
			return i64, nil
		}
	case json.Number:
		i64, err := v.Int64()
		if nil == err {
			return i64, nil
		}
		f64, err := v.Float64()
		if nil == err {
			return int64(f64), nil
		}
	}
	if nil == value {
		return 0, ErrValueNull
	}
	return 0, CreateTypeError(value, "int64")
}

func Int64WithDefault(v interface{}, defaultValue int64) int64 {
	s, e := Int64(v)
	if nil != e {
		return defaultValue
	}
	return s
}

func Int32(value interface{}) (int32, error) {
	i64, err := Int64(value)
	if nil != err {
		return 0, CreateTypeError(value, "int32")
	}
	if -2147483648 > i64 || 2147483647 < i64 {
		return 0, errOverflow(value, "int32")
	}
	return int32(i64), nil
}

func Int32WithDefault(value interface{}, defaultValue int32) int32 {
	i32, e := Int32(value)
	if nil == e {
		return i32
	}
	return defaultValue
}

func Int16(value interface{}) (int16, error) {
	i64, err := Int64(value)
	if nil != err {
		return 0, CreateTypeError(value, "int16")
	}
	if -32768 > i64 || 32767 < i64 {
		return 0, errOverflow(value, "int16")
	}
	return int16(i64), nil
}

func Int8(value interface{}) (int8, error) {
	i64, err := Int64(value)
	if nil != err {
		return 0, CreateTypeError(value, "int8")
	}
	if -128 > i64 || 127 < i64 {
		return 0, errOverflow(value, "int8")
	}
	return int8(i64), nil
}

// Uint type AsSerts to `float64` then converts to `int`
func Uint64(value interface{}) (uint64, error) {
	switch v := value.(type) {
	case []byte:
		i64, err := strconv.ParseUint(string(v), 10, 64)
		if nil == err {
			return i64, nil
		}
	case string:
		i64, err := strconv.ParseUint(v, 10, 64)
		if nil == err {
			return i64, nil
		}
		return i64, CreateTypeError(value, "uint64")

	case json.Number:
		i64, err := strconv.ParseUint(v.String(), 10, 64)
		if nil == err {
			return i64, nil
		}
		f64, err := v.Float64()
		if nil == err {
			if f64 >= 0 {
				return uint64(f64), nil
			}
			if f64 < 0 {
				if math.IsNaN(f64) {
					return 0, nil
				}
				if int64(f64) == 0 {
					return 0, nil
				}
			}
		} else {
			return 0, CreateTypeError(value, "uint64")
		}
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case int:
		if v >= 0 {
			return uint64(v), nil
		}
	case int8:
		if v >= 0 {
			return uint64(v), nil
		}
	case int16:
		if v >= 0 {
			return uint64(v), nil
		}
	case int32:
		if v >= 0 {
			return uint64(v), nil
		}
	case int64:
		if v >= 0 {
			return uint64(v), nil
		}
	case float32:
		if v >= 0 && 18446744073709551615 >= v {
			return uint64(v), nil
		}

		if v < 0 {
			if math.IsNaN(float64(v)) {
				return 0, nil
			}
			if int64(v) == 0 {
				return 0, nil
			}
		}
	case float64:
		if v >= 0 && 18446744073709551615 >= v {
			return uint64(v), nil
		}

		if v < 0 {
			if math.IsNaN(v) {
				return 0, nil
			}
			if int64(v) == 0 {
				return 0, nil
			}
		}
	}
	if nil == value {
		return 0, ErrValueNull
	}
	return 0, CreateTypeError(value, "uint64")
}

func Uint32(value interface{}) (uint32, error) {
	ui64, err := Uint64(value)
	if nil != err {
		return 0, CreateTypeError(value, "uint32")
	}
	if 4294967295 < ui64 {
		return 0, errOverflow(value, "uint32")
	}
	return uint32(ui64), nil
}

func Uint16(value interface{}) (uint16, error) {
	ui64, err := Uint64(value)
	if nil != err {
		return 0, CreateTypeError(value, "uint16")
	}
	if 65535 < ui64 {
		return 0, errOverflow(value, "uint16")
	}
	return uint16(ui64), nil
}

func Uint8(value interface{}) (uint8, error) {
	ui64, err := Uint64(value)
	if nil != err {
		return 0, CreateTypeError(value, "uint8")
	}
	if 255 < ui64 {
		return 0, errOverflow(value, "uint8")
	}
	return uint8(ui64), nil
}

// Uint type AsSerts to `float64` then converts to `int`
func Float64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case string:
		f64, err := strconv.ParseFloat(v, 64)
		if nil == err {
			return f64, nil
		}
		return f64, CreateTypeError(value, "float64")
	case json.Number:
		return v.Float64()
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	}
	if nil == value {
		return 0, ErrValueNull
	}
	return 0, CreateTypeError(value, "float64")
}

func Float64WithDefault(value interface{}, defaultValue float64) float64 {
	f64, err := Float64(value)
	if nil != err {
		return defaultValue
	}
	return f64
}

func Float32(value interface{}) (float32, error) {
	f64, err := Float64(value)
	if nil != err {
		return 0, CreateTypeError(value, "float32")
	}

	if f64 > math.MaxFloat32 {
		return 0, CreateTypeError(value, "float32")
	}
	return float32(f64), nil
}

func Float32WithDefault(value interface{}, defaultValue float32) float32 {
	f64, err := Float64(value)
	if nil != err {
		return defaultValue
	}
	return float32(f64)
}

// String type AsSerts to `string`
func String(value interface{}) (string, error) {
	if nil == value {
		return "", ErrValueNull
	}

	switch v := value.(type) {
	case string:
		return v, nil
	case json.Number:
		return v.String(), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case int:
		return strconv.FormatInt(int64(v), 10), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'e', -1, 64), nil
	case float64:
		return strconv.FormatFloat(v, 'e', -1, 64), nil
	case bool:
		if v {
			return "true", nil
		} else {
			return "false", nil
		}
	}

	return "", CreateTypeError(value, "string")
}

func StringWithDefault(v interface{}, defaultStr string) string {
	s, e := String(v)
	if nil != e {
		return defaultStr
	}
	return s
}

func DurationArray(value interface{}) ([]time.Duration, error) {
	switch svalue := value.(type) {
	case []string:
		results := make([]time.Duration, 0, len(svalue))
		for idx := range svalue {
			if svalue[idx] == "" {
				continue
			}

			addr, err := Duration(svalue[idx])
			if err != nil {
				return nil, CreateTypeError(value, "DurationArray")
			}
			results = append(results, addr)
		}
		return results, nil
	case []time.Duration:
		return svalue, nil
	case []interface{}:
		results := make([]time.Duration, 0, len(svalue))
		for idx := range svalue {
			if svalue[idx] == nil {
				continue
			}

			switch tvalue := svalue[idx].(type) {
			case string:
				addr, err := Duration(tvalue)
				if err != nil {
					return nil, CreateTypeError(value, "DurationArray")
				}
				results = append(results, addr)
			case time.Duration:
				results = append(results, tvalue)
			default:
				return nil, CreateTypeError(value, "DurationArray")
			}
		}
		return results, nil
	}
	return nil, CreateTypeError(value, "DurationArray")
}

func Duration(v interface{}) (time.Duration, error) {
	if t, ok := v.(time.Duration); ok {
		return t, nil
	}

	if i, e := Int64(v); nil == e {
		return time.Duration(i), nil
	}

	s, ok := v.(string)
	if !ok {
		return 0, CreateTypeError(v, "duration")
	}

	m, e := time.ParseDuration(s)
	if nil == e {
		return m, nil
	}
	return 0, CreateTypeError(v, "duration")
}

func DurationWithDefault(v interface{}, defValue time.Duration) time.Duration {
	if t, ok := v.(time.Duration); ok {
		return t
	}

	if i, e := Int64(v); nil == e {
		return time.Duration(i)
	}

	s, ok := v.(string)
	if !ok {
		return defValue
	}

	m, e := time.ParseDuration(s)
	if nil == e {
		return m
	}
	return defValue
}

func Time(v interface{}) (time.Time, error) {
	if t, ok := v.(time.Time); ok {
		return t, nil
	}

	s, ok := v.(string)
	if !ok {
		return time.Time{}, CreateTypeError(v, "Time")
	}

	for _, layout := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		m, e := time.ParseInLocation(layout, s, time.Local)
		if nil == e {
			return m, nil
		}
	}

	return time.Time{}, CreateTypeError(v, "Time")
}

func TimeWithDefault(v interface{}, defValue time.Time) time.Time {
	if t, ok := v.(time.Time); ok {
		return t
	}

	s, ok := v.(string)
	if !ok {
		return defValue
	}

	for _, layout := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		m, e := time.ParseInLocation(layout, s, time.Local)
		if nil == e {
			return m
		}
	}

	return defValue
}

func BoolWithDefault(v interface{}, defaultValue bool) bool {
	b, e := Bool(v)
	if nil != e {
		return defaultValue
	}
	return b
}

func IntWithDefault(v interface{}, defaultValue int) int {
	i, e := Int(v)
	if nil != e {
		return defaultValue
	}
	return i
}

func UintWithDefault(v interface{}, defaultValue uint) uint {
	u, e := Uint(v)
	if nil != e {
		return defaultValue
	}
	return u
}

func Uint32WithDefault(v interface{}, defaultValue uint32) uint32 {
	u32, e := Uint32(v)
	if nil != e {
		return defaultValue
	}
	return u32
}

func Uint64WithDefault(v interface{}, defaultValue uint64) uint64 {
	u64, e := Uint64(v)
	if nil != e {
		return defaultValue
	}
	return u64
}

func ObjectWithDefault(v interface{}, defaultValue map[string]interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	if m, ok := v.(map[string]string); ok {
		result := make(map[string]interface{}, len(m))
		for key, value := range m {
			result[key] = value
		}
		return result
	}
	return defaultValue
}

func ObjectsWithDefault(v interface{}, defaultValue []map[string]interface{}) []map[string]interface{} {
	if o, ok := v.([]map[string]interface{}); ok {
		return o
	}

	a, ok := v.([]interface{})
	if !ok {
		return defaultValue
	}

	res := make([]map[string]interface{}, 0, len(a))
	for _, value := range a {
		m, ok := value.(map[string]interface{})
		if !ok {
			return defaultValue
		}
		res = append(res, m)
	}
	return res
}

func Biginteger(value interface{}) (big.Int, error) {
	var intValue big.Int
	switch v := value.(type) {
	case string:
		_, ok := intValue.SetString(v, 10)
		if ok {
			return intValue, nil
		}
	case json.Number:
		_, ok := intValue.SetString(v.String(), 10)
		if ok {
			return intValue, nil
		}
	case int:
		intValue.SetInt64(int64(v))
		return intValue, nil
	case int64:
		intValue.SetInt64(v)
		return intValue, nil
	case uint:
		intValue.SetUint64(uint64(v))
		return intValue, nil
	case uint64:
		intValue.SetUint64(v)
		return intValue, nil
	case uint8:
		intValue.SetUint64(uint64(v))
		return intValue, nil
	case uint16:
		intValue.SetUint64(uint64(v))
		return intValue, nil
	case uint32:
		intValue.SetUint64(uint64(v))
		return intValue, nil
	case int8:
		intValue.SetInt64(int64(v))
		return intValue, nil
	case int16:
		intValue.SetInt64(int64(v))
		return intValue, nil
	case int32:
		intValue.SetInt64(int64(v))
		return intValue, nil
	case float32:
		if v >= 0 && math.MaxUint64 >= v {
			intValue.SetUint64(uint64(v))
			return intValue, nil
		} else if v < 0 && math.MinInt64 <= v {
			intValue.SetInt64(int64(v))
			return intValue, nil
		}
	case float64:
		if v >= 0 && math.MaxUint64 >= v {
			intValue.SetUint64(uint64(v))
			return intValue, nil
		} else if v < 0 && math.MinInt64 <= v {
			intValue.SetInt64(int64(v))
			return intValue, nil
		}
	}
	if nil == value {
		return intValue, ErrValueNull
	}
	return intValue, CreateTypeError(value, "big.Int")
}

func IPString(value interface{}) (string, error) {
	switch svalue := value.(type) {
	case string:
		return svalue, nil
	case net.IP:
		return svalue.String(), nil
	case *net.IP:
		return svalue.String(), nil
	}
	return "", CreateTypeError(value, "IP")
}

func IPStrings(value interface{}) ([]string, error) {
	switch svalue := value.(type) {
	case []string:
		return svalue, nil
	case []net.IP:
		results := make([]string, len(svalue))
		for idx := range svalue {
			results[idx] = svalue[idx].String()
		}
		return results, nil
	case []*net.IP:
		results := make([]string, len(svalue))
		for idx := range svalue {
			results[idx] = svalue[idx].String()
		}
		return results, nil
	case []interface{}:
		results := make([]string, len(svalue))
		for idx := range svalue {
			addr, err := IPString(svalue[idx])
			if err != nil {
				return nil, err
			}
			results[idx] = addr
		}
		return results, nil
	}
	return nil, CreateTypeError(value, "IPStringArray")
}

func IPAddresses(value interface{}) ([]net.IP, error) {
	switch svalue := value.(type) {
	case []string:
		results := make([]net.IP, 0, len(svalue))
		for idx := range svalue {
			if svalue[idx] == "" {
				continue
			}

			addr := net.ParseIP(svalue[idx])
			if addr == nil {
				return nil, CreateTypeError(value, "IPArray")
			}
			results = append(results, addr)
		}
		return results, nil
	case []net.IP:
		return svalue, nil
	case []*net.IP:
		results := make([]net.IP, 0, len(svalue))
		for idx := range svalue {
			if svalue[idx] != nil {
				results = append(results, *svalue[idx])
			}
		}
		return results, nil
	case []interface{}:
		results := make([]net.IP, 0, len(svalue))
		for idx := range svalue {
			if svalue[idx] == nil {
				continue
			}

			switch tvalue := svalue[idx].(type) {
			case string:
				addr := net.ParseIP(tvalue)
				if addr == nil {
					return nil, CreateTypeError(value, "IPArray")
				}
				results = append(results, addr)
			case net.IP:
				results = append(results, tvalue)
			case *net.IP:
				if tvalue != nil {
					results = append(results, *tvalue)
				}
			default:
				return nil, CreateTypeError(value, "IPArray")
			}
		}
		return results, nil
	}
	return nil, CreateTypeError(value, "IPArray")
}

func MacString(value interface{}) (string, error) {
	switch svalue := value.(type) {
	case string:
		return svalue, nil
	case net.HardwareAddr:
		return svalue.String(), nil
	case *net.HardwareAddr:
		return svalue.String(), nil
	}
	return "", CreateTypeError(value, "HardwareAddr")
}

func MacStrings(value interface{}) ([]string, error) {
	switch svalue := value.(type) {
	case []string:
		return svalue, nil
	case []net.HardwareAddr:
		results := make([]string, len(svalue))
		for idx := range svalue {
			results[idx] = svalue[idx].String()
		}
		return results, nil
	case []*net.HardwareAddr:
		results := make([]string, len(svalue))
		for idx := range svalue {
			results[idx] = svalue[idx].String()
		}
		return results, nil
	case []interface{}:
		results := make([]string, len(svalue))
		for idx := range svalue {
			addr, err := MacString(svalue[idx])
			if err != nil {
				return nil, err
			}
			results[idx] = addr
		}
		return results, nil
	}
	return nil, CreateTypeError(value, "MacStringArray")
}

func HardwareAddresses(value interface{}) ([]net.HardwareAddr, error) {
	switch svalue := value.(type) {
	case []string:
		results := make([]net.HardwareAddr, 0, len(svalue))
		for idx := range svalue {
			if svalue[idx] == "" {
				continue
			}

			addr, err := net.ParseMAC(svalue[idx])
			if err != nil {
				return nil, CreateTypeError(value, "MacArray")
			}
			results = append(results, addr)
		}
		return results, nil
	case []net.HardwareAddr:
		return svalue, nil
	case []*net.HardwareAddr:
		results := make([]net.HardwareAddr, 0, len(svalue))
		for idx := range svalue {
			if svalue[idx] != nil {
				results = append(results, *svalue[idx])
			}
		}
		return results, nil
	case []interface{}:
		results := make([]net.HardwareAddr, 0, len(svalue))
		for idx := range svalue {
			if svalue[idx] == nil {
				continue
			}

			switch tvalue := svalue[idx].(type) {
			case string:
				addr, err := net.ParseMAC(tvalue)
				if err != nil {
					return nil, CreateTypeError(value, "MacArray")
				}
				results = append(results, addr)
			case net.HardwareAddr:
				results = append(results, tvalue)
			case *net.HardwareAddr:
				if tvalue != nil {
					results = append(results, *tvalue)
				}
			default:
				return nil, CreateTypeError(value, "MacArray")
			}
		}
		return results, nil
	}
	return nil, CreateTypeError(value, "MacArray")
}

func Get(values map[string]interface{}, name string) interface{} {
	if values == nil {
		return nil
	}
	return values[name]
}
