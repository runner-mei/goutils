package httpauth

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptrace"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/runner-mei/goutils/httputil"
	"github.com/runner-mei/goutils/urlutil"
	"github.com/runner-mei/goutils/util"
	rutil "github.com/runner-mei/resty/util"
	"golang.org/x/net/html"
)

type LoginParams struct {
	Timeout  time.Duration `json:"timeout,omitempty"`
	Protocol string        `json:"protocol,omitempty"`
	Address  string        `json:"address,omitempty"`
	Port     int64         `json:"port,omitempty"`

	// 登录方法， 可取值 baseauth 和 web, 或 无
	// 值为无时，下面的参数不用填，应该隐藏
	// 值为 baseauth 时，下面的参数只要填Username和Password，其它应该隐藏
	// 值为 web 时，下面的参数全部要填
	AuthMethod string `json:"auth_meth,omitempty"`

	// 可选，登录时需要访问的页面的 http 方法， 缺省值为 GET
	WelcomeMethod string `json:"welcome_method,omitempty"`
	// 可选，登录时需要访问的页面的 URL
	WelcomeURL string `json:"welcome_url,omitempty"`
	// 可选，登录时发送用户信息的页面 http 方法， 缺省值为 POST
	LoginMethod string `json:"login_method,omitempty"`
	// 必选选，登录时发送用户信息的页面 URL
	LoginURL string `json:"login_url,omitempty"`
	// 可选，登录时发送用户名时的字段名， 缺省值为 username
	UsernameArgname string `json:"user_arg_name,omitempty"`
	// 可选，登录时发送密码时的字段名， 缺省值为 password
	PasswordArgname string `json:"password_arg_name,omitempty"`
	// 必选，用户名
	Username string `json:"username,omitempty"`
	// 必选，密码
	Password string `json:"password,omitempty"`
	// 可选，密码加密方式，可取值 base64
	PasswordCrypto string `json:"password_crypto,omitempty"`
	// 可选，是否解析 WelcomeURL 的返回页面
	ReadForm bool `json:"readform,omitempty"`
	// 可选， WelcomeURL 的返回页面中 form 的 selection
	FormLocation string `json:"form_location,omitempty"`
	// 可选值，登录请求的 Content-Type, 可取值， json, urlencoded, 缺省值 urlencoded
	ContentType string `json:"content_type,omitempty"`
	// 可选， 登录请求的其它参数
	Values map[string]string `json:"values,omitempty"`
	// 可选， 登录请求的 header
	Headers map[string]string `json:"headers,omitempty"`
	// 可选， 登录成功时的返回状态
	ExceptedStatusCode int `json:"excepted_status_code,omitempty"`
	// 可选， 登录成功时的返回内容
	ExceptedContent string `json:"excepted_content,omitempty"`

	AutoRedirectEnabled string `json:"auto_redirect_enabled"`
	AutoRedirectURL     string `json:"auto_redirect_url"`
	Referrer            string `json:"referrer"`

	LogoutMethod          string `json:"logout_method,omitempty"`
	LogoutURL             string `json:"logout_url,omitempty"`
	LogoutContentType     string `json:"logout_content_type,omitempty"`
	LogoutBody            string `json:"logout_body,omitempty"`
	ExceptedLogoutStatus  int    `json:"excepted_logout_status_code,omitempty"`
	ExceptedLogoutContent string `json:"excepted_logout_content,omitempty"`
}

func (params *LoginParams) BaseURL() string {
	protocol := params.Protocol
	if protocol == "" {
		protocol = "http"
	}

	address := params.Address
	if params.Port != 0 {
		address = net.JoinHostPort(address, strconv.FormatInt(params.Port, 10))
	}
	return protocol + "://" + address
}

func WithClientTrace(ctx context.Context, dumpOut io.Writer) context.Context {
	if dumpOut == nil {
		return ctx
	}

	var trace *httptrace.ClientTrace
	if ctx != nil {
		trace = httptrace.ContextClientTrace(ctx)
		if trace != nil {
			return ctx
		}
	}

	trace = &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			fmt.Fprintf(dumpOut, "Got Conn: %+v\r\n", connInfo)
		},
		WroteHeaderField: func(key string, value []string) {
			fmt.Fprintf(dumpOut, "WroteHeaderField: %s:%v\r\n", key, value)
		},
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return httptrace.WithClientTrace(ctx, trace)
}

