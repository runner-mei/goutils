package split

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

// InplaceReader a inplace reader for bufio.Scanner
type InplaceReader int

func (p *InplaceReader) Read([]byte) (int, error) {
	if *p == 0 {
		return 0, io.EOF
	}
	ret := int(*p)
	*p = 0
	return ret, io.EOF
}

func Strings(bs []byte) ([]string, error) {
	if len(bs) == 0 {
		return nil, nil
	}

	var ipList []string
	if err := json.Unmarshal(bs, &ipList); err == nil {
		return ipList, nil
	}

	r := InplaceReader(len(bs))
	scanner := bufio.NewScanner(&r)
	scanner.Buffer(bs, len(bs))

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		for _, field := range bytes.Split(line, []byte(",")) {
			if len(field) == 0 {
				continue
			}
			field = bytes.TrimSpace(field)
			if len(field) == 0 {
				continue
			}
			ipList = append(ipList, string(field))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ipList, nil
}

func Lines(bs []byte, ignoreEmpty, trimEmpty bool) [][]byte {
	if len(bs) == 0 {
		return nil
	}

	r := InplaceReader(len(bs))
	scanner := bufio.NewScanner(&r)
	scanner.Buffer(bs, len(bs))

	lines := make([][]byte, 0, 10)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) > 0 {
			if line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}
		}
		if len(line) > 0 {
			if line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
		}
		if trimEmpty {
			line = bytes.TrimSpace(line)
		}
		if ignoreEmpty && len(line) == 0 {
			continue
		}

		lines = append(lines, line)
	}

	if nil != scanner.Err() {
		panic(scanner.Err())
	}
	return lines
}

func StringLines(bs []byte, ignoreEmpty, trimEmpty bool) []string {
	if len(bs) == 0 {
		return nil
	}
	lines := Lines(bs, ignoreEmpty, trimEmpty)
	ss := make([]string, 0, len(lines))
	for idx := range lines {
		ss = append(ss, string(lines[idx]))
	}
	return ss
}

func Split(s, sep string, ignoreEmpty, trimSpace bool) []string {
	if len(s) == 0 {
		return nil
	}

	lines := strings.Split(s, sep)
	if !ignoreEmpty && !trimSpace {
		return lines
	}

	offset := 0
	for idx := range lines {
		if trimSpace {
			lines[idx] = strings.TrimSpace(lines[idx])
		}

		if ignoreEmpty && len(lines[idx]) == 0 {
			continue
		}
		if offset != idx {
			lines[offset] = lines[idx]
		}
		offset++
	}
	return lines[:offset]
}
