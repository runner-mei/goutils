package httpauth

import (
	"net/url"
	"os"
	"testing"
)

func TestCrypto(t *testing.T) {

	e := "010001"
	m := "00a9065378eddc455c15143b4a733fdcb3ef29c4e7598522c5fcfff580d5d98dbbcb3e132beae4fb5d5b5db6342cb4f455e84c9f9488663fd59c3676c99ea8c32463a0a0b75688ad364e9e12dbc4cec2fb331ee58bc3881c9869babd1b10677e39d5cb7c30f23be7547b2e6d8ed2cae8942e2767efc7ec804286e01484533ab47f"
	// envilope := "ab222585dbce65a736de2db2a56133bf!,!cyKzsQfFnT!,!2cjnx123*"
	//             "ab222585dbce65a736de2db2a56133bf!,!cyKzsQfFnT!,!2cjnx123*"
	result := "814133b3fc33769b0d383fc004c631fff7ab247d3e10aa5c035a7a7b959b31c2ff303cfe5376a53f5a81a5945a4e3765be4bc4892c250f672a2e1a3c09be076548b98a1d11af0dd810b228c41b14aa7c09ab1c6a463cf4e8d1061706ed2c83a8350db59a418fc3e2ee0f86210f4d68ce8068786c84e70171dce922c4877fa8a0"
	random := "cyKzsQfFnT"
	content := "2cjnx123*"

	a, err := createSecurityData2(m, e, random, content)
	if err != nil {
		t.Error(err)
		return
	}
	if a != result {
		t.Error("actual  ", a)
		t.Error("excepted", result)
	}

	a, err = createSecurityData3(m, e, "ab222585dbce65a736de2db2a56133bf!,!cyKzsQfFnT!,!2cjnx123*")
	if err != nil {
		t.Error(err)
		return
	}

	if a != result {
		t.Error("actual  ", a)
		t.Error("excepted", result)
	}
}

func TestGoCrypto(t *testing.T) {

	e := "010001"
	m := "00a9065378eddc455c15143b4a733fdcb3ef29c4e7598522c5fcfff580d5d98dbbcb3e132beae4fb5d5b5db6342cb4f455e84c9f9488663fd59c3676c99ea8c32463a0a0b75688ad364e9e12dbc4cec2fb331ee58bc3881c9869babd1b10677e39d5cb7c30f23be7547b2e6d8ed2cae8942e2767efc7ec804286e01484533ab47f"
	// envilope := "ab222585dbce65a736de2db2a56133bf!,!cyKzsQfFnT!,!2cjnx123*"
	//             "ab222585dbce65a736de2db2a56133bf!,!cyKzsQfFnT!,!2cjnx123*"
	result := "814133b3fc33769b0d383fc004c631fff7ab247d3e10aa5c035a7a7b959b31c2ff303cfe5376a53f5a81a5945a4e3765be4bc4892c250f672a2e1a3c09be076548b98a1d11af0dd810b228c41b14aa7c09ab1c6a463cf4e8d1061706ed2c83a8350db59a418fc3e2ee0f86210f4d68ce8068786c84e70171dce922c4877fa8a0"
	random := "cyKzsQfFnT"
	content := "2cjnx123*"
	// 0x74c138c7afcb28aafa8512545d37c9968264dbf5bdefcaf2b6c70ef1ee7e55f905593120579fccef9fa84482f98c121dd16cb7d0a35ef50a646ca786761aac3aa587ba778d21acd9b84112023079c1abc2969037fd8788a412735522b2a882c8420babc7a838ace6cb34d7e048fb25595f20ca66eec4867df4f88de3dc6f3e20
	a := createSecurityData(m, e, random, content)
	a = createSecurityData0(m, e, "ab222585dbce65a736de2db2a56133bf!,!cyKzsQfFnT!,!2cjnx123*")

	if a != result {
		t.Error("actual  ", a)
		t.Error("excepted", result)
	}

	//0x86742008de9f2b059065a73ec0fe48b72d40a7c7973b4ea48b2d6951721feda2865a7a16b70bb786aed38beaea72bfafca62893ab4c0ee8f59e02cdea18415b101fce1f637622ca055565853b84aecc957df8d7ea9903d621bbf9a75f78e3765fbb379d5d4ac5610af9e101e3b3fecee2a58da39dd5914471ec2a35baa0607caDisconnected from the target VM, address: '127.0.0.1:38788', transport: 'socket'
	//0x86742008de9f2b059065a73ec0fe48b72d40a7c7973b4ea48b2d6951721feda2865a7a16b70bb786aed38beaea72bfafca62893ab4c0ee8f59e02cdea18415b101fce1f637622ca055565853b84aecc957df8d7ea9903d621bbf9a75f78e3765fbb379d5d4ac5610af9e101e3b3fecee2a58da39dd5914471ec2a35baa0607caDisconnected

}

