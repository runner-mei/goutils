package httpauth

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func makeClientA() *http.ServeMux {
	var mux = &http.ServeMux{}

	mux.Handle("/irr/portal", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "IPCZQX037arbc",
			Value: "0000df010a80040f97f58c013fa485e1f4b2d6a9",
			Path:  "/",
		})
		http.Redirect(w, r, "/nxxx/app/plogin?c=name/password/uri&%22/irr/portal%22", http.StatusTemporaryRedirect)
	}))

	mux.Handle("/nxxx/app/plogin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "JSESSIONID",
			Value: "7AA9A9B86A56CDFC58FF20AB1DDB3C3C",
			Path:  "/nxxx",
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "UrnNovellNidpClusterMemberId",
			Value: "~03~02fff~0D~1A~1F~7B~7Ct",
			Path:  "/nxxx",
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "urn:novell:nidp:cluster:member:id",
			Value: "~03~02fff~0D~1A~1F~7B~7C",
			Path:  "/nxxx",
		})

		cookie, err := r.Cookie("IPCZQX037arbc")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if cookie == nil {
			http.Error(w, "IPCZQX037arbc is missing", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/nixx/idabc/sso?RequestID=idObfkayACJA6KXJH8rr-mpreLmsQ&MajorVersion=1&MinorVersion=2&IssueInstant=2019-11-21T02%3A49%3A19Z&ProviderID=http%3A%2F%2Fssoproxy1.ec.hengwei.com.cn%3A80%2Fnesp%2Fidff%2Fmetadata&RelayState=MA%3D%3D&consent=urn%3Aliberty%3Aconsent%3Aunavailable&ForceAuthn=false&IsPassive=false&NameIDPolicy=onetime&ProtocolProfile=http%3A%2F%2Fprojectliberty.org%2Fprofiles%2Fbrws-art&target=http%3A%2F%2Fauth.hengwei.com.cn%2Firr%2Fportal&AuthnContextStatementRef=name%2Fpassword%2Furi", http.StatusTemporaryRedirect)
	}))

	mux.Handle("/pwdmgt/CheckLogin.jsp", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method is invalid", http.StatusInternalServerError)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "ParseForm:"+err.Error(), http.StatusInternalServerError)
			return
		}
		values := r.PostForm

		if actual, except := values.Get("user"), "00020023"; actual != except {
			//t.Log(values)
			http.Error(w, "user: want "+except+" got "+actual, http.StatusInternalServerError)
			return
		}

		if actual, except := values.Get("url_path"), "/nixx/jsp/main.jsp"; !strings.HasSuffix(actual, except) {
			http.Error(w, "url_path: want "+except+" got "+actual, http.StatusInternalServerError)
			return
		}

		if actual, except := values.Get("RequestID"), "idObfkayACJA6KXJH8rr-mpreLmsQ"; actual != except {
			http.Error(w, "RequestID: want "+except+" got "+actual, http.StatusInternalServerError)
			return
		}

		if actual, except := values.Get("IssueInstant"), "2019-11-21T02:49:19Z"; actual != except {
			http.Error(w, "IssueInstant: want "+except+" got "+actual, http.StatusInternalServerError)
			return
		}

		if actual, except := values.Get("RelayState"), "MA=="; actual != except {
			http.Error(w, "RelayState: want "+except+" got "+actual, http.StatusInternalServerError)
			return
		}

		if actual, except := values.Get("loginurl"), "/nixx/jsp/login.jsp"; !strings.HasSuffix(actual, except) {
			http.Error(w, "loginurl: want "+except+" got "+actual, http.StatusInternalServerError)
			return
		}

		if actual, except := values.Get("Ecom_User_ID"), "00020023"; actual != except {
			http.Error(w, "Ecom_User_ID: want "+except+" got "+actual, http.StatusInternalServerError)
			return
		}

		password := values.Get("Ecom_Password")
		bs, _ := base64.StdEncoding.DecodeString(password)

		if actual, except := string(bs), "qwer123$"; actual != except {
			http.Error(w, "Ecom_Password: want "+except+" got "+actual, http.StatusInternalServerError)
			return
		}
		//&Ecom_User_ID=00020023&Ecom_Password=cXdlcjEyMyQ%3D&Submit=%E7%99%BB%E5%BD%95

		http.SetCookie(w, &http.Cookie{
			Name:  "JSESSIONID",
			Value: "984B8A02F13CA72D3179AC779742DE1D",
			Path:  "/nixx",
		})
		io.WriteString(w, `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd">
      <html>
      <head>
      <meta http-equiv="Content-Type" content="text/html; charset=US-ASCII">
      <title></title>
      </head>
      <body>

            <form name="IDPLogin" id="IDPLogin"
              enctype="application/x-www-form-urlencoded" method="POST"
              action="/nixx/idabc/sso?MajorVersion=1&MinorVersion=2&ProviderID=http%3A%2F%2Fssoauth1.ec.hengwei.com.cn%3A80%2Fnesp%2Fidff%2Fmetadata&consent=urn%3Aliberty%3Aconsent%3Aunavailable&ForceAuthn=false&IsPassive=false&NameIDPolicy=onetime&ProtocolProfile=http%3A%2F%2Fprojectliberty.org%2Fprofiles%2Fbrws-art&AuthnContextStatementRef=name%2Fpassword%2FuriRequestID=idObfkayACJA6KXJH8rr-mpreLmsQ&IssueInstant=2019-11-21T02%3A49%3A19Z&RelayState=MA==" AUTOCOMPLETE="off">
              <input  type="hidden" id="user" name="user" value="00020023" /> 
              <input  type="hidden" id="event" name="event" value="1" /><!--<form name="IDPLogin" id="IDPLogin" enctype="application/x-www-form-urlencoded" method="POST" action="http://ssoserver3.ec.hengwei.com.cn:8080/nixx/app/login" AUTOCOMPLETE="on">-->
              <input type="hidden" class="smalltext" id="Ecom_User_ID" name="Ecom_User_ID" size="30" value="00020023" /> 
              <input type="hidden" class="smalltext" name="Ecom_Password" id="Ecom_Password"  size="30" value="qwer123$" /> 
              <!--<input type="submit" value="submit"/>-->
            </form>
            <script language="javascript">
            function isNumAndStr (str) {
              
              var regexpUperStr=/[A-Z]+/;
              var reexpLowerStr=/[a-z]+/;
              var regexpNum=/\d+/;
              var regexpSpecialStr=/[^a-zA-Z0-9]/;
              var uperStrFlag = regexpUperStr.test(str);
              var lowerStrFlag = reexpLowerStr.test(str);
              var numFlag = regexpNum.test(str);
              var specialStrFlag = regexpSpecialStr.test(str);
              if((lowerStrFlag&&numFlag&&specialStrFlag)||(uperStrFlag&&numFlag&&specialStrFlag)||(lowerStrFlag&&numFlag&&uperStrFlag)||(lowerStrFlag&&specialStrFlag&&uperStrFlag)) {
                
                return true;
              }
              
                return false;
            }
            
            var jspwd = "qwer123$";
            
            /*
            alert(jspwd);
            jspwd = jspwd.replace(/\//g,'\\/');
            jspwd = jspwd.replace(/\'/g,"\\'");
            jspwd = jspwd.replace(/\"/g,'\\"');
            */
            var pwdchangedays ="346";
            
            if (pwdchangedays>365) {
              alert('您的密码修改时间已超过一年,根据国网通知要求，需要您重新修改登录密码。');
              document.IDPLogin.action='ChangePassword.jsp';
              
            }else{
              if (!isNumAndStr(jspwd)) {
                alert('您的密码不符合国网密码策略要求，密码必须是大小写字母、数字、特殊字符组合,请修改您的密码。');
                document.IDPLogin.action='ChangePassword.jsp';
                
              }else{
                
              }
            }
            document.getElementById('IDPLogin').submit();
            
            </script>
          
      </body>
      </html>`)
	}))

	mux.Handle("/nixx/idabc/sso", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			http.SetCookie(w, &http.Cookie{
				Name:  "JSESSIONID",
				Value: "90D2DFCC46EDD22EF658DA45F6714D21",
				Path:  "/nixx",
			})
			http.SetCookie(w, &http.Cookie{
				Name:  "UrnNovellNidpClusterMemberId",
				Value: "~03~02fff~0D~1A~1F~7B~7C",
				Path:  "/nixx",
			})
			http.SetCookie(w, &http.Cookie{
				Name:  "urn:novell:nidp:cluster:member:id",
				Value: "~03~02fff~0D~1A~1F~7B~7C",
				Path:  "/nixx",
			})

			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
            <html xmlns="http://www.w3.org/1999/xhtml">
            <head>
            <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
            <title>用户登陆</title>
            <script language="javascript">
              
              function Base64() {
              
                // private property
                _keyStr = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=";
              
                // public method for encoding
                this.encode = function(input) {
                  var output = "";
                  var chr1, chr2, chr3, enc1, enc2, enc3, enc4;
                  var i = 0;
                  input = _utf8_encode(input);
                  while (i < input.length) {
                    chr1 = input.charCodeAt(i++);
                    chr2 = input.charCodeAt(i++);
                    chr3 = input.charCodeAt(i++);
                    enc1 = chr1 >> 2;
                    enc2 = ((chr1 & 3) << 4) | (chr2 >> 4);
                    enc3 = ((chr2 & 15) << 2) | (chr3 >> 6);
                    enc4 = chr3 & 63;
                    if (isNaN(chr2)) {
                      enc3 = enc4 = 64;
                    } else if (isNaN(chr3)) {
                      enc4 = 64;
                    }
                    output = output + _keyStr.charAt(enc1)
                        + _keyStr.charAt(enc2) + _keyStr.charAt(enc3)
                        + _keyStr.charAt(enc4);
                  }
                  return output;
                }
              
                // public method for decoding
                this.decode = function(input) {
                  var output = "";
                  var chr1, chr2, chr3;
                  var enc1, enc2, enc3, enc4;
                  var i = 0;
                  input = input.replace(/[^A-Za-z0-9\+\/\=]/g, "");
                  while (i < input.length) {
                    enc1 = _keyStr.indexOf(input.charAt(i++));
                    enc2 = _keyStr.indexOf(input.charAt(i++));
                    enc3 = _keyStr.indexOf(input.charAt(i++));
                    enc4 = _keyStr.indexOf(input.charAt(i++));
                    chr1 = (enc1 << 2) | (enc2 >> 4);
                    chr2 = ((enc2 & 15) << 4) | (enc3 >> 2);
                    chr3 = ((enc3 & 3) << 6) | enc4;
                    output = output + String.fromCharCode(chr1);
                    if (enc3 != 64) {
                      output = output + String.fromCharCode(chr2);
                    }
                    if (enc4 != 64) {
                      output = output + String.fromCharCode(chr3);
                    }
                  }
                  output = _utf8_decode(output);
                  return output;
                }
              
                // private method for UTF-8 encoding
                _utf8_encode = function(string) {
                  string = string.replace(/\r\n/g, "\n");
                  var utftext = "";
                  for (var n = 0; n < string.length; n++) {
                    var c = string.charCodeAt(n);
                    if (c < 128) {
                      utftext += String.fromCharCode(c);
                    } else if ((c > 127) && (c < 2048)) {
                      utftext += String.fromCharCode((c >> 6) | 192);
                      utftext += String.fromCharCode((c & 63) | 128);
                    } else {
                      utftext += String.fromCharCode((c >> 12) | 224);
                      utftext += String.fromCharCode(((c >> 6) & 63) | 128);
                      utftext += String.fromCharCode((c & 63) | 128);
                    }
              
                  }
                  return utftext;
                }
              
                // private method for UTF-8 decoding
                _utf8_decode = function(utftext) {
                  var string = "";
                  var i = 0;
                  var c = c1 = c2 = 0;
                  while (i < utftext.length) {
                    c = utftext.charCodeAt(i);
                    if (c < 128) {
                      string += String.fromCharCode(c);
                      i++;
                    } else if ((c > 191) && (c < 224)) {
                      c2 = utftext.charCodeAt(i + 1);
                      string += String.fromCharCode(((c & 31) << 6)
                          | (c2 & 63));
                      i += 2;
                    } else {
                      c2 = utftext.charCodeAt(i + 1);
                      c3 = utftext.charCodeAt(i + 2);
                      string += String.fromCharCode(((c & 15) << 12)
                          | ((c2 & 63) << 6) | (c3 & 63));
                      i += 3;
                    }
                  }
                  return string;
                }
              }

              function checkPwd()
              {
                var loginForm = document.getElementById('IDPLogin');
                var user = document.getElementById('Ecom_User_ID').value;
                document.getElementById('user').value = user;
                //loginForm.target="_blank";
                  
              }
                  function onLoginLoaded() 
                { 
                  GetLastUser(); 
                } 
                function GetLastUser() 
                { 
                  var id = "49BAC005-7D5B-4231-8CEA-16939BEACD67"; 
                  var usr = GetCookie(id); 
                  if(usr != null) 
                  { 
                    document.getElementById('Ecom_User_ID').value = usr; 
                  } 
                  else 
                  { 
                    document.getElementById('Ecom_User_ID').value = ""; 
                  } 
                  GetPwdAndChk(); 
                } 
                //点击登录时触发客户端事件 
                function SetPwdAndChk() 
                { 
                  //取用户名 
                  var usr = document.getElementById('Ecom_User_ID').value; 
                  //将最后一个用户信息写入到Cookie 
                  SetLastUser(usr); 
                  
                  var pwd = document.getElementById('Ecom_Password').value; 
                  
                  //如果记住密码选项被选中 
                  if(document.getElementById('saveme').checked == true) 
                  { 
                    //取密码值 
                    //var pwd = document.getElementById('Ecom_Password').value; 
                    var expdate = new Date(); 
                    expdate.setTime(expdate.getTime() + 14 * (24 * 60 * 60 * 1000)); 
                    //将用户名和密码写入到Cookie 
                    SetCookie(usr,pwd, expdate); 

                  } 
                  else 
                  { 
                    //如果没有选中记住密码,则立即过期 
                    ResetCookie(); 
                  } 
                  var b = new Base64();
                  var epwd = b.encode(pwd); 
                  document.getElementById('Ecom_Password').value = epwd;
                  //alert(document.getElementById('Ecom_Password').value);
                  
                  document.forms[0].submit();
                }     
                function SetLastUser(usr) 
                { 
                  var id = "49BAC005-7D5B-4231-8CEA-16939BEACD67"; 
                  var expdate = new Date(); 
                  //当前时间加上两周的时间 
                  expdate.setTime(expdate.getTime() + 14 * (24 * 60 * 60 * 1000)); 
                  SetCookie(id, usr, expdate); 
                } 
                //用户名失去焦点时调用该方法 
                function GetPwdAndChk() 
                { 
                  var usr = document.getElementById('Ecom_User_ID').value; 
                  var pwd = GetCookie(usr); 
                  if(pwd != null) 
                  { 
                    document.getElementById('saveme').checked = true; 
                    document.getElementById('Ecom_Password').value = pwd; 
                  } 
                  else 
                  { 
                    document.getElementById('saveme').checked = false; 
                    document.getElementById('Ecom_Password').value = ""; 
                  } 
                } 
                //取Cookie的值 
                function GetCookie (name) 
                { 
                  var arg = name + "="; 
                  var alen = arg.length; 
                  var clen = document.cookie.length; 
                  var i = 0; 
                  while (i < clen) 
                  { 
                    var j = i + alen; 
                    //alert(j); 
                    if (document.cookie.substring(i, j) == arg) 
                    return getCookieVal (j); 
                    i = document.cookie.indexOf(" ", i) + 1; 
                    if (i == 0) break; 
                  } 
                  return null; 
                }  
                function getCookieVal (offset) 
                { 
                  var endstr = document.cookie.indexOf (";", offset); 
                  if (endstr == -1) 
                  endstr = document.cookie.length; 
                  return unescape(document.cookie.substring(offset, endstr)); 
                } 
                //写入到Cookie 
                function SetCookie(name, value, expires) 
                { 
                  var argv = SetCookie.arguments; 
                  //本例中length = 3 
                  var argc = SetCookie.arguments.length; 
                  var expires = (argc > 2) ? argv[2] : null; 
                  var path = (argc > 3) ? argv[3] : null; 
                  var domain = (argc > 4) ? argv[4] : null; 
                  var secure = (argc > 5) ? argv[5] : false; 
                  document.cookie = name + "=" + escape (value) + 
                  ((expires == null) ? "" : ("; expires=" + expires.toGMTString())) + 
                  ((path == null) ? "" : ("; path=" + path)) + 
                  ((domain == null) ? "" : ("; domain=" + domain)) + 
                  ((secure == true) ? "; secure" : ""); 
                } 
                function ResetCookie() 
                { 
                  var usr = document.getElementById('Ecom_User_ID').value; 
                  var expdate = new Date(); 
                  SetCookie(usr, null, expdate); 
                }
            </script>
            <style type="text/css">
            <!--
            body {
              background-color: #42bd8f;
              background-image: url(/nixx/images/bg_line.jpg);
              margin-left: 0px;
              margin-top: 0px;
              margin-right: 0px;
              margin-bottom: 0px;
            }
            td {
              font-size: 12px;
            }
            .STYLE1 {color: #FFFFFF}
            a:link {
              font-family: "宋体";
              font-size: 12px;
              color: #FFFFFF;
              text-decoration: underline;
            }
            .login_bg {
              background-repeat: no-repeat;
            }
            -->
            </style>
            </head>

            <body onload="onLoginLoaded()">

            <form name="IDPLogin" id="IDPLogin"
              enctype="application/x-www-form-urlencoded" method="POST"
              action="http://ssoauth1.ec.hengwei.com.cn:8080/pwdmgt/CheckLogin.jsp" AUTOCOMPLETE="off"
              onSubmit="checkPwd()">
            <input type="hidden" id="user" name="user"/>
            <input type="hidden" id="url_path" name="url_path" value="http://ssoauth1.ec.hengwei.com.cn:8080/nixx/jsp/main.jsp"/>
            <input type="hidden" id="RequestID" name="RequestID" value="idObfkayACJA6KXJH8rr-mpreLmsQ"/>
            <input type="hidden" id="IssueInstant" name="IssueInstant" value="2019-11-21T02:49:19Z"/>
            <input type="hidden" id="RelayState" name="RelayState" value="MA=="/>
            <input type="hidden" id="loginurl"  name="loginurl" value="http://ssoauth1.ec.hengwei.com.cn:8080/nixx/jsp/login.jsp">
            <table width="485" border="0" align="center" cellpadding="0" cellspacing="0">
              <tr>
                <td height="80">&nbsp;</td>
                <td>&nbsp;</td>
                <td>&nbsp;</td>
                <td>&nbsp;</td>
                <td>&nbsp;</td>
              </tr>
              <tr>
                <td colspan="5" valign="top" background="/nixx/images/bg.jpg" class="login_bg"><table width="485" height="386" border="0" cellpadding="0" cellspacing="0">
                  <tr>
                    <td width="50" height="120">&nbsp;</td>
                    <TD width="140" valign="top"><a href="http://portal.hengwei.com.cn/" target="_blank"><img align="right" src="/nixx/images/tm1.gif" width="140" height="55"  border="0"/></TD>
                    <td colspan="3">&nbsp;</td>
                  </tr>
                  <tr>
                    <td>&nbsp;</td>
                    <td height="50" align="right">用户名：</td>
                    <td colspan="3" align="left"><input id="Ecom_User_ID" name="Ecom_User_ID" type="text" size="25" onblur="GetPwdAndChk()" style="width:150px"/></td>
                  </tr>
                  <tr>
                    <td>&nbsp;</td>
                    <td height="25" align="right">密&nbsp;&nbsp;&nbsp;&nbsp;码：</td>
                    <td colspan="3" align="left"><input name="Ecom_Password" id="Ecom_Password" type="password" size="25" style="width:150px"/></td>
                  </tr>
                
              <tr>
                      <td colspan="5" valign="top">
                  <table width="60%" border="0" align="center" cellpadding="0" cellspacing="0">
                      <tr>
                        <td width="70" align="right" height="45"><input type="checkbox" id = "saveme" name="saveme"/></td>
                          <td width="70" align="left">记住密码</td>
                    <td width="70" align="left"><a href="http://192.168.1.62:17001/isc_mp_auth/modifypwd/index.jsp" target="_blank">密码重置</a></td>
                    <td width="70" align="left"><a href="http://192.168.1.62:17001/isc_mp_auth/modifypwd/inforeset/index.jsp" target="_blank">个人密保</a></td>
                    <td width="100" align="left"><a href="http://iscmp.hengwei.com.cn/isc_mp_auth/ownregister/regist/commonAccountRegist.jsp" target="_blank">公共账号注册</a></td>
                      </tr>
                    </table>
                </td>
              </tr>

                  <tr>
                <td >&nbsp;</td>
                <td colspan="4" valign="top">
                  <table width="60%" border="0" align="center" cellpadding="0" cellspacing="0">
                      <tr>
                        <td >&nbsp;</td>
                    <td width="125" align="left"><input type="submit" name="Submit" value="登录" onclick="SetPwdAndChk()" style="width:90px"/></td>
                          <td colspan="2" align="left"><input type="button" name="Submit2" value="用户注册" onclick="window.open('http://iscmp.hengwei.com.cn/isc_mp_auth/ownregister/regist/reg.jsp?prvid=EC3D68241FFA0C97E0430100007F78C9')" style="width:90px"/></td>
                      </tr>
                    </table>
                </td>
                  </tr>

                <tr>
                <td>&nbsp;</td>
                <td align=right></td>
                <td width="6">&nbsp;</td>
                <td width="200"></td>
                <td>&nbsp;</td>
                </tr>       

                  

                <tr>
                <td>&nbsp;</td>
                <td align=right></td>
                <td width="6">&nbsp;</td>
                <td width="200"></td>
                <td>&nbsp;</td>
                </tr> 
             
                  <tr>
                    <td height="165">&nbsp;</td>
                    <td colspan="4" valign="top">
                  <table width="99%" border="0" align="center" cellpadding="0" cellspacing="0">
                      <tr>
                        <td width="10%"><br /></td>
                        <td width="90%">&nbsp;注意：1、修改门户密码后，请在报销、考勤、膳食系统<br />
                        &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;使用原帐号和新密码。<br />
                        &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;2、修改门户密码后，不需要修改本地邮件客户端<br />
                        &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;帐号密码。<br />
                        </td> 
                      </tr>
                  </table>
                </td>
                  </tr>
                </table></td>
              </tr>
            </table>
            </form>
            </body>
                
            </html>`)
			return
		}

		// http://ssoauth1.ec.hengwei.com.cn:8080/nixx/idabc/sso?
		// MajorVersion=1
		// MinorVersion=2
		// ProviderID=http%3A%2F%2Fssoauth1.ec.hengwei.com.cn%3A80%2Fnesp%2Fidff%2Fmetadata
		// &consent=urn%3Aliberty%3Aconsent%3Aunavailable
		// &ForceAuthn=false
		// &IsPassive=false
		// &NameIDPolicy=onetime
		// &ProtocolProfile=http%3A%2F%2Fprojectliberty.org%2Fprofiles%2Fbrws-art
		// &AuthnContextStatementRef=name%2Fpassword%2FuriRequestID=idObfkayACJA6KXJH8rr-mpreLmsQ
		// &IssueInstant=2019-11-21T02%3A49%3A19Z
		// &RelayState=MA==

		if r.Method != "POST" {
			http.Error(w, "Method is invalid", http.StatusInternalServerError)
			return
		}

		referer := r.Header.Get("Referer")
		if !strings.Contains(referer, "nixx/idabc/sso") {
			http.Error(w, "referer is invalid", http.StatusInternalServerError)
			return
		}

		//user=00020023&event=1&Ecom_User_ID=00020023&Ecom_Password=qwer123%24

		io.WriteString(w, `<!DOCTYPE HTML PUBLIC "-//W3C//Dtd HTML 4.0 transitional//zh">
        <html lang="zh">
          <head>
            <META HTTP-EQUIV="expires" CONTENT="0">
          </head>
          <body>
            <script language="JavaScript">
                                <!--
                                        parent.location.href="http://auth.hengwei.com.cn/irr";

                                -->     
                        </script>

          </body>
        </html>`)
	}))
	return mux
}

func TestLoginClientA1(t *testing.T) {

	mux := makeClientA()
	hsrv := httptest.NewServer(mux)
	defer hsrv.Close()

	u, _ := url.Parse(hsrv.URL)

	params := &LoginParams{
		Protocol:            "http",
		Address:             u.Host,
		WelcomeURL:          "/irr/portal",
		LoginURL:            "/pwdmgt/CheckLogin.jsp",
		Username:            "00020023",
		Password:            "qwer123$",
		PasswordCrypto:      "base64",
		UsernameArgname:     "Ecom_User_ID",
		PasswordArgname:     "Ecom_Password",
		ReadForm:            true,
		ExceptedContent:     `parent.location.href="http://auth.hengwei.com.cn/irr"`,
		Values:              map[string]string{"user": "00020023"},
		AutoRedirectEnabled: "false",
	}

	client := New()
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
