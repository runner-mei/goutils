package render

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/runner-mei/errors"
	"github.com/runner-mei/goutils/as"
	"github.com/runner-mei/goutils/human"
	"github.com/runner-mei/goutils/tid"
	"github.com/runner-mei/goutils/util"
	"golang.org/x/text/transform"
)

func QueryEscape(charset, content string) string {
	encoding := util.GetCharset(charset)
	new_content, _, err := transform.String(encoding.NewEncoder(), content)
	if err != nil {
		return content
	}
	return url.QueryEscape(new_content)
}

func parseInterval(s string, defValue time.Duration) time.Duration {
	minus := false
	if strings.HasPrefix(s, "-") {
		minus = true
		s = strings.TrimPrefix(s, "-")
	} else if strings.HasPrefix(s, "+") {
		s = strings.TrimPrefix(s, "+")
	}

	a, err := time.ParseDuration(s)
	if err != nil {
		return defValue
	}

	if minus {
		return -a
	}
	return a
}

var TemplateFuncs = template.FuncMap{
	"add": func(a interface{}, b ...interface{}) interface{} {
		fa, err := as.Float64(a)
		if err != nil {
			panic(err)
		}

		for _, v := range b {
			fb, err := as.Float64(v)
			if err != nil {
				panic(err)
			}

			fa += fb
		}
		return fa
	},
	"sub": func(a interface{}, b ...interface{}) interface{} {
		fa, err := as.Float64(a)
		if err != nil {
			panic(err)
		}

		for _, v := range b {
			fb, err := as.Float64(v)
			if err != nil {
				panic(err)
			}

			fa -= fb
		}
		return fa
	},
	"div": func(a interface{}, b ...interface{}) interface{} {
		fa, err := as.Float64(a)
		if err != nil {
			panic(err)
		}

		for _, v := range b {
			fb, err := as.Float64(v)
			if err != nil {
				panic(err)
			}

			fa /= fb
		}
		return fa
	},
	"mul": func(a interface{}, b ...interface{}) interface{} {
		fa, err := as.Float64(a)
		if err != nil {
			panic(err)
		}

		for _, v := range b {
			fb, err := as.Float64(v)
			if err != nil {
				panic(err)
			}

			fa *= fb
		}
		return fa
	},

	"concat": func(values ...interface{}) string {
		var buf bytes.Buffer
		for _, v := range values {
			fmt.Fprint(&buf, v)
		}
		return buf.String()
	},
	"toString": func(v interface{}) string {
		return fmt.Sprint(v)
	},
	"timeFormat": func(format string, t interface{}) string {
		now, err := as.Time(t)
		if err != nil {
			return fmt.Sprint(t)
		}
		switch {
		case strings.HasPrefix(format, "unix"):
			interval := time.Duration(0)
			if len(format) >= 2 {
				interval = parseInterval(strings.TrimSpace(strings.TrimPrefix(format, "unix")), 0)
			}

			return strconv.FormatInt(now.UTC().Add(interval).Unix(), 10)
		case strings.HasPrefix(format, "unix_ms"):
			interval := time.Duration(0)
			if len(format) >= 2 {
				interval = parseInterval(strings.TrimSpace(strings.TrimPrefix(format, "unix_ms")), 0)
			}
			return strconv.FormatInt(now.UTC().Add(interval).UnixNano()/int64(time.Millisecond), 10)
		}
		return now.Format(format)
	},
	"nowUnix": func() int64 {
		return time.Now().Unix()
	},
	"timeUnix": func(t time.Time) int64 {
		return t.Unix()
	},
	"generateID": tid.GenerateID,
	"toLower":    strings.ToLower,
	"toUpper":    strings.ToUpper,
	"toTitle":    strings.ToTitle,
	"replace": func(old_s, new_s, content string) string {
		return strings.Replace(content, old_s, new_s, -1)
	},
	"now": func(format ...string) interface{} {
		if len(format) == 0 {
			return time.Now()
		}

		interval := time.Duration(0)
		if len(format) >= 2 {
			interval = parseInterval(format[1], 0)
		}
		switch format[0] {
		case "unix":
			return strconv.FormatInt(time.Now().Add(interval).UTC().Unix(), 10)
		case "unix_ms":
			return strconv.FormatInt(time.Now().Add(interval).UTC().UnixNano()/int64(time.Millisecond), 10)
		}
		return time.Now().Format(format[0])
	},
	"md5": func(s string) string {
		bs := md5.Sum([]byte(s))
		return hex.EncodeToString(bs[:])
	},
	"base64": func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	},
	"queryEscape": QueryEscape,
	"hash": func(t, content string) string {
		var h hash.Hash
		switch strings.ToUpper(t) {
		case "MD4":
			h = crypto.MD4.New()
		case "MD5":
			h = crypto.MD5.New()
		case "SHA1":
			h = crypto.SHA1.New()
		case "SHA224":
			h = crypto.SHA224.New()
		case "SHA256":
			h = crypto.SHA256.New()
		case "SHA384":
			h = crypto.SHA384.New()
		case "SHA512":
			h = crypto.SHA512.New()
		case "MD5SHA1":
			h = crypto.MD5SHA1.New()
		case "RIPEMD160":
			h = crypto.RIPEMD160.New()
		case "SHA3_224":
			h = crypto.SHA3_224.New()
		case "SHA3_256":
			h = crypto.SHA3_256.New()
		case "SHA3_384":
			h = crypto.SHA3_384.New()
		case "SHA3_512":
			h = crypto.SHA3_512.New()
		default:
			panic(errors.New("'" + t + "' is unsupported hash."))
		}

		if _, e := io.WriteString(h, content); nil != e {
			panic(e)
		}

		return hex.EncodeToString(h.Sum(nil))
	},
	"encrypt": func(t, pwd, content string) string {
		switch t {
		case "aes_cbc":
			return hex.EncodeToString(aes_cbc_encrypt([]byte(pwd), []byte(content)))
		case "aes_cfb":
			return hex.EncodeToString(aes_cfb_encrypt([]byte(pwd), []byte(content)))
		case "des_cbc":
			return hex.EncodeToString(des_cbc_encrypt([]byte(pwd), []byte(content)))
		case "des_cfb":
			return hex.EncodeToString(des_cfb_encrypt([]byte(pwd), []byte(content)))
		default:
			panic(errors.New("'" + t + "' is unsupported."))
		}
	},

	"toHumableBytes": func(v interface{}) string {
		if s, ok := v.(string); ok {
			if f64, e := strconv.ParseFloat(s, 64); nil == e {
				if f64 >= 0 {
					return human.ToHumanByteString(uint64(f64))
				}
				return "-" + human.ToHumanByteString(uint64(-f64))
			}
		}

		u64, e := as.Uint64(v)
		if nil != e {
			if f64, e := as.Float64(v); nil == e {
				if f64 >= 0 {
					return human.ToHumanByteString(uint64(f64))
				}
				return "-" + human.ToHumanByteString(uint64(-f64))
			}

			return fmt.Sprint(v)
		}
		return human.ToHumanByteString(u64)
	},

	"toBitsFromBytes": func(v interface{}) interface{} {
		u64, e := as.Uint64(v)
		if nil != e {
			if s, ok := v.(string); ok {
				if f64, e := strconv.ParseFloat(s, 64); nil == e {
					return f64 * 8
				}
			}
			return v
		}
		return u64 * 8
	},

	"toBytesFromBits": func(v interface{}) interface{} {
		u64, e := as.Uint64(v)
		if nil != e {
			if s, ok := v.(string); ok {
				if f64, e := strconv.ParseFloat(s, 64); nil == e {
					return f64 / 8
				}
			}
			return v
		}
		return u64 / 8
	},
	"formatFloat": formatFloat,
	"formatTime":  util.TimeFormatWithJavaStyle,
	"formatGoTime": func(t time.Time, layout string) string {
		return t.Format(layout)
	},
	"toInt": toInt,
	"isError": func(v interface{}) bool {
		if v == nil {
			return false
		}
		_, ok := v.(error)
		return ok
	},
	"isApplicationError": func(v interface{}) bool {
		if v == nil {
			return false
		}
		_, ok := v.(*errors.Error)
		return ok
	},
	"keyExists": func(v map[string]interface{}, key string) bool {
		_, ok := v[key]
		return ok
	},
	"keyExist": func(v map[string]interface{}, key string) bool {
		_, ok := v[key]
		return ok
	},
	"charset_encode": func(charset, content string) string {
		encoding := util.GetCharset(charset)
		newContent, _, err := transform.String(encoding.NewEncoder(), content)
		if err != nil {
			return content
		}
		return newContent
	},
}

