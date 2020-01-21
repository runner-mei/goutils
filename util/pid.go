package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	ps "github.com/mitchellh/go-ps"
)

func CreatePidFile(pidFile, image string) error {
	if pidString, err := ioutil.ReadFile(pidFile); err == nil {
		pid, err := strconv.Atoi(string(pidString))
		if err == nil {
			if pr, e := ps.FindProcess(pid); nil != e || (nil != pr &&
				strings.Contains(strings.ToLower(pr.Executable()), strings.ToLower(image))) {
				return fmt.Errorf("pid file is already exists, ensure "+image+" is not running or delete %s.", pidFile)
			}
		}
	}
	if e := os.MkdirAll(filepath.Dir(pidFile), 0777); e != nil {
		if !os.IsExist(e) {
			log.Println("[warn] mkdir '"+filepath.Dir(pidFile)+"' fail:", e)
		}
	}
	file, err := os.Create(pidFile)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = fmt.Fprintf(file, "%d", os.Getpid())
	return err
}

func RemovePidFile(pidFile string) {
	if err := os.Remove(pidFile); err != nil {
		fmt.Printf("Error removing %s: %s\r\n", pidFile, err)
	}
}
