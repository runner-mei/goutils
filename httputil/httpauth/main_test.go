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
		WelcomeURL: "/xxxx/xx/sso/login",
		LoginURL:   "/xxxx/xx/sso/login",
		Username:   "admin",
		Password:   "Admin",
		ReadForm:   true,
	}

	client := New()
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
	client = New()
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
	}
}
