package harness

import (
	"cn/com/hengwei/pkg/ds_client"
	"context"
	"io"
	"os"
	"time"

	"github.com/runner-mei/errors"
	"github.com/runner-mei/goutils/shell"
)

var dumpSSH = false

func DailSSH(ctx context.Context, params *ds_client.SSHParam, args ...Option) (shell.Conn, []byte, error) {
	var opts options
	for _, o := range args {
		o.apply(&opts)
	}

	if opts.questions == nil {
		opts.questions = noQuestions
	}

	if dumpSSH {
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

	if params.UseExternalSSH {
		c, err := shell.ConnectPlink(params.Host(), params.Username, params.Password, params.PrivateKey, opts.sWriter, opts.cWriter)
		if err != nil {
			return nil, nil, err
		}

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

		return sshLoginWithExternSSH(ctx, c, params, &opts)
	}

	c, err := shell.ConnectSSH(params.Host(), params.Username, params.Password, params.PrivateKey, opts.sWriter, opts.cWriter)
	if err != nil {
		return nil, nil, err
	}

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
	return sshLogin(ctx, c, params, &opts)
}

func sshLogin(ctx context.Context, c shell.Conn, params *ds_client.SSHParam, opts *options) (shell.Conn, []byte, error) {
	var prompts [][]byte
	if params.Prompt != "" {
		prompts = [][]byte{[]byte(params.Prompt)}
	}

	if opts.skipEnable && opts.skipPrompt {
		return c, nil, nil
	}

	prompt, err := shell.ReadPrompt(ctx, c, prompts, opts.questions...)
	if err != nil {
		c.Close()
		return nil, nil, err
	}

	if opts.skipEnable {
		return c, prompt, nil
	}
	return sshEnableLogin(ctx, c, params, prompt, opts)
}

func sshLoginWithExternSSH(ctx context.Context, c shell.Conn, params *ds_client.SSHParam, opts *options) (shell.Conn, []byte, error) {
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
		return nil, nil, errors.New("便用 UseExternalSSH 时不支持 skipPrompt 选项")
	}

	if opts.skipEnable {
		return c, prompt, nil
	}
	return sshEnableLogin(ctx, c, params, prompt, opts)
}

func sshEnableLogin(ctx context.Context, c shell.Conn, params *ds_client.SSHParam, prompt []byte, opts *options) (shell.Conn, []byte, error) {
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
