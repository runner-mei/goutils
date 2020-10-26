package cmdx

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/runner-mei/errors"
	"github.com/runner-mei/goutils/tid"
	"github.com/runner-mei/goutils/util"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type Cmd struct {
	Execute         string   `json:"execute"`
	Directory       string   `json:"directory,omitempty"`
	Arguments       []string `json:"arguments,omitempty"`
	Environments    []string `json:"environments,omitempty"`
	ExceptedReturn  string   `json:"excepted_return,omitempty"`
	NoCheckExitCode bool     `json:"no_check_exit_code,omitempty"`
}

func (cmd *Cmd) Run(ctx context.Context, field func(string) string) error {
	execExt := ".exe"
	batExt := ".bat"
	if !util.IsWindows {
		execExt = ""
		batExt = ".sh"
	}
	expendFunc := func(name string) string {
		switch name {
		case "os":
			return runtime.GOOS
		case "os_ext":
			return execExt
		case "sh_ext":
			return batExt
		case "arch":
			return runtime.GOARCH
		}
		return field(name)
	}

	innercmd := exec.CommandContext(ctx, cmd.Execute, cmd.Arguments...)
	innercmd.Dir = os.Expand(cmd.Directory, expendFunc)
	innercmd.Path = os.Expand(innercmd.Path, expendFunc)
	for idx := range innercmd.Args {
		innercmd.Args[idx] = os.Expand(innercmd.Args[idx], expendFunc)
	}

	sysEnv := os.Environ()
	innercmd.Env = make([]string, len(sysEnv)+len(cmd.Environments)+1)
	copy(innercmd.Env, sysEnv)
	for idx := range cmd.Environments {
		innercmd.Env[len(sysEnv)+idx] = os.Expand(cmd.Environments[idx], expendFunc)
	}
	innercmd.Env[len(sysEnv)+len(cmd.Environments)] = "uuid=" + tid.GenerateID()

	if util.IsWindows {
		if strings.Contains(innercmd.Path, " ") && (strings.HasSuffix(innercmd.Path, ".bat") || strings.HasSuffix(innercmd.Path, ".cmd")) {
			// https://github.com/golang/go/issues/9084
			// 这是一个巨坑。

			systemDrive := os.Getenv("SystemDrive")
			if systemDrive == "" {
				systemDrive = "C:"
			}
			filename := strings.Replace(filepath.Base(innercmd.Path), " ", "_", -1)
			if strings.HasSuffix(innercmd.Path, ".cmd") {
				filename = strings.TrimSuffix(filename, ".cmd") + "_" + tid.GenerateID() + ".cmd"
			} else {
				filename = strings.TrimSuffix(filename, ".bat") + "_" + tid.GenerateID() + ".bat"
			}
			filename = filepath.Join(systemDrive+"\\hw_tmp", filename)
			if dirName := filepath.Dir(filename); !util.DirExists(dirName) {
				err := os.MkdirAll(dirName, 0666)
				if err != nil {
					return errors.Wrap(err, "请个 bat 路中有空格，将 '"+innercmd.Path+"' 拷到 '"+filename+"' 失败，你可将它移到一个没有空格的位置试试")
				}
			}

			err := util.CopyFile(innercmd.Path, filename)
			if err != nil {
				return errors.Wrap(err, "请个 bat 路中有空格，将 '"+innercmd.Path+"' 拷到 '"+filename+"' 失败，你可将它移到一个没有空格的位置试试")
			}
			innercmd.Path = filename
			innercmd.Args[0] = filename
			defer os.Remove(filename)
		}
	}

	out, err := innercmd.CombinedOutput()
	if util.IsWindows && len(out) > 0 {
		bs, _, err := transform.Bytes(simplifiedchinese.GB18030.NewDecoder(), out)
		if err == nil {
			out = bs
		}
	}
	if !cmd.NoCheckExitCode {
		if err != nil {
			if len(out) == 0 {
				return err
			}
			return errors.WrapWithSuffix(err, string(out))
		}
	}
	if cmd.ExceptedReturn == "" {
		return nil
	}
	if bytes.Contains(out, []byte(cmd.ExceptedReturn)) {
		return nil
	}
	return errors.New("程序执行成功，但没有找到关键字 '" + cmd.ExceptedReturn + "'\r\n" + string(out))
}

func Read(filename string) (*Cmd, error) {
	var cmd Cmd
	err := util.FromHjsonFile(filename, filename)
	return &cmd, err
}