func NewTransport(insecureSkipVerify bool, minTlsVersion, maxTlsVersion string) (*http.Transport, bool) {
	min := parseTlsVersion(minTlsVersion)
	max := parseTlsVersion(maxTlsVersion)

	if min == 0 && max == 0 {
		return httputil.InsecureHttpTransport, false
	}
	cfg := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}

	if min > 0 {
		cfg.MinVersion = min
	}

	if max > 0 {
		cfg.MaxVersion = max
	}

	return &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: cfg,
	}, true
}

func New(minTlsVersion, maxTlsVersion string) http.Client {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	min := parseTlsVersion(minTlsVersion)
	max := parseTlsVersion(maxTlsVersion)

	transport := httputil.InsecureHttpTransport
	if min > 0 || max > 0 {
		cfg := &tls.Config{
			InsecureSkipVerify: true,
		}
		if transport.TLSClientConfig != nil {
			*cfg = *transport.TLSClientConfig
		}
		cfg.MinVersion = min
		cfg.MaxVersion = max

		transport.TLSClientConfig = cfg
	}
	return http.Client{
		Transport: transport,
		Jar:       cookieJar,
	}
}

func NewWithTransport(transport *http.Transport) http.Client {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	return http.Client{
		Transport: transport,
		Jar:       cookieJar,
	}
}

func parseTlsVersion(s string) uint16 {
	switch s {
	case "tls10":
		return tls.VersionTLS10
	case "tls11":
		return tls.VersionTLS11
	case "tls12":
		return tls.VersionTLS12
	case "tls13":
		return tls.VersionTLS13
	default:
		return 0
	}
}

