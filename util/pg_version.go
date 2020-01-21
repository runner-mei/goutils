package util

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func splitVersion(v string) (major, minor, patch int, arch string, err error) {
	_, err = fmt.Sscanf(v, "%d.%d.%d_%s", &major, &minor, &patch, &arch)
	return
}

func PGVersionEqual(v1, v2 string) bool {
	if v1 == v2 {
		return true
	}

	if v1 == "9.3" {
		return strings.HasPrefix(v2, "9.3.")
	}

	if v2 == "9.3" {
		return strings.HasPrefix(v1, "9.3.")
	}

	major1, minor1, _, arch1, err := splitVersion(v1)
	if err != nil {
		return false
	}
	major2, minor2, _, arch2, err := splitVersion(v2)
	if err != nil {
		return false
	}

	return major1 == major2 && minor1 == minor2 && arch1 == arch2
}

func PGVersionRead(path string) string {
	if FileExists(path) {
		s, e := ioutil.ReadFile(path)
		if nil != e {
			panic(e)
		}
		return string(s)
	}

	return "9.3"
}

func PGArchEqual(v1, v2 string) bool {
	if v1 == v2 {
		return true
	}

	if v1 == "9.3" {
		return strings.HasPrefix(v2, "9.3.")
	}

	if v2 == "9.3" {
		return strings.HasPrefix(v1, "9.3.")
	}

	_, _, _, arch1, err := splitVersion(v1)
	if err != nil {
		return false
	}
	_, _, _, arch2, err := splitVersion(v2)
	if err != nil {
		return false
	}

	return arch1 == arch2
}
