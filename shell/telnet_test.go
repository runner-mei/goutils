package shell

import (
	"cn/com/hengwei/sim/sshd"
	"cn/com/hengwei/sim/telnetd"
	"context"
	"net"
	"testing"
	"time"
)

func TestTelnetSimSimple(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "","abc#", sshd.Echo)
	options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)

	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil)
	if err != nil {
		t.Error(err)
		return
	}

	testSimSimple(t, ctx, conn, prompt)
}

func TestTelnetSimWithEnable(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithEnable("ABC>", "enable", "password:", "testsx", "", "abc#", sshd.Echo)
	//options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)
	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimWithEnable(t, ctx, conn, prompt, "enable", "testsx")
}

func TestTelnetSimWithEnableNonePassword(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithEnable("ABC>", "enable", "password:", "<<none>>", "", "abc#", sshd.Echo)
	//options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}
	testSimWithEnable(t, ctx, conn, prompt, "enable", "<<none>>")
}

func TestTelnetSimWithEnableEmptyPassword(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithEnable("ABC>", "enable", "password:", "<<empty>>", "", "abc#", sshd.Echo)
	//options.WithNoEnable("ABC>", telnetd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimWithEnable(t, ctx, conn, prompt, "enable", "<<empty>>")
}

func TestTelnetSimWithYesNo(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "","abc#", sshd.Echo)
	options.WithQuest("abc? [Y/N]:", "N", "ABC>", sshd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimSimple(t, ctx, conn, prompt)
}

func TestTelnetSimWithEnableWithYesNo(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithQuest("abc? [Y/N]:", "N", "ABC>",
		sshd.WithEnable("enable", "password:", "testsx", "", "abc#", sshd.Echo))
	//options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimWithEnable(t, ctx, conn, prompt, "enable", "testsx")
}