func readWelcome(ctx context.Context, client *http.Client, params *LoginParams, dumpOut io.Writer) (*http.Response, string, string, url.Values, []string, error) {
	baseurl := params.BaseURL()

	action := http.MethodGet
	if params.WelcomeMethod != "" {
		action = params.WelcomeMethod
	}

	last := params.WelcomeURL
	rawWelcomeURL := params.WelcomeURL

	var logMessages []string
	for retry := 0; retry < 10; retry++ {
		welcomeURL := strings.ToLower(rawWelcomeURL)
		if strings.HasPrefix(welcomeURL, "https://") || strings.HasPrefix(welcomeURL, "http://") {
			welcomeURL = rawWelcomeURL
		} else if retry == 0 {
			welcomeURL = urlutil.Join(baseurl, rawWelcomeURL)
		} else {
			if strings.HasPrefix(welcomeURL, "/") {
				welcomeURL = urlutil.Join(baseurl, last, rawWelcomeURL)
			} else {
				welcomeURL = urlutil.Join(baseurl, path.Dir(last), rawWelcomeURL)
			}
		}

		welcomeReq, err := http.NewRequest(action, welcomeURL, nil)
		if err != nil {
			return nil, "", "", nil, []string{"创建登录首页请求失败", err.Error()}, err
		}
		for key, value := range params.Headers {
			welcomeReq.Header.Set(key, value)
		}
		welcomeReq = welcomeReq.WithContext(ctx)
		if params.Referrer == "" {
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				params.Referrer = req.URL.String()
				if req.URL.User != nil {
					// This is not very efficient, but is the best we can
					// do without:
					// - introducing a new method on URL
					// - creating a race condition
					// - copying the URL struct manually, which would cause
					//   maintenance problems down the line
					auth := req.URL.User.String() + "@"
					params.Referrer = strings.Replace(params.Referrer, auth, "", 1)
				}
				return nil
			}
		}

		welcomeResp, err := client.Do(welcomeReq)
		if nil != err {
			return nil, "","", nil, []string{"创建登录首页请求失败", err.Error()}, err
		}

		client.CheckRedirect = nil

		logMessages = append(logMessages, "访问"+rawWelcomeURL+" 成功")

		welcomeResp, err = rutil.WrapUncompress(welcomeResp, false)
		if nil != err {
			return nil,"", "", nil, []string{"判断登录首页响应是否要解压失败", err.Error()}, err
		}

		body, err := ioutil.ReadAll(welcomeResp.Body)
		welcomeResp.Body.Close()
		if err != nil {
			return nil, "","", nil, append(logMessages, "读登录首页内容失败", err.Error()), err
		}

		httputil.Dump(dumpOut,
			"========= 1 DumpRequest =========\r\n", welcomeReq, nil,
			"\r\n========= 1 DumpResponse =========\r\n", welcomeResp, bytes.NewReader(body))

		if http.StatusOK != welcomeResp.StatusCode {
			return nil, "","", nil, append(logMessages, "登录首页的响应码不正确"),
				errors.New("status code is '" + welcomeResp.Status + "' - " + string(body))
		}

		if !params.ReadForm {
			return nil, "","", nil, append(logMessages, "跳过登录首页的解析"), nil
		}

		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if nil != err {
			return nil, "","", nil, append(logMessages, "解析登录首页失败", err.Error()), errors.New("failed to parse login page, " + err.Error())
		}

		if params.FormLocation == "" {
			params.FormLocation = "form"
		}

		var method string
		var submitURL string
		values := url.Values{}
		formCount := 0
		doc.Find(params.FormLocation).Each(func(idx int, form *goquery.Selection) {
			if nodes := form.Nodes; len(nodes) > 0 {
				action = strings.ToUpper(attributeValueWithDefaultValue(nodes[0], "method", "POST"))
				method = action
				submitURL = attributeValueWithDefaultValue(nodes[0], "action", "")
			}

			formCount++
			form.Find("input[type=\"hidden\"]").Each(func(idx int, input *goquery.Selection) {
				for _, node := range input.Nodes {
					formName := attributeValue(node, "name")
					if formName == "" {
						continue
					}
					values[formName] = []string{attributeValue(node, "value")}
				}
			})
		})
		if formCount == 1 {
			welcomeResp.Body = ioutil.NopCloser(bytes.NewReader(body))

			urlLow := strings.ToLower(submitURL)
			if !strings.HasPrefix(urlLow, "https://") && !strings.HasPrefix(urlLow, "http://") {
					submitURL = urlutil.Join(welcomeResp.Request.URL.Scheme +"://" + welcomeResp.Request.URL.Host, submitURL)
			}

			return welcomeResp, method, submitURL, values, append(logMessages, "解析登录首页时找到"+strconv.Itoa(len(values))+"个表单项"), nil
		}

		if formCount > 1 {
			if err == nil {
				return nil, "","", nil, append(logMessages, "解析登录首页时找到多个表单"), errors.New("'" + params.FormLocation + "' is muti choice")
			}
			return nil,"", "", nil, append(logMessages, "解析登录首页时找到多个表单", err.Error()), err
		}

		redirectURL := ParseJsRedirect(body, []string{
			"parent.location.href",
			"window.location.href",
			"window.location",
		})
		if redirectURL == "" {
			if err == nil {
				return nil,"","",  nil, append(logMessages, "解析登录首页时没有找到表单"), errors.New("'" + params.FormLocation + "' isn't found")
			}
			return nil, "", "", nil, append(logMessages, "解析登录首页时没有找到表单", err.Error()), err
		}

		logMessages = append(logMessages, "解析内容时发现有重定向，开始重定向")
		last = rawWelcomeURL
		rawWelcomeURL = redirectURL
		action = http.MethodGet
	}

	return nil,"", "", nil, logMessages, errors.New("重定向次数太多了")
}

