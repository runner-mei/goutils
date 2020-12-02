package scopy

import (
	"io"
	"log"
	"path/filepath"
)

type Session interface {
	io.Closer

	Upload(localPath string, remotePath string) (bytes int64, err error)
	Download(remotePath string, localPath string) (bytes int64, err error)
}

type Target interface {
	io.Closer

	Copy(srcPath, destPath string) error
}

func Upload(sess Session, currentdir string) Target {
	return &UploadCopyer{
		Session:    sess,
		CurrentDir: currentdir,
	}
}

func Download(sess Session, currentdir string) Target {
	return &DownloadCopyer{
		Session:    sess,
		CurrentDir: currentdir,
	}
}

type UploadCopyer struct {
	Session    Session
	CurrentDir string
}

func (cp *UploadCopyer) Close() error {
	return cp.Session.Close()
}

func (cp *UploadCopyer) Copy(srcPath, destPath string) error {
	remotePath := destPath
	if cp.CurrentDir != "" {
		if !filepath.IsAbs(destPath) {
			remotePath = filepath.Join(cp.CurrentDir, destPath)
		}
	}
	remotePath = filepath.ToSlash(remotePath)
	_, err := cp.Session.Upload(srcPath, remotePath)
	if err == nil {
		log.Println("copy", srcPath, "to", remotePath)
	}
	return err
}

type DownloadCopyer struct {
	Session    Session
	CurrentDir string
}

func (cp *DownloadCopyer) Close() error {
	return cp.Session.Close()
}

func (cp *DownloadCopyer) Copy(srcPath, destPath string) error {
	remotePath := srcPath
	if cp.CurrentDir != "" {
		if !filepath.IsAbs(srcPath) {
			remotePath = filepath.Join(cp.CurrentDir, srcPath)
		}
	}
	remotePath = filepath.ToSlash(remotePath)
	_, err := cp.Session.Download(remotePath, destPath)
	if err == nil {
		log.Println("copy", remotePath, "to", srcPath)
	}
	return err
}
