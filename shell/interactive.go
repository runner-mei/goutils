package shell

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"unicode"

	"github.com/runner-mei/errors"
	"github.com/runner-mei/goutils/util"
)

const YesOrNo = "? [Y/N]:"

// MorePrompts 为 terminal 中 more 的各种格式
var MorePrompts = [][]byte{[]byte("- More -"),
	[]byte("-- More --"),
	[]byte("- more -"),
	[]byte("-- more --"),
	[]byte("-More-"),
	[]byte("--More--"),
	[]byte("-more-"),
	[]byte("--more--"),
	[]byte("-MORE-"),
	[]byte("--MORE--"),
	[]byte("- MORE -"),
	[]byte("-- MORE --"),
}

var (
	h3cSuperResponse  = []byte("User privilege level is")
	anonymousPassword = []byte("<<anonymous>>")
	nonePassword      = []byte("<<none>>")
	emptyPassword     = []byte("<<empty>>")

	defaultUserPrompts     = [][]byte{[]byte("Username:"), []byte("username:"), []byte("login:"), []byte("login as:")}
	defaultPasswordPrompts = [][]byte{[]byte("Password:"), []byte("password:")}
	defaultPrompts         = [][]byte{[]byte(">"), []byte("$"), []byte("#")}
	defaultErrorPrompts    = [][]byte{[]byte("Bad secrets"), []byte("Login invalid"), []byte("Access denied"), []byte("Login failed!"), []byte("Error:")}
	defaultEnableCmd       = []byte("enable")
)

var SayYesCRLF = func(conn Conn, idx int) (bool, error) {
	conn.Sendln([]byte("y"))
	return true, nil
}
var SayNoCRLF = func(conn Conn, idx int) (bool, error) {
	conn.Sendln([]byte("N"))
	return true, nil
}

var SayYes = func(conn Conn, idx int) (bool, error) {
	conn.Send([]byte("y"))
	return true, nil
}
var SayNo = func(conn Conn, idx int) (bool, error) {
	conn.Send([]byte("N"))
	return true, nil
}

var ReturnOK = func(conn Conn, idx int) (bool, error) {
	return false, nil
}

var (
	StorKeyInCache  = Match("Store key in cache? (y/n)", SayYes)
	More            = Match(MorePrompts, SayYes)
	UpdateCachedKey = Match("Update cached key? (y/n, Return cancels connection)", SayYes)

	DefaultMatchers = []Matcher{
		StorKeyInCache,
		UpdateCachedKey,
		More,
	}
)

func IsNonePassword(password []byte) bool {
	return bytes.Equal(password, nonePassword) || bytes.Equal(password, anonymousPassword)
}

func IsEmptyPassword(password []byte) bool {
	return bytes.Equal(password, emptyPassword)
}

type Matcher interface {
	Strings() []string
	Prompts() [][]byte
	Do() DoFunc
}

type stringMatcher struct {
	prompts []string
	do      DoFunc
}

func (s *stringMatcher) Strings() []string {
	return s.prompts
}

func (s *stringMatcher) Prompts() [][]byte {
	var prompts [][]byte
	for idx := range s.prompts {
		prompts = append(prompts, []byte(s.prompts[idx]))
	}
	return prompts
}

func (s *stringMatcher) Do() DoFunc {
	return s.do
}

type bytesMatcher struct {
	prompts [][]byte
	do      DoFunc
}

func (s *bytesMatcher) Strings() []string {
	var prompts []string
	for idx := range s.prompts {
		prompts = append(prompts, string(s.prompts[idx]))
	}
	return prompts
}

func (s *bytesMatcher) Prompts() [][]byte {
	return s.prompts
}

func (s *bytesMatcher) Do() DoFunc {
	return s.do
}

func Match(prompts interface{}, cb func(Conn, int) (bool, error)) Matcher {
	switch values := prompts.(type) {
	case []string:
		return &stringMatcher{
			prompts: values,
			do:      cb,
		}
	case [][]byte:
		return &bytesMatcher{
			prompts: values,
			do:      cb,
		}
	case []byte:
		return &bytesMatcher{
			prompts: [][]byte{values},
			do:      cb,
		}
	case string:
		return &bytesMatcher{
			prompts: [][]byte{[]byte(values)},
			do:      cb,
		}
	default:
		panic(fmt.Errorf("want []string or [][]byte got %T", prompts))
	}
}

const maxRetryCount = 100

