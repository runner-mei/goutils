package cfg

import (
	"strings"
	"time"

	"github.com/runner-mei/goutils/as"
)

// Config 配置
type Config struct {
	settings map[string]interface{}
}

// PasswordWithDefault 读配置
func (self *Config) PasswordWithDefault(key, defValue string) string {
	if s, ok := self.settings[key]; ok {
		return as.StringWithDefault(s, defValue)
	}
	return defValue
}

// StringWithDefault 读配置
func (self *Config) StringWithDefault(key, defValue string) string {
	if s, ok := self.settings[key]; ok {
		return as.StringWithDefault(s, defValue)
	}
	return defValue
}

// StringsWithDefault 读配置
func (self *Config) StringsWithDefault(key string, defValue []string) []string {
	if o, ok := self.settings[key]; ok {
		if s, ok := o.(string); ok {
			return strings.Split(s, ",")
		}
		return as.StringsWithDefault(o, defValue)
	}
	return defValue
}

// UintWithDefault 读配置
func (self *Config) UintWithDefault(key string, defValue uint) uint {
	if s, ok := self.settings[key]; ok {
		return as.UintWithDefault(s, defValue)
	}
	return defValue
}

// Uint64WithDefault 读配置
func (self *Config) Uint64WithDefault(key string, defValue uint64) uint64 {
	if s, ok := self.settings[key]; ok {
		return as.Uint64WithDefault(s, defValue)
	}
	return defValue
}

// IntWithDefault 读配置
func (self *Config) IntWithDefault(key string, defValue int) int {
	if s, ok := self.settings[key]; ok {
		return as.IntWithDefault(s, defValue)
	}
	return defValue
}

// Int64WithDefault 读配置
func (self *Config) Int64WithDefault(key string, defValue int64) int64 {
	if s, ok := self.settings[key]; ok {
		return as.Int64WithDefault(s, defValue)
	}
	return defValue
}

// BoolWithDefault 读配置
func (self *Config) BoolWithDefault(key string, defValue bool) bool {
	if s, ok := self.settings[key]; ok {
		return as.BoolWithDefault(s, defValue)
	}
	return defValue
}

// DurationWithDefault 读配置
func (self *Config) DurationWithDefault(key string, defValue time.Duration) time.Duration {
	if s, ok := self.settings[key]; ok {
		return as.DurationWithDefault(s, defValue)
	}
	return defValue
}

// Has 是否有指定的配置
func (self *Config) Has(key string) bool {
	_, ok := self.settings[key]
	return ok
}

// Set 写配置
func (self *Config) Set(key string, value interface{}) {
	self.settings[key] = value
}

// Set 写配置
func (self *Config) SetDefault(key string, value interface{}) {
	if _, ok := self.settings[key]; !ok {
		self.settings[key] = value
	}
}

// Unset 写配置
func (self *Config) Unset(key string) interface{} {
	old := self.settings[key]
	delete(self.settings, key)
	return old
}

// Get 读配置
func (self *Config) Get(key string, subKeys ...string) interface{} {
	o := self.settings[key]
	if len(subKeys) == 0 {
		return o
	}

	if o == nil {
		return nil
	}

	for _, subKey := range subKeys {
		m, ok := o.(map[string]interface{})
		if !ok {
			return nil
		}
		o = m[subKey]
		if o == nil {
			return nil
		}
	}
	return o
}

// GetAsString 读配置
func (self *Config) GetAsString(keys []string, defaultValue string) string {
	o := self.Get(keys[0], keys[1:]...)
	return as.StringWithDefault(o, defaultValue)
}

// GetAsInt 读配置
func (self *Config) GetAsInt(keys []string, defaultValue int) int {
	o := self.Get(keys[0], keys[1:]...)
	return as.IntWithDefault(o, defaultValue)
}

// GetAsBool 读配置
func (self *Config) GetAsBool(keys []string, defaultValue bool) bool {
	o := self.Get(keys[0], keys[1:]...)
	return as.BoolWithDefault(o, defaultValue)
}

// GetAsDuration 读配置
func (self *Config) GetAsDuration(keys []string, defaultValue time.Duration) time.Duration {
	o := self.Get(keys[0], keys[1:]...)
	return as.DurationWithDefault(o, defaultValue)
}

// GetAsTime 读配置
func (self *Config) GetAsTime(keys []string, defaultValue time.Time) time.Time {
	o := self.Get(keys[0], keys[1:]...)
	return as.TimeWithDefault(o, defaultValue)
}

// ForEach 读配置
func (self *Config) ForEach(cb func(key string, value interface{})) {
	for k, v := range self.settings {
		cb(k, v)
	}
}

// ForEach 读配置
func (self *Config) ForEachWithPrefix(prefix string, cb func(key string, value interface{})) {
	for k, v := range self.settings {
		if strings.HasPrefix(k, prefix) {
			cb(k, v)
		}
	}
}

func NewConfig(settings map[string]interface{}) *Config {
	return &Config{settings: settings}
}
