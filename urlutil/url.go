package urlutil

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
)

// SplitURLPath 分隔 url path, 取出 url path 的第一部份
func SplitURLPath(pa string) (string, string) {
	return Split(pa)
}

// JoinURLPath 拼接 url
func JoinURLPath(paths ...string) string {
	return Join(paths...)
}

// JoinURLPathWith 拼接 url
func JoinURLPathWith(base string, paths []string) string {
	return JoinWith(base, paths)
}

// Split 分隔 url path, 取出 url path 的第一部份
func Split(pa string) (string, string) {
	if "" == pa {
		return "", ""
	}

	if '/' == pa[0] {
		pa = pa[1:]
	}

	idx := strings.IndexRune(pa, '/')
	if -1 == idx {
		return pa, ""
	}
	return pa[:idx], pa[idx:]
}

// Join 拼接 url
func Join(paths ...string) string {
	switch len(paths) {
	case 0:
		return ""
	case 1:
		return paths[0]
	default:
		return JoinWith(paths[0], paths[1:])
	}
}

// JoinWith 拼接 url
func JoinWith(base string, paths []string) string {
	var buf bytes.Buffer
	buf.WriteString(base)

	lastSplash := strings.HasSuffix(base, "/")
	for _, pa := range paths {
		if 0 == len(pa) {
			continue
		}

		if lastSplash {
			if '/' == pa[0] {
				buf.WriteString(pa[1:])
			} else {
				buf.WriteString(pa)
			}
		} else {
			if '/' != pa[0] {
				buf.WriteString("/")
			}
			buf.WriteString(pa)
		}

		lastSplash = strings.HasSuffix(pa, "/")
	}
	return buf.String()
}

// URLBuilder 创建 url 的小工具
type URLBuilder struct {
	bytes.Buffer
	hasQuest  bool
	hasParams bool
}

// NewURLBuilder 创建 url builder
func NewURLBuilder(base string) *URLBuilder {
	builder := &URLBuilder{hasQuest: false, hasParams: false}
	if 0 < len(base) && '/' == base[len(base)-1] {
		builder.WriteString(base[:len(base)-1])
	} else {
		builder.WriteString(base)
	}
	if strings.ContainsRune(base, '?') {
		builder.hasQuest = true
	}
	if !strings.HasSuffix(base, "?") {
		builder.hasParams = true
	}
	return builder
}

func (self *URLBuilder) Clone() *URLBuilder {
	url := &URLBuilder{
		hasQuest:  self.hasQuest,
		hasParams: self.hasParams,
	}
	url.Buffer.Write(self.Buffer.Bytes())
	return url
}

func (self *URLBuilder) Concat(paths ...string) *URLBuilder {
	if self.hasQuest {
		panic("[panic] don`t append path to the query")
	}

	for _, pa := range paths {
		if 0 == len(pa) {
			continue
		}

		if '/' != pa[0] {
			self.WriteString("/")
		}

		if '/' == pa[len(pa)-1] {
			self.WriteString(pa[:len(pa)-1])
		} else {
			self.WriteString(pa)
		}
	}
	return self
}

func (self *URLBuilder) closePath() *URLBuilder {
	if !self.hasQuest {
		self.WriteString("?")
		self.hasQuest = true
	} else if self.hasParams {
		self.WriteString("&")
	} else {
		self.hasParams = true
	}
	return self
}

func (self *URLBuilder) WithQuery(key, value string) *URLBuilder {
	if 0 == len(key) {
		return self
	}
	self.closePath()

	self.WriteString(key)
	self.WriteString("=")
	self.WriteString(url.QueryEscape(value))
	return self
}

func (self *URLBuilder) WithQueries(params map[string]string, prefix string) *URLBuilder {
	if 0 == len(params) {
		return self
	}
	self.closePath()

	for k, v := range params {
		self.WriteString(prefix)
		self.WriteString(k)
		self.WriteString("=")
		self.WriteString(url.QueryEscape(v))
		self.WriteString("&")
	}
	self.Truncate(self.Len() - 1)
	return self
}

func (self *URLBuilder) WithAnyQueries(params map[string]interface{}, prefix string) *URLBuilder {
	if 0 == len(params) {
		return self
	}
	self.closePath()

	for k, v := range params {
		self.WriteString(prefix)
		self.WriteString(k)
		self.WriteString("=")
		if s, ok := v.(string); ok {
			self.WriteString(url.QueryEscape(s))
		} else {
			self.WriteString(url.QueryEscape(fmt.Sprint(v)))
		}
		self.WriteString("&")
	}
	self.Truncate(self.Len() - 1)
	return self
}

func (self *URLBuilder) ToUrl() string {
	return self.String()
}

func (self *URLBuilder) Build() string {
	return self.String()
}
