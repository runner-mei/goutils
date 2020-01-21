package util

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type TimeInDay struct {
	Hour, Minute int
}

func ParseTimeInDay(s string) (TimeInDay, error) {
	if "" == s {
		return TimeInDay{}, errors.New("'" + s + "' isn't a valid TimeInDay.")
	}
	ss := strings.Split(s, ":")
	if 2 != len(ss) {
		return TimeInDay{}, errors.New("'" + s + "' isn't a valid TimeInDay.")
	}
	hour, e := strconv.ParseInt(ss[0], 10, 0)
	if nil != e {
		return TimeInDay{}, errors.New("'" + s + "' isn't a valid TimeInDay.")
	}
	minute, e := strconv.ParseInt(ss[1], 10, 0)
	if nil != e {
		return TimeInDay{}, errors.New("'" + s + "' isn't a valid TimeInDay.")
	}
	return TimeInDay{Hour: int(hour), Minute: int(minute)}, nil
}

type RangeInDay struct {
	BeginAt, EndAt TimeInDay
}

func (rn *RangeInDay) In(hour, min int) bool {
	if hour > rn.BeginAt.Hour || (hour == rn.BeginAt.Hour && min >= rn.BeginAt.Minute) {
		if hour < rn.EndAt.Hour || (hour == rn.EndAt.Hour && min < rn.EndAt.Minute) {
			return true
		}
	}
	return false
}

func ParseRangeInDay(s string, rn *RangeInDay) error {
	ss := strings.Split(s, "-")
	if 2 != len(ss) {
		return errors.New("'" + s + "' isn't a valid RangeInDay.")
	}
	beginAt, e := ParseTimeInDay(ss[0])
	if nil != e {
		return errors.New("'" + s + "' isn't a valid RangeInDay.")
	}
	endAt, e := ParseTimeInDay(ss[1])
	if nil != e {
		return errors.New("'" + s + "' isn't a valid RangeInDay.")
	}

	if beginAt.Hour > endAt.Hour || (beginAt.Hour == endAt.Hour && beginAt.Minute >= endAt.Minute) {
		return errors.New("'" + s + "' isn't a valid RangeInDay.")
	}

	rn.BeginAt = beginAt
	rn.EndAt = endAt
	return nil
}

func ParseRangesInDay(s string) ([]RangeInDay, error) {
	ss := strings.Split(s, ",")
	res := make([]RangeInDay, len(ss))
	offset := 0
	for _, str := range ss {
		if "" == str {
			continue
		}

		if e := ParseRangeInDay(str, &res[offset]); nil != e {
			return nil, errors.New("'" + s + "' isn't a valid []RangeInDay.")
		}

		offset++
	}

	if 0 == offset {
		return nil, nil
	}

	return res[:offset], nil
}

func InTimeRangesInDay(t time.Time, ranges []RangeInDay) bool {
	if 0 == len(ranges) {
		return false
	}
	hour, min, _ := t.Clock()
	for _, rn := range ranges {
		if rn.In(hour, min) {
			return true
		}
	}
	return false
}
