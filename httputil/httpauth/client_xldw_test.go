package httpauth

import (
  "bytes"
  "io"
  "net/http"
  "net/http/httptest"
  "net/url"
  "fmt"
  "testing"
  "strings"

  "github.com/runner-mei/goutils/urlutil"
)

func TestLoginClientXldw(t *testing.T) {
  var redirectAddressFunc func(string)string

  var redirectAddress = func(s string)string {
    return redirectAddressFunc(s)
  }
  outResult := map[string]int{}
  portalMux := makeClientXldwPortal(t, outResult, redirectAddress)
  portalHsrv := httptest.NewServer(portalMux)
  defer portalHsrv.Close()

  ssomux := makeClientXldwSSO(t, outResult, redirectAddress)
  ssoHsrv := httptest.NewServer(ssomux)
  defer ssoHsrv.Close()

  fmt.Println("sso=", ssoHsrv.URL)
  fmt.Println("portal=", portalHsrv.URL)

 redirectAddressFunc = func(s string)string {
  switch s {
  case "", "http://21.11.40.8:8080":
    return portalHsrv.URL
  case  "http://21.11.13.6:17002":
    return ssoHsrv.URL
  default:
    return s
  }
 }


  u, _ := url.Parse(portalHsrv.URL)

  params := &LoginParams{
    AuthMethod: "web",
    Protocol:   "http",
    Address:    u.Host,
    WelcomeURL: "/portal",
    LoginURL:   "<auto>",
    Username:   "sunwei",
    Password:   "qwer123$",
    // PasswordCrypto:      "base64",
    UsernameArgname:     "username",
    PasswordArgname:     "password",
    ReadForm:            true,
    ExceptedContent:     `var loginName = "sunwei"`,
    // Values:              map[string]string{"tree": "10.128.7.120"},
    AutoRedirectEnabled: "false",
  }



  client := New("", "")
  var out bytes.Buffer
  _, msgs, err := Login(nil, &client, params, &out)
  if err != nil {
    t.Log(msgs)
    t.Log(out.String())
    t.Error(err)
    return
  }
  t.Log(msgs)
  t.Log(out.String())
}

