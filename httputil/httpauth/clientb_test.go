package httpauth

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func makeClientB(out map[string]int) *http.ServeMux {
	var mux = &http.ServeMux{}

	mux.Handle("/bizxxxx/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `<HTML>
  <HEAD>
    <script>
               window.location="servlet/portal?render=on";
         </script>
  </HEAD>
</HTML>`)
	}))

	mux.Handle("/bizxxxx/servlet/portal", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "username",
			Value: "65635F786E6A6332",
			Path:  "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "tree",
			Value: "31302E3132382E372E313230",
			Path:  "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "rank",
			Value: "7072696D617279",
			Path:  "/",
		})

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, formHtml)
	}))
	mux.Handle("/bizxxxx/servlet/webacc", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		queryParams := r.URL.Query()
		switch queryParams.Get("taskId") {
		case "fw.Header":
			for key, value := range map[string]string{
				"username": "ABC65635F786E6A6332",
				"tree":     "ABC31302E3132382E372E313230",
				"rank":     "ABC7072696D617279",
			} {
				actual, err := r.Cookie(key)
				if err != nil {
					http.Error(w, key+": get "+err.Error(), http.StatusInternalServerError)
					return
				}
				if except := value; actual.Value != except {
					//t.Log(values)
					http.Error(w, key+": want "+except+" got "+actual.Value, http.StatusInternalServerError)
					return
				}
			}

			if r.Method != "GET" {
				http.Error(w, "Method is invalid", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			io.WriteString(w, clientb_header)

			out["header"] = out["header"] + 1
		case "dev.Empty":

			for key, value := range map[string]string{
				"username": "ABC65635F786E6A6332",
				"tree":     "ABC31302E3132382E372E313230",
				"rank":     "ABC7072696D617279",
			} {
				actual, err := r.Cookie(key)
				if err != nil {
					http.Error(w, key+": get "+err.Error(), http.StatusInternalServerError)
					return
				}
				if except := value; actual.Value != except {
					//t.Log(values)
					http.Error(w, key+": want "+except+" got "+actual.Value, http.StatusInternalServerError)
					return
				}
			}

			if r.Method != "GET" {
				http.Error(w, "Method is invalid", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, clientb_tasks)
			out["body"] = out["body"] + 1
		case "":
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

			for key, value := range map[string]string{
				"username": "65635F786E6A6332",
				"tree":     "31302E3132382E372E313230",
				"rank":     "7072696D617279",
			} {
				actual, err := r.Cookie(key)
				if err != nil {
					http.Error(w, key+": get "+err.Error(), http.StatusInternalServerError)
					return
				}
				if except := value; actual.Value != except {
					//t.Log(values)
					http.Error(w, key+": want "+except+" got "+actual.Value, http.StatusInternalServerError)
					return
				}
			}

			for key, value := range map[string]string{
				"DoLogin":     "true",
				"forceMaster": "false",
				"Login_Key":   "1585624158387",
				"password":    "qwer123$",
				"rank":        "primary",
				"tree":        "10.128.7.120",
				"username":    "00020023",
				// "登录.x":        "55",
				// "登录.y":        "13",
			} {
				if actual, except := values.Get(key), value; actual != except {
					//t.Log(values)
					http.Error(w, key+": want "+except+" got "+actual, http.StatusInternalServerError)
					return
				}
			}

			http.SetCookie(w, &http.Cookie{
				Name:  "username",
				Value: "ABC65635F786E6A6332",
				Path:  "/",
			})
			http.SetCookie(w, &http.Cookie{
				Name:  "tree",
				Value: "ABC31302E3132382E372E313230",
				Path:  "/",
			})
			http.SetCookie(w, &http.Cookie{
				Name:  "rank",
				Value: "ABC7072696D617279",
				Path:  "/",
			})

			w.WriteHeader(http.StatusOK)
			io.WriteString(w, okHtml)
			return
		default:
			http.Error(w, "taskid arguments is invalid", http.StatusInternalServerError)
			return
		}
	}))
	mux.Handle("/bizxxxx/Empty.html", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method is invalid", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `<HTML></HTML>`)
		out["empty"] = out["empty"] + 1
	}))
	return mux
}

func TestLoginClientB1(t *testing.T) {
	outResult := map[string]int{}
	mux := makeClientB(outResult)
	hsrv := httptest.NewServer(mux)
	defer hsrv.Close()

	u, _ := url.Parse(hsrv.URL)

	params := &LoginParams{
		Protocol:   "http",
		Address:    u.Host,
		WelcomeURL: "/bizxxxx/servlet/portal",
		LoginURL:   "/bizxxxx/servlet/webacc",
		Username:   "00020023",
		Password:   "qwer123$",
		// PasswordCrypto:      "base64",
		UsernameArgname:     "username",
		PasswordArgname:     "password",
		ReadForm:            true,
		ExceptedContent:     `控制中心`,
		Values:              map[string]string{"tree": "10.128.7.120"},
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

func TestLoginClientB2(t *testing.T) {
	outResult := map[string]int{}
	mux := makeClientB(outResult)
	hsrv := httptest.NewServer(mux)
	defer hsrv.Close()

	u, _ := url.Parse(hsrv.URL)

	params := &LoginParams{
		Protocol:   "http",
		Address:    u.Host,
		WelcomeURL: "/bizxxxx/",
		LoginURL:   "/bizxxxx/servlet/webacc",
		Username:   "00020023",
		Password:   "qwer123$",
		// PasswordCrypto:      "base64",
		UsernameArgname:     "username",
		PasswordArgname:     "password",
		ReadForm:            true,
		ExceptedContent:     `控制中心`,
		Values:              map[string]string{"tree": "10.128.7.120"},
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

const (
	okHtml = `

                                                                             

    <HTML>
    <HEAD>
    <TITLE>Novell iManager</TITLE>
    </HEAD>

    <FRAMESET rows="80,*,0" framespacing=0 frameborder=0 ONLOAD="top.name='eMFrame'" id=iManagerRootFrame>
        <FRAME SRC="webacc?taskId=fw.Header&merge=fw.iManagerHeader"
               NAME="Branding"
               title="${requestScope['FrameTitle.HeaderFrame']}"
               NORESIZE SCROLLING="no"
               marginheight=0 marginwidth=0>
        <FRAME SRC="webacc?taskId=dev.Empty&amp;merge=fw.Body&amp;view=Tasks"
               NAME="Mainscreen"
               title="${requestScope['FrameTitle.Mainscreen']}">
         <FRAME SRC="/bizxxxx/Empty.html"
               NAME="Bitbucket"
               title="Bitbucket">
    </FRAMESET>

    </HTML>`

	formHtml = `






  <!-- ========= START imaneMFrameScripts tag ========== -->
<SCRIPT>
BrowserCharset='utf-8';
ParentWindowChangedErrorAlertMessage = '未保存更改。';
fw_modulesPath = '/bizxxxx/portal/modules';
fw_lang="zh_CN";
fw_taskId="null";
fw_respId="9419"
fw_isDebugEnabled=false;
var djConfig = {isDebug: false, bindEncoding: "utf-8"};
function exists(s)
{
   var ret=false;
   if (s!=null && s.length>0)
   {
      try
      {
         eval(s);
         ret=true;
      }
      catch(e)
      {
      }
   }
   return ret;
}
</SCRIPT>
<SCRIPT type="text/javascript" src="/bizxxxx/dojo/dojo.js"></SCRIPT>
<script language="JavaScript" type="text/javascript">
</script>
<SCRIPT language="JavaScript" src="/bizxxxx/portal/modules/dev/javascripts/BrowserVersions.js"></SCRIPT>
<SCRIPT language="JavaScript" src="/bizxxxx/portal/modules/dev/javascripts/eMFrameScripts.js"></SCRIPT>
<!-- ========= END imaneMFrameScripts tag ========== -->

<SCRIPT src='/bizxxxx/portal/modules/fw/javascripts/iManDialogScripts.js'></SCRIPT>


<html>
   <head>
      <title>Novell iManager</title>
      <LINK href="/bizxxxx/portal/modules/dev/css/hf_style.css" rel="styleSheet" type="text/css">
      <style type= "text/css" media="screen">
         #Headgraphic  { position: absolute; z-index: 0; top: 0px; left: 0px; width: 499px; visibility: visible }
     #Head1        { position: absolute; z-index: 1; top: 0px; left: 0px; width: 499px; visibility: visible }
     #Filler1       { position: absolute; z-index: 0; top: 0px; left: 499px; right: 60px; width: 380px; visibility: visible }
         #logo          { position: absolute; z-index: 2; top: 0px; right: 0px; width: 60px; height: 80px; visibility: visible }
 
     #logo1         { position: absolute; z-index: 3; top: 0px; right: 0px; width: 60px; height: 80px; visibility: hidden }
 
         #title   { position: absolute; z-index: 1; top: 25px; left: 12px; width: 208px; visibility: visible }
         #Apptitle1  { position: absolute; z-index: 1; top: 0px; left: 0px; visibility: visible }
     #Apptitle   { position: absolute; overflow:hidden; z-index: 1; top: 0px; left: 275px; right: 60px; height: 80px; visibility: visible; }
         #Nimage  { position: absolute; top: 10px; left: 10px; width: 100px; height: 103px; visibility: visible }
         #BodyPane  { position: absolute; width: 470px; visibility: visible }
         #FailPane  { position: absolute; width: 470px; visibility: visible }
         .errorhead  { color: #c82727; font-style: normal; font-weight: 800; font-size: 14px; line-height: 18px; }
         .emframeformhead1 { color: white;  font-size: 1.2em; line-height: 1.2em; letter-spacing: 0.0em; background-color: #458ab9; text-align: left; text-indent: 0.5em}
         .emframerulebelow { padding-bottom: 10px; border-bottom: 2px solid #458ab9 }
         .apptitle1 {color: white; text-decoration: none; font-weight: normal; font-size: 2.2em; line-height: 1.9em; background: url(/bizxxxx/portal/modules/fw/images/iMan27_H1_L1_bg.gif) repeat-x 0% 0%  }
         body { background: white url(/bizxxxx/portal/modules/fw/images/iMan27_H1_L0_bg.gif) repeat-x 0% 0% }
      </style>
      


<!-- BEGIN HelpScripts.inc -->
<!-- requires eMFrameScripts -->
<script>
function launchHelp(helpFile)
{
   var slashIndex = helpFile.indexOf("/");
   if(slashIndex == -1 || slashIndex == helpFile.length - 1)
   {
      return;
   }
   var module = helpFile.substring(0, slashIndex);
   var page = helpFile.substring(slashIndex + 1, helpFile.length);
   var w=window.open('webacc?taskId=fw.HelpRedirect&merge=fw.GoUrl&module=' + urlEncode(module) + '&page=' + urlEncode(page) + '&type=help', "帮助", 'toolbar=no,location=no,directories=no,menubar=no,scrollbars=yes,resizable=yes,width=500,height=500');
   if(w != null)
   {
      w.focus();
   }
}

function launchError(errorFile)
{
   var slashIndex = errorFile.indexOf("/");
   if(slashIndex == -1 || slashIndex == errorFile.length - 1)
   {
      return;
   }
   var module = errorFile.substring(0, slashIndex);
   var page = errorFile.substring(slashIndex + 1, errorFile.length);
   var w=window.open('webacc?taskId=fw.HelpRedirect&merge=fw.GoUrl&module=' + urlEncode(module) + '&page=' + urlEncode(page) + '&type=errors', "帮助", 'toolbar=no,location=no,directories=no,menubar=no,scrollbars=yes,resizable=yes,width=500,height=500');
   if(w != null)
   {
      w.focus();
   }
}

</script>
<!-- END HelpScripts.inc -->
      
   </head>
   <body bgcolor="white" marginwidth="0" marginheight="0" leftmargin="0" topmargin="0" onLoad="activateImanDialog('authenticate');pageInit();">
      <div id="Headgraphic" style="height:80px; z-index:0; background-color:white;">
         <IMG src="/bizxxxx/portal/modules/fw/images/iMan27_H1_L0.gif" width="499" height="80" border="0"></div>
      <div id="Head1"><IMG src="/bizxxxx/portal/modules/fw/images/iMan27_H1_L1.gif" width="499" height="80" border="0"/></div> 




      <div id="Apptitle" style="left:275px; right:60px; top:0px; height: 80px">
         <a id="AppTitleA" style="text-decoration: none">
            <div id="AppTitleDiv" class="apptitle1" style="padding-top: 3px; padding-bottom: 3px; width: 100%">
               Novell iManager
            </div>
         </a>
      </div>

      <div id="logo" style="z-index:2">
         <a href="http://www.novell.com/products/consoles/imanager/" target="_blank"><IMG src="/bizxxxx/portal/modules/fw/images/iMan27_logo_L0.gif" width="60" height="80"  border="0"></a>
      </div>
      <div id="logo1" style="z-index:3">
         <a href="http://www.novell.com/products/consoles/imanager/" target="_blank"><IMG src="/bizxxxx/portal/modules/fw/images/iMan27_logo_L1.gif" width="60" height="80"  border="0"></a>
      </div>

      <table height="100%" border=0 width="600" cellpadding=0 cellspacing=0>        
      <tr>
         <td bgcolor="#edeeec" width="125" valign="top">
         <br>
         <br>
         <br>
         <br>
         <br>
         &nbsp;&nbsp;<IMG src="/bizxxxx/portal/modules/fw/images/nlogo_100.gif" border="0">
      </td>
      <td width="10">
         &nbsp;
      </td>
      <td valign="top" width="465">   
      <br>
      <br>
      <br>
      <br><p/>
      <br>
      <script>var authenticate_center = false;</script>
<div id=authenticate style="display:none; z-index:100; width:500px; position:relative; border:0px solid black; background-color:#ffffff;">
<div style='clear:both;display:block;height:0px;visibility:hidden'></div>
<div id="authenticate_body" style="height:500px; width:500px; color:#000000; font-size:11pt; vertical-align:middle; overflow:auto ">
      <script>
         function pageInit()
         {
              if(top.name != "eMFrame")
              {
                 if(this.top.opener)
                 {
                    //this.top.opener.top.location = "frameservice?taskId=fw.Startup";
                    //Commenting out the line to fix the  Bug 331856 - When using IE if 
                    //iManager is opened from a child window it closes the child window and takes over the parent window.
                    //this.top.close();
                 }
              }
              else if(this != top)
              {
                 top.location = "frameservice?taskId=fw.Startup";
              }
            top.name="eMFrame";
            if(document.AuthenticateForm != null)
            {
               if(document.AuthenticateForm.username.value.length > 0)
               {
                   document.AuthenticateForm.password.focus();
               }
               else
               {
                    document.AuthenticateForm.username.focus();
               }
            }
            onChangeProtocol();
        }
        nvdsUsernameExample = "（示例：admin 或 cn=admin,o=novell）";
        eDirUsernameExample = "（示例：admin 或 admin.novell）";
        function onChangeProtocol()
        {
            // change the example string to match the protocal selected
            var usernameExample = document.getElementById("usernameExample");
        }
        
        var req;
      function webacc(){

        var username = document.AuthenticateForm.username;
        var password = document.AuthenticateForm.password;
        var rank = document.AuthenticateForm.rank;
         var protocol = "";
         if(document.AuthenticateForm.protocol)
         {
            protocol = "&protocol="+escape(document.AuthenticateForm.protocol.value); 
         }
         var tree = document.AuthenticateForm.tree;
        var url = "webacc?taskId=fw.Empty&username=" + escape(username.value)+"&password="+escape(password.value)+"&tree="+escape(tree.value)+"&rank="+escape(rank.value)+protocol;
            
        if (window.XMLHttpRequest) 
         {
          req = new XMLHttpRequest();
        } 
         else if (window.ActiveXObject) 
         {
          req = new ActiveXObject("Microsoft.XMLHTTP");
        }
          
        req.open("POST", url);
        req.onreadystatechange = callback;
        req.send(url);
      }

      function callback() {
         if (req.readyState == 4) 
         {
            if (req.status == 200) 
            {
               parseMessage(req);
            }
         }
      }

      function parseMessage(req) 
      {
         var message = req.responseXML.getElementsByTagName("authentications")[0];
        var error = req.responseXML.getElementsByTagName("error")[0];
          
        if(error)
        {
            var errorText = '错误';
          var html = "<img style='float:left;' class='margin8' src='/bizxxxx/portal/modules/dev/images/error32.gif' alt='"+errorText+"' title='"+errorText+"'  align='absmiddle'/><FONT FACE='Arial,Helvetica'>"+error.firstChild.firstChild.nodeValue+"</FONT>";
          var errorMsgDiv = document.getElementById("errorpane");
          errorMsgDiv.innerHTML = html;
        }
        else
        {
          var errorMsgDiv = document.getElementById("errorpane");
          errorMsgDiv.innerHTML = "";
          //deactivateImanDialog('authenticate');
          if(document.getElementById("authenticationDataTable"))
          {
            var authenticationDataTableDiv = document.getElementById("authenticationDataTable");
            authenticationDataTableDiv.innerHTML = req.responseText;
          }
            
            if(document.getElementById("loginList"))
            {
               var loginListDiv = document.getElementById("loginList");
               
               var callBackFunctionName = 'noCallBackDefined';
               
               var i = 0;
               var html = "";
               while(message.childNodes[i]){
                  var friendlyName = message.childNodes[i].attributes[0].value;
                  var keyName = message.childNodes[i].childNodes[0].nodeValue;
                  html = html+"<div class=\"task3\"><a href=\"javascript:"+callBackFunctionName+"('"+keyName+"')\" >"+friendlyName+"</a></div>"
                  i++;
               }
               loginListDiv.innerHTML = html;
            }
        }
      }
      
    function checkDefaultKey(evt)
      {
         var form = document.forms[0];
         var keyCode = evt.which ? evt.which : evt.keyCode;
         // 13 is the Enter key code
         if (keyCode == 13)
         {
              document.forms[0].submit();
              return false;
         }
         return true;
      }
      
      function clearError() {
        var errorMsgDiv = document.getElementById("errorpane");
        errorMsgDiv.innerHTML = "";
      }
      </script>
      <form name="AuthenticateForm" method=post action="webacc" autocomplete="off">
            <input type=hidden name="rank" value="primary"/>
      <input type=hidden name="DoLogin" value="true">
      <input type=hidden name="forceMaster" value="false"/>
      <input type=hidden name="Login_Key" value="1585624158387">
      <table border="0" cellpadding="2" cellspacing="0" >
         <tr bgcolor="#458ab9">
            <td valign=middle colspan="2">
               <div class="emframeformhead1"><a href="#" onClick="javascript:window.open('http://www.novell.com/documentation/imanager27/imanager_admin_272/data/bu6vkj8.html', '帮助', 'toolbar=no,location=no,directories=no,menubar=no,scrollbars=yes,resizable=yes,width=700,height=500');"><IMG class="floatright" id=help src='/bizxxxx/portal/modules/dev/images/help_16.gif' border=0 width=16 height=16 alt="帮助" title="帮助"></a>登录</div>               
            </td>
         </tr>
         <tr>
              <td colspan="2" style="padding: 5px;" >
               <div id="errorpane" ></div>
            </td>
         </tr>
        <tr>
           <td height="50" nowrap valign="bottom" class="mediumtext">
                  用户名：
              <br>
              <input TABINDEX=1 type="text" name="username" maxLength="256" autocomplete="off" width="200" style="width: 200px" value="acccbbbb" size="40" onkeypress="return checkDefaultKey(event)" />
                  <div id="usernameExample" class="instructions">（示例：admin 或 admin.novell）
                  </div>
           </td>  
         </td>
        </tr>
        <tr>
           <td valign="bottom" class="mediumtext">
              口令：
              <br>
              <input TABINDEX=2 type="password" name="password" maxLength="256" autocomplete="off" width="200" style="width: 200px" value="" size="40" onkeypress="return checkDefaultKey(event)" >
           </td>
        </tr>
        <tr>
            <td valign="bottom" class="mediumtext">
                  <!-- Tree text -->
                  树：
                  <br>
                  <input TABINDEX=3 type="text" name="tree" width="200" maxLength="256" style="width: 200px" value="10.128.7.120" size="40" onkeypress="return checkDefaultKey(event)">
                  <div class="instructions">(192.168.14.199, mytree, myserver.company.com)</div>
            </td>
         </tr>
         <tr>
            <td valign="bottom" class="mediumtext">
            </td>
         </tr>
         <tr>
            <td valign="bottom" height="34" class="emframerulebelow" style="padding-top: 10px;">
                  <INPUT TABINDEX=5 type=image name="登录" alt="登录" title="登录" src="/bizxxxx/portal/modules/dev/images/zh_CN/btnlogin_zh_CN.gif" border="0" />&nbsp;
            </td>
         </tr>
         <tr>
            <td colspan="2" class="smalltext" width="100%"><p>&copy 版权所有 1999-2010 Novell, Inc. 保留所有权利。</p></td>
         </tr>
      </table>
      </form>
<p style='text-align:center; vertical-align:baseline;'>
</p>
</div>
</div>
      <br>
      <p>
      </td>
   </tr>
</table>
</body>
</html>`

	clientb_tasks = `



<!-- //File:Body.jsp -->     
<HTML>
<HEAD></HEAD>
   <FRAMESET COLS="230,*,0">
      <FRAME NAME="Nav"     
             TITLE="导航帧"
             SRC="webacc?taskId=fw.Nav&merge=fw.Nav&view=Tasks" 
             FRAMEBORDER="0" SCROLLING="on">
      <FRAME NAME="Content"
             TITLE="内容帧" 
             SRC="webacc?NPService=fw.LaunchService&NPAction=CompleteReturn&returnID=fw.HomePage&firstTime=true"
             MARGINHEIGHT="15" MARGINWIDTH="15" FRAMEBORDER="0">
      <FRAME NAME="Empty"   
             SRC="webacc?taskId=dev.Empty&merge=fw.Empty&firstTime=true"  
             MARGINHEIGHT="0" MARGINWIDTH="0" FRAMEBORDER="0" NORESIZE>
   </FRAMESET>
</HTML>
`
	clientb_header = `















<html>
<head>
  <title>Novell iManager</title>
  <LINK href="/bizxxxx/portal/modules/dev/css/hf_style.css" rel="styleSheet" type="text/css"/>

   <style type="text/css" media="screen">
      #Head0          { position: absolute; z-index: 0; top: 0px; left: 0px; width: 499px; visibility: visible }
      #Head1          { position: absolute; z-index: 1; top: 0px; left: 0px; width: 499px; visibility: visible }
      #logo0          { position: absolute; z-index: 1; top: 0px; right: 0px; width: 60px; height: 60px; visibility: visible }
 
      #logo1         { position: absolute; z-index: 3; top: 0px; right: 0px; width: 60px; height: 80px; visibility: hidden }
 
          #quickviewlog  { position: absolute; z-index: 3; top: 61px; right: 88px; height: 16px; visibility: hidden }
          #quickclearlog { position: absolute; z-index: 3; top: 61px; right: 68px; height: 16px; visibility: hidden }
 
      #Apptitle     { position: absolute; overflow:hidden; z-index: 5; top: 13px; left: 55px; width: 186px; height: 23px; visibility: visible; }
      #Apptitlehint  { color: white; position: absolute;  overflow:hidden; z-index: 6; top: 2px; left: 55px; width: 150px; height: 17px; visibility: hidden }
      #logincontext  { position: absolute; z-index: 4; top: 40px; left: 55px; visibility: visible }
          #mode          { font-size: 11px; position: absolute; z-index: 2; top: 53px; left: 55px; width:600px; visibility: visible; display: block }
          .apptitle1 {color: white; overflow:hidden; text-decoration: none; font-weight: normal; font-size: 18px }
          .apptitle2 {color: white; overflow:hidden; text-decoration: none; font-weight: bold; font-size: 18px }
 
      #buthelp       { position: absolute; z-index: 2; top: 22px; width: 34px; height: 48px; visibility: visible;}
      #buthelp2      { position: absolute; z-index: 3; top: 22px; width: 34px; height: 48px; visibility: hidden;}
      #helphint      { color: #fff; position: absolute; z-index: 5; top: 4px; width: 200px; visibility: hidden}
          #homebut       { position: absolute; z-index: 2; left: 249px; width: 34px; height: 48px; visibility: visible }
      #homebut2      { position: absolute; z-index: 3; left: 249px; width: 34px; height: 48px; visibility: hidden }
      #homebut3      { position: absolute; z-index: 4; left: 249px; width: 34px; height: 48px; visibility: hidden }
      #homehint      { color: #fff; position: absolute; z-index: 5; top: 4px; left: 249px; width: 200px; visibility: hidden }
      #exitbut       { position: absolute; z-index: 2; top: 22px; left: 283px; width: 34px; height: 48px; visibility: visible }
      #exitbut2      { position: absolute; z-index: 3; top: 22px; left: 283px; width: 34px; height: 48px; visibility: hidden }
      #exithint      { color: #fff; position: absolute; z-index: 1; top: 4px; left: 290px; width: 150px; visibility: hidden }
      #FillDiv      { position: absolute; z-index: 1; overflow:hidden; top: 0px; left: 375px; right: 60px; height: 80px; visibility: visible; }
      body          { background: white url(/bizxxxx/portal/modules/fw/images/iMan27_H1_L0_bg.gif) repeat-x 0% 0% }
          .username      { color: black; font-weight:bold; font-size:0.7em; text-transform:uppercase; white-space:nowrap; width:186px; overflow-x:hidden;}
          .fillclass {color: white; text-decoration: none; overflow:hidden; font-weight: normal; font-size: 2.2em; line-height: 1.9em; background: url(/bizxxxx/portal/modules/fw/images/iMan27_H1_L1_bg.gif) repeat-x 0% 0%  }
   </style>
   
   <!-- ========= START imaneMFrameScripts tag ========== -->
<SCRIPT>
BrowserCharset='utf-8';
ParentWindowChangedErrorAlertMessage = '未保存更改。';
fw_modulesPath = '/bizxxxx/portal/modules';
fw_lang="zh_CN";
fw_taskId="fw.Header";
fw_respId="9421"
fw_isDebugEnabled=false;
var djConfig = {isDebug: false, bindEncoding: "utf-8"};
function exists(s)
{
   var ret=false;
   if (s!=null && s.length>0)
   {
      try
      {
         eval(s);
         ret=true;
      }
      catch(e)
      {
      }
   }
   return ret;
}
</SCRIPT>
<SCRIPT type="text/javascript" src="/bizxxxx/dojo/dojo.js"></SCRIPT>
<script language="JavaScript" type="text/javascript">
</script>
<SCRIPT language="JavaScript" src="/bizxxxx/portal/modules/dev/javascripts/BrowserVersions.js"></SCRIPT>
<SCRIPT language="JavaScript" src="/bizxxxx/portal/modules/dev/javascripts/eMFrameScripts.js"></SCRIPT>
<!-- ========= END imaneMFrameScripts tag ========== -->

   


<!-- BEGIN HelpScripts.inc -->
<!-- requires eMFrameScripts -->
<script>
function launchHelp(helpFile)
{
   var slashIndex = helpFile.indexOf("/");
   if(slashIndex == -1 || slashIndex == helpFile.length - 1)
   {
      return;
   }
   var module = helpFile.substring(0, slashIndex);
   var page = helpFile.substring(slashIndex + 1, helpFile.length);
   var w=window.open('webacc?taskId=fw.HelpRedirect&merge=fw.GoUrl&module=' + urlEncode(module) + '&page=' + urlEncode(page) + '&type=help', "帮助", 'toolbar=no,location=no,directories=no,menubar=no,scrollbars=yes,resizable=yes,width=500,height=500');
   if(w != null)
   {
      w.focus();
   }
}

function launchError(errorFile)
{
   var slashIndex = errorFile.indexOf("/");
   if(slashIndex == -1 || slashIndex == errorFile.length - 1)
   {
      return;
   }
   var module = errorFile.substring(0, slashIndex);
   var page = errorFile.substring(slashIndex + 1, errorFile.length);
   var w=window.open('webacc?taskId=fw.HelpRedirect&merge=fw.GoUrl&module=' + urlEncode(module) + '&page=' + urlEncode(page) + '&type=errors', "帮助", 'toolbar=no,location=no,directories=no,menubar=no,scrollbars=yes,resizable=yes,width=500,height=500');
   if(w != null)
   {
      w.focus();
   }
}

</script>
<!-- END HelpScripts.inc -->

   
   <script>
   NN4 = ((parseInt(navigator.appVersion)>=4 && parseInt(navigator.appVersion)<5)&&(navigator.appName.indexOf("Netscape")!=-1))? 1:0;
   NN6 = ((parseInt(navigator.appVersion)>=5)&&(navigator.appName.indexOf("Netscape")!=-1))? 1:0;
   MS4 = ((parseInt(navigator.appVersion)>=4)&&(navigator.appName.indexOf("Microsoft")!=-1))? 1:0;
  
   

      var curImageID = "fw.TasksView";
   

   function show_hint(hintID, imageID)
   {
      var browser = navigator.appName;

      if (curImageID != imageID)
      {
         document.getElementById(imageID+'2').style.visibility = "visible";
      }
      document.getElementById(hintID).style.visibility = "visible";
   }

   function hide_hint(hintID, imageID)
   {
      var browser = navigator.appName;

      if (curImageID != imageID)
      {
         document.getElementById(imageID+'2').style.visibility = "hidden";
      }
      document.getElementById(hintID).style.visibility = "hidden";
   }

   function show_App_hint(hintID)
   {
      var browser = navigator.appName;

      document.getElementById('AppTitleDiv').className="apptitle2";
      document.getElementById(hintID).style.visibility = "visible";
   }

   function hide_App_hint(hintID)
   {
      var browser = navigator.appName;

      document.getElementById('AppTitleDiv').className="apptitle1";
      document.getElementById(hintID).style.visibility = "hidden";
   }

   function switch_icon(hintID, imageID)
   {

      var fullCurID = curImageID + '2'
      var fullID = imageID + '3';

      if (document.getElementById(curImageID) !=null)
      {
         document.getElementById(curImageID+'3').style.visibility = "hidden";
         document.getElementById(curImageID).style.visibility = "visible";
      }
      document.getElementById(imageID).style.visibility = "hidden";
      document.getElementById(imageID+'2').style.visibility = "hidden";
      document.getElementById(imageID+'3').style.visibility = "visible";
      document.getElementById(hintID).style.visibility = "hidden";

      curImageID = imageID;
   }

   function setCollectionModeText(newText)
   {
      var full = "acccbbbb.ActiveUser.ECGCHR.ECGCTREE";
      var treeName = full.substring(full.lastIndexOf(".") + 1);
        var collTextSpan = document.getElementById( 'mode' );
        if (collTextSpan != null && newText != null)
      {
        collTextSpan.innerHTML = treeName;
        }
   }
   
   </script>
</head>

<body bgcolor="white" marginwidth="0" marginheight="0" leftmargin="0" topmargin="0">
   <div id="logincontext">
      <div class="username" title="acccbbbb.ActiveUser.ECGCHR.ECGCTREE">acccbbbb</div>
   </div>

   <div id="Apptitle">
      <a id="AppTitleA" style="text-decoration: none" href="webacc?taskId=fw.About&amp;merge=fw.About" target="Content" onmouseover="show_App_hint('Apptitlehint')" onmouseout="hide_App_hint('Apptitlehint')" >
        <div id="AppTitleDiv" class="apptitle1">
          Novell iManager<br>
        </div>
      </a>
   </div>

   <div id="Head0"><IMG src="/bizxxxx/portal/modules/fw/images/iMan27_H1_L0.gif" width="499" height="80" border="0"/></div> 
   <div id="Head1"><IMG src="/bizxxxx/portal/modules/fw/images/iMan27_H1_L1.gif" width="499" height="80" border="0"/></div> 

   <div id="FillDiv" style="left:375px; right:60px; top:0px; height: 80px">
      <a id="FillDivA" style="text-decoration: none">
         <div id="FillDivB" class="fillclass" style="padding-top: 3px; padding-bottom: 3px; width: 100%">
            <br>
         </div>
      </a>
   </div>

   <div id="Apptitlehint" class="hint1">
      关于信息
   </div>
   
   <div id="mode" title="acccbbbb.ActiveUser.ECGCHR.ECGCTREE">ECGCTREE</div>
   <div id="logo0">
      
         
         
            <a href="http://www.novell.com/products/consoles/imanager" target="_blank">
               <img height="80" width="60" src="/bizxxxx/portal/modules/fw/images/iMan27_logo_L0.gif" border="0" title="Novell" alt="Novell"/></a>
         
      
   </div>
   <div id="logo1">
      
         
         
            <a href="http://www.novell.com/products/consoles/imanager" target="_blank">
               <img height="80" width="60" src="/bizxxxx/portal/modules/fw/images/iMan27_logo_L1.gif" border="0" title="Novell" alt="Novell"/></a>
         
      
   </div>
 
   <div id="quickviewlog">
      <a href="webacc?taskId=fw.iManager Configuration&amp;merge=fw.ViewLog&amp;nextState=QuickLogViewing" target="_blank">
         <img src="/bizxxxx/portal/modules/fw/images/log_show.gif" border="0" title="查看调试日志" alt="查看调试日志"/></a></div>
         
   <div id="quickclearlog">
      <a href="webacc?taskId=fw.iManager Configuration&amp;nextState=QuickLogClearing" target="Bitbucket">
         <img src="/bizxxxx/portal/modules/fw/images/log_clear.gif" border="0" title="清除调试日志" alt="清除调试日志"/></a></div>
 
         
   <script>
      var bDebugMode = "";
      if (bDebugMode == "true")
      {
         document.getElementById("quickviewlog").style.visibility = "visible";
         document.getElementById("quickclearlog").style.visibility = "visible";
      }
   </script>
   
  
    
      
    
    
    

   <div id="homebut" style="top:22px;">
      <a href="webacc?NPService=fw.LaunchService&amp;NPAction=Return&amp;returnID=fw.HomePage&amp;merge=fw.HomePage" target="Content" onclick="switch_icon('homehint','homebut')" onmouseover="show_hint('homehint','homebut')" onmouseout="hide_hint('homehint','homebut')">
         <IMG height="48" width="34" id="home" src="/bizxxxx/portal/modules/fw/images/but_nl_home1.gif" border="0" title="主页" alt="%%%Header.HomeHint}%%%"/>
      </a>
   </div>
   <div id="homebut2" style="top:22px;">
      <a href="webacc?NPService=fw.LaunchService&amp;NPAction=Return&amp;returnID=fw.HomePage&amp;merge=fw.HomePage" target="Content" onclick="switch_icon('homehint','homebut')" onmouseover="show_hint('homehint','homebut')" onmouseout="hide_hint('homehint','homebut')">
         <IMG height="48" width="34" id="home" src="/bizxxxx/portal/modules/fw/images/but_nl_home2.gif" border="0" title="主页" alt="%%%Header.HomeHint}%%%"/>
      </a>
   </div>
   <div id="homebut3" style="top:22px;">
      <a href="webacc?NPService=fw.LaunchService&amp;NPAction=Return&amp;returnID=fw.HomePage&amp;merge=fw.HomePage" target="Content" onclick="switch_icon('homehint','homebut')" onmouseover="show_hint('homehint','homebut')" onmouseout="hide_hint('homehint','homebut')">
         <IMG height="48" width="34" id="home" src="/bizxxxx/portal/modules/fw/images/but_nl_home3.gif" border="0" title="主页" alt="%%%Header.HomeHint}%%%"/>
      </a>
   </div>
   <div id="homehint" class="hint1">
      主页
   </div>

   
      <div id="exitbut">
         <a href="portalservice?fw.exit=true" target="_top" onmouseover="show_hint('exithint','exitbut')" onmouseout="hide_hint('exithint','exitbut')">
            <IMG height="48" width="34" src="/bizxxxx/portal/modules/fw/images/but_nl_exit1.gif" border="0" title="退出" alt="退出"/>
         </a>
      </div>
      <div id="exitbut2">
         <a href="portalservice?fw.exit=true" target="_top" onmouseover="show_hint('exithint','exitbut')" onmouseout="hide_hint('exithint','exitbut')">
            <IMG height="48" width="34" src="/bizxxxx/portal/modules/fw/images/but_nl_exit2.gif" border="0" title="退出" alt="退出"/>
         </a>
      </div>
      <div id="exithint" class="hint1">
         退出
      </div>
   


   
   

























   
   
   
      
      <div id="fw.TasksView" style="position: absolute; z-index: 2; top: 22px; left: 328px; width: 34px; height: 48px; visibility: visible">
         <a href="webacc?taskId=dev.Empty&amp;merge=fw.Body&amp;view=Tasks" target="Mainscreen" onclick="switch_icon('fw.TasksViewhint', 'fw.TasksView')" onmouseover="show_hint('fw.TasksViewhint', 'fw.TasksView')" onmouseout="hide_hint('fw.TasksViewhint', 'fw.TasksView')">
            <IMG src="/bizxxxx/portal/modules/fw/images/but_nl_tasks1.gif" border="0" title="职能和任务" alt="职能和任务"/>
         </a>
      </div>
      <div id="fw.TasksView2" style="position: absolute; z-index: 3; top: 22px; left: 328px; width: 34px; height: 48px; visibility: hidden">
         <a href="webacc?taskId=dev.Empty&amp;merge=fw.Body&amp;view=Tasks" target="Mainscreen" onclick="switch_icon('fw.TasksViewhint', 'fw.TasksView')" onmouseover="show_hint('fw.TasksViewhint', 'fw.TasksView')" onmouseout="hide_hint('fw.TasksViewhint', 'fw.TasksView')">
            <IMG src="/bizxxxx/portal/modules/fw/images/but_nl_tasks2.gif" border="0" title="职能和任务" alt="职能和任务"/>
         </a>
      </div>
      <div id="fw.TasksView3" style="position: absolute; z-index: 4; top: 22px; left: 328px; width: 34px; height: 48px; visibility: hidden">
         <a href="webacc?taskId=dev.Empty&amp;merge=fw.Body&amp;view=Tasks" target="Mainscreen" onclick="switch_icon('fw.TasksViewhint', 'fw.TasksView')" onmouseover="show_hint('fw.TasksViewhint', 'fw.TasksView')" onmouseout="hide_hint('fw.TasksViewhint', 'fw.TasksView')">
            <IMG src="/bizxxxx/portal/modules/fw/images/but_nl_tasks3.gif" border="0" title="已选择“职能和任务”" alt="已选择“职能和任务”"/>
         </a>
      </div>
      <div class="hint1" id="fw.TasksViewhint" style=" color: #fff; position: absolute; z-index: 6; top: 4px; left: 325px; width: 200px; visibility: hidden">
         职能和任务
      </div>
      
   
      
      <div id="fw.ObjectViewView" style="position: absolute; z-index: 2; top: 22px; left: 362px; width: 34px; height: 48px; visibility: visible">
         <a href="webacc?taskId=dev.Empty&amp;merge=fw.OV.ObjectView" target="Mainscreen" onclick="switch_icon('fw.ObjectViewViewhint', 'fw.ObjectViewView')" onmouseover="show_hint('fw.ObjectViewViewhint', 'fw.ObjectViewView')" onmouseout="hide_hint('fw.ObjectViewViewhint', 'fw.ObjectViewView')">
            <IMG src="/bizxxxx/portal/modules/fw/images/but_nl_obj1.gif" border="0" title="查看对象" alt="查看对象"/>
         </a>
      </div>
      <div id="fw.ObjectViewView2" style="position: absolute; z-index: 3; top: 22px; left: 362px; width: 34px; height: 48px; visibility: hidden">
         <a href="webacc?taskId=dev.Empty&amp;merge=fw.OV.ObjectView" target="Mainscreen" onclick="switch_icon('fw.ObjectViewViewhint', 'fw.ObjectViewView')" onmouseover="show_hint('fw.ObjectViewViewhint', 'fw.ObjectViewView')" onmouseout="hide_hint('fw.ObjectViewViewhint', 'fw.ObjectViewView')">
            <IMG src="/bizxxxx/portal/modules/fw/images/but_nl_obj2.gif" border="0" title="查看对象" alt="查看对象"/>
         </a>
      </div>
      <div id="fw.ObjectViewView3" style="position: absolute; z-index: 4; top: 22px; left: 362px; width: 34px; height: 48px; visibility: hidden">
         <a href="webacc?taskId=dev.Empty&amp;merge=fw.OV.ObjectView" target="Mainscreen" onclick="switch_icon('fw.ObjectViewViewhint', 'fw.ObjectViewView')" onmouseover="show_hint('fw.ObjectViewViewhint', 'fw.ObjectViewView')" onmouseout="hide_hint('fw.ObjectViewViewhint', 'fw.ObjectViewView')">
            <IMG src="/bizxxxx/portal/modules/fw/images/but_nl_obj3.gif" border="0" title="已选择“查看对象”" alt="已选择“查看对象”"/>
         </a>
      </div>
      <div class="hint1" id="fw.ObjectViewViewhint" style=" color: #fff; position: absolute; z-index: 6; top: 4px; left: 359px; width: 200px; visibility: hidden">
         查看对象
      </div>
      
   
      
      <div id="fw.ConfigureView" style="position: absolute; z-index: 2; top: 22px; left: 396px; width: 34px; height: 48px; visibility: visible">
         <a href="webacc?taskId=dev.Empty&amp;merge=fw.Body&amp;view=Config" target="Mainscreen" onclick="switch_icon('fw.ConfigureViewhint', 'fw.ConfigureView')" onmouseover="show_hint('fw.ConfigureViewhint', 'fw.ConfigureView')" onmouseout="hide_hint('fw.ConfigureViewhint', 'fw.ConfigureView')">
            <IMG src="/bizxxxx/portal/modules/fw/images/but_nl_admin1.gif" border="0" title="配置" alt="配置"/>
         </a>
      </div>
      <div id="fw.ConfigureView2" style="position: absolute; z-index: 3; top: 22px; left: 396px; width: 34px; height: 48px; visibility: hidden">
         <a href="webacc?taskId=dev.Empty&amp;merge=fw.Body&amp;view=Config" target="Mainscreen" onclick="switch_icon('fw.ConfigureViewhint', 'fw.ConfigureView')" onmouseover="show_hint('fw.ConfigureViewhint', 'fw.ConfigureView')" onmouseout="hide_hint('fw.ConfigureViewhint', 'fw.ConfigureView')">
            <IMG src="/bizxxxx/portal/modules/fw/images/but_nl_admin2.gif" border="0" title="配置" alt="配置"/>
         </a>
      </div>
      <div id="fw.ConfigureView3" style="position: absolute; z-index: 4; top: 22px; left: 396px; width: 34px; height: 48px; visibility: hidden">
         <a href="webacc?taskId=dev.Empty&amp;merge=fw.Body&amp;view=Config" target="Mainscreen" onclick="switch_icon('fw.ConfigureViewhint', 'fw.ConfigureView')" onmouseover="show_hint('fw.ConfigureViewhint', 'fw.ConfigureView')" onmouseout="hide_hint('fw.ConfigureViewhint', 'fw.ConfigureView')">
            <IMG src="/bizxxxx/portal/modules/fw/images/but_nl_admin3.gif" border="0" title="已选择“配置”" alt="已选择“配置”"/>
         </a>
      </div>
      <div class="hint1" id="fw.ConfigureViewhint" style=" color: #fff; position: absolute; z-index: 6; top: 4px; left: 393px; width: 200px; visibility: hidden">
         配置
      </div>
      
   
      
      <div id="fw.FavoritesView" style="position: absolute; z-index: 2; top: 22px; left: 430px; width: 34px; height: 48px; visibility: visible">
         <a href="webacc?taskId=dev.Empty&amp;merge=fw.Body&amp;view=Favorites" target="Mainscreen" onclick="switch_icon('fw.FavoritesViewhint', 'fw.FavoritesView')" onmouseover="show_hint('fw.FavoritesViewhint', 'fw.FavoritesView')" onmouseout="hide_hint('fw.FavoritesViewhint', 'fw.FavoritesView')">
            <IMG src="/bizxxxx/portal/modules/fw/images/but_nl_favorite1.gif" border="0" title="书签" alt="书签"/>
         </a>
      </div>
      <div id="fw.FavoritesView2" style="position: absolute; z-index: 3; top: 22px; left: 430px; width: 34px; height: 48px; visibility: hidden">
         <a href="webacc?taskId=dev.Empty&amp;merge=fw.Body&amp;view=Favorites" target="Mainscreen" onclick="switch_icon('fw.FavoritesViewhint', 'fw.FavoritesView')" onmouseover="show_hint('fw.FavoritesViewhint', 'fw.FavoritesView')" onmouseout="hide_hint('fw.FavoritesViewhint', 'fw.FavoritesView')">
            <IMG src="/bizxxxx/portal/modules/fw/images/but_nl_favorite2.gif" border="0" title="书签" alt="书签"/>
         </a>
      </div>
      <div id="fw.FavoritesView3" style="position: absolute; z-index: 4; top: 22px; left: 430px; width: 34px; height: 48px; visibility: hidden">
         <a href="webacc?taskId=dev.Empty&amp;merge=fw.Body&amp;view=Favorites" target="Mainscreen" onclick="switch_icon('fw.FavoritesViewhint', 'fw.FavoritesView')" onmouseover="show_hint('fw.FavoritesViewhint', 'fw.FavoritesView')" onmouseout="hide_hint('fw.FavoritesViewhint', 'fw.FavoritesView')">
            <IMG src="/bizxxxx/portal/modules/fw/images/but_nl_favorite3.gif" border="0" title="已选择“书签”" alt="已选择“书签”"/>
         </a>
      </div>
      <div class="hint1" id="fw.FavoritesViewhint" style=" color: #fff; position: absolute; z-index: 6; top: 4px; left: 427px; width: 200px; visibility: hidden">
         书签
      </div>
      
   
      
      <div id="fw.PreferencesView" style="position: absolute; z-index: 2; top: 22px; left: 464px; width: 34px; height: 48px; visibility: visible">
         <a href="webacc?taskId=dev.Empty&amp;merge=fw.Body&amp;view=Preferences" target="Mainscreen" onclick="switch_icon('fw.PreferencesViewhint', 'fw.PreferencesView')" onmouseover="show_hint('fw.PreferencesViewhint', 'fw.PreferencesView')" onmouseout="hide_hint('fw.PreferencesViewhint', 'fw.PreferencesView')">
            <IMG src="/bizxxxx/portal/modules/fw/images/but_nl_prefs1.gif" border="0" title="自选设置" alt="自选设置"/>
         </a>
      </div>
      <div id="fw.PreferencesView2" style="position: absolute; z-index: 3; top: 22px; left: 464px; width: 34px; height: 48px; visibility: hidden">
         <a href="webacc?taskId=dev.Empty&amp;merge=fw.Body&amp;view=Preferences" target="Mainscreen" onclick="switch_icon('fw.PreferencesViewhint', 'fw.PreferencesView')" onmouseover="show_hint('fw.PreferencesViewhint', 'fw.PreferencesView')" onmouseout="hide_hint('fw.PreferencesViewhint', 'fw.PreferencesView')">
            <IMG src="/bizxxxx/portal/modules/fw/images/but_nl_prefs2.gif" border="0" title="自选设置" alt="自选设置"/>
         </a>
      </div>
      <div id="fw.PreferencesView3" style="position: absolute; z-index: 4; top: 22px; left: 464px; width: 34px; height: 48px; visibility: hidden">
         <a href="webacc?taskId=dev.Empty&amp;merge=fw.Body&amp;view=Preferences" target="Mainscreen" onclick="switch_icon('fw.PreferencesViewhint', 'fw.PreferencesView')" onmouseover="show_hint('fw.PreferencesViewhint', 'fw.PreferencesView')" onmouseout="hide_hint('fw.PreferencesViewhint', 'fw.PreferencesView')">
            <IMG src="/bizxxxx/portal/modules/fw/images/but_nl_prefs3.gif" border="0" title="已选择“自选设置”" alt="已选择“自选设置”"/>
         </a>
      </div>
      <div class="hint1" id="fw.PreferencesViewhint" style=" color: #fff; position: absolute; z-index: 6; top: 4px; left: 461px; width: 200px; visibility: hidden">
         自选设置
      </div>
      
   
      
      <div id="IDM.View.Administration" style="position: absolute; z-index: 2; top: 22px; left: 498px; width: 34px; height: 48px; visibility: visible">
         <a href="webacc?taskId=IDM.Task.AdministrationView" target="Mainscreen" onclick="switch_icon('IDM.View.Administrationhint', 'IDM.View.Administration')" onmouseover="show_hint('IDM.View.Administrationhint', 'IDM.View.Administration')" onmouseout="hide_hint('IDM.View.Administrationhint', 'IDM.View.Administration')">
            <IMG src="/bizxxxx/portal/modules/DirXML/images/IDMViewNormal.gif" border="0" title="Identity Manager 管理" alt="Identity Manager 管理"/>
         </a>
      </div>
      <div id="IDM.View.Administration2" style="position: absolute; z-index: 3; top: 22px; left: 498px; width: 34px; height: 48px; visibility: hidden">
         <a href="webacc?taskId=IDM.Task.AdministrationView" target="Mainscreen" onclick="switch_icon('IDM.View.Administrationhint', 'IDM.View.Administration')" onmouseover="show_hint('IDM.View.Administrationhint', 'IDM.View.Administration')" onmouseout="hide_hint('IDM.View.Administrationhint', 'IDM.View.Administration')">
            <IMG src="/bizxxxx/portal/modules/DirXML/images/IDMViewMouseOver.gif" border="0" title="Identity Manager 管理" alt="Identity Manager 管理"/>
         </a>
      </div>
      <div id="IDM.View.Administration3" style="position: absolute; z-index: 4; top: 22px; left: 498px; width: 34px; height: 48px; visibility: hidden">
         <a href="webacc?taskId=IDM.Task.AdministrationView" target="Mainscreen" onclick="switch_icon('IDM.View.Administrationhint', 'IDM.View.Administration')" onmouseover="show_hint('IDM.View.Administrationhint', 'IDM.View.Administration')" onmouseout="hide_hint('IDM.View.Administrationhint', 'IDM.View.Administration')">
            <IMG src="/bizxxxx/portal/modules/DirXML/images/IDMViewSelected.gif" border="0" title="Identity Manager 管理" alt="Identity Manager 管理"/>
         </a>
      </div>
      <div class="hint1" id="IDM.View.Administrationhint" style=" color: #fff; position: absolute; z-index: 6; top: 4px; left: 495px; width: 200px; visibility: hidden">
         Identity Manager 管理
      </div>
      
   
      
      <div id="nrm.NrmView" style="position: absolute; z-index: 2; top: 22px; left: 532px; width: 34px; height: 48px; visibility: visible">
         <a href="webacc?taskId=nrm.NrmViewTask" target="Mainscreen" onclick="switch_icon('nrm.NrmViewhint', 'nrm.NrmView')" onmouseover="show_hint('nrm.NrmViewhint', 'nrm.NrmView')" onmouseout="hide_hint('nrm.NrmViewhint', 'nrm.NrmView')">
            <IMG src="/bizxxxx/portal/modules/nrm/images/but_nl_zen1.gif" border="0" title="ZENworks 控制中心" alt="ZENworks 控制中心"/>
         </a>
      </div>
      <div id="nrm.NrmView2" style="position: absolute; z-index: 3; top: 22px; left: 532px; width: 34px; height: 48px; visibility: hidden">
         <a href="webacc?taskId=nrm.NrmViewTask" target="Mainscreen" onclick="switch_icon('nrm.NrmViewhint', 'nrm.NrmView')" onmouseover="show_hint('nrm.NrmViewhint', 'nrm.NrmView')" onmouseout="hide_hint('nrm.NrmViewhint', 'nrm.NrmView')">
            <IMG src="/bizxxxx/portal/modules/nrm/images/but_nl_zen2.gif" border="0" title="ZENworks 控制中心" alt="ZENworks 控制中心"/>
         </a>
      </div>
      <div id="nrm.NrmView3" style="position: absolute; z-index: 4; top: 22px; left: 532px; width: 34px; height: 48px; visibility: hidden">
         <a href="webacc?taskId=nrm.NrmViewTask" target="Mainscreen" onclick="switch_icon('nrm.NrmViewhint', 'nrm.NrmView')" onmouseover="show_hint('nrm.NrmViewhint', 'nrm.NrmView')" onmouseout="hide_hint('nrm.NrmViewhint', 'nrm.NrmView')">
            <IMG src="/bizxxxx/portal/modules/nrm/images/but_nl_zen3.gif" border="0" title="Novell ZENworks 控制中心" alt="Novell ZENworks 控制中心"/>
         </a>
      </div>
      <div class="hint1" id="nrm.NrmViewhint" style=" color: #fff; position: absolute; z-index: 6; top: 4px; left: 529px; width: 200px; visibility: hidden">
         ZENworks 控制中心
      </div>
      
   


   
      <script language="javascript">
           document.getElementById(curImageID).style.visibility = "hidden";
           document.getElementById(curImageID+'2').style.visibility = "hidden";
           document.getElementById(curImageID+'3').style.visibility = "visible";
      </script>
   

   <div id="buthelp" style="left: 576px;">
      <a href="javascript:launchHelp('fw/helpindex.html');" onmouseover="show_hint('helphint','buthelp')" onmouseout="hide_hint('helphint','buthelp')">
         <IMG height="48" width="34" id="help" src="/bizxxxx/portal/modules/fw/images/but_nl_help1.gif" border="0" title="帮助" alt="帮助"/>
      </a>
   </div>
   <div id="buthelp2" style="left: 576px;">
      <a href="javascript:launchHelp('fw/helpindex.html');" onmouseover="show_hint('helphint','buthelp')" onmouseout="hide_hint('helphint','buthelp')">
         <IMG height="48" width="34" id="help" src="/bizxxxx/portal/modules/fw/images/but_nl_help2.gif" border="0" title="帮助" alt="帮助"/>
      </a>
   </div>
   <div id="helphint" class="hint1" style="left: 576px;">
      帮助
   </div>
</body>
</html>
`
)
