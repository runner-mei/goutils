package httpauth

import "testing"

func TestCrypto(t *testing.T) {

	e := "010001"
	m := "00a9065378eddc455c15143b4a733fdcb3ef29c4e7598522c5fcfff580d5d98dbbcb3e132beae4fb5d5b5db6342cb4f455e84c9f9488663fd59c3676c99ea8c32463a0a0b75688ad364e9e12dbc4cec2fb331ee58bc3881c9869babd1b10677e39d5cb7c30f23be7547b2e6d8ed2cae8942e2767efc7ec804286e01484533ab47f"
	// envilope := "ab222585dbce65a736de2db2a56133bf!,!cyKzsQfFnT!,!2cjnx123*"
	//             "ab222585dbce65a736de2db2a56133bf!,!cyKzsQfFnT!,!2cjnx123*"
	result := "814133b3fc33769b0d383fc004c631fff7ab247d3e10aa5c035a7a7b959b31c2ff303cfe5376a53f5a81a5945a4e3765be4bc4892c250f672a2e1a3c09be076548b98a1d11af0dd810b228c41b14aa7c09ab1c6a463cf4e8d1061706ed2c83a8350db59a418fc3e2ee0f86210f4d68ce8068786c84e70171dce922c4877fa8a0"
	random := "cyKzsQfFnT"
	content := "2cjnx123*"

	a := createSecurityData2(m, e, random, content)

	if a != result {
		t.Error("actual  ", a)
		t.Error("excepted", result)
	}
}
