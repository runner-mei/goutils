package scopy

import (
	"io"
	"os"

	"github.com/runner-mei/ftp"
)

func FTP(host, username, password, currentdir string, disableEPSV bool) (Session, error) {
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

	return &ftpTarget{
		client: conn,
	}, nil
}

type ftpTarget struct {
	client *ftp.ServerConn
}

func (st *ftpTarget) Close() error {
	return st.client.Quit()
}

func (st *ftpTarget) Upload(localPath string, remotePath string) (int64, error) {
	// create source file
	srcFile, err := os.Open(localPath)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	// create destination file
	err = st.client.Stor(remotePath, srcFile)
	if err != nil {
		return 0, err
	}
	return -1, nil
}

func (st *ftpTarget) Download(remotePath string, localPath string) (int64, error) {
	// create destination file
	dstFile, err := os.Create(localPath)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()

	// open source file
	srcFile, err := st.client.Retr(remotePath)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	// copy source file to destination file
	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return bytes, err
	}

	return bytes, dstFile.Close()
}