func Expect(ctx context.Context, conn Conn, matchs ...Matcher) error {
	var matchIdxs = make([]int, 0, len(matchs)+len(DefaultMatchers))
	var prompts = make([][]byte, 0, len(matchs)+len(DefaultMatchers))

	for idx := range matchs {
		matchIdxs = append(matchIdxs, len(prompts))
		prompts = append(prompts, matchs[idx].Prompts()...)
	}
	for idx := range DefaultMatchers {
		matchIdxs = append(matchIdxs, len(prompts))
		prompts = append(prompts, DefaultMatchers[idx].Prompts()...)
	}

	more := false
	for retryCount := 0; retryCount < maxRetryCount; retryCount++ {
		idx, recvBytes, err := conn.Expect(prompts)
		if err != nil {
			err = errors.Wrap(err, "read util '"+string(bytes.Join(prompts, []byte(",")))+"' failed")
			return errors.WrapWithSuffix(err, "\r\n"+ToHexStringIfNeed(recvBytes))
		}

		foundMatchIndex := -1

		for i := 0; i < len(matchIdxs); i++ {
			if matchIdxs[i] <= idx && (i == len(matchIdxs)-1 || idx < matchIdxs[i+1]) {
				foundMatchIndex = i
				break
			}
		}

		if foundMatchIndex < 0 {
			return errors.New("read util '" + string(bytes.Join(prompts, []byte(","))) + "' failed, return index is '" + strconv.Itoa(idx) + "'")
		}

		var cb DoFunc
		if len(matchs) > foundMatchIndex {
			cb = matchs[foundMatchIndex].Do()
		} else {
			cb = DefaultMatchers[foundMatchIndex-len(matchs)].Do()
		}
		more, err = cb(conn, idx-matchIdxs[foundMatchIndex])
		if err != nil {
			return err
		}

		if !more {
			return nil
		}
	}

	return errors.New("read util '" + string(bytes.Join(prompts, []byte(","))) + "' failed, retry count > " + strconv.FormatInt(maxRetryCount, 10))
}

func UserLogin(ctx context.Context, conn Conn, userPrompts [][]byte, username []byte, passwordPrompts [][]byte, password []byte, prompts [][]byte, matchs ...Matcher) ([]byte, error) {
	if len(userPrompts) == 0 {
		userPrompts = defaultUserPrompts
	}
	if len(passwordPrompts) == 0 {
		passwordPrompts = defaultPasswordPrompts
	}
	if len(prompts) == 0 {
		prompts = defaultPrompts
	}

	var buf bytes.Buffer
	cancel := conn.SetTeeOutput(&buf)
	defer cancel()

	status := 0

	copyed := make([]Matcher, len(matchs)+4)
	copyed[0] = Match(userPrompts, func(c Conn, nidx int) (bool, error) {
		if e := conn.Sendln(username); e != nil {
			return false, errors.Wrap(e, "send username failed")
		}
		status = 1
		return false, nil
	})
	copyed[1] = Match(passwordPrompts, func(c Conn, nidx int) (bool, error) {
		if IsEmptyPassword(password) {
			password = []byte{}
		}
		if e := conn.SendPassword(password); e != nil {
			return false, errors.Wrap(e, "send user password failed")
		}

		status = 2
		return false, nil
	})
	copyed[2] = Match(prompts, func(c Conn, nidx int) (bool, error) {
		status = 3
		return false, nil
	})

	copyed[3] = Match(defaultErrorPrompts, func(c Conn, nidx int) (bool, error) {
		status = 4
		return false, nil
	})

	copy(copyed[4:], matchs)

	for i := 0; ; i++ {
		if i >= 10 {
			return nil, errors.New("user login fail:\r\n" + ToHexStringIfNeed(buf.Bytes()))
		}
		err := Expect(ctx, conn, copyed...)
		if err != nil {
			return nil, errors.Wrap(err, "user login fail")
		}

		if status == 3 {
			received := buf.Bytes()
			if len(received) == 0 {
				return nil, errors.New("read prompt failed, received is empty")
			}

			prompt := GetPrompt(received, prompts)
			if len(prompt) == 0 {
				return nil, errors.New("read prompt '" + string(bytes.Join(prompts, []byte(","))) + "' failed: \r\n" + ToHexStringIfNeed(received))
			}
			return prompt, nil
		}

		if status == 4 {
			received := buf.Bytes()
			if len(received) == 0 {
				return nil, errors.New("invalid password")
			}

			return nil, errors.New("invalid password: \r\n" + ToHexStringIfNeed(received))
		}
	}
}

