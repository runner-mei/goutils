package httpauth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/runner-mei/goutils/httputil"
	"github.com/runner-mei/goutils/urlutil"
	"github.com/runner-mei/goutils/util"
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
func New() http.Client {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	return http.Client{
		Transport: httputil.InsecureHttpTransport,
		Jar:       cookieJar,
	}
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
	if params.Timeout > 0 {
		var c func()
		ctx, c = context.WithTimeout(ctx, params.Timeout)
		defer c()
	}

	values := url.Values{}
	var action string
	var logMessages []string

	if params.WelcomeURL != "" {
		action = "GET"
		if params.WelcomeMethod != "" {
			action = params.WelcomeMethod
		}

		welcomeURL := strings.ToLower(params.WelcomeURL)
		if strings.HasPrefix(welcomeURL, "https://") || strings.HasPrefix(welcomeURL, "http://") {
			welcomeURL = params.WelcomeURL
		} else {
			welcomeURL = urlutil.Join(baseurl, params.WelcomeURL)
		}

		welcomeReq, err := http.NewRequest(action, welcomeURL, nil)
		if err != nil {
			logMessages = append(logMessages, "创建登录首页请求失败", err.Error())
			return nil, logMessages, err
		}
		welcomeReq = welcomeReq.WithContext(ctx)
		action = ""

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
			logMessages = append(logMessages, "访问登录首页失败", err.Error())
			return nil, logMessages, err
		}

		logMessages = append(logMessages, "访问登录首页成功")
		client.CheckRedirect = nil

		body, err := ioutil.ReadAll(welcomeResp.Body)
		welcomeResp.Body.Close()
		if err != nil {
			logMessages = append(logMessages, "解析登录首页失败", err.Error())
			return nil, logMessages, err
		}

		if http.StatusOK != welcomeResp.StatusCode {
			logMessages = append(logMessages, "解析登录首页响应码不正确")
			return nil, logMessages, errors.New("status code is '" + welcomeResp.Status + "' - " + string(body))
		}

		httputil.Dump(dumpOut,
			"========= 1 DumpRequest =========\r\n", welcomeReq, nil,
			"\r\n========= 1 DumpResponse =========\r\n", welcomeResp, bytes.NewReader(body))

		if params.ReadForm {
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
			if nil != err {
				logMessages = append(logMessages, "解析登录首页失败", err.Error())
				return nil, logMessages, errors.New("failed to parse login page, " + err.Error())
			}
			//dumpNodesFromDocument(doc, "#user_login")
			//dumpNodesFromDocument(doc, "#user_password")

			if params.FormLocation == "" {
				params.FormLocation = "form"
			}

			//<form accept-charset="UTF-8" action="/login" class="login-box" id="new_user" method="post">
			// var formNode *html.Node
			// if sl := doc.Find(params.FormLocation); nil == sl || 0 == sl.Size() {
			//  return nil, errors.New("'"+params.FormLocation+"' isn't found")
			// } else if nodes := sl.Nodes; 1 != len(nodes) {
			//  return nil, errors.New("'"+params.FormLocation+"' is muti choice")
			// } else {
			//  formNode = nodes[0]
			// }

			count := 0
			doc.Find(params.FormLocation).Each(func(idx int, form *goquery.Selection) {
				if nodes := form.Nodes; len(nodes) > 0 {
					action = strings.ToUpper(attributeValueWithDefaultValue(nodes[0], "method", "POST"))
				}

				count++
				form.Find("input[type=\"hidden\"]").Each(func(idx int, input *goquery.Selection) {
					for _, node := range input.Nodes {
						values[attributeValue(node, "name")] = []string{attributeValue(node, "value")}
					}
				})
			})

			switch {
			case count == 0:
				if err == nil {
					logMessages = append(logMessages, "解析登录首页时没有找到表单")
				} else {
					logMessages = append(logMessages, "解析登录首页时没有找到表单", err.Error())
				}
				return nil, logMessages, errors.New("'" + params.FormLocation + "' isn't found")
			case count > 1:
				if err == nil {
					logMessages = append(logMessages, "解析登录首页时找到多个表单")
				} else {
					logMessages = append(logMessages, "解析登录首页时找到多个表单", err.Error())
				}
				return nil, logMessages, errors.New("'" + params.FormLocation + "' is muti choice")
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
	values[passwordform] = []string{params.Password}

	if params.PasswordCrypto == "base64" {
		values[passwordform] = []string{base64.StdEncoding.EncodeToString([]byte(params.Password))}
	}

	if 0 != len(params.Values) {
		for k, v := range params.Values {
			values[k] = []string{v}
		}
	}
	if params.LoginMethod != "" {
		action = params.LoginMethod
	}
	if action == "" {
		action = "POST"
	}

	loginURL := strings.ToLower(params.LoginURL)
	if strings.HasPrefix(loginURL, "https://") || strings.HasPrefix(loginURL, "http://") {
		loginURL = params.LoginURL
	} else {
		loginURL = urlutil.Join(baseurl, params.LoginURL)
	}

	return PostLogin(ctx, client, params, action, loginURL, params.ContentType, values, logMessages, 0, dumpOut)
}

func PostLogin(ctx context.Context, client *http.Client, params *LoginParams, loginMethod, loginURL, contentType string, values url.Values, logMessages []string, tryCount int, dumpOut io.Writer) (*http.Response, []string, error) {

	// loginReq, loginResp, err := Do(ctx, client, loginMethod, loginURL, contentType, values)
	// if nil != err {
	// 	return nil, err
	// }
	// defer loginResp.Body.Close()

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

	loginReq, err := http.NewRequest(loginMethod, loginURL, bytes.NewReader(loginBody))
	if err != nil {
		logMessages = append(logMessages, "创建登录请求失败", err.Error())
		return nil, logMessages, err
	}

	loginReq.Header.Set("Content-Type", contentType)
	loginReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	loginReq.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	loginReq.Header.Set("Cache-Control", "max-age=0")
	loginReq.Header.Set("Connection", "keep-alive")
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
				posturl = params.BaseURL() + posturl
			}

			logMessages = append(logMessages, "发送登录请求成功， 但响应不包含指定的字符，但表单尝试再试一次")
			return PostLogin(ctx, client, params, method, posturl, "urlencoded", values, logMessages, 1, dumpOut)
		}

		if params.AutoRedirectEnabled == "" || params.AutoRedirectEnabled == "auto" {
			for _, locationHref := range []string{"parent.location.href=\"", "window.location.href=\""} {
				if strings.HasPrefix(params.ExceptedContent, locationHref) {
					urlStr := strings.TrimPrefix(params.ExceptedContent, locationHref)
					urlStr = strings.TrimSuffix(strings.TrimSpace(urlStr), ";")
					urlStr = strings.TrimSuffix(urlStr, "\"")

					params.AutoRedirectURL = urlStr
					params.AutoRedirectEnabled = "true"

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
				values[attributeValue(node, "name")] = []string{attributeValue(node, "value")}
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