func toInt(value interface{}) interface{} {
	if nil == value {
		return value
	}
	switch v := value.(type) {
	case []byte:
		v = bytes.Trim(v, "\"")
		i64, err := strconv.ParseInt(string(v), 10, 64)
		if nil == err {
			return i64
		}

		u64, err := strconv.ParseUint(string(v), 10, 64)
		if nil == err {
			return u64
		}
		return value
	case string:
		v = strings.Trim(v, "\"")
		i64, err := strconv.ParseInt(v, 10, 64)
		if nil == err {
			return i64
		}
		u64, err := strconv.ParseUint(v, 10, 64)
		if nil == err {
			return u64
		}
		return value
	case json.Number:
		i64, err := strconv.ParseInt(v.String(), 10, 64)
		if nil == err {
			return i64
		}
		u64, err := strconv.ParseUint(v.String(), 10, 64)
		if nil == err {
			return u64
		}
		return value
	case *json.Number:
		i64, err := strconv.ParseInt(v.String(), 10, 64)
		if nil == err {
			return i64
		}
		u64, err := strconv.ParseUint(v.String(), 10, 64)
		if nil == err {
			return u64
		}
		return value
	case uint:
		return value
	case uint8:
		return value
	case uint16:
		return value
	case uint32:
		return value
	case uint64:
		return value
	case int:
		return value
	case int8:
		return value
	case int16:
		return value
	case int32:
		return value
	case int64:
		return value
	case float32:
		if v < 0 && math.MinInt64 <= v {
			return int64(v)
		}

		if v >= 0 && math.MaxUint64 >= v {
			return uint64(v)
		}
	case float64:
		if v < 0 && math.MinInt64 <= v {
			return int64(v)
		}

		if v >= 0 && math.MaxUint64 >= v {
			return uint64(v)
		}
	}

	if ar, ok := value.([]interface{}); ok {
		for idx, a := range ar {
			ar[idx] = toInt(a)
		}
		return ar
	} else {
		return value
	}
}

