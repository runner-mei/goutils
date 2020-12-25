package harness

import (
	"bytes"
	"cn/com/hengwei_private/sim/telnetd"
	"context"
	"fmt"
	_ "net/http/pprof"
	"strings"
	"testing"

	"github.com/runner-mei/goutils/shell"
)

var h3cConfigurations = []string{
	`#
 sysname H3C
#
 super password level 3 cipher $c$3$/MfGj/XsuoyLY4gwyW7wVWNz1b0Vig==
#
 level 3
#
`,
	`acl number 2000
#
 voice vlan mac-address 0001-e300-0000 mask ffff-ff00-0000
`,
	`#
 snmp-agent
 authentication-mode scheme
 protocol inbound ssh`,
	`#
return
`,
}

func TestH3C(t *testing.T) {

	// go http.ListenAndServe(":12445", nil)

	moreAfter := append([]byte{0x1b}, []byte("[42D                                          ")...)
	moreAfter = append(moreAfter, 0x1b)
	moreAfter = append(moreAfter, []byte("[42D")...)

	welcome := []byte{0xFF, 0xFB, 0x01, 0xFF, 0xFB, 0x01, 0xFF, 0xFB, 0x01, 0xFF, 0xFB, 0x03, 0xFF, 0xFD, 0x18, 0xFF, 0xFD, 0x1F, 0x0D, 0x0A}
	welcome = append(welcome, []byte("********************************************************************************\r\n"+
		"*  Copyright(c) 2004-2014 Hangzhou H3C Tech. Co., Ltd. All rights reserved.    *\r\n"+
		"*  Without the owner's prior written consent,                                  *\r\n"+
		"*  no decompiling or reverse-engineering shall be allowed.                     *\r\n"+
		"********************************************************************************\r\n")...)
	welcome = append(welcome, 0xFF, 0xFE, 0x1F, 0xFF, 0xFA, 0x18, 0x01, 0xFF, 0xF0, 0xFF, 0xFA, 0x18, 0x01, 0xFF, 0xF0)
	welcome = append(welcome, []byte("\r\n\r\nLogin authentication\r\n\r\n")...)

	options := &telnetd.Options{}
	options.SetWelcome(welcome)
	options.SetUserQuest(append([]byte("Username: "), 0xFF, 0xFA, 0x18, 0x01, 0xFF, 0xF0), []byte("Password:"))
	options.SetUserPassword("admin1", "admin2")

	options.WithPrompt([]byte("<H3C>"),
		telnetd.WithEnable("super", "Password: ", "admin", "User privilege level is 3, and only those commands can be used \r\nwhose level is equal or less than this.\r\nPrivilege note: 0-VISIT, 1-MONITOR, 2-SYSTEM, 3-MANAGE", "<H3C>", telnetd.OS(telnetd.Commands{
			"display": telnetd.WithCommands(telnetd.Commands{
				"current-configuration": telnetd.WithMore(h3cConfigurations, []byte(" ---- More ----"), moreAfter),
			}),
		})))

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	params := &TelnetParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "admin1",
		// PasswordQuest: "",
		Password:            "admin2",
		Prompt:              "",
		EnableCommand:       "super",
		EnablePasswordQuest: "",
		EnablePassword:      "admin",
		EnablePrompt:        "",
		UseCRLF:             true,
	}

	testTelnetH3C(t, ctx, params)
}

func TestH3CWithNonePassword(t *testing.T) {

	// go http.ListenAndServe(":12445", nil)

	moreAfter := append([]byte{0x1b}, []byte("[42D                                          ")...)
	moreAfter = append(moreAfter, 0x1b)
	moreAfter = append(moreAfter, []byte("[42D")...)

	welcome := []byte{0xFF, 0xFB, 0x01, 0xFF, 0xFB, 0x01, 0xFF, 0xFB, 0x01, 0xFF, 0xFB, 0x03, 0xFF, 0xFD, 0x18, 0xFF, 0xFD, 0x1F, 0x0D, 0x0A}
	welcome = append(welcome, []byte("********************************************************************************\r\n"+
		"*  Copyright(c) 2004-2014 Hangzhou H3C Tech. Co., Ltd. All rights reserved.    *\r\n"+
		"*  Without the owner's prior written consent,                                  *\r\n"+
		"*  no decompiling or reverse-engineering shall be allowed.                     *\r\n"+
		"********************************************************************************\r\n")...)
	welcome = append(welcome, 0xFF, 0xFE, 0x1F, 0xFF, 0xFA, 0x18, 0x01, 0xFF, 0xF0, 0xFF, 0xFA, 0x18, 0x01, 0xFF, 0xF0)
	welcome = append(welcome, []byte("\r\n\r\nLogin authentication\r\n\r\n")...)

	options := &telnetd.Options{}
	options.SetWelcome(welcome)
	options.SetUserQuest(append([]byte("Username: "), 0xFF, 0xFA, 0x18, 0x01, 0xFF, 0xF0), []byte("Password:"))
	options.SetUserPassword("admin1", "admin2")

	options.WithPrompt([]byte("<H3C>"),
		telnetd.WithEnable("super", "Password: ", "<<none>>", "User privilege level is 3, and only those commands can be used \r\nwhose level is equal or less than this.\r\nPrivilege note: 0-VISIT, 1-MONITOR, 2-SYSTEM, 3-MANAGE", "<H3C>", telnetd.OS(telnetd.Commands{
			"display": telnetd.WithCommands(telnetd.Commands{
				"current-configuration": telnetd.WithMore(h3cConfigurations, []byte(" ---- More ----"), moreAfter),
			}),
		})))

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	params := &TelnetParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "admin1",
		// PasswordQuest: "",
		Password:            "admin2",
		Prompt:              "",
		EnableCommand:       "super",
		EnablePasswordQuest: "",
		EnablePassword:      "<<none>>",
		EnablePrompt:        "",
		UseCRLF:             true,
	}

	testTelnetH3C(t, ctx, params)
}

func testTelnetH3C(t *testing.T, ctx context.Context, params *TelnetParam) {
	var buf bytes.Buffer
	c, prompt, err := DailTelnet(ctx, params, ClientWriter(&buf), ServerWriter(&buf), Question(AbcQuestion.Prompts(), AbcQuestion.Do()))

	if err != nil {
		t.Error(err)
		// t.Error(buf.Len(), buf.String())

		s := shell.ToHexStringIfNeed(buf.Bytes())
		t.Log(s)
		fmt.Println(s)
		return
	}

	conn := &Shell{Conn: c, Prompt: prompt}
	defer conn.Close()

	result, err := Exec(ctx, conn, "display current-configuration")
	if err != nil {
		t.Error(err)
		return
	}
	conn.Close()

	if !strings.Contains(result.Incomming, "super password level 3 cipher $c$3$/MfGj/XsuoyLY4gwyW7wVWNz1b0Vig==") {
		t.Errorf("want 'super password level 3 cipher' got %s", result.Incomming)
	}
	t.Log(result.Incomming)

	s := shell.ToHexStringIfNeed(buf.Bytes())
	t.Log(s)
	//fmt.Println(s)
}
