package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// var block cipher.Block

// func init() {
// 	var err error
// 	block, err = aes.NewCipher([]byte("123sdfkl79345drg987ne4tr"))
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func pkcs5Padding(src []byte, blockSize int) []byte {
// 	padding := blockSize - len(src)%blockSize
// 	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
// 	return append(src, padtext...)
// }

// func pkcs5UnPadding(src []byte) []byte {
// 	length := len(src)
// 	unpadding := int(src[length-1])
// 	return src[:(length - unpadding)]
// }

// func decryptoString(t string) string {
// 	if strings.HasPrefix(t, "[1]") {
// 		s, e := hex.DecodeString(strings.TrimPrefix(t, "[1]"))
// 		if nil != e {
// 			return ""
// 		}

// 		if len(s) < aes.BlockSize {
// 			return ""
// 		}

// 		iv := s[:aes.BlockSize]
// 		s = s[aes.BlockSize:]

// 		// CBC mode always works in whole blocks.
// 		if len(s)%aes.BlockSize != 0 {
// 			return ""
// 		}

// 		mode := cipher.NewCBCDecrypter(block, iv)
// 		mode.CryptBlocks(s, s)
// 		return string(pkcs5UnPadding(s))
// 	}
// 	return t
// }

// // PasswordEqual 判断两个密码是否相等，actual 是 Aes 加密的， excepted 可能是用 sha256 或 sha512 加密过的
// func PasswordEqual(actual, excepted string) bool {
// 	token := decryptoString(actual)
// 	if strings.HasPrefix(excepted, "sha256") {
// 		tokenSha := sha256.Sum256([]byte(token))
// 		// exceptedSha, err := hex.DecodeString(strings.TrimPrefix(excepted, "sha256"))
// 		// if err != nil {
// 		// 	panic(err)
// 		// }

// 		return ("sha256" + hex.EncodeToString(tokenSha[:])) == excepted
// 	} else if strings.HasPrefix(excepted, "sha512") {
// 		tokenSha := sha512.Sum512([]byte(token))
// 		return ("sha512" + hex.EncodeToString(tokenSha[:])) == excepted
// 	}
// 	return token == excepted
// }

// // func CryptoPassword(actual string) string {
// // 	return actual
// // }

// // PasswordHash 生成一个 hash 之后的密码
// func PasswordHash(pwd string) string {
// 	tokenSha := sha512.Sum512([]byte(pwd))
// 	return "sha512" + hex.EncodeToString(tokenSha[:])
// }

func Encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(text))
	return ciphertext, nil
}

func Decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	return text, nil
}
