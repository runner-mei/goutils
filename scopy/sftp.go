package scopy

import (
	"github.com/melbahja/goph"
)

func SFTPWithKey(host, username, keyfile, passphrase string) (Target, error) {
	auth, err := goph.Key(keyfile, passphrase)
	if err != nil {
		return nil, err
	}

	return goph.New(username, host, auth)
}

func SFTPWithPassword(host, username, password string) (Target, error) {
	auth := goph.Password(password)

	return goph.New(username, host, auth)
}
