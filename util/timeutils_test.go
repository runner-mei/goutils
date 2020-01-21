package util

import (
	"strings"
	"testing"
	"time"
)

func TestParseRangesInDay(t *testing.T) {
	assert := func(s string, hour1, min1, hour2, min2 int) {
		ranges, e := ParseRangesInDay(s)
		if nil != e {
			t.Error("'"+s+"' - ", e)
			return
		}
		if 1 != len(ranges) {
			t.Error("'" + s + "' is not one")
			return
		}

		if hour1 != ranges[0].BeginAt.Hour {
			t.Error("'"+s+"' - exceptd hour1 is", hour1, ", actual is", ranges[0].BeginAt.Hour)
		}

		if min1 != ranges[0].BeginAt.Minute {
			t.Error("'"+s+"' - exceptd min1 is", min1, ", actual is", ranges[0].BeginAt.Minute)
		}

		if hour2 != ranges[0].EndAt.Hour {
			t.Error("'"+s+"' - exceptd hour2 is", hour2, ", actual is", ranges[0].EndAt.Hour)
		}

		if min2 != ranges[0].EndAt.Minute {
			t.Error("'"+s+"' - exceptd min2 is", min2, ", actual is", ranges[0].EndAt.Minute)
		}
	}

	assert2 := func(s string, count, idx, hour1, min1, hour2, min2 int) {
		ranges, e := ParseRangesInDay(s)
		if nil != e {
			t.Error("'"+s+"' - ", e)
			return
		}
		if count != len(ranges) {
			t.Error("'" + s + "' is not one")
			return
		}

		if hour1 != ranges[idx].BeginAt.Hour {
			t.Error("'"+s+"' - exceptd hour1 is", hour1, ", actual is", ranges[idx].BeginAt.Hour)
		}

		if min1 != ranges[idx].BeginAt.Minute {
			t.Error("'"+s+"' - exceptd min1 is", min1, ", actual is", ranges[idx].BeginAt.Minute)
		}

		if hour2 != ranges[idx].EndAt.Hour {
			t.Error("'"+s+"' - exceptd hour2 is", hour2, ", actual is", ranges[idx].EndAt.Hour)
		}

		if min2 != ranges[idx].EndAt.Minute {
			t.Error("'"+s+"' - exceptd min2 is", min2, ", actual is", ranges[idx].EndAt.Minute)
		}
	}

	assert("16:05-16:10", 16, 05, 16, 10)
	assert("12:23-12:34", 12, 23, 12, 34)
	assert("12:23-13:34", 12, 23, 13, 34)
	assert("12:23-13:14", 12, 23, 13, 14)
	assert("12:23-13:14,", 12, 23, 13, 14)
	assert(",12:23-13:14", 12, 23, 13, 14)
	assert(",12:23-13:14,,", 12, 23, 13, 14)

	assert2(",12:23-13:14,14:23-15:14,", 2, 0, 12, 23, 13, 14)
	assert2(",12:23-13:14,14:23-15:14,", 2, 1, 14, 23, 15, 14)
	assert2(",12:23-13:14,14:23-15:14,,15:23-16:14", 3, 2, 15, 23, 16, 14)

	assert_error := func(s string, err string) {
		_, e := ParseRangesInDay(s)
		if nil == e {
			t.Error("'" + s + "' - error is nil")
			return
		}

		if !strings.Contains(e.Error(), err) {
			t.Error("'"+s+"' - exceptd error contains", err, ", actual is", e)
		}
	}

	assert_error("12:23", "12:23")
	assert_error("12:23-", "12:23-")
	assert_error("12:23-12", "12:23-")
	assert_error("12:23-12:", "12:23-")
	assert_error("12:23-12:12", "12:23-")
}

func TestInTimeRangesInDay(t *testing.T) {
	assert := func(s string, hour, min, sec int, result bool) {
		ranges, e := ParseRangesInDay(s)
		if nil != e {
			t.Error("'"+s+"' - ", e)
			return
		}

		if result != InTimeRangesInDay(time.Date(0, time.Month(1), 1, hour, min, sec, 0, time.Local), ranges) {
			t.Error("'"+s+"' not match ", hour, ":", min, ":", sec)
		}
	}

	assert("12:23-12:34", 12, 23, 12, true)
	assert("12:23-13:34", 13, 23, 12, true)
	assert("00:12-02:12,12:23-13:34", 13, 23, 12, true)

	assert("12:23-13:34", 13, 34, 12, false)
	assert("12:23-12:34", 12, 22, 12, false)
	assert("12:23-12:34", 11, 32, 12, false)
}
