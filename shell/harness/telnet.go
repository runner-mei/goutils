package harness

import (
	"cn/com/hengwei/pkg/ds_client"
	"cn/com/hengwei/pkg/goutils/shell"
	"context"
	"io"
	"os"
	"time"

	"github.com/runner-mei/errors"
)

var dumpTelnet = false

func DailTelnet(ctx context.Context, params *ds_client.TelnetParam, args ...Option) (shell.Conn, []byte, error) {
	var opts options
	for _, o := range args {
		o.apply(&opts)
	}
	if opts.questions == nil {
		opts.questions = noQuestions
	}

	telnetConn, err := shell.DialTelnetTimeout("tcp", params.Host(), 30*time.Second)
	if err != nil {
		return nil, nil, err
	}

	if dumpTelnet {
		sw := shell.WriteFunc(func(p []byte) (int, error) {
			io.WriteString(os.Stdout, "s:")
			io.WriteString(os.Stdout, shell.ToHexStringIfNeed(p))
			io.WriteString(os.Stdout, "\r\n")
			return len(p), nil
		})

		if opts.sWriter == nil {
			opts.sWriter = sw
		} else {
			opts.sWriter = io.MultiWriter(opts.sWriter, sw)
		}

		cw := shell.WriteFunc(func(p []byte) (int, error) {
			io.WriteString(os.Stdout, "c:")
			io.WriteString(os.Stdout, shell.ToHexStringIfNeed(p))
			io.WriteString(os.Stdout, "\r\n")
			return len(p), nil
		})

		if opts.cWriter == nil {
			opts.cWriter = cw
		} else {
			opts.cWriter = io.MultiWriter(opts.cWriter, cw)
		}
	}

	c := shell.TelnetWrap(telnetConn, opts.sWriter, opts.cWriter)
	if params.UseCRLF {
		c.UseCRLF()
	}
	c.SetReadDeadline(30 * time.Second)
	c.SetWriteDeadline(10 * time.Second)

	if opts.skipLogin {
		return c, nil, nil
	}
	c1 := c.SetTeeReader(opts.inWriter)
	c2 := c.SetTeeWriter(opts.outWriter)

	defer func() {
		c1()
		c2()
	}()

	return telnetLogin(ctx, c, params, &opts)
}

func telnetLogin(ctx context.Context, c shell.Conn, params *ds_client.TelnetParam, opts *options) (shell.Conn, []byte, error) {

	var prompts [][]byte
	if params.Prompt != "" {
		prompts = [][]byte{[]byte(params.Prompt)}
	}

	var userPrompts [][]byte
	if params.UserQuest != "" {
		userPrompts = [][]byte{[]byte(params.UserQuest)}
	}
	var passwordPrompts [][]byte
	if params.PasswordQuest != "" {
		passwordPrompts = [][]byte{[]byte(params.PasswordQuest)}
	}

	prompt, err := shell.UserLogin(ctx, c, userPrompts, []byte(params.Username), passwordPrompts, []byte(params.Password), prompts, opts.questions...)
	if err != nil {
		c.Close()
		return nil, nil, err
	}

	if opts.skipPrompt {
		c.Close()
		return nil, nil, errors.New("便用 Telnet 时不支持 skipPrompt 选项")
	}

	if opts.skipEnable {
		return c, prompt, nil
	}

	return telnetEnableLogin(ctx, c, params, prompt, opts)
}

func telnetEnableLogin(ctx context.Context, c shell.Conn, params *ds_client.TelnetParam, prompt []byte, opts *options) (shell.Conn, []byte, error) {
	var enablePasswordPrompts [][]byte
	if params.EnablePasswordQuest != "" {
		enablePasswordPrompts = [][]byte{[]byte(params.EnablePasswordQuest)}
	}
	var enablePrompts [][]byte
	if params.EnablePrompt != "" {
		enablePrompts = [][]byte{[]byte(params.EnablePrompt)}
	}

	if params.EnablePassword != "" {
		var err error
		prompt, err = shell.WithEnable(ctx, c, []byte(params.EnableCommand), enablePasswordPrompts, []byte(params.EnablePassword), enablePrompts)
		if err != nil {
			c.Close()
			return nil, nil, err
		}
	}

	return c, prompt, nil
}
