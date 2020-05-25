package generic

import (
	"sync/atomic"
	"time"

	"github.com/cheekybits/genny/generic"
)

// ValueType 用于泛型替换的类型
type ValueType generic.Type

type CachedValue struct {
	MaxAge int64
	data   atomic.Value
}

type cachedData struct {
	value     ValueType
	timestamp int64
}

func (cv *CachedValue) Get() ValueType {
	var v ValueType
	return cv.Read(v, cv.MaxAge)
}

func (cv *CachedValue) GetWithDefault(v ValueType) ValueType {
	return cv.Read(v, cv.MaxAge)
}

func (cv *CachedValue) Read(v ValueType, maxAge int64) ValueType {
	o := cv.data.Load()
	if o == nil {
		return v
	}
	cdata, ok := o.(*cachedData)
	if !ok {
		return v
	}
	if (cdata.timestamp + maxAge) < time.Now().Unix() {
		return v
	}
	return cdata.value
}

func (cv *CachedValue) Set(v ValueType, t time.Time) {
	cv.data.Store(&cachedData{
		value:     v,
		timestamp: t.Unix(),
	})
}
