package csv

import (
	"bytes"
	"encoding/csv"
	"io"
	"strings"
)

func RemoveBOM(in io.Reader) (io.Reader, error) {
	var magic [3]byte
	n, err := io.ReadFull(in, magic[:])
	if err != nil {
		return nil, err
	}
	if n < 3 {
		return nil, io.ErrUnexpectedEOF
	}
	if bytes.HasPrefix(magic[:], []byte{0xEF, 0xBB, 0xBF}) {
		return in, nil
	}
	return io.MultiReader(bytes.NewReader(magic[:]), in), nil
}

func ReadCSV(in io.Reader, alias map[string]string, cb func(line int, header, values []string) error) error {
	nr, err := RemoveBOM(in)
	if err != nil {
		return err
	}

	r := csv.NewReader(nr)
	header, err := r.Read()
	if err != nil {
		return err
	}
	for idx := range header {
		header[idx] = strings.TrimSpace(header[idx])
	}

	if len(alias) > 0 {
		for idx := range header {
			name := header[idx]
			a, ok := alias[name]
			if ok {
				header[idx] = a
			}
		}
	}

	return readCSV(r, header, 1, cb)
}

func ReadCSVWithNoHead(in io.Reader, header []string, cb func(line int, header, values []string) error) error {
	nr, err := RemoveBOM(in)
	if err != nil {
		return err
	}

	r := csv.NewReader(nr)

	return readCSV(r, header, 0, cb)
}

func readCSV(r *csv.Reader, header []string, line int, cb func(line int, header, values []string) error) error {
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		line++

		for idx := range record {
			record[idx] = strings.TrimSpace(record[idx])
		}
		err = cb(line, header, record)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return nil
}