func ReadPrompt(ctx context.Context, conn Conn, prompts [][]byte, matchs ...Matcher) ([]byte, error) {
	if len(prompts) == 0 {
		prompts = defaultPrompts
	}

	var buf bytes.Buffer
	cancel := conn.SetTeeOutput(&buf)
	defer cancel()

	isPrompt := false

	copyed := make([]Matcher, len(matchs)+1)
	copyed[0] = Match(prompts, func(conn Conn, idx int) (bool, error) {
		isPrompt = true
		return false, nil
	})
	copy(copyed[1:], matchs)

	for retryCount := 0; ; retryCount++ {
		if retryCount >= 10 {
			return nil, errors.New("read prompt failed, retry count > 10")
		}
		e := Expect(ctx, conn, copyed...)
		if nil != e {
			return nil, e
		}

		if isPrompt {
			break
		}
	}

	received := buf.Bytes()
	if len(received) == 0 {
		return nil, errors.New("read prompt failed, received is empty")
	}

	prompt := GetPrompt(received, prompts)
	if len(prompt) == 0 {
		return nil, errors.New("read prompt '" + string(bytes.Join(prompts, []byte(","))) + "' failed: \r\n" + ToHexStringIfNeed(received))
	}
	return prompt, nil
}

func GetPrompt(bs []byte, prompts [][]byte) []byte {
	if len(bs) == 0 {
		return nil
	}
	lines := util.SplitLines(bs)

	var fullPrompt []byte
	for i := len(lines) - 1; i >= 0; i-- {
		fullPrompt = bytes.TrimFunc(lines[i], func(r rune) bool {
			if r == 0 {
				return true
			}
			return unicode.IsSpace(r)
		})
		if len(fullPrompt) > 0 {
			break
		}
	}

	for _, prompt := range prompts {
		if bytes.HasSuffix(fullPrompt, prompt) {
			if 2 <= len(fullPrompt) && ']' == fullPrompt[len(fullPrompt)-2] {
				idx := bytes.LastIndex(fullPrompt, []byte("["))
				if idx > 0 {
					fullPrompt = fullPrompt[idx:]
				}
			}
			return fullPrompt
		}
	}

	return nil
}

func Exec(ctx context.Context, conn Conn, prompt, cmd []byte) ([]byte, error) {
	if len(prompt) == 0 {
		return nil, errors.New("prompt is missing")
	}
	if len(cmd) == 0 {
		return nil, errors.New("cmd is missing")
	}

	if bytes.HasPrefix(prompt, []byte("\\n")) {
		prompt[1] = '\n'
		prompt = prompt[1:]
	}

	var buf bytes.Buffer
	cancel := conn.SetTeeOutput(&buf)
	defer cancel()

	err := conn.Sendln(cmd)
	if err != nil {
		return nil, err
	}

	err = Expect(ctx, conn, Match(prompt, func(Conn, int) (bool, error) {
		return false, nil
	}))

	if err != nil {
		return nil, err
	}
	bs := buf.Bytes()
	bs = bs[:len(bs)-len(prompt)]
	return bs, nil
}

func WithEnable(ctx context.Context, conn Conn, enableCmd []byte, passwordPrompts [][]byte, password []byte, enablePrompts [][]byte) ([]byte, error) {
	if len(enableCmd) == 0 {
		enableCmd = defaultEnableCmd
	}

	if e := conn.Sendln(enableCmd); nil != e {
		return nil, errors.Wrap(e, "send enable '"+string(enableCmd)+"' failed")
	}

	if len(passwordPrompts) == 0 {
		passwordPrompts = defaultPasswordPrompts
	}
	if len(enablePrompts) == 0 {
		enablePrompts = defaultPrompts
	}

	// fmt.Println("send enable '" + string(enableCmd) + "' ok and read enable password prompt")

	if !IsNonePassword(password) {
		err := Expect(ctx, conn,
			Match(append(passwordPrompts, h3cSuperResponse), func(c Conn, nidx int) (bool, error) {
				if IsEmptyPassword(password) {
					password = []byte{}
				}
				if e := conn.SendPassword(password); e != nil {
					return false, errors.Wrap(e, "send enable password failed")
				}
				return false, nil
			}))
		if err != nil {
			return nil, err
		}
	}
	return ReadPrompt(ctx, conn, enablePrompts)
}
