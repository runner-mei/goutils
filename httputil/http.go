package httputil

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"io/ioutil"
	"net/http"
	nhttputil "net/http/httputil"
	"net/url"

	"github.com/runner-mei/goutils/netutil"
	"github.com/runner-mei/goutils/util"
	"github.com/runner-mei/goutils/crypto"
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

func Do(req *http.Request) (resp *http.Response, err error) {
	return InsecureHttpClent.Do(req)
}

func Dump(dumpOut io.Writer, reqPrefix string, req *http.Request, reqBody io.Reader, respPrefix string, resp *http.Response, respBody io.Reader) {
	if dumpOut == nil {
		return
	}

	io.WriteString(dumpOut, reqPrefix)
	if bs, e := nhttputil.DumpRequest(req, false); nil != e {
		io.WriteString(dumpOut, e.Error())
	} else {
		dumpOut.Write(bs)
		if reqBody != nil {
			io.Copy(dumpOut, reqBody)
			dumpOut.Write([]byte("\r\n"))
		}
	}

	io.WriteString(dumpOut, respPrefix)
	if bs, e := nhttputil.DumpResponse(resp, false); nil != e {
		io.WriteString(dumpOut, e.Error())
	} else {
		dumpOut.Write(bs)

		if respBody != nil {
			io.Copy(dumpOut, respBody)
			dumpOut.Write([]byte("\r\n"))
		} else {

			var body []byte
			switch resp.Header.Get("Content-Encoding") {
			case "gzip":
				reader, _ := gzip.NewReader(resp.Body)
				defer reader.Close()
				body, e = ioutil.ReadAll(reader)
			default:
				body, e = ioutil.ReadAll(resp.Body)
			}
			if e != nil {
				io.WriteString(dumpOut, "***")
				io.WriteString(dumpOut, e.Error())
			} else {
				dumpOut.Write(body)
				dumpOut.Write([]byte("\r\n"))

				resp.Body = util.ToReadCloser(bytes.NewReader(body))
			}
		}
		// dumpOut.Write(body)
	}
}

func EncryptWrap(key string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newWriter := NewResponse(w)
		handler.ServeHTTP(newWriter, r)

		if newWriter.Buffer.Len() > 0 {
			bs, err := crypto.Encrypt([]byte(key), newWriter.Buffer.Bytes())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, err.Error())
			} else {
				w.WriteHeader(newWriter.Status)
				_, err = w.Write(bs)
				if err != nil {
					log.Println(err)
				}
			}
		}
	})
}



