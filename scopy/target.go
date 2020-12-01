package scopy

import "io"

type Target interface {
	io.Closer

	Upload(localPath string, remotePath string) (err error)
	Download(remotePath string, localPath string) (err error)
}
