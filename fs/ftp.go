package fs

import (
	"io"
	"os"

	"github.com/runner-mei/ftp"
)

func FTP(host, username, password, currentdir string, disableEPSV bool) (FS, error) {
	conn, err := ftp.Dial(host)
	if err != nil {
		return nil, err
	}

	if err := conn.Login(username, password); err != nil {
		conn.Quit()
		return nil, err
	}
	if currentdir != "" {
		if err := conn.ChangeDir(currentdir); err != nil {
			conn.Quit()
			return nil, err
		}
	}
	conn.DisableEPSV = disableEPSV

	return &ftpFS{
		client: conn,
	}, nil
}

type ftpFS struct {
	client *ftp.ServerConn
}

func (st *ftpFS) Close() error {
	return st.client.Quit()
}

func (fs *ftpFS) ReadDir() ([]os.FileInfo, error) {
	// entries, err := fs.client.List(".")
	// if err != nil {
	// 	return nil, err
	// }
	// var results []os.FileInfo
	// for _, f := range entries {
	// 	results = append(results, f)
	// }
	//  return results
	panic("not implement")
}
func (fs *ftpFS) Open(filename string) (io.ReadCloser, error) {
	response, err := fs.client.Retr(filename)
	if err != nil {
		return nil, err
	}
	return response, nil
}
func (fs *ftpFS) Create(filename string) (io.WriteCloser, error) {
	r, w := io.Pipe()
	err := fs.client.Stor(filename, r)
	if err != nil {
		return nil, err
	}
	return w, nil
}
func (fs *ftpFS) Delete(filename string) error {
	return fs.client.Delete(filename)
}