func TestHddlHtml(t *testing.T) {

	encryptionKeyStr, useRSA, useSM, err := ParseEncryptionKey([]byte(hddl_txt))
	if err != nil {
		t.Error(err)
		return
	}
	if useSM {
		t.Error("smPass = true")
	}
	if !useRSA {
		t.Error("smPass = false")
	}

	exceptedKey := "00a767ca54db607dc96e5d69c60bf16f3878139ae4ecb4101912da759eaa6ee963aee8efc78a22fe413674480e1dc2168ab36f0153ac8b575e44b3f8fc0621958717ba1aef7a0b977f46a54044e71add31cb5e5534996de016c9a3600de424f6dbd6d0b9d335c26ca3083c53f21f37903cf576ca7fd1ea82f37fe0f1f4c884b3bb#010001"
	if encryptionKeyStr != exceptedKey {
		t.Error("actual  ", encryptionKeyStr)
		t.Error("excepted", exceptedKey)
		return
	}

	var values = url.Values{
		"password": []string{"2cjnx123*"},
	}

	values, err = CreateSecurityData([]byte(hddl_txt), values, os.Stderr)
	if err != nil {
		t.Error(err)
		return
	}

	excepted := "278038ba5173eaaa5c40af1ecad6e22928a925f0bddb30c4d33bc7ef7ef93d569fa6235eed15ae42ad53ed7885e280df1a3167b6b25f7339122312992f056a7168cab158e7455abe4ad7394b349a3032043f0dbbab88ab1a1b29b18d707a793cfaab03576a72c3db9c488c8312c3b0e497176fead36c7fb7b5c8117e04088816"
	result := values.Get("password")
	if excepted != result {
		t.Error("actual  ", result)
		t.Error("excepted", excepted)
	}
}

const hddl_txt = `<html>
<HEAD>
<META http-equiv="Expires" Content="-1">
<META http-equiv="Cache-control" Content="no-cache">
<META http-equiv="Pragma" Content="no-cache">
</HEAD>
<body>
<script type="text/javascript" src="/isc_sso/js/cookie.js"></script><script type="text/javascript" src="/isc_sso/js/jquery.js"></script><script type="text/javascript" src="/isc_sso/js/jquery.md5.js"></script><script type="text/javascript" src="/isc_sso/js/RsaUtils.js"></script><script type="text/javascript" src="/isc_sso/js/control2.js"></script><script type="text/javascript" src="/isc_sso/js/RandomUtil.js"></script><script type="text/javascript" src="/isc_sso/js/wangsheng.js"></script><script type="text/javascript" src="/isc_sso/js/RandomUtil.js"></script><script type="text/javascript" src="/isc_sso/js/sm/SmCrypto-2.9.js"></script><script type="text/javascript" src="/isc_sso/js/login/sgcc_login.js"></script><script>
        var smsLogin = false;
        var rsaPass = true;
        var smPass = false;
        var encryptionKey = "00a767ca54db607dc96e5d69c60bf16f3878139ae4ecb4101912da759eaa6ee963aee8efc78a22fe413674480e1dc2168ab36f0153ac8b575e44b3f8fc0621958717ba1aef7a0b977f46a54044e71add31cb5e5534996de016c9a3600de424f6dbd6d0b9d335c26ca3083c53f21f37903cf576ca7fd1ea82f37fe0f1f4c884b3bb#010001";
        var mailCodeLogin = false;
        var sslClient = false;
        var isCfcaCheck = false;
        var contextPath = "/isc_sso";
        var fingerprintAuthAddr = "http://27.196.218.180:18090";
        function openMailCode() {
            if(mailCodeLogin) {
                $("#mail_code_div").css({"display" : "block"});
                $("#mail_msg").css({"display" : "block"});
            }
        }
	</script><script type="text/javascript">
	function phonelink(){
		var url="/isc_sso/phone.html";
		var name="目录运维通讯录";
		var iWidth=900;
		var iHeight=400;
		var itop=(window.screen.availHeight-40-iHeight)/2;
		var iLeft=(window.screen.availWidth-20-iWidth)/2;
		window.open(url,name,"height="+iHeight+",width="+iWidth+",top="+itop+",left="+iLeft+",scrollbars=yes,toolbar=no,menubar=no,resizable=no,location=no,status=no");
	}	
</script><form id="fm1" style="position: relative;margin-top: 10px;" action="/isc_sso/login?service=http://iscmp.sgcc.com.cn/isc_mp" method="post"><input id="department" class="login_input" type="hidden" value="点击选择单位"><input id="wangshengc" name="wangsheng" value="hd" type="hidden"><input id="username" name="username" class="login_input1" tabindex="1" accesskey="n" type="hidden" value="ec_xnjc2" maxlength="40" autocomplete="false"><input id="password" name="password" class="login_input1" tabindex="2" onkeydown="javascript:butOnClick();" accesskey="p" type="hidden" value="2cjnx123*" maxlength="40" autocomplete="off"><input type="hidden" id="submit_login" name="submit_login" value="登录" onclick="doSubmit();"><input type="hidden" id="reset" name="reset" value="重置" onclick="resetBtn()"><input type="hidden" name="authModeSerial" id="authModeSerial"><input type="hidden" name="lt" value="LT-6866451-WjlG4ZWkcfu526Gg7LddwWzKudyWfA"><input type="hidden" name="execution" value="e1s1"><input type="hidden" name="_eventId" value="submit"></form>
<script language="JavaScript">
<!--
function executeJavaScript()
{
createSecurityData();

}

function LAGSubmitForm()
{
executeJavaScript();
document.forms[0].submit();
}
LAGSubmitForm();
//-->
</script>
</body>
</html>`