func makeClientXldwPortal(t testing.TB, out map[string]int, redirectAddress func(string) string) *http.ServeMux {
    var mux = &http.ServeMux{}

//   mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//     w.WriteHeader(http.StatusOK)
//     io.WriteString(w, `<script>
// window.location.href="sw/index.shtml";
// </script>`)
//   }))

  mux.Handle("/portal", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      JSESSIONID, _ := r.Cookie("JSESSIONID")
      excepted := "BD84B0BF9DF1C16DC0F0557077762DF4"
      if JSESSIONID == nil || JSESSIONID.Value != excepted {
        urlstr := urlutil.Join(redirectAddress("http://21.11.13.6:17002"), "isc_sso/login?service=http%3A%2F%2F21.11.40.8%3A8080%2Fportal%2Fportal_um%2Frest%2Flogin%2F60F724B8FA46D631")
        http.Redirect(w, r, urlstr, http.StatusFound)
        fmt.Println(urlstr)
        // Location: http://21.11.13.6:17002/
        return
      }


      http.SetCookie(w, &http.Cookie{
        Name: "JSESSIONID",
        Value: "BD84B0BF9DF1C16DC0F0557077762DF4",

        Path: "/portal",
        HttpOnly: true,
      })


      http.Redirect(w, r, urlutil.Join(redirectAddress("http://21.11.40.8:8080"), "/portal/portal-web/rest/portal-face"), http.StatusFound)
      // Location: http://21.11.13.6:17002/
      return
  }))
  mux.Handle("/portal/portal_um/rest/login/60F724B8FA46D631", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    
      queryParams := r.URL.Query()
      ticket := queryParams.Get("ticket")
      excepted := "ST-1701-XQLwcaHcrZOHzHp2tF1i-10.242.0.1"
      if ticket != excepted {
        fmt.Println("Post, ticket")
        fmt.Println("want:", excepted)
        fmt.Println(" got:", ticket)
        http.Error(w,  "ticket fail\r\nwant:"+ excepted+"\r\n got:"+ ticket, http.StatusInternalServerError)
        return
      }

      JSESSIONID, _ := r.Cookie("JSESSIONID")
      excepted = "BD84B0BF9DF1C16DC0F0557077762DF4"
      if JSESSIONID == nil || JSESSIONID.Value != excepted {
        fmt.Println("Post, ticket")
        fmt.Println("want:", excepted)
        fmt.Println(" got:", JSESSIONID)
        if JSESSIONID == nil {
          http.Error(w,  "JSESSIONID missing", http.StatusInternalServerError)
        } else {
          http.Error(w,  "JSESSIONID fail\r\nwant:"+ excepted+"\r\n got:"+ JSESSIONID.Value, http.StatusInternalServerError) 
        }
        return
      }
      http.SetCookie(w, &http.Cookie{
        Name: "JSESSIONID",
        Value: "BD84B0BF9DF1C16DC0F0557077762DF4",

        Path: "/portal",
        HttpOnly: true,
      })
      http.SetCookie(w, &http.Cookie{
        Name: "portalsid",
        Value: "BD84B0BF9DF1C16DC0F0557077762DF4",

        Path: "/portal",
        HttpOnly: true,
      })      
      http.SetCookie(w, &http.Cookie{
        Name: "portalstk",
        Value: "rKgsMK7yuUMsBgi9Iop0aecGsRZkpKql_-IcljEosUmNiU0wVGUx_WUC3h097QEB7WOVxFZH1RSsq7Nl_8RgujWZRtUus5PcZ5Ztg6j-Njs|hEuPqH4rTWIWXs1zytYryrYSylQ",
        Path: "/portal",
        HttpOnly: true,
      })

      urlstr := "http://21.11.40.8:8080/portal/portal_um/rest/login/60F724B8FA46D631"
      http.Redirect(w, r, strings.Replace(urlstr, "http://21.11.40.8:8080", redirectAddress("http://21.11.40.8:8080"), -1), http.StatusFound)

    // Location: 
  }))


  mux.Handle("/portal/portal-web/rest/portal-face", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      // queryParams := r.URL.Query()
      JSESSIONID, _ := r.Cookie("JSESSIONID")
      excepted := "BD84B0BF9DF1C16DC0F0557077762DF4"
      if JSESSIONID.Value != excepted {
        fmt.Println("Post, JSESSIONID")
        fmt.Println("want:", excepted)
        fmt.Println(" got:", JSESSIONID)
        http.Error(w, "JSESSIONID fail\r\nwant:"+ excepted+"\r\n got:"+ JSESSIONID.Value, http.StatusInternalServerError)
        return
      }

      urlstr := "http://21.11.40.8:8080/portal/portal-web/release/ff80808160244850016024574f5100c9/index.jsp"
      http.Redirect(w, r, strings.Replace(urlstr, "http://21.11.40.8:8080", redirectAddress("http://21.11.40.8:8080"), -1), http.StatusTemporaryRedirect)

    // Location: http://21.11.13.6:17002/
  }))


  mux.Handle("/portal/portal-web/release/ff80808160244850016024574f5100c9/index.jsp", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
     // queryParams := r.URL.Query()
      JSESSIONID, _ := r.Cookie("JSESSIONID")
      excepted := "BD84B0BF9DF1C16DC0F0557077762DF4"
      if JSESSIONID.Value != excepted {
        fmt.Println("Post, JSESSIONID")
        fmt.Println("want:", excepted)
        fmt.Println(" got:", JSESSIONID)
        http.Error(w, "JSESSIONID fail\r\nwant:"+ excepted+"\r\n got:"+ JSESSIONID.Value, http.StatusInternalServerError)
        return
      }

      w.WriteHeader(http.StatusOK)
      io.WriteString(w, xldwPortalIndexJSPAfterLogin)
    // Location: http://21.11.13.6:17002/
  }))
  return mux
}


