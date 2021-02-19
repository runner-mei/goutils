package util

import (
	"github.com/runner-mei/errors"
	"github.com/runner-mei/goutils/ioext"
)

var ErrStopped = errors.ErrStopped

func ReadLines(filename string) ([][]byte, error) {
	return ioext.ReadLines(filename)
}

func ReadStringLines(filename string, ignoreEmpty bool) ([]string, error) {
	return ioext.ReadStringLines(filename, ignoreEmpty)
}

func ReadEachLines(filename string, cb func(int, []byte) error) error {
	return ioext.ReadEachLines(filename, cb)
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func CopyFile(src, dst string) (err error) {
	return ioext.CopyFile(src, dst)
}

func FileAppend(filename string, content []byte) error {
	return ioext.FileAppend(filename, content)
}

// FileExists 文件是否存在
func FileExists(dir string, e ...*error) bool {
	return ioext.FileExists(dir, e...)
}

// DirExists 目录是否存在
func DirExists(dir string, e ...*error) bool {
	return ioext.DirExists(dir, e...)
}

func IsDirectory(dir string, e ...*error) bool {
	return ioext.IsDirectory(dir, e...)
}

func EnumerateFiles(pa string) ([]string, error) {
	return ioext.EnumerateFiles(pa)
}
