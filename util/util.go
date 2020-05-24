package util

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/runner-mei/goutils/syncx"
)

var IsWindows = runtime.GOOS == "windows"

type CloseFunc = syncx.CloseFunc

func CloseBatch(closeList ...io.Closer) error {
	var errList []error
	for _, c := range closeList {
		if err := c.Close(); err != nil {
			errList = append(errList, err)
		}
	}
	if len(errList) == 0 {
		return nil
	}

	if len(errList) == 1 {
		return errList[0]
	}
	var buffer bytes.Buffer
	isFirst := true
	for _, e := range errList {
		if isFirst {
			isFirst = false
		} else {
			buffer.WriteString("\r\n")
		}
		buffer.WriteString(e.Error())
	}
	return errors.New(buffer.String())
}

func TryClose(v interface{}) {
	if c, ok := v.(interface {
		Close()
	}); ok {
		c.Close()
	}
	if c, ok := v.(io.Closer); ok {
		c.Close()
	}
}

// CloseWith 捕获错误并打印
func CloseWith(closer io.Closer) {
	if err := closer.Close(); err != nil {
		if err == sql.ErrTxDone {
			return
		}

		log.Println("[WARN]", err)
		panic(err)
	}
}

// RollbackWith 捕获错误并打印
func RollbackWith(closer interface {
	Rollback() error
}, noPanic ...bool) {
	if err := closer.Rollback(); err != nil {
		if err == sql.ErrTxDone {
			return
		}

		log.Println("[WARN]", err)
		if len(noPanic) == 0 || !noPanic[0] {
			panic(err)
		}
	}
}

func ToJSON(a interface{}) string {
	bs, _ := json.Marshal(a)
	if len(bs) == 0 {
		return ""
	}
	return string(bs)
}

func decodeHook(from reflect.Kind, to reflect.Kind, v interface{}) (interface{}, error) {
	if from == reflect.String && to == reflect.Bool {
		s := v.(string)
		if s == "off" || s == "false" || s == "FALSE" || s == "False" {
			return false, nil
		}
		return s == "on" || s == "true" || s == "TRUE" || strings.ToLower(s) == "True", nil
	}
	return v, nil
}

func ToStruct(rawVal interface{}, row map[string]interface{}) (err error) {
	config := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(decodeHook,
			stringToTimeHookFunc(time.RFC3339,
				time.RFC3339Nano,
				"2006-01-02 15:04:05Z07:00",
				"2006-01-02 15:04:05",
				"2006-01-02")),
		Metadata:         nil,
		Result:           rawVal,
		TagName:          "json",
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(row)
}

func stringToTimeHookFunc(layouts ...string) mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}
		s := data.(string)
		if s == "" {
			return time.Time{}, nil
		}
		for _, layout := range layouts {
			t, err := time.Parse(layout, s)
			if err == nil {
				return t, nil
			}
		}
		// Convert it by parsing
		return data, nil
	}
}

func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func CopyFrom(froms ...map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for _, from := range froms {
		for k, v := range from {
			res[k] = v
		}
	}
	return res
}
