package util

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/runner-mei/errors"
)

var ErrStopped = errors.ErrStopped

func ReadLines(filename string) ([][]byte, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return SplitLines(bs), nil
}

func ReadStringLines(filename string, ignoreEmpty bool) ([]string, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := SplitLines(bs)
	ss := make([]string, 0, len(lines))
	for idx := range lines {
		if ignoreEmpty {
			if len(lines[idx]) == 0 {
				continue
			}
		}

		ss = append(ss, string(lines[idx]))
	}
	return ss, nil
}

func ReadEachLines(filename string, cb func(int, []byte) error) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer CloseWith(f)

	count := 0
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		count++
		err := cb(count, scan.Bytes())
		if err != nil {
			if errors.IsStopped(err) {
				return nil
			}
			return err
		}
	}

	return scan.Err()
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func CopyFile(src, dst string) (err error) {
	root := filepath.Dir(dst)
	if err := os.MkdirAll(root, 0777); err != nil {
		return err
	}
	var in, out *os.File

	if in, err = os.Open(src); err != nil {
		return err
	}
	defer in.Close()

	if out, err = os.Create(dst); err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func FileAppend(filename string, content []byte) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(content)
	return err
}

// FileExists 文件是否存在
func FileExists(dir string, e ...*error) bool {
	info, err := os.Stat(dir)
	if err != nil {
		if len(e) != 0 {
			*e[0] = err
		}
		return false
	}

	return !info.IsDir()
}

// DirExists 目录是否存在
func DirExists(dir string, err ...*error) bool {
	d, e := os.Stat(dir)
	switch {
	case e != nil:
		if len(err) != 0 {
			*err[0] = e
		}
		return false
	case !d.IsDir():
		return false
	}

	return true
}

func IsDirectory(dir string, e ...*error) bool {
	info, err := os.Stat(dir)
	if err != nil {
		if len(e) != 0 {
			*e[0] = err
		}
		return false
	}

	return info.IsDir()
}

func EnumerateFiles(pa string) ([]string, error) {
	if "" == pa {
		return nil, errors.New("path is empty.")
	}

	dir, serr := os.Stat(pa)
	if serr != nil {
		return nil, serr
	}

	if !dir.IsDir() {
		return nil, errors.New(pa + " is not a directory")
	}

	fd, err := os.Open(pa)
	if nil != err {
		return nil, err
	}
	defer fd.Close()

	paths := make([]string, 0, 30)
	for {
		dirs, err := fd.Readdir(10)
		if nil != err {
			if io.EOF == err {
				return paths, nil
			}
			return nil, err
		}
		for _, dir := range dirs {
			if dir.IsDir() {
				subPaths, err := EnumerateFiles(path.Join(pa, dir.Name()))
				if nil != err {
					return nil, err
				}
				for _, sp := range subPaths {
					paths = append(paths, sp)
				}
			} else {
				paths = append(paths, path.Join(pa, dir.Name()))
			}
		}
	}
}
