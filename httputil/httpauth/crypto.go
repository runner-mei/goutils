package httpauth

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/url"
)

func MD5(src string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(src))
	cipherStr := md5Ctx.Sum(nil)
	return base64.StdEncoding.EncodeToString(cipherStr)
}

func CreateSecurityData(values url.Values) string {
	m := values.Get("m")
	e := values.Get("e")

	content := values.Get("password")

	return createSecurityData(m, e, "abc", content)
}

func createSecurityData(m, e, random, content string) string {
	// var key = RSAUtils.getKeyPair(e, '', m);
	pub := &rsa.PublicKey{
		N: fromBase16(m),
		E: fromBase16(e),
	}

	md5 := MD5(content)
	var envilope = md5 + "!,!" + random + "!,!" + content

	data, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(envilope))
	if err != nil {
		panic(fmt.Errorf("m=%s,e=%s, random=%s, content=%content", m, e, random, content))
	}
	return hex.EncodeToString(data)
	// return RSAUtils.encryptedString(key, envilope)
}

func fromBase16(base16 string) *big.Int {
	i, ok := new(big.Int).SetString(base16, 16)
	if !ok {
		panic("bad number: " + base16)
	}
	return i
}