func makeClientXldwSSO(t testing.TB, out map[string]int, redirectAddress func(string) string) *http.ServeMux {
  var mux = &http.ServeMux{}

  mux.Handle("/isc_sso/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
    w.WriteHeader(http.StatusOK)
    io.WriteString(w, xldwLoginHtml)
    // case http.MethodPost:

    //   POST /isc_sso/login;jsessionid=Thk3gZ6Gty227fXLJqG6DTvBSS2kFfgh23GSQtGtGy4npRqcMbGf!1342605650?service=http%3A%2F%2F21.11.40.8%3A8080%2Fportal%2Fportal_um%2Frest%2Flogin%2F60F724B8FA46D631 HTTP/1.1

    //   w.WriteHeader(http.StatusOK)
    //   io.WriteString(w, strings.Replace(xldwLoginHtml, "http://21.11.40.8:8080", redirectAddress(), -1))
    default:
      http.Error(w, "method not allow - "+ r.RequestURI, http.StatusInternalServerError)
    }
  }))

  mux.Handle("/isc_sso/login;jsessionid=Thk3gZ6Gty227fXLJqG6DTvBSS2kFfgh23GSQtGtGy4npRqcMbGf!1342605650", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodPost:

      err := r.ParseForm()
      if err != nil {
        http.Error(w, "service fail:"+ err.Error(), http.StatusInternalServerError)
        return
      }
      form := r.PostForm
      
      if value := form.Get("username"); value != "sunwei" {
        t.Error("want:", "sunwei")
        t.Error("want:", value)
      }

      if value := form.Get("password"); value != "sunwei" {
        t.Error("want:", "72ef758a29c5e4cb4ccfc5239e737a7b16c29c65d672c5c618f15aedeb9c9c240eeb11ed9caf3da57b8b7a1f0067720659914ec6ced967f79278b5655df5c9e1a106936c4a8f0a2ffd9b35d305d64ee9c83d0b9c793cc39e15eb22063e4ce7a88aca23a7822822f13c2704cf8745a3d144a5fec54a453e69cd6d85a1e7b5f7b9")
        t.Error(" got:", value)
      }
      if value := form.Get("messageCode"); value != "" {
        t.Error("want:", "")
        t.Error(" got:", value)
      }
      if value := form.Get("signature"); value != "" {
        t.Error("want:", "")
        t.Error(" got:", value)
      }
      if value := form.Get("lt"); value != "LT-41187-fPI7C7QoTABXRjNTSe5J1CgKdUyWEJ" {
        t.Error("want:", "LT-41187-fPI7C7QoTABXRjNTSe5J1CgKdUyWEJ")
        t.Error(" got:", value)
      }
      if value := form.Get("execution"); value != "e1s1" {
        t.Error("want:", "e1s1")
        t.Error(" got:", value)
      }
      if value := form.Get("token"); value != "" {
        t.Error("want:", "")
        t.Error(" got:", value)
      }
      if value := form.Get("_eventId"); value != "submit" {
        t.Error("want:", "submit")
        t.Error(" got:", value)
      }
      queryParams := r.URL.Query()
      service := queryParams.Get("service")

      excepted, _ := url.QueryUnescape("http%3A%2F%2F21.11.40.8%3A8080%2Fportal%2Fportal_um%2Frest%2Flogin%2F60F724B8FA46D631")
      if service != excepted {

        fmt.Println("Post, login")
        fmt.Println("want:", excepted)
        fmt.Println(" got:", service)
        http.Error(w, "service fail\r\nwant:"+ excepted+"\r\n got:"+ service, http.StatusInternalServerError)
        return
      }



      urlstr := "http://21.11.40.8:8080/portal/portal_um/rest/login/60F724B8FA46D631?ticket=ST-1701-XQLwcaHcrZOHzHp2tF1i-10.242.0.1"
      http.Redirect(w, r, strings.Replace(urlstr, "http://21.11.40.8:8080", redirectAddress("http://21.11.40.8:8080"), -1), http.StatusTemporaryRedirect)
    default:
      http.Error(w,  "method not allow - "+ r.RequestURI, http.StatusInternalServerError)
    }
  }))

  return mux
}

