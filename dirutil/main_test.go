package dirutil

import "testing"

func TestDiff(t *testing.T) {
	res, e := Diff("ftp://mfk:mfk@192.168.1.141:2221/unittests/a", "ftp://mfk:mfk@192.168.1.141:2221/unittests/b")
	if nil != e {
		t.Error(e)
		return
	}
	t.Log(res)

	if 5 != len(res) {
		t.Error("size is error - ", len(res))
	}

	assert := func(name, result string) {
		for _, r := range res {
			if r.Filename == name {
				if r.Result != result {
					t.Error(name, "result isn't excepted, ", r.Result, "-", result)
				}
				return
			}
		}

		t.Error(name, "is not found")
	}

	assert("a.txt", "leftOnly")
	assert("b.txt", "different")
	assert("c.txt", "")
	assert("d.txt", "")
	assert("e.txt", "rightOnly")
}
