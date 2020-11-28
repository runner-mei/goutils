package atomicfile

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

const (
	onWindows = runtime.GOOS == "windows"
)

// PendingFile is a pending temporary file, waiting to replace the destination
// path in a call to CloseAtomicallyReplace.
type PendingFile struct {
	*os.File

	path   string
	done   bool
	closed bool
}

// Cleanup is a no-op if CloseAtomicallyReplace succeeded, and otherwise closes
// and removes the temporary file.
func (t *PendingFile) Cleanup() error {
	if t.done {
		return nil
	}
	// An error occurred. Close and remove the tempfile. Errors are returned for
	// reporting, there is nothing the caller can recover here.
	var closeErr error
	if !t.closed {
		closeErr = t.Close()
	}
	if err := os.Remove(t.Name()); err != nil {
		return err
	}
	return closeErr
}

// CloseAtomicallyReplace closes the temporary file and atomically replaces
// the destination file with it, i.e., a concurrent open(2) call will either
// open the file previously located at the destination path (if any), or the
// just written file, but the file will always be present.
func (t *PendingFile) CloseAtomicallyReplace() error {
	// Even on an ordered file system (e.g. ext4 with data=ordered) or file
	// systems with write barriers, we cannot skip the fsync(2) call as per
	// Theodore Ts'o (ext2/3/4 lead developer):
	//
	// > data=ordered only guarantees the avoidance of stale data (e.g., the previous
	// > contents of a data block showing up after a crash, where the previous data
	// > could be someone's love letters, medical records, etc.). Without the fsync(2)
	// > a zero-length file is a valid and possible outcome after the rename.
	if err := t.Sync(); err != nil {
		return err
	}
	t.closed = true
	if err := t.Close(); err != nil {
		return err
	}
	if err := os.Rename(t.Name(), t.path); err != nil {
		return err
	}
	t.done = true
	return nil
}

// CreateFile wraps ioutil.TempFile for the use case of atomically creating or
// replacing the destination file at path.
func CreateFile(path string, perm os.FileMode) (*PendingFile, error) {
	f, err := ioutil.TempFile(filepath.Dir(path), filepath.Base(path)+".tmp")
	if err != nil {
		return nil, err
	}

	// Set permissions before writing data, in case the data is sensitive.
	if !onWindows {
		if err := f.Chmod(perm); err != nil {
			tmpfilename := f.Name()
			f.Close()
			os.Remove(tmpfilename)
			return nil, err
		}
	}

	return &PendingFile{File: f, path: path}, nil
}

func WriteFile(filename string, data []byte, perm os.FileMode) error {
	t, err := CreateFile(filename, perm)
	if err != nil {
		return err
	}
	defer t.Cleanup()

	if _, err := t.Write(data); err != nil {
		return err
	}

	return t.CloseAtomicallyReplace()
}
