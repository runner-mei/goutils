package cfg

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestUpdate(t *testing.T) {
	input := `
# asdfsd
# a=b
# sdfasdf
c=d
noupdate=a
`

	excepted := "\r\n" +
		"# asdfsd\r\n" +
		"a=1\r\n" +
		"# sdfasdf\r\n" +
		"c=2\r\n" +
		"noupdate=a\r\n" +
		"notExists=3\r\n"

	var buf bytes.Buffer

	e := UpdateWith(strings.NewReader(input), &buf, map[string]string{"a": "1", "c": "2", "notExists": "3"})
	if e != nil {
		t.Error(e)
		return
	}

	if excepted != buf.String() {
		fmt.Println("actual   =============")
		fmt.Println(buf.String())
		fmt.Println("excepted =============")
		fmt.Println(excepted)
		t.Error(buf.String())
	}
}

func TestRead(t *testing.T) {
	input := `
# asdfsd
# a=b
#sdfasdf
c=b  #abc
d1=d1
d2 =d2
d3= d3
d4= d#4
d#5= d#5
"aaaa
"d6"= d6
"d7" = d7
"d 8" = d8
"d#9" = d9
						daemon.host=127.0.0.1
						daemon.port=51791
  d10 = d10  
  d11 = a${d10}b
  d12 = a${d11}b
  pa=c:\a.txt

phone =  "24小时服务热线：63846590×776295 、776702;15301688819"
email = MAIL地址：service@shsmu.edu.cn

d13 = d 13
d14 = d 1 4
d15 = d 1  5
`

	excepted := map[string]string{"c": "b",
		"d1":          "d1",
		"d2":          "d2",
		"d3":          "d3",
		"d4":          "d#4",
		"d#5":         "d#5",
		"d6":          "d6",
		"d7":          "d7",
		"d 8":         "d8",
		"d#9":         "d9",
		"daemon.host": "127.0.0.1",
		"daemon.port": "51791",
		"d10":         "d10",
		"d11":         "ad10b",
		"d12":         "aad10bb",
		"pa":          "c:\\a.txt",

		"phone": "24小时服务热线：63846590×776295 、776702;15301688819",
		"email": "MAIL地址：service@shsmu.edu.cn",

		"d13" : "d 13",
		"d14" : "d 1 4",
		"d15" : "d 1  5",
	}

	actual, e := Read(strings.NewReader(input))
	if e != nil {
		t.Error(e)
		return
	}

	if len(actual) != len(excepted) {
		t.Error("actual   count is", len(actual))
		t.Error("excepted count is", len(excepted))

		t.Error("actual   is", actual)
		t.Error("excepted is", excepted)
	}
	for k, v := range excepted {
		if a, ok := actual[k]; ok {
			if v != a {
				t.Error(k, "actual   is", a)
				t.Error(k, "excepted is", v)
			}
		} else {
			t.Error(k, "actual   is not exists")
			t.Error(k, "excepted is", v)
		}
	}
}
