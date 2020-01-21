package util

import (
	"testing"
	"time"
)

func TestReplaceTime(t *testing.T) {
	now := time.Date(2014, time.July, 21, 4, 12, 2, 1, time.Local)

	for idx, test := range []struct {
		Format, Excepted string
	}{
		{
			Format:   "asfasdf{yyyy-MMM-dd HH:mm:ss}aa{yyyyMMdd}",
			Excepted: "asfasdf2014-Jul-21 04:12:02aa20140721",
		},

		{
			Format:   "asfasdf{now()-24h|yyyy-MMM-dd HH:mm:ss}aa{yyyyMMdd}",
			Excepted: "asfasdf2014-Jul-20 04:12:02aa20140721",
		},

		{
			Format:   "asfasdf",
			Excepted: "asfasdf",
		},
	} {
		s := string(ReplaceTime(now, []byte(test.Format)))
		if test.Excepted != s {
			t.Error(idx, "excepted is", test.Excepted)
			t.Error(idx, "actual   is", s)
		} else {
			t.Log(s)
		}

		s = ReplaceTimeString(now, test.Format)
		if test.Excepted != s {
			t.Error(idx, "excepted is", test.Excepted)
			t.Error(idx, "actual   is", s)
		} else {
			t.Log(s)
		}
	}
}