func Logout(ctx context.Context, client *http.Client, params *LoginParams, dumpOut io.Writer) (*http.Response, []string, error) {
	if params.LogoutURL == "" {
		return nil, nil, nil
	}
	baseurl := params.BaseURL()

	// cookieJar, err := cookiejar.New(nil)
	// if err != nil {
	//   return nil, err
	// }

	// client := http.Client{
	//    Transport: httputil.InsecureHttpTransport,
	//    Jar: cookieJar,
	//  }

	if ctx == nil {
		ctx = context.Background()
	}
	ctx = WithClientTrace(ctx, dumpOut)

	if params.Timeout > 0 {
		var c func()
		ctx, c = context.WithTimeout(ctx, params.Timeout)
		defer c()
	}

	var logMessages []string
	var err error

	var logoutMethod = "GET"
	if params.LogoutMethod != "" {
		logoutMethod = params.LogoutMethod
	}

	logoutURL := strings.ToLower(params.LogoutURL)
	if strings.HasPrefix(logoutURL, "https://") || strings.HasPrefix(logoutURL, "http://") {
		logoutURL = params.LogoutURL
	} else {
		logoutURL = urlutil.Join(baseurl, params.LogoutURL)
	}

	var body io.Reader
	if logoutMethod != "GET" {
		body = strings.NewReader(params.LogoutBody)
	}
	logoutReq, err := http.NewRequest(logoutMethod, logoutURL, body)
	if err != nil {
		logMessages = append(logMessages, "创建登出请求失败", err.Error())
		return nil, logMessages, err
	}

	logoutReq.Header.Set("Content-Type", params.LogoutContentType)
	logoutReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	logoutReq.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	logoutReq.Header.Set("Cache-Control", "max-age=0")
	logoutReq.Header.Set("Connection", "keep-alive")

	logoutResp, err := client.Do(logoutReq)
	if nil != err {
		logMessages = append(logMessages, "发送登出请求失败", err.Error())
		return nil, logMessages, err
	}
	if logoutResp != nil {
		body := logoutResp.Body
		defer body.Close()
	}

	logoutResp, err = rutil.WrapUncompress(logoutResp, false)
	if nil != err {
		return nil, []string{"判断登出响应是否要解压失败", err.Error()}, err
	}
	logoutResp, err = rutil.WrapCharset(logoutResp, false)
	if nil != err {
		return nil, []string{"判断登出响应是否要转码失败", err.Error()}, err
	}

	httputil.Dump(dumpOut,
		"\r\n========= logout DumpRequest =========\r\n", logoutReq, strings.NewReader(params.LogoutBody),
		"\r\n========= logout DumpResponse =========\r\n", logoutResp, nil)

	if (params.ExceptedLogoutStatus == 0 && (logoutResp.StatusCode < 200 || logoutResp.StatusCode > 299)) ||
		(params.ExceptedLogoutStatus > 0 && params.ExceptedLogoutStatus != logoutResp.StatusCode) {
		body, err := ioutil.ReadAll(logoutResp.Body)
		if err != nil {
			logMessages = append(logMessages, "发送登出请求成功， 但响应码不正确")
			return nil, logMessages, errors.New("status code is '" + logoutResp.Status + "' - failed to read body")
		}

		logMessages = append(logMessages, "发送登出请求成功， 但响应码不正确")
		return nil, logMessages, errors.New("status code is '" + logoutResp.Status + "' - " + string(body))
	}

	if params.ExceptedLogoutContent != "" {
		body, err := ioutil.ReadAll(logoutResp.Body)
		if err != nil {
			logMessages = append(logMessages, "发送登出请求成功， 但读响应失败")
			return nil, logMessages, errors.New("status code is '" + logoutResp.Status + "' - failed to read body")
		}
		logoutResp.Body = util.ToReadCloser(bytes.NewReader(body))

		if !bytes.Contains(body, []byte(params.ExceptedLogoutContent)) {
			logMessages = append(logMessages, "发送登出请求成功， 但响应不包含指定的字符")
			return nil, logMessages, errors.New("excepted content not found in http body")
		}
	}

	return logoutResp, logMessages, nil
}

