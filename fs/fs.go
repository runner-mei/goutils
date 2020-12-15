package fs

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/runner-mei/errors"
	"github.com/runner-mei/goutils/as"
)

type FS interface {
	io.Closer

	ReadDir() ([]os.FileInfo, error)
	Open(filename string) (io.ReadCloser, error)
	Create(filename string) (io.WriteCloser, error)
	Delete(filename string) error
}

func MoveFile(src, dst FS, name string) error {
	_, err := CopyFile(src, dst, name)
	if err != nil {
		return errors.Wrap(err, "拷贝文件失败")
	}

	err = src.Delete(name)
	if err != nil {
		return errors.Wrap(err, "拷贝文件成功后，删除源文件失败")
	}
	return nil
}

func CopyFile(src, dst FS, name string) (int64, error) {
	dstFile, err := dst.Create(name)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()

	srcFile, err := src.Open(name)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	// copy source file to destination file
	nbytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return nbytes, err
	}

	return nbytes, dstFile.Close()
}

func OpenFS(target string) (FS, error) {
	if filepath.IsAbs(target) {
		return &osFs{dir: target}, nil
	}

	u, err := url.Parse(target)
	if err != nil {
		return &osFs{dir: target}, nil
	}

	switch strings.ToLower(u.Scheme) {
	case "":
		fmt.Println("scheme", u.Scheme, target)
		return &osFs{dir: target}, nil
	case "file":
		fmt.Println(u.Path)
		return &osFs{dir: u.Path}, nil
	case "ssh", "sftp":
		userinfo := u.User
		if userinfo == nil {
			return nil, errors.New("SSH 缺少用户名或密码")
		}
		password, _ := userinfo.Password()
		fmt.Println(u.Host, userinfo.Username(), password, u.Path)

		address, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			address = u.Host
		}
		if port == "" {
			port = "22"
		}
		return OpenSftp(net.JoinHostPort(address, port), userinfo.Username(), password, u.Path)
	case "ftp":
		userinfo := u.User
		if userinfo == nil {
			return nil, errors.New("SSH 缺少用户名或密码")
		}
		password, _ := userinfo.Password()
		params := u.Query()

		var disableEPSV bool
		if s := params.Get("disableEPSV"); s != "" {
			disableEPSV = as.BoolWithDefault(s, false)
		} else if s := params.Get("disableEpsv"); s != "" {
			disableEPSV = as.BoolWithDefault(s, false)
		} else if s := params.Get("disable_epsv"); s != "" {
			disableEPSV = as.BoolWithDefault(s, false)
		}
		return FTP(u.Host, userinfo.Username(), password, u.Path, disableEPSV)
	default:
		return nil, errors.New("不支持 '" + u.Scheme + "'")
	}
}
