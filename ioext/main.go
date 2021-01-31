package ioext

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func ReadLine(r io.Reader) ([]byte, error) {
	var line = make([]byte, 0, 64)

	for i := 0; ; i++ {
		if i > 8*1024 {
			return nil, errors.New("read too much")
		}

		var c [1]byte
		n, err := r.Read(c[:])
		if err != nil {
			fmt.Println("readline:", len(line), "'"+string(line)+"'")
			return nil, err
		}

		if n == 0 {
			continue
		}

		if c[0] == '\n' {
			if len(line) > 0 && line[len(line)-1] == '\r' {
				return line[:len(line)-1], nil
			}
			return line, nil
		}
		line = append(line, c[0])
	}
	return line, nil
}

func WriteFull(w io.Writer, bs []byte) error {
	for len(bs) > 0 {
		n, e := w.Write(bs)
		if nil != e {
			return e
		}
		bs = bs[n:]
	}
	return nil
}

// CopyFile the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
