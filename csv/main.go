package csv

import (
	"encoding/csv"
	"io"
	"strings"
)

func ReadCSV(in io.Reader, alias map[string]string, cb func(line int, header, values []string) error) error {
	r := csv.NewReader(in)
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
	r := csv.NewReader(in)
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