func Login(ctx context.Context, client *http.Client, params *LoginParams, dumpOut io.Writer) (*http.Response, []string, error) {
	baseurl := params.BaseURL()

	// cookieJar, err := cookiejar.New(nil)
	// if err != nil {
	//   return nil, err
	// }

	// client := http.Client{
	//    Transport: httputil.InsecureHttpTransport,
	//    Jar: cookieJar,
	//  }

	if ctx == nil {
		ctx = context.Background()
	}
	ctx = WithClientTrace(ctx, dumpOut)

	if params.Timeout > 0 {
		var c func()
		ctx, c = context.WithTimeout(ctx, params.Timeout)
		defer c()
	}

	values := url.Values{}
	var logMessages []string
	var err error
	var hasLoginForm bool
	var cryptoPasswordByWebjs bool
	var loginMethod string
	var welcomeResp *http.Response
	var welcomeRespBytes []byte

	if params.WelcomeURL != "" {
		var loginPostUrl string

		welcomeResp, loginMethod, loginPostUrl, values, logMessages, err = readWelcome(ctx, client, params, dumpOut)
		if err != nil {
			return nil, logMessages, err
		}
		if values == nil {
			values = url.Values{}
		}
		welcomeRespBytes, _ = ioutil.ReadAll(welcomeResp.Body)
		if bytes.Contains(welcomeRespBytes, []byte("encryptionKey")) {
			cryptoPasswordByWebjs = true
		}
		hasLoginForm = true

		if loginPostUrl != "" && strings.ToLower(params.LoginURL) == "<auto>"{
			loginPostUrlLow := strings.ToLower(loginPostUrl)
			if strings.HasPrefix(loginPostUrlLow, "https://") || strings.HasPrefix(loginPostUrlLow, "http://") {
				params.LoginURL = loginPostUrl
			} else {
				params.LoginURL = urlutil.Join(baseurl, loginPostUrl)
			}
		}
	}

	usernameform := "username"
	if params.UsernameArgname != "" {
		usernameform = params.UsernameArgname
	}
	passwordform := "password"
	if params.PasswordArgname != "" {
		passwordform = params.PasswordArgname
	}


	values[usernameform] = []string{params.Username}

	if params.PasswordCrypto == "base64" {
		values[passwordform] = []string{base64.StdEncoding.EncodeToString([]byte(params.Password))}
	} else if cryptoPasswordByWebjs {
		logMessages = append(logMessages, "需要加密密码")
		password, err := CryptoPassword(welcomeRespBytes, params.Password, dumpOut)
		if err != nil {
			if dumpOut != nil {
				io.WriteString(dumpOut, "\r\n 加密密码失败 ")
				io.WriteString(dumpOut, err.Error())
			}
			logMessages = append(logMessages, "加密密码失败")
			return nil, logMessages, err
		}
		values[passwordform] = []string{password}
	} else {
		values[passwordform] = []string{params.Password}
	}


	for k, v := range params.Values {
		values[k] = []string{v}
	}

	var action = "POST"
	if hasLoginForm {
		if loginMethod != "" {
			action = loginMethod
		}
	}
	if params.LoginMethod != "" {
		action = params.LoginMethod
	}

	loginURL := strings.ToLower(params.LoginURL)
	if strings.HasPrefix(loginURL, "https://") || strings.HasPrefix(loginURL, "http://") {
		loginURL = params.LoginURL
	} else {
		loginURL = urlutil.Join(baseurl, params.LoginURL)
	}

	fmt.Println("===", loginURL)

	return PostLogin(ctx, client, params, action, loginURL, params.ContentType, values, logMessages, 0, dumpOut)
}

