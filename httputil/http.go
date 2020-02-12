package httputil

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	nhttputil "net/http/httputil"
	"net/url"

	"github.com/runner-mei/goutils/netutil"
	"github.com/runner-mei/goutils/util"
	"github.com/runner-mei/resty"
)

var InsecureHttpTransport = resty.InsecureHttpTransport
var InsecureHttpClent = resty.InsecureHttpClent

func init() {
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		t.DialContext = netutil.WrapDialContext(t.DialContext)
		InsecureHttpTransport.DialContext = t.DialContext
	}
}

func Get(url string) (resp *http.Response, err error) {
	return InsecureHttpClent.Get(url)
}

func Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	return InsecureHttpClent.Post(url, contentType, body)
}

func PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return InsecureHttpClent.PostForm(url, data)
}

func Head(url string) (resp *http.Response, err error) {
	return InsecureHttpClent.Head(url)
}

func Dump(dumpOut io.Writer, reqPrefix string, req *http.Request, respPrefix string, resp *http.Response) {
	if dumpOut == nil {
		return
	}

	io.WriteString(dumpOut, reqPrefix)
	if bs, e := nhttputil.DumpRequest(req, false); nil != e {
		io.WriteString(dumpOut, e.Error())
	} else {
		dumpOut.Write(bs)
	}

	io.WriteString(dumpOut, respPrefix)
	if bs, e := nhttputil.DumpResponse(resp, false); nil != e {
		io.WriteString(dumpOut, e.Error())
	} else {
		dumpOut.Write(bs)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			io.WriteString(dumpOut, "***")
			io.WriteString(dumpOut, err.Error())
		} else {
			dumpOut.Write(body)
			dumpOut.Write([]byte("\r\n"))

			resp.Body = util.ToReadCloser(bytes.NewReader(body))
		}
		// dumpOut.Write(body)
	}
}