const (
  xldwLoginHtml = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">










<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head>

<link type="text/css" rel="stylesheet" href="/isc_sso/css/login_province.css;jsessionid=Thk3gZ6Gty227fXLJqG6DTvBSS2kFfgh23GSQtGtGy4npRqcMbGf!1342605650" />
<script type="text/javascript" src="/isc_sso/js/cookie.js"></script>
<script type="text/javascript" src="/isc_sso/js/jquery.js"></script>
<script type="text/javascript" src="/isc_sso/js/ca_util.js"></script>
<script type="text/javascript" src="/isc_sso/js/jquery.md5.js"></script>
<script type="text/javascript" src="/isc_sso/js/RandomUtil.js"></script>
<script type="text/javascript" src="/isc_sso/js/RsaUtils.js"></script>
<script type="text/javascript" src="/isc_sso/js/sm/SmCrypto-2.9.js"></script>
<script type="text/javascript" src="/isc_sso/js/login/province_login.js"></script>

<meta content="edge" http-equiv="X-UA-Compatible"/>
<title>ISC-SSO</title>
<script>
var smsLogin = false;
var rsaPass = true;
var smPass = false;
var encryptionKey = "00a767ca54db607dc96e5d69c60bf16f3878139ae4ecb4101912da759eaa6ee963aee8efc78a22fe413674480e1dc2168ab36f0153ac8b575e44b3f8fc0621958717ba1aef7a0b977f46a54044e71add31cb5e5534996de016c9a3600de424f6dbd6d0b9d335c26ca3083c53f21f37903cf576ca7fd1ea82f37fe0f1f4c884b3bb#010001";
var mailCodeLogin = false;
var sslClient = false;
var isCfcaCheck = false;
var contextPath = "/isc_sso";
var fingerprintAuthAddr = "http://27.196.218.180:18090";
if(sslClient) {
    clearCookies();
}
function changeLoginMode() {
    var className = $("#login_box_class").attr('class');
    if(className == "login_bg") {
        $("#login_box_class").attr('class', 'login_bg login_bg_sms');
        $("#send_sms").css("display","inline-block");
        document.forms[0].username.value = "";
    }
    if(className == "login_bg login_bg_sms") {
        $("#login_box_class").attr('class','login_bg');
        $("#send_sms").css("display","none");

  }
}
function openMailCode() {
    if(mailCodeLogin) {
        $("#mail_code_div").css({"display" : "block"});
        $("#mail_msg").css({"display" : "block"});
    }
  //开启cfca登录
    if(isCfcaCheck) {
        loadObject();
        $("#submit_login").css({"display" : "none"});
        $("#submit_cfca").css({"display" : "block"});
    }
    if(smsLogin) {
        $("#change-code").css("display","block");
    }
}
function getEncryptPwd(pwd) {
  var encryptPwd = "";
    if(smPass) {
        var sm2Utils = new Sm2Utils(CipherMode.c1c3c2);
        var sm3Pwd = Sm3Utils.encryptFromText(pwd)+getRandomString(8)+pwd;
        encryptPwd = sm2Utils.encryptFromText(encryptionKey,sm3Pwd);
  }
  if(rsaPass) {
        var keys = encryptionKey.split("#");
        var modulus = keys[0];
        var exponent = keys[1];
        //生成0-100之间的随机数
        var random=getRandomString(8);
        //获取key秘钥
        var key = RSAUtils.getKeyPair(exponent, '', modulus);
        //对密码进行md5信息摘要
        var envilope=$.md5(pwd)+random+pwd;
        //进行完整的信息连接，生成数字签名元数据
        encryptPwd = RSAUtils.encryptedString(key, envilope);
  }
  return encryptPwd;
}
var doSubmit = function() {
  //使用国密加密算法加密传输
  var pwd = document.forms[0].password.value;
    var username = document.forms[0].username.value;
    if(username == null || username == "") {
        alert("请输入用户名!");
        return;
    }
    if(pwd == null || pwd == "") {
        alert("请输入密码!");
        return;
    }
    //如果是短信验证，则短信验证码和密码相同
    if("login_bg login_bg_sms" == $("#login_box_class").attr('class')){
        document.forms[0].messageCode.value = pwd;
    }
  //判断是否有check属性,记住密码功能
    var hasCheck  = document.getElementById("checkAcc");
  if(hasCheck != null) {
        if ($("#checkAcc").is(":checked")) {
            setUserCookie(username);
        } else {
            removeCacheCookieUsername();
    }
  }
  document.forms[0].password.value = getEncryptPwd(pwd);
  disabledButton();
  document.forms[0].submit();
};

</script>
</head>
<body onload="javascript:setCookieUsername();">
  <div class="login_box" id="login_box">
    <div class="logo"></div>
    <div id="login_box_class" class="login_bg">
      <div id="change-code" style="display: none;" class="change-code" onclick="changeLoginMode();"></div>
      <form id="fm1" action="/isc_sso/login;jsessionid=Thk3gZ6Gty227fXLJqG6DTvBSS2kFfgh23GSQtGtGy4npRqcMbGf!1342605650?service=http%3A%2F%2F21.11.40.8%3A8080%2Fportal%2Fportal_um%2Frest%2Flogin%2F60F724B8FA46D631" method="post">
        <li>
           
          
            
            <input id="username" name="username" class="login_input1" tabindex="1" accesskey="n" type="text" value="" size="25" maxlength="40" autocomplete="false"/>
          
              &nbsp;&nbsp; <input type="checkbox" id="checkAcc" name="checkAcc"><font color=white>记住账号</font>   
        </li>
        <li id="password_normal" style="padding-top: 5px; padding-top: 0px; *padding-top: 5px;">
          
          <input id="password" name="password" class="login_input1" tabindex="2" onkeydown="javascript:butOnClick();" accesskey="p" type="password" value="" size="25" maxlength="40" autocomplete="off"/>
          <a id="send_sms" style="display: none;" class="send-mes" onclick="sendSmsCode(this)">发送验证码</a>

        </li>
        <li id="mail_code_div" style="display:none;">
          <input id="telOwnLoginName" type="hidden"/>
          <input id="messageCode" name="messageCode" style="width:80px;height:20px;
            border-radius:5px;background-color:#FAFFBD;" placeholder="验证码" value="" />
          <input type="button" id="sendMail" name="sendMail" style="width:110px;height:25px;background:url('');
            border-radius:5px;background-color:#021D19;" value="获取邮件验证码" onclick="sendMailCode();"/>
        </li>
        <li class="login_div" style="height: 25px;">
          <input type="button" id="submit_login" value="登录" onclick="doSubmit();" style="float: left;"/>
          <input type="button" id="submit_cfca" value="证书认证" onclick="submitCfca();" style="float: left;display:none" />
          <input type="button" id="reset" value="重置" style="float: left;" onclick="resetBtn();"/>
        </li>
        <li>
          
          <div id="mail_msg" style="color:#FFF600;float:left;display: none"/>
        </li>
        <input type="hidden" id="signature" name="signature" />
        <input type="hidden" name="lt" value="LT-41187-fPI7C7QoTABXRjNTSe5J1CgKdUyWEJ" />
        <input type="hidden" name="execution" value="e1s1" />
        <input type="hidden" name="token"/>
        <input type="hidden" name="_eventId" value="submit" />
      </form>
    </div>
    <div id="loading" style="position:absolute;top:50%;left:50%;z-index:29;"></div>
    <div class="login_info">Copyright&copy;2008-2018 国家电网公司</div>
  </div>
</body>
</html>`


xldwLoginPost = `<html><head><title>302 Moved Temporarily</title></head>
<body bgcolor="#FFFFFF">
<p>This document you requested has moved temporarily.</p>
<p>It's now at <a href="http://21.11.40.8:8080/portal/portal_um/rest/login/60F724B8FA46D631?ticket=ST-1701-XQLwcaHcrZOHzHp2tF1i-10.242.0.1">http://21.11.40.8:8080/portal/portal_um/rest/login/60F724B8FA46D631?ticket=ST-1701-XQLwcaHcrZOHzHp2tF1i-10.242.0.1</a>.</p>
</body></html>`



xldwPortalIndexJSPAfterLogin = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head><meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<title>西南分部门户</title>
<script type="text/javascript" src="/portal/portal-web/script/frame/jquery.js?v=1623754108488"></script>
<script type="text/javascript" src="/portal/portal-web/script/frame/qs-framework.js?v=1623754108488"></script>
<script type="text/javascript" src="/portal/portal-web/script/frame/qui-run-frame.js?v=1623754108488"></script>
<script type="text/javascript" src="/portal/portal-web/script/frame/jquery-ui-1.9.2.mini.js?v=1623754108488"></script>
<script type="text/javascript" src="/portal/portal-web/release/ff80808160244850016024574f5100c9/script/PortalRunData.js?v=1623754108488"></script>
<script type="text/javascript" src="/portal/portal-web/script/frame/specialTopic.js?v=1623754108488"></script><script type="text/javascript" src="/portal/portal-web/release/ff80808160244850016024574f5100c9/style/portal_ext/resources/menu/menu201403271600/script/Menu.js?v=1623754108488"></script>
<script type="text/javascript" src="/portal/portal-web/release/ff80808160244850016024574f5100c9/style/portal_ext/resources/script/page1/PageEvent.js?v=1623754108488"></script>

<script>
var loginName = "sunwei",styles = '<link id="skinColorLink" type="text/css" href="/portal/portal-web/release/ff80808160244850016024574f5100c9/style/skincolor/default/style.css" rel="stylesheet" /><link id="qscss" type="text/css" href="/portal/portal-web/style/qs/default/style.css" rel="stylesheet" /><link id="compcss" type="text/css" href="/portal/portal-web/release/ff80808160244850016024574f5100c9/style/compcss/default/style.css" rel="stylesheet" />';$(function(){PortalRunMgr.initPage();});</script></head>
<body style="*position:relative;">
<div class="pf_skin pf_skin_201403271600"><DIV class="pf_shell pf_shell_201403271600 clearfix" sizset="2" sizcache019311233104134834="0" sizcache00648110293643388="0">
<DIV class="pf_shell_top div-class-25 clearfix" sizset="3" sizcache019311233104134834="0" sizcache00648110293643388="0">
<DIV class="pf_shell_header div-class-26 clearfix" sizset="4" sizcache019311233104134834="0" sizcache00648110293643388="0">
<DIV class="pf_shell_header1 div-class-27 clearfix" sizset="5" sizcache019311233104134834="0" sizcache00648110293643388="0">
<DIV class="pf_shell_tools pf_shell_tools_top div-class-28 clearfix">
<DIV class="pf_shell_con div-class-29"></DIV></DIV>
<DIV class="pf_shell_nav div-class-30 clearfix">
<DIV class="pf_shell_con div-class-31"></DIV></DIV></DIV></DIV>
<DIV class="pf_shell_header2 div-class-32 clearfix" sizset="8" sizcache019311233104134834="0" sizcache00648110293643388="0">
<DIV style="BACKGROUND-IMAGE: url(/portal/portal-web/image/upload/7038891cf5564e258fcfe5397b477378.jpg); BACKGROUND-COLOR: rgb(255,255,255); BACKGROUND-REPEAT: no-repeat" class="pf_shell_logo div-class-33 clearfix">
<DIV class="pf_shell_con div-class-34"></DIV></DIV>
<DIV class="pf_shell_tools pf_shell_tools01 div-class-35 clearfix">
<DIV class="pf_shell_con div-class-36"> </DIV></DIV>
<DIV class="pf_shell_tools pf_shell_tools02 div-class-37 clearfix">
<DIV class="pf_shell_con div-class-38"> </DIV></DIV></DIV>
<DIV class="pf_shell_header3 div-class-39 clearfix" sizset="12" sizcache019311233104134834="0" sizcache00648110293643388="0">
<DIV class="pf_shell_tools_left_pins div-class-40 clearfix"><A class=current href="javascript:void(0);"></A></DIV>
<DIV class="pf_shell_tools pf_shell_tools_left_pins_con div-class-41 clearfix">
<DIV class="pf_shell_con div-class-42"></DIV></DIV></DIV></DIV>
<DIV class="pf_shell_main div-class-43 clearfix">
<DIV class="pf_shell_con div-class-44"></DIV></DIV>
<DIV class="pf_shell_bottom div-class-45 clearfix" sizset="16" sizcache019311233104134834="0" sizcache00648110293643388="0">
<DIV class="pf_shell_con div-class-46" sizset="16" sizcache019311233104134834="0" sizcache00648110293643388="0">
<DIV class="pf_shell_bottom_info1 div-class-47 clearfix">国家电网公司版权所有 企业邮箱：E_mail:system-info@sgcc.com.cn 四川中电启明星信息技术有限公司制作维护 </DIV>
<DIV class="pf_shell_bottom_info2 div-class-48 clearfix"></DIV></DIV></DIV></DIV></div>
</body>
<script>
if(window.localstore){
var color;try{color = localstore.get("skinColor-" + loginName);}catch(e){
}
color = color || "default"; styles = styles.replace(/\/(default|green|blue|azure)\//g, "/" + color + "/");}document.write(styles); </script></html>`
)