func PostLogin(ctx context.Context, client *http.Client, params *LoginParams, loginMethod, loginURL, contentType string, values url.Values, logMessages []string, tryCount int, dumpOut io.Writer) (*http.Response, []string, error) {

	// loginReq, loginResp, err := Do(ctx, client, loginMethod, loginURL, contentType, values)
	// if nil != err {
	// 	return nil, err
	// }
	// defer loginResp.Body.Close()

	ctx = WithClientTrace(ctx, dumpOut)
	
	if loginMethod == "" {
		loginMethod = "POST"
	}

	var loginBody []byte
	if loginMethod != "GET" {
		switch contentType {
		case "", "urlencoded", "application/x-www-form-urlencoded":
			contentType = "application/x-www-form-urlencoded"
			s := values.Encode()
			loginBody = []byte(s)
		case "json", "application/json":
			contentType = "application/json"
			var a = make(map[string]string, len(values))
			for k, v := range values {
				a[k] = v[len(v)-1]
			}
			bs, _ := json.Marshal(a)
			loginBody = bs
		default:
			logMessages = append(logMessages, "contentType 不支持")
			return nil, logMessages, errors.New("ContentType '" + contentType + "' is unsupported")
		}
	}

	var body io.Reader
	if loginMethod != "GET" {
		body = bytes.NewReader(loginBody)
	}
	loginReq, err := http.NewRequest(loginMethod, loginURL, body)
	if err != nil {
		logMessages = append(logMessages, "创建登录请求失败", err.Error())
		return nil, logMessages, err
	}

	loginReq.Header.Set("Content-Type", contentType)
	if accept := loginReq.Header.Get("Accept"); accept == "" {
		loginReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	}
	if acceptLang := loginReq.Header.Get("Accept-Language"); acceptLang == "" {
		loginReq.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	}
	if cacheControl := loginReq.Header.Get("Cache-Control"); cacheControl == "" {
		loginReq.Header.Set("Cache-Control", "max-age=0")
	}
	if connection := loginReq.Header.Get("Connection"); connection == "" {
		loginReq.Header.Set("Connection", "keep-alive")
	}
	if params.Referrer != "" {
		loginReq.Header.Set("Referer", params.Referrer)
	}

	for k, v := range params.Headers {
		loginReq.Header.Set(k, v)
	}

	loginResp, err := client.Do(loginReq)
	if nil != err {
		logMessages = append(logMessages, "发送登录请求失败", err.Error())
		return nil, logMessages, err
	}
	if loginResp != nil {
		body := loginResp.Body
		defer body.Close()
	}

	loginResp, err = rutil.WrapUncompress(loginResp, false)
	if nil != err {
		return nil, []string{"判断登录响应是否要解压失败", err.Error()}, err
	}
	loginResp, err = rutil.WrapCharset(loginResp, false)
	if nil != err {
		return nil, []string{"判断登录响应是否要转码失败", err.Error()}, err
	}

	httputil.Dump(dumpOut,
		"\r\n========= "+strconv.Itoa(2+tryCount)+" DumpRequest =========\r\n", loginReq, bytes.NewReader(loginBody),
		"\r\n========= "+strconv.Itoa(2+tryCount)+" DumpResponse =========\r\n", loginResp, nil)

	exceptedStatusCode := params.ExceptedStatusCode
	if exceptedStatusCode == 0 {
		exceptedStatusCode = http.StatusOK
	}

	if exceptedStatusCode != loginResp.StatusCode {
		body, err := ioutil.ReadAll(loginResp.Body)
		if err != nil {
			logMessages = append(logMessages, "发送登录请求成功， 但响应码不正确")
			return nil, logMessages, errors.New("status code is '" + loginResp.Status + "' - failed to read body")
		}

		logMessages = append(logMessages, "发送登录请求成功， 但响应码不正确")
		return nil, logMessages, errors.New("status code is '" + loginResp.Status + "' - " + string(body))
	}

	if params.ExceptedContent != "" {
		//var body []byte

		// switch loginResp.Header.Get("Content-Encoding") {
		// case "gzip":
		// 	reader, e := gzip.NewReader(loginResp.Body)
		// 	if e != nil {

		// 		logMessages = append(logMessages, "发送登录请求成功， 但读响应失败:"+e.Error())
		// 		return nil, logMessages, errors.New("status code is '" + loginResp.Status + "' - failed to read body: " + e.Error())
		// 	}
		// 	defer reader.Close()
		// 	body, err = ioutil.ReadAll(reader)
		// default:
		// body, err := ioutil.ReadAll(loginResp.Body)
		//}

		// FIXME: 这里有点坑，httputil.Dump 有对内容解压缩，这里就不能再解压缩了
		body, err := ioutil.ReadAll(loginResp.Body)
		if err != nil {
			logMessages = append(logMessages, "发送登录请求成功， 但读响应失败")
			return nil, logMessages, errors.New("status code is '" + loginResp.Status + "' - failed to read body")
		}
		loginResp.Body = util.ToReadCloser(bytes.NewReader(body))

		if !bytes.Contains(body, []byte(params.ExceptedContent)) {
			if tryCount != 0 {
				logMessages = append(logMessages, "发送登录请求成功， 但响应不包含指定的字符")
				return nil, logMessages, errors.New("excepted content not found in http body")
			}

			method, posturl, values, err := ParseForm(bytes.NewReader(body))
			if err != nil {
				if bytes.Contains(body, []byte("FRAMESET")) {
					urls, err := ParseFrameset(bytes.NewReader(body))
					if err == nil {
						for _, aUrl := range urls {
							if !strings.HasPrefix(aUrl, "https://") && !strings.HasPrefix(aUrl, "http://") {
								u, _ := url.Parse(loginURL)
								u.RawQuery = ""
								if quest := strings.IndexRune(aUrl, '?'); quest >= 0 {
									u.RawQuery = aUrl[quest+1:]
									aUrl = aUrl[:quest]
								}

								if strings.HasPrefix(aUrl, "/") {
									u.Path = aUrl
								} else {
									u.Path = path.Join(path.Dir(u.Path), aUrl)
									// dumpOut.Write([]byte("\r\n===========" + path.Dir(u.Path) + "," + aUrl))
								}
								aUrl = u.String()
							}

							req, err := http.NewRequest(http.MethodGet, aUrl, nil)
							if err != nil {
								logMessages = append(logMessages, "创建登录请求失败", err.Error())
								return nil, logMessages, err
							}

							resp, err := client.Do(req)
							if nil != err {
								logMessages = append(logMessages, "发送登录请求失败", err.Error())
								return nil, logMessages, err
							}

							body, err := ioutil.ReadAll(resp.Body)
							if err != nil {
								logMessages = append(logMessages, "发送登录请求成功， 但读响应失败")
								return nil, logMessages, errors.New("status code is '" + resp.Status + "' - failed to read body")
							}

							tryCount++
							httputil.Dump(dumpOut,
								"\r\n========= "+strconv.Itoa(2+tryCount)+" DumpRequest =========\r\n", req, nil,
								"\r\n========= "+strconv.Itoa(2+tryCount)+" DumpResponse =========\r\n", resp, bytes.NewReader(body))

							if bytes.Contains(body, []byte(params.ExceptedContent)) {
								logMessages = append(logMessages, "发送登录请求成功")
								return loginResp, logMessages, nil
							}
						}

						logMessages = append(logMessages, "发送登录请求成功， 但响应不包含指定的字符")
						return nil, logMessages, errors.New("excepted content not found in http body")
					}
				}
			}

			if err != nil {
				if dumpOut != nil {
					io.WriteString(dumpOut, err.Error())
				}

				logMessages = append(logMessages, "发送登录请求成功， 但响应不包含指定的字符")
				return nil, logMessages, errors.New("excepted content not found in http body")
			}
			if strings.HasPrefix(posturl, "/") {
				u, _ := url.Parse(loginURL)
				if u != nil {
					posturl = u.Scheme + "://" + u.Host + posturl
				} else {
					posturl = params.BaseURL() + posturl
				}
			}

			logMessages = append(logMessages, "发送登录请求成功， 但响应不包含指定的字符，但表单尝试再试一次")
			return PostLogin(ctx, client, params, method, posturl, "urlencoded", values, logMessages, 1, dumpOut)
		}

		if params.AutoRedirectEnabled == "" || params.AutoRedirectEnabled == "auto" {
			for _, locationHref := range []string{"parent.location.href=\"",
				"window.location.href=\"",
				"window.location.href =\"",
				"window.location.href= \"",
				"window.location.href = \"",
				"window.location=\"",
				"window.location =\"",
				"window.location= \"",
				"window.location = \""} {
				if strings.HasPrefix(params.ExceptedContent, locationHref) {
					urlStr := strings.TrimPrefix(params.ExceptedContent, locationHref)
					urlStr = strings.TrimSuffix(strings.TrimSpace(urlStr), ";")
					urlStr = strings.TrimSuffix(urlStr, "\"")

					params.AutoRedirectURL = urlStr
					params.AutoRedirectEnabled = "true"

					if strings.HasPrefix(params.AutoRedirectURL, "/") {
						u, _ := url.Parse(loginURL)
						if u != nil {
							params.AutoRedirectURL = u.Scheme + "://" + u.Host + params.AutoRedirectURL
						} else {
							params.AutoRedirectURL = params.BaseURL() + params.AutoRedirectURL
						}
					}

					logMessages = append(logMessages, "发送登录请求成功， 在响应中读到重定向")
					break
				}
			}
		}
	}

	if params.AutoRedirectEnabled == "true" {
		req, err := http.NewRequest("GET", params.AutoRedirectURL, nil)
		if err != nil {

			logMessages = append(logMessages, "创建重定向请求失败", err.Error())
			return nil, logMessages, err
		}
		if params.Referrer != "" {
			loginReq.Header.Set("Referer", params.Referrer)
		}
		resp, err := client.Do(req)
		if err != nil {
			logMessages = append(logMessages, "发送重定向请求失败", err.Error())
			return nil, logMessages, err
		}
		if resp.Body != nil {
			defer func() {
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()
			}()
		}

		logMessages = append(logMessages, "发送重定向请求成功")

		tryCount++

		httputil.Dump(dumpOut,
			"\r\n========= "+strconv.Itoa(2+tryCount)+" DumpRequest =========\r\n", req, nil,
			"\r\n========= "+strconv.Itoa(2+tryCount)+" DumpResponse =========\r\n", resp, nil)
	}

	logMessages = append(logMessages, "发送登录请求成功")
	return loginResp, logMessages, nil
}