func formatFloat(prec int, current_value interface{}) interface{} {
	if f, ok := current_value.(float64); ok {
		return strconv.FormatFloat(f, 'f', prec, 64)
	} else if f, ok := current_value.(float32); ok {
		return strconv.FormatFloat(float64(f), 'f', prec, 64)
	} else if f, ok := current_value.(*json.Number); ok && strings.ContainsRune(f.String(), '.') {
		if f64, e := f.Float64(); nil == e {
			return strconv.FormatFloat(f64, 'f', prec, 64)
		}
	} else if f, ok := current_value.(json.Number); ok && strings.ContainsRune(f.String(), '.') {
		if f64, e := f.Float64(); nil == e {
			return strconv.FormatFloat(f64, 'f', prec, 64)
		}
	}
	if array, ok := current_value.([]interface{}); ok {
		for idx, a := range array {
			array[idx] = formatFloat(prec, a)
		}
		return array
	} else {
		return current_value
	}
}

func aes_cbc_encrypt(pwd, src []byte) []byte {
	if len(src)%aes.BlockSize != 0 {
		src = append(src, bytes.Repeat([]byte{0}, aes.BlockSize-len(src)%aes.BlockSize)...)
	}

	encryptText := make([]byte, aes.BlockSize+len(src))
	iv := encryptText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(pwd)
	if err != nil {
		panic(err)
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encryptText[aes.BlockSize:], src)
	return encryptText
}

func aes_cfb_encrypt(pwd, src []byte) []byte {
	encryptText := make([]byte, aes.BlockSize+len(src))
	iv := encryptText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(pwd)
	if err != nil {
		panic(err)
	}
	mode := cipher.NewCFBEncrypter(block, iv)
	mode.XORKeyStream(encryptText[aes.BlockSize:], src)
	return encryptText
}

func des_cbc_encrypt(pwd, src []byte) []byte {
	if len(src)%des.BlockSize != 0 {
		src = append(src, bytes.Repeat([]byte{0}, des.BlockSize-len(src)%des.BlockSize)...)
	}

	encryptText := make([]byte, des.BlockSize+len(src))
	iv := encryptText[:des.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	block, err := des.NewCipher(pwd)
	if err != nil {
		panic(err)
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encryptText[des.BlockSize:], src)
	return encryptText
}

func des_cfb_encrypt(pwd, src []byte) []byte {
	encryptText := make([]byte, des.BlockSize+len(src))
	iv := encryptText[:des.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	block, err := des.NewCipher(pwd)
	if err != nil {
		panic(err)
	}
	mode := cipher.NewCFBEncrypter(block, iv)
	mode.XORKeyStream(encryptText[des.BlockSize:], src)
	return encryptText
}

func ParseFile(nm string, funcs template.FuncMap) (*template.Template, error) {
	s, e := ioutil.ReadFile(nm)
	if nil != e {
		return nil, e
	}
	return ParseString(filepath.Base(nm), string(s), funcs)
}

func ParseString(name, content string, funcs template.FuncMap) (*template.Template, error) {
	if 0 == len(funcs) {
		return template.New(name).Funcs(TemplateFuncs).Parse(content)
	}
	for k, v := range TemplateFuncs {
		funcs[k] = v
	}
	return template.New(name).Funcs(funcs).Parse(content)
}

func NewTemplate(name string) *template.Template {
	t := template.New(name)
	if len(TemplateFuncs) > 0 {
		t.Funcs(TemplateFuncs)
	}
	return t
}

func RenderText(content string, args interface{}, funcs template.FuncMap) string {
	t, e := ParseString("default", content, funcs)
	if nil != e {
		log.Println("[warn] failed to merge '"+content+"' with ", args, " - ", e)
		return content
	}

	var buffer bytes.Buffer
	e = t.Execute(&buffer, args)
	if nil != e {
		log.Println("[warn] failed to merge '"+content+"' with ", args, " - ", e)
		return content
	}
	return buffer.String()
}
