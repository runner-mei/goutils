package netutil

import (
	"testing"
)

func assertPH(t *testing.T, addr string, ok bool) {
	r := IsValidPhysical(addr)
	if r == ok {
		return
	}
	if ok {
		t.Error("except '" + addr + "' is valid, actual is invalid.")
	} else {
		t.Error("except '" + addr + "' is invalid, actual is valid.")
	}
}

func TestIsValidPhysicalAddress(t *testing.T) {
	assertPH(t, "00:00:00:00:00:00", false)
	assertPH(t, "00:00:00:00:00:01", false)
	assertPH(t, "FF:FF:FF:FF:FF:FF", false)
	assertPH(t, "00:AB:AB:AB:AB:AB", false)
	assertPH(t, "12:34:56:78:9A:BC", false)
	assertPH(t, "00:00:5E:00:00:00", false)
	assertPH(t, "00:00:5E:FF:FF:FF", false)
	assertPH(t, "00:00:00:FF:FF:FF", false)
	assertPH(t, "CC:CC:CC:CC:CC:CC", false)
	assertPH(t, "01:00:00:FF:FF:F3", false)
	assertPH(t, "33:00:00:FF:FF:F3", false)
	assertPH(t, "37:00:00:FF:FF:F3", false)
	assertPH(t, "C1:00:00:FF:FF:F3", false)
	assertPH(t, "C3:00:00:FF:FF:F3", false)
	assertPH(t, "00:10:00:FF:FF:F3", true)
	assertPH(t, "01:10:00:FF:FF:F3", false)
	assertPH(t, "03:10:00:FF:FF:F3", false)
	assertPH(t, "33:10:00:FF:FF:F3", false)
	assertPH(t, "13:10:00:FF:FF:F3", false)
}
