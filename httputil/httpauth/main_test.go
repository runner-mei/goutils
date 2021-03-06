package httpauth

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestLoginMy(t *testing.T) {
	params := &LoginParams{
		Protocol:   "https",
		Address:    "127.0.0.1",
		WelcomeURL: "/hengwei/sso/login",
		LoginURL:   "/hengwei/sso/login",
		Username:   "admin",
		Password:   "Admin",
		ReadForm:   true,

		LogoutURL:             "/hengwei/sso/logout",
		ExceptedLogoutContent: "登出已成功",
	}

	client := New("", "")
	var out bytes.Buffer
	resp, msgs, err := Login(nil, &client, params, &out)
	if err != nil {
		t.Log(msgs)
		t.Error(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Log(out.String())
		t.Error(err)
		return
	}

	if !bytes.Contains(body, []byte("登录已成功")) {
		t.Log(out.String())
		t.Error(string(body))
	}

	params.WelcomeURL = ""
	client = New("", "")
	out.Reset()
	resp, msgs, err = Login(nil, &client, params, &out)
	if err != nil {
		t.Log(msgs)
		t.Log(out.String())
		t.Error(err)
		return
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Log(out.String())
		t.Error(err)
		return
	}

	if !bytes.Contains(body, []byte("登录已成功")) {
		t.Error(string(body))
		t.Log(out.String())
	}

	resp, msgs, err = Logout(nil, &client, params, &out)
	if err != nil {
		t.Log(msgs)
		t.Log(out.String())
		t.Error(err)
		return
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Log(out.String())
		t.Error(err)
		return
	}

	if !bytes.Contains(body, []byte("登出已成功")) {
		t.Error(string(body))
		t.Log(out.String())
	}
}

func TestLoginHPIlo(t *testing.T) {
	t.Skip("hp ilo2")
	params := &LoginParams{
		Protocol:        "https",
		Address:         "192.168.1.15",
		WelcomeURL:      "https://192.168.1.15/",
		LoginURL:        "https://192.168.1.15/",
		UsernameArgname: "UN",
		Username:        "Administrator",
		PasswordArgname: "PW",
		ExceptedContent: "",
		Password:        "iLO 2 Log",
		ReadForm:        true,
	}

	client := New("", "tls10")

	var out bytes.Buffer
	resp, msgs, err := Login(nil, &client, params, &out)
	if err != nil {
		t.Log(out.String())
		t.Log(msgs)
		t.Error(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Log(out.String())
		t.Error(err)
		return
	}

	if !bytes.Contains(body, []byte("登录已成功")) {
		t.Log(out.String())
		t.Error(string(body))
	}
}
