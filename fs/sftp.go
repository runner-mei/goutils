package fs

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"github.com/runner-mei/errors"
	"github.com/runner-mei/goutils/shell"
	"golang.org/x/crypto/ssh"
)

type sftpFs struct {
	conn   *ssh.Client
	client *sftp.Client
	dir    string
}

func (st *sftpFs) Close() error {
	err1 := st.client.Close()
	err2 := st.conn.Close()

	if err1 == nil {
		return err2
	}
	if err2 == nil {
		return err1
	}
	return errors.ErrArray(err1, err2)
}

func (fs *sftpFs) ReadDir() ([]os.FileInfo, error) {
	return fs.client.ReadDir(fs.dir)
}
func (fs *sftpFs) Open(filename string) (io.ReadCloser, error) {
	return fs.client.Open(filepath.ToSlash(filepath.Join(fs.dir, filename)))
}
func (fs *sftpFs) Create(filename string) (io.WriteCloser, error) {
	fullfilename := filepath.Join(fs.dir, filename)
	w, err := fs.client.Create(filepath.ToSlash(fullfilename))
	if err == nil {
		return w, nil
	}
	if os.IsNotExist(err) {
		e := fs.MkdirAll(filepath.ToSlash(filepath.Dir(fullfilename)))
		if e == nil {
			return fs.client.Create(filepath.ToSlash(fullfilename))
		}
	}
	return nil, err
}

func (fs *sftpFs) MkdirAll(dir string) error {
	return fs.client.MkdirAll(dir)
}

func (fs *sftpFs) Delete(filename string) error {
	return fs.client.Remove(filepath.ToSlash(filepath.Join(fs.dir, filename)))
}

func OpenSftp(host, username, password, dir string) (FS, error) {
	conn, err := shell.DialSSH(host, username, password, "")
	if err != nil {
		return nil, err
	}

	// create new SFTP client
	client, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &sftpFs{
		conn:   conn,
		client: client,
		dir:    dir,
	}, nil
}
