package util

import (
	"bytes"
	"strings"
	"sync/atomic"
	"time"
)

type UnixTime int64

func (t *UnixTime) UnixNano() int64 {
	return int64(*t)
}

func (t *UnixTime) Unix() int64 {
	return int64(*t) / int64(time.Second)
}

func (t *UnixTime) Set(now time.Time) {
	*(*int64)(t) = now.UnixNano()
}

func (t *UnixTime) AtomicSet(now time.Time) {
	atomic.StoreInt64((*int64)(t), now.UnixNano())
}

func (t *UnixTime) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *UnixTime) Format(f string) string {
	return time.Unix(int64(*t)/int64(time.Second), int64(*t)%int64(time.Second)).Format(f)
}

func (t *UnixTime) String() string {
	return t.Format(time.RFC3339Nano)
}

func ToUnixTime(now time.Time) UnixTime {
	var t UnixTime
	t.Set(now)
	return t
}

func TimeFormatWithJavaStyle(t time.Time, layout string) string {
	layout = strings.Replace(layout, "yyyy", "2006", -1)
	layout = strings.Replace(layout, "MMMM", "January", -1)
	layout = strings.Replace(layout, "MMM", "Jan", -1)
	layout = strings.Replace(layout, "MM", "01", -1)
	layout = strings.Replace(layout, "dd", "02", -1)
	layout = strings.Replace(layout, "HH", "15", -1)
	layout = strings.Replace(layout, "mm", "04", -1)
	layout = strings.Replace(layout, "ss", "05", -1)
	layout = strings.Replace(layout, "z", "MST", -1)
	layout = strings.Replace(layout, "Z", "-0700", -1)
	layout = strings.Replace(layout, "'T'", "T", -1)

	layout = strings.Replace(layout, "EEE", "Mon", -1)
	return t.Format(layout)
}

func ReplaceTimeString(current time.Time, s string) string {
	idx := strings.IndexByte(s, '{')
	if idx < 0 {
		return s
	}
	idx = strings.IndexByte(s[idx:], '}')
	if idx < 0 {
		return s
	}

	bs := ReplaceTime(current, []byte(s))
	return string(bs)
}

func ReplaceTime(current time.Time, bs []byte) []byte {
	var buf bytes.Buffer
	for {
		now := current
		start := bytes.IndexByte(bs, '{')
		if start < 0 {
			if buf.Len() == 0 {
				return bs
			}
			buf.Write(bs)
			return buf.Bytes()
		}

		end := bytes.IndexByte(bs[start:], '}')
		if end < 0 {
			if buf.Len() == 0 {
				return bs
			}
			buf.Write(bs)
			return buf.Bytes()
		}

		buf.Write(bs[:start])

		timeFormat := bs[start+1 : start+end]
		//fmt.Println(start+1, end, len(bs), string(bs[start+1:start+end]))
		if bytes.Contains(timeFormat, []byte("now()")) {
			bb := bytes.SplitN(timeFormat, []byte("|"), 2)
			if len(bb) == 2 {
				modifier := bytes.TrimPrefix(bytes.TrimSpace(bb[0]), []byte("now()"))
				modifier = bytes.TrimSpace(modifier)
				isMinus := false
				if bytes.HasPrefix(modifier, []byte("-")) {
					isMinus = true
					modifier = bytes.TrimPrefix(modifier, []byte("-"))
				} else if bytes.HasPrefix(modifier, []byte("+")) {
					modifier = bytes.TrimPrefix(modifier, []byte("+"))
				}

				interval, err := time.ParseDuration(string(modifier))
				if err == nil {
					if isMinus {
						now = now.Add(-1 * interval)
					} else {
						now = now.Add(interval)
					}
				}

				timeFormat = bb[1]
			}
		}
		buf.WriteString(TimeFormatWithJavaStyle(now, string(timeFormat)))
		bs = bs[start+end+1:]
	}
}
