package util

import "testing"

func TestSplitPGVersion(t *testing.T) {
	assert := func(s string, ma, mi, pa int, arch string) {
		major, minor, patch, ar, err := splitVersion(s)

		if err != nil {
			t.Error("["+s+"]", err)
			return
		}
		if ma != major {
			t.Error("["+s+"]", ma, major)
		}
		if mi != minor {
			t.Error("["+s+"]", mi, minor)
		}
		if pa != patch {
			t.Error("["+s+"]", pa, patch)
		}
		if ar != arch {
			t.Error("["+s+"]", ar, arch)
		}
	}

	assert("9.3.1_x86", 9, 3, 1, "x86")
	assert("9.3.1_x64", 9, 3, 1, "x64")

	if PGArchEqual("9.3.1_x86", "9.3.1_x64") {
		t.Error("arch is error")
	}
}
