package scopy

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/pkg/sftp"
	"github.com/runner-mei/errors"
	"github.com/runner-mei/goutils/shell"
	"golang.org/x/crypto/ssh"
)

func SFTPWithKey(host, username, keyfile, passphrase string) (Session, error) {
	var privateKey string
	bs, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, errors.Wrap(err, "load keyfile fail")
	}
	privateKey = string(bs)

	conn, err := shell.DialSSH(host, username, passphrase, privateKey)
	if err != nil {
		return nil, err
	}

	// create new SFTP client
	client, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &sftpTarget{
		conn:   conn,
		client: client,
	}, nil
}

func SFTPWithPassword(host, username, password string) (Session, error) {
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

	return &sftpTarget{
		conn:   conn,
		client: client,
	}, nil
}

type sftpTarget struct {
	conn   *ssh.Client
	client *sftp.Client
}

func (st *sftpTarget) Close() error {
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

func (st *sftpTarget) Upload(localPath string, remotePath string) (int64, error) {
	// create destination file
	dstFile, err := st.client.Create(remotePath)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()

	// create source file
	srcFile, err := os.Open(localPath)
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

func (st *sftpTarget) Download(remotePath string, localPath string) (int64, error) {
	// create destination file
	dstFile, err := os.Create(localPath)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()

	// open source file
	srcFile, err := st.client.Open(remotePath)
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
