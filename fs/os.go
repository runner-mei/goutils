package fs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type osFs struct {
	dir string
}

func (fs *osFs) Close() error {
	return nil
}
func (fs *osFs) ReadDir() ([]os.FileInfo, error) {
	return ioutil.ReadDir(fs.dir)
}
func (fs *osFs) Open(filename string) (io.ReadCloser, error) {
	fmt.Println(filepath.Join(fs.dir, filename))
	return os.Open(filepath.Join(fs.dir, filename))
}
func (fs *osFs) Create(filename string) (io.WriteCloser, error) {
	if err := os.MkdirAll(filepath.Dir(filepath.Join(fs.dir, filename)), 0777); err != nil {
		if !os.IsExist(err) {
			return nil, err
		}
	}
	return os.Create(filepath.Join(fs.dir, filename))
}
func (fs *osFs) Delete(filename string) error {
	fmt.Println(filepath.Join(fs.dir, filename))
	return os.Remove(filepath.Join(fs.dir, filename))
}
