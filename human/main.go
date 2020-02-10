package human

import (
	"fmt"
	"strings"
)

const (
	HUMAN_BYTE     = 1.0
	HUMAN_KILOBYTE = 1024 * HUMAN_BYTE
	HUMAN_MEGABYTE = 1024 * HUMAN_KILOBYTE
	HUMAN_GIGABYTE = 1024 * HUMAN_MEGABYTE
	HUMAN_TERABYTE = 1024 * HUMAN_GIGABYTE
)

func ToHumanByteString(bytes uint64) string {
	unit := ""
	value := float32(bytes)

	switch {
	case bytes >= HUMAN_TERABYTE:
		unit = "T"
		value = value / HUMAN_TERABYTE
	case bytes >= HUMAN_GIGABYTE:
		unit = "G"
		value = value / HUMAN_GIGABYTE
	case bytes >= HUMAN_MEGABYTE:
		unit = "M"
		value = value / HUMAN_MEGABYTE
	case bytes >= HUMAN_KILOBYTE:
		unit = "K"
		value = value / HUMAN_KILOBYTE
	case bytes >= HUMAN_BYTE:
		unit = "B"
	case bytes == 0:
		return "0"
	}

	return strings.TrimSuffix(fmt.Sprintf("%.2f", value), ".00") + unit
}