func attributeValue(node *html.Node, nm string) string {
	if nil == node {
		return ""
	}
	if 0 == len(node.Attr) {
		return ""
	}
	for _, attr := range node.Attr {
		if attr.Key == nm {
			return attr.Val
		}
	}
	return ""
}

func attributeValueWithDefaultValue(node *html.Node, nm, defaultValue string) string {
	v := attributeValue(node, nm)
	if "" == v {
		return defaultValue
	}
	return v
}

func ParseForm(body io.Reader) (method, posturl string, values url.Values, err error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if nil != err {
		err = errors.New("failed to parse login page, " + err.Error())
		return
	}

	//var method, posturl string
	values = url.Values{}

	hasNotHidden := false
	count := 0
	doc.Find("form").Each(func(idx int, form *goquery.Selection) {
		if nodes := form.Nodes; len(nodes) > 0 {
			method = strings.ToUpper(attributeValueWithDefaultValue(nodes[0], "method", "POST"))
			posturl = attributeValueWithDefaultValue(nodes[0], "action", "")
		}

		count++
		form.Find("input").Each(func(idx int, input *goquery.Selection) {
			for _, node := range input.Nodes {
				inputType := attributeValue(node, "type")
				if inputType != "hidden" {
					hasNotHidden = true
					continue
				}
				name := attributeValue(node, "name")
				if name == "" {
					continue
				}
				values[name] = []string{attributeValue(node, "value")}
			}
		})
	})

	if hasNotHidden {
		err = errors.New("尝试解析响应，看看是不是可以自动提交请求, 发现 form 中 input 不是 hidden 类型")
	}
	if count != 1 {
		err = errors.New("尝试解析响应，看看是不是可以自动提交请求, 发现没有 form 或多个 form")
	}
	if posturl == "" {
		err = errors.New("尝试解析响应，看看是不是可以自动提交请求, 发现 url 为空")

	}
	if method == "" {
		err = errors.New("尝试解析响应，看看是不是可以自动提交请求, 发现 method 为空")
	}
	return
}

