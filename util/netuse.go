package util

import (
	"bytes"
	"context"
	"log"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var (
	success_of_gbk    []byte
	success_of_utf8   []byte
	success_of_ascii  []byte
	multconn_of_gbk   []byte
	multconn_of_utf8  []byte
	multconn_of_ascii []byte
)

func init() {
	var gbk_encoder = simplifiedchinese.GB18030.NewEncoder()

	success_of_utf8 = []byte("成功")
	success_of_ascii = []byte("success")
	bs, _, e := transform.Bytes(gbk_encoder, success_of_utf8)
	if nil != e {
		panic(e)
	}
	success_of_gbk = bs

	multconn_of_utf8 = []byte("不允许一个用户使用一个以上用户名与服务器或共享资源的多重连接")
	multconn_of_ascii = []byte("不允许一个用户使用一个以上用户名与服务器或共享资源的多重连接")
	bs, _, e = transform.Bytes(gbk_encoder, multconn_of_utf8)
	if nil != e {
		panic(e)
	}
	multconn_of_gbk = bs
}

func Netuse(dir, username, password string) {
	ctx, cancal := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancal()

	if strings.HasSuffix(dir, "\\") {
		dir = strings.TrimSuffix(dir, "\\")
	}
	if strings.HasSuffix(dir, "/") {
		dir = strings.TrimSuffix(dir, "/")
	}

	if strings.Contains(username, "@") {
		names := strings.SplitN(username, "@", 2)
		if len(names) == 2 {
			username = names[1] + "\\" + names[0]
		}
	}
	//net use \192.168.1.2\software_backup tpt_8498b2c7 /USER:administrator /PERSISTENT:YES
	cmd := exec.CommandContext(ctx, "net", "use", dir, password, "/USER:"+username, "/PERSISTENT:YES")
	bs, err := cmd.CombinedOutput()
	if err != nil {
		if len(bs) == 0 || !(bytes.Contains(bs, multconn_of_gbk) || bytes.Contains(bs, multconn_of_utf8) || bytes.Contains(bs, multconn_of_ascii)) {
			cmd.Args[3] = "******"
			if len(bs) > 0 {
				log.Println("[windows]", dir, cmd.Path, cmd.Args, string(ToSimplifiedChinese(bs)), err)
			} else {
				log.Println("[windows]", dir, cmd.Path, cmd.Args, err)
			}
		}
	} else if len(bs) > 0 && !(bytes.Contains(bs, success_of_gbk) || bytes.Contains(bs, success_of_utf8) || bytes.Contains(bs, success_of_ascii)) {
		cmd.Args[3] = "******"
		log.Println("[windows]", dir, cmd.Path, cmd.Args, string(ToSimplifiedChinese(bs)))
	}
}
