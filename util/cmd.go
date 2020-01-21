package util

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kardianos/osext"
)

var (
	Commands         = map[string]string{}
	ExecutableFolder string
)

func init() {
	executableFolder, e := osext.ExecutableFolder()
	if nil != e {
		return
	}
	ExecutableFolder = executableFolder
}

func abs(s string) string {
	r, e := filepath.Abs(s)
	if e != nil {
		return s
	}
	return r
}

func LookPath(executableFolder string, alias ...string) (string, bool) {
	var names []string
	for _, aliasName := range alias {
		if runtime.GOOS == "windows" {
			names = append(names, aliasName, aliasName+".bat", aliasName+".com", aliasName+".exe")
		} else {
			names = append(names, aliasName, aliasName+".sh")
		}
	}

	for _, nm := range names {
		files := []string{nm,
			filepath.Join("bin", nm),
			filepath.Join("tools", nm),
			filepath.Join("runtime_env", nm),
			filepath.Join("..", nm),
			filepath.Join("..", "bin", nm),
			filepath.Join("..", "tools", nm),
			filepath.Join("..", "runtime_env", nm),
			filepath.Join(executableFolder, nm),
			filepath.Join(executableFolder, "bin", nm),
			filepath.Join(executableFolder, "tools", nm),
			filepath.Join(executableFolder, "runtime_env", nm),
			filepath.Join(executableFolder, "..", nm),
			filepath.Join(executableFolder, "..", "bin", nm),
			filepath.Join(executableFolder, "..", "tools", nm),
			filepath.Join(executableFolder, "..", "runtime_env", nm)}
		for _, file := range files {
			// fmt.Println("====", file)
			file = abs(file)
			if st, e := os.Stat(file); nil == e && nil != st && !st.IsDir() {
				//fmt.Println("1=====", file, e)
				return file, true
			}
		}
	}

	for _, nm := range names {
		_, err := exec.LookPath(nm)
		if nil == err {
			return nm, true
		}
	}
	return "", false
}

func LoadCommands(executableFolder string) {
	for _, nm := range []string{"snmpget", "snmpgetnext", "snmpdf", "snmpbulkget",
		"snmpbulkwalk", "snmpdelta", "snmpnetstat", "snmpset", "snmpstatus",
		"snmptable", "snmptest", "snmptools", "snmptranslate", "snmptrap", "snmpusm",
		"snmpvacm", "snmpwalk", "wshell"} {
		if pa, ok := LookPath(executableFolder, nm); ok {
			Commands[nm] = pa
		} else if pa, ok := LookPath(executableFolder, "netsnmp/"+nm); ok {
			Commands[nm] = pa
		} else if pa, ok := LookPath(executableFolder, "net-snmp/"+nm); ok {
			Commands[nm] = pa
		}
	}

	if pa, ok := LookPath(executableFolder, "tpt"); ok {
		Commands["tpt"] = pa
	}
	if pa, ok := LookPath(executableFolder, "nmap/nping"); ok {
		Commands["nping"] = pa
	}
	if pa, ok := LookPath(executableFolder, "nmap/nmap"); ok {
		Commands["nmap"] = pa
	}
	if pa, ok := LookPath(executableFolder, "putty/plink", "ssh"); ok {
		Commands["plink"] = pa
		Commands["ssh"] = pa
	}
	if pa, ok := LookPath(executableFolder, "dig/dig", "dig"); ok {
		Commands["dig"] = pa
		Commands["runtime_env/dig/dig"] = pa
	}
	if pa, ok := LookPath(executableFolder, "ping"); ok {
		Commands["ping"] = pa
	}
	if pa, ok := LookPath(executableFolder, "tracert"); ok {
		Commands["tracert"] = pa
	}
	if pa, ok := LookPath(executableFolder, "traceroute"); ok {
		Commands["traceroute"] = pa
	}
}

// Job is an interface for submitted cron jobs.
type Command struct {
	Execute   string
	Arguments []string
	Dir       string

	Logfile      string
	Environments []string
}

const maxBytes = 5 * 1024 * 1024
const maxNum = 5

func (job *Command) rotateLogFile() error {
	st, err := os.Stat(job.Logfile)
	if nil != err { // file exists
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if st.Size() < maxBytes {
		return nil
	}

	fname2 := job.Logfile + fmt.Sprintf(".%04d", maxNum)
	_, err = os.Stat(fname2)
	if nil == err {
		err = os.Remove(fname2)
		if err != nil {
			return err
		}
	}

	fname1 := fname2
	for num := maxNum - 1; num > 0; num-- {
		fname2 = fname1
		fname1 = job.Logfile + fmt.Sprintf(".%04d", num)

		_, err = os.Stat(fname1)
		if nil != err {
			continue
		}
		err = os.Rename(fname1, fname2)
		if err != nil {
			return err
		}
	}

	err = os.Rename(job.Logfile, fname1)
	if err != nil {
		return err
	}
	return nil
}

func (job *Command) Run(ctx context.Context) error {
	if e := job.rotateLogFile(); e != nil {
		return errors.New("rotate log file(" + job.Logfile + ") failed, " + e.Error())
	}
	out, e := os.OpenFile(job.Logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if nil != e {
		logfile := filepath.Base(job.Logfile)
		logfile = strings.Replace(logfile, "/", "_", -1)
		logfile = strings.Replace(logfile, "\\", "_", -1)
		logfile = strings.Replace(logfile, "*", "_", -1)
		logfile = strings.Replace(logfile, ":", "_", -1)
		logfile = strings.Replace(logfile, "\"", "_", -1)
		logfile = strings.Replace(logfile, "|", "_", -1)
		logfile = strings.Replace(logfile, "?", "_", -1)
		logfile = strings.Replace(logfile, ">", "_", -1)
		logfile = strings.Replace(logfile, "<", "_", -1)
		job.Logfile = filepath.Join(filepath.Dir(job.Logfile), logfile)

		out, e = os.OpenFile(job.Logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if nil != e {
			return errors.New("open log file(" + job.Logfile + ") failed, " + e.Error())
		}
	}
	defer out.Close()

	io.WriteString(out, "=============== begin ===============\r\n")
	defer io.WriteString(out, "===============  end  ===============\r\n")

	execPath := job.Execute
	if s := Commands[job.Execute]; s != "" {
		s = execPath
	}

	executePath, found := LookPath(ExecutableFolder, execPath)
	if !found {
		executePath = job.Execute
	}

	cmd := exec.CommandContext(ctx, executePath, job.Arguments...)
	cmd.Stderr = out
	cmd.Stdout = out
	cmd.Dir = job.Dir

	if len(job.Environments) > 0 {
		osEnv := os.Environ()
		var environments []string
		environments = make([]string, 0, len(job.Environments)+len(osEnv))
		environments = append(environments, osEnv...)
		environments = append(environments, job.Environments...)
		cmd.Env = environments
	}

	io.WriteString(out, cmd.Path)
	for idx, s := range cmd.Args {
		if 0 == idx {
			continue
		}
		io.WriteString(out, "\r\n \t\t")
		io.WriteString(out, s)
	}
	io.WriteString(out, "\r\n===============  out  ===============\r\n")
	if e = cmd.Start(); nil != e {
		io.WriteString(out, "start failed, "+e.Error()+"\r\n")
		return nil
	}

	if e = cmd.Wait(); nil != e {
		io.WriteString(out, "run failed, "+e.Error()+"\r\n")
	} else if nil != cmd.ProcessState {
		io.WriteString(out, "run ok, exit with "+cmd.ProcessState.String()+".\r\n")
	}
	return nil
}