func ParseFrameset(body io.Reader) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if nil != err {
		return nil, errors.New("failed to parse login page, " + err.Error())
	}

	var urls []string
	doc.Find("FRAME").Each(func(idx int, frame *goquery.Selection) {
		for _, node := range frame.Nodes {
			src := attributeValue(node, "SRC")
			if src == "" {
				src = attributeValue(node, "src")
			}
			if src != "" {
				urls = append(urls, src)
			}
		}
	})
	return urls, nil
}

func ParseJsRedirect(body []byte, tokens []string) string {
	foundIdx := -1
	for _, token := range tokens {
		location := []byte(token)
		idx := bytes.Index(body, location)
		if idx >= 0 {
			foundIdx = idx + len(location)
			break
		}
	}
	if foundIdx < 0 {
		return ""
	}
	body = body[foundIdx:]
	body = bytes.TrimLeftFunc(body, unicode.IsSpace)
	if len(body) == 0 {
		return ""
	}

	if body[0] != '=' {
		return ""
	}

	body = body[1:]
	body = bytes.TrimLeftFunc(body, unicode.IsSpace)
	if len(body) == 0 {
		return ""
	}

	idx := bytes.Index(body, []byte("\n"))
	if idx < 0 {
		return ""
	}
	body = bytes.TrimSpace(body[:idx])
	if len(body) == 0 {
		return ""
	}

	body = bytes.Trim(body, ";")
	body = bytes.TrimSpace(body)
	body = bytes.Trim(body, "\"")
	return string(body)
}
