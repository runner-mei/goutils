package shell

import (
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/runner-mei/errors"
	"github.com/runner-mei/goutils/util"
)

var (
	PlinkPath = "runtime_env/putty/plink.exe"
	binDir    string
)

func init() {
	pa, _ := os.Executable()
	binDir = filepath.Dir(pa)
}

func init() {
	if plinkPath := os.Getenv("hengwei_plink"); plinkPath != "" {
		PlinkPath = plinkPath
	}

	if fi, err := os.Stat(PlinkPath); err != nil && os.IsNotExist(err) {
		var files []string
		if util.IsWindows {
			files = []string{
				"C:\\Program Files\\hengwei\\runtime_env\\putty\\plink.exe",
				filepath.Join(binDir, "plink.exe"),

				"C:\\Program Files\\hengwei\\runtime_env\\putty\\plink_old.exe",
				filepath.Join(binDir, "plink_old.exe"),
			}
		} else {
			files = []string{
				"/usr/local/tpt/runtime_env/putty/plink",
				filepath.Join(binDir, "plink"),
				"/usr/local/tpt/runtime_env/putty/plink_old",
				filepath.Join(binDir, "plink_old"),
			}
		}
		for _, pa := range files {
			if fi, err = os.Stat(pa); err == nil && !fi.IsDir() {
				PlinkPath = pa
				break
			}
		}
	}
}

type PlinkClient struct {
	cmd *exec.Cmd
	ConnWrapper
}

func (c *PlinkClient) Close() error {
	if e := c.cmd.Process.Kill(); nil != e {
		return e
	}

	c.ConnWrapper.Close()
	return c.cmd.Wait()
}

func ConnectPlink(host, username, password, privateKey string, sWriter, cWriter io.Writer) (*PlinkClient, error) {
	if privateKey != "" {
		return nil, errors.New("兼容模式不支持 证书登录")
	}
	p := MakePipe(2048)
	address, port, err := net.SplitHostPort(host)

	var cmd *exec.Cmd
	if err != nil {
		cmd = exec.Command(PlinkPath, "-t", username+"@"+host)
	} else {
		cmd = exec.Command(PlinkPath, "-t", username+"@"+address, "-P", port)
	}

	if sWriter != nil {
		cmd.Stderr = MultWriters(p, sWriter)
	} else {
		cmd.Stderr = p //MultWriters(w, os.Stdout)
	}
	cmd.Stdout = cmd.Stderr

	stdin, e := cmd.StdinPipe()
	if nil != e {
		return nil, e
	}
	if e := cmd.Start(); nil != e {
		return nil, e
	}

	if cWriter != nil {
		cWriter = MultWriters(stdin, cWriter)
	} else {
		cWriter = stdin
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			p.CloseWithError(err)
		}
	}()

	pClient := &PlinkClient{
		cmd: cmd,
	}
	pClient.Init(closeFunc(func() error {
		if e := cmd.Process.Kill(); nil != e {
			return e
		}
		return nil
	}), cWriter, p)
	return pClient, nil
}
