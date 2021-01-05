package httpauth

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/dop251/goja"
)

func MD5(src string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(src))
	cipherStr := md5Ctx.Sum(nil)
	// return  base64.StdEncoding.EncodeToString(cipherStr)
	return hex.EncodeToString(cipherStr)
}

var rsaPass = regexp.MustCompile(`var\s*rsaPass\s*=\s*true;`)
var smPass = regexp.MustCompile(`var\s*smPass\s*=\s*true;`)
var encryptionKey = regexp.MustCompile(`var\s+encryptionKey\s*=\s*"([^"]*)"\s*;`)

func ParseEncryptionKey(responseBody []byte) (key string, useRsa, useSM bool, err error) {
	aa := encryptionKey.FindSubmatch(responseBody)
	if len(aa) != 2 || len(aa[0]) == 0 {
		err = errors.New("没有找到 encryptionKey")
		return
	}

	key = string(aa[1])
	// fmt.Println(string(aa[1]))

	useRsa = rsaPass.Match(responseBody)
	useSM = smPass.Match(responseBody)
	return
}

func CreateSecurityData(responseBody []byte, values url.Values) (url.Values, error) {
	encryptionKeyStr, useRsa, _, err := ParseEncryptionKey(responseBody)
	if err != nil {
		return nil, err
	}
	if useRsa {
		return CreateSecurityDataWithRsa(values, encryptionKeyStr)
	}

	return CreateSecurityDataWithSM(values, encryptionKeyStr)
}

func CreateSecurityDataWithSM(values url.Values, encryptionKeyStr string) (url.Values, error) {
	pwd := values.Get("password")

	// function getEncryptPwd(pwd) {
	//     var encryptPwd = "";
	//     if(smPass) {
	//         var sm2Utils = new Sm2Utils(CipherMode.c1c3c2);
	//         var sm3Pwd = Sm3Utils.encryptFromText(pwd)+getRandomString(8)+pwd;
	//         encryptPwd = sm2Utils.encryptFromText(encryptionKey,sm3Pwd);
	//     }
	//     if(rsaPass) {
	//         var keys = encryptionKey.split("#");
	//         var modulus = keys[0];
	//         var exponent = keys[1];
	//         //生成0-100之间的随机数
	//         var random=getRandomString(8);
	//         //获取key秘钥
	//         var key = RSAUtils.getKeyPair(exponent, '', modulus);
	//         //对密码进行md5信息摘要
	//         var envilope=$.md5(pwd)+random+pwd;
	//         //进行完整的信息连接，生成数字签名元数据
	//         encryptPwd = RSAUtils.encryptedString(key, envilope);
	//     }
	//     return encryptPwd;
	// }

	vm := goja.New()

	navigator := vm.NewObject()
	navigator.Set("appName", "Microsoft Internet Explorer")
	vm.Set("navigator", navigator)

	vm.Set("print", func(msg goja.Value, args ...goja.Value) {
		var a []interface{}
		for _, arg := range args {
			a = append(a, arg.Export())
		}

		fmt.Println(msg, a)
	})

	_, err := vm.RunString(smJs)
	if err != nil {
		return nil, err
	}

	var random = "cyKzsQfFnT"
	a, err := vm.RunString(`var sm2Utils = new Sm2Utils(CipherMode.c1c3c2);
        var sm3Pwd = Sm3Utils.encryptFromText("` + pwd + `")+"` + random + `"+ "` + pwd + `";
        sm2Utils.encryptFromText("` + encryptionKeyStr + `",sm3Pwd);`)
	if err != nil {
		return nil, err
	}

	s := a.Export()
	if s == nil {
		return nil, errors.New("加密失败")
	}
	fmt.Println(s)
	values.Set("password", s.(string))
	return values, nil
}

func CreateSecurityDataWithRsa(values url.Values, encryptionKeyStr string) (url.Values, error) {
	pwd := values.Get("password")

	var keys = strings.Split(encryptionKeyStr, "#")
	if len(keys) != 2 {
		return nil, errors.New("encryptionKeyStr is invalid - '" + encryptionKeyStr + "'")
	}
	var modulus = keys[0]
	var exponent = keys[1]

	if modulus == "" {
		return nil, errors.New("m is empty")
	}
	if exponent == "" {
		return nil, errors.New("e is empty")
	}
	if pwd == "" {
		return nil, errors.New("pwd is empty")
	}

	//生成0-100之间的随机数
	var random = "cyKzsQfFnT"

	newPwd, err := createSecurityData3(modulus, exponent, MD5(pwd)+random+pwd)
	if err != nil {
		return nil, err
	}
	values.Set("password", newPwd)
	return values, nil
}

func CreateSecurityDataOld(responseBody []byte, values url.Values) (url.Values, error) {
	m := values.Get("m")
	e := values.Get("e")

	content := values.Get("password")

	var err error
	content, err = createSecurityData2(m, e, "cyKzsQfFnT", content)
	if err != nil {
		return nil, err
	}
	values.Set("password", content)
	return values, nil
}

func createSecurityData(m, e, random, content string) string {
	// var key = RSAUtils.getKeyPair(e, '', m);
	pub := &rsa.PublicKey{
		N: fromBase16BigInt(m),
		E: fromBase16Int(e),
	}

	md5 := MD5(content)
	var envilope = md5 + "!,!" + random + "!,!" + content

	data, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(envilope))
	if err != nil {
		panic(fmt.Errorf("m=%s,e=%s, random=%s, content=%s", m, e, random, content))
	}
	return hex.EncodeToString(data)
	// return RSAUtils.encryptedString(key, envilope)
}

func createSecurityData0(m, e, content string) string {
	// var key = RSAUtils.getKeyPair(e, '', m);
	pub := &rsa.PublicKey{
		N: fromBase16BigInt(m),
		E: fromBase16Int(e),
	}

	data, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(content))
	if err != nil {
		panic(fmt.Errorf("m=%s,e=%s, content=%s", m, e, content))
	}
	return hex.EncodeToString(data)
	// return RSAUtils.encryptedString(key, envilope)
}

func fromBase16BigInt(base16 string) *big.Int {
	i, ok := new(big.Int).SetString(base16, 16)
	if !ok {
		panic("bad number1: " + base16)
	}
	return i
}

func fromBase16Int(base16 string) int {
	i, err := strconv.ParseInt(base16, 16, 0)
	if err != nil {
		fmt.Println(err)
		panic("bad number2: " + base16)
	}
	return int(i)
}

func createSecurityData2(m, e, random, content string) (string, error) {
	if m == "" {
		return "", errors.New("m is empty")
	}
	if e == "" {
		return "", errors.New("e is empty")
	}
	if content == "" {
		return "", errors.New("content is empty")
	}

	vm := goja.New()
	vm.Set("print", func(msg goja.Value, args ...goja.Value) {
		var a []interface{}
		for _, arg := range args {
			a = append(a, arg.Export())
		}

		fmt.Println(msg, a)
	})

	_, err := vm.RunString(rsaJS)
	if err != nil {
		return "", err
	}

	a, err := vm.RunString(`createSecurityData('` + m + `', '` + e + `', '` + random + `', '` + MD5(content) + `', '` + content + `')`)
	if err != nil {
		return "", err
	}

	s := a.Export()
	return s.(string), nil
}

func createSecurityData3(m, e, content string) (string, error) {
	vm := goja.New()

	vm.Set("print", func(msg goja.Value, args ...goja.Value) {
		var a []interface{}
		for _, arg := range args {
			a = append(a, arg.Export())
		}

		fmt.Println(msg, a)
	})

	_, err := vm.RunString(rsaJS)
	if err != nil {
		return "", err
	}

	a, err := vm.RunString(`createSecurityData3('` + m + `', '` + e + `', '` + content + `')`)
	if err != nil {
		return "", err
	}

	s := a.Export()
	return s.(string), nil
}

const rsaJS = `/*
			 * RSA, a suite of routines for performing RSA public-key computations in JavaScript.
			 * Copyright 1998-2005 David Shapiro.
			 * Dave Shapiro
			 * dave@ohdave.com 
			 * changed by Fuchun, 2010-05-06
			 * fcrpg2005@gmail.com
			 */

			var window = {};
			window.RSAUtils = {};

			(function($w) {
				//if (typeof $w.RSAUtils === 'undefined')
				var RSAUtils = $w.RSAUtils = {};

				var biRadixBase = 2;
				var biRadixBits = 16;
				var bitsPerDigit = biRadixBits;
				//var biRadix = 1 < < 16; // = 2^16 = 65536
				var biRadix = 65536
				var biHalfRadix = 65536/2;
				var biRadixSquared = biRadix * biRadix;
				var maxDigitVal = biRadix - 1;
				var maxInteger = 9999999999999998;

				//maxDigits:
				//Change this to accommodate your largest number size. Use setMaxDigits()
				//to change it!
				//
				//In general, if you're working with numbers of size N bits, you'll need 2*N
				//bits of storage. Each digit holds 16 bits. So, a 1024-bit key will need
				//
				//1024 * 2 / 16 = 128 digits of storage.
				//
				var maxDigits;
				var ZERO_ARRAY;
				var bigZero, bigOne;

				var BigInt = $w.BigInt = function(flag) {
					if (typeof flag == "boolean" && flag == true) {
						this.digits = null;
					} else {
						this.digits = ZERO_ARRAY.slice(0);
					}
					this.isNeg = false;
				};

				RSAUtils.setMaxDigits = function(value) {
					maxDigits = value;
					ZERO_ARRAY = new Array(maxDigits);
					for ( var iza = 0; iza < ZERO_ARRAY.length; iza++)
						ZERO_ARRAY[iza] = 0;
					bigZero = new BigInt();
					bigOne = new BigInt();
					bigOne.digits[0] = 1;
				};
				RSAUtils.setMaxDigits(20);

				//The maximum number of digits in base 10 you can convert to an
				//integer without JavaScript throwing up on you.
				var dpl10 = 15;

				RSAUtils.biFromNumber = function(i) {
					var result = new BigInt();
					result.isNeg = i < 0;
					i = Math.abs(i);
					var j = 0;
					while (i > 0) {
						result.digits[j++] = i & maxDigitVal;
						i = Math.floor(i / biRadix);
					}
					return result;
				};

				//lr10 = 10 ^ dpl10
				var lr10 = RSAUtils.biFromNumber(1000000000000000);

				RSAUtils.biFromDecimal = function(s) {
					var isNeg = s.charAt(0) == '-';
					var i = isNeg ? 1 : 0;
					var result;
					// Skip leading zeros.
					while (i < s.length && s.charAt(i) == '0')
						++i;
					if (i == s.length) {
						result = new BigInt();
					} else {
						var digitCount = s.length - i;
						var fgl = digitCount % dpl10;
						if (fgl == 0)
							fgl = dpl10;
						result = RSAUtils.biFromNumber(Number(s.substr(i, fgl)));
						i += fgl;
						while (i < s.length) {
							result = RSAUtils.biAdd(RSAUtils.biMultiply(result, lr10),
									RSAUtils.biFromNumber(Number(s.substr(i, dpl10))));
							i += dpl10;
						}
						result.isNeg = isNeg;
					}
					return result;
				};

				RSAUtils.biCopy = function(bi) {
					var result = new BigInt(true);
					result.digits = bi.digits.slice(0);
					result.isNeg = bi.isNeg;
					return result;
				};

				RSAUtils.reverseStr = function(s) {
					var result = "";
					for ( var i = s.length - 1; i > -1; --i) {
						result += s.charAt(i);
					}
					return result;
				};

				var hexatrigesimalToChar = [ '0', '1', '2', '3', '4', '5', '6', '7',
						'8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
						'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v',
						'w', 'x', 'y', 'z' ];

				RSAUtils.biToString = function(x, radix) { // 2 <= radix <= 36
					var b = new BigInt();
					b.digits[0] = radix;
					var qr = RSAUtils.biDivideModulo(x, b);
					var result = hexatrigesimalToChar[qr[1].digits[0]];
					while (RSAUtils.biCompare(qr[0], bigZero) == 1) {
						qr = RSAUtils.biDivideModulo(qr[0], b);
						digit = qr[1].digits[0];
						result += hexatrigesimalToChar[qr[1].digits[0]];
					}
					return (x.isNeg ? "-" : "") + RSAUtils.reverseStr(result);
				};

				RSAUtils.biToDecimal = function(x) {
					var b = new BigInt();
					b.digits[0] = 10;
					var qr = RSAUtils.biDivideModulo(x, b);
					var result = String(qr[1].digits[0]);
					while (RSAUtils.biCompare(qr[0], bigZero) == 1) {
						qr = RSAUtils.biDivideModulo(qr[0], b);
						result += String(qr[1].digits[0]);
					}
					return (x.isNeg ? "-" : "") + RSAUtils.reverseStr(result);
				};

				var hexToChar = [ '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
						'a', 'b', 'c', 'd', 'e', 'f' ];

				RSAUtils.digitToHex = function(n) {
					var mask = 0xf;
					var result = "";
					for (var i = 0; i < 4; ++i) {
						result += hexToChar[n & mask];
						n >>>= 4;
					}
					return RSAUtils.reverseStr(result);
				};

				RSAUtils.biToHex = function(x) {
					var result = "";
					var n = RSAUtils.biHighIndex(x);
					for ( var i = RSAUtils.biHighIndex(x); i > -1; --i) {
						result += RSAUtils.digitToHex(x.digits[i]);
					}
					return result;
				};

				RSAUtils.charToHex = function(c) {
					var ZERO = 48;
					var NINE = ZERO + 9;
					var littleA = 97;
					var littleZ = littleA + 25;
					var bigA = 65;
					var bigZ = 65 + 25;
					var result;

					if (c >= ZERO && c <= NINE) {
						result = c - ZERO;
					} else if (c >= bigA && c <= bigZ) {
						result = 10 + c - bigA;
					} else if (c >= littleA && c <= littleZ) {
						result = 10 + c - littleA;
					} else {
						result = 0;
					}
					return result;
				};

				RSAUtils.hexToDigit = function(s) {
					var result = 0;
					var sl = Math.min(s.length, 4);
					for ( var i = 0; i < sl; ++i) {
						result <<= 4;
						result |= RSAUtils.charToHex(s.charCodeAt(i));
					}
					return result;
				};

				RSAUtils.biFromHex = function(s) {
					var result = new BigInt();
					var sl = s.length;
					for ( var i = sl, j = 0; i > 0; i -= 4, ++j) {
						result.digits[j] = RSAUtils.hexToDigit(s.substr(Math.max(i - 4,
								0), Math.min(i, 4)));
					}
					return result;
				};

				RSAUtils.biFromString = function(s, radix) {
					var isNeg = s.charAt(0) == '-';
					var istop = isNeg ? 1 : 0;
					var result = new BigInt();
					var place = new BigInt();
					place.digits[0] = 1; // radix^0
					for ( var i = s.length - 1; i >= istop; i--) {
						var c = s.charCodeAt(i);
						var digit = RSAUtils.charToHex(c);
						var biDigit = RSAUtils.biMultiplyDigit(place, digit);
						result = RSAUtils.biAdd(result, biDigit);
						place = RSAUtils.biMultiplyDigit(place, radix);
					}
					result.isNeg = isNeg;
					return result;
				};

				RSAUtils.biDump = function(b) {
					return (b.isNeg ? "-" : "") + b.digits.join(" ");
				};

				RSAUtils.biAdd = function(x, y) {
					var result;

					if (x.isNeg != y.isNeg) {
						y.isNeg = !y.isNeg;
						result = RSAUtils.biSubtract(x, y);
						y.isNeg = !y.isNeg;
					} else {
						result = new BigInt();
						var c = 0;
						var n;
						for ( var i = 0; i < x.digits.length; ++i) {
							n = x.digits[i] + y.digits[i] + c;
							result.digits[i] = n % biRadix;
							c = Number(n >= biRadix);
						}
						result.isNeg = x.isNeg;
					}
					return result;
				};

				RSAUtils.biSubtract = function(x, y) {
					var result;
					if (x.isNeg != y.isNeg) {
						y.isNeg = !y.isNeg;
						result = RSAUtils.biAdd(x, y);
						y.isNeg = !y.isNeg;
					} else {
						result = new BigInt();
						var n, c;
						c = 0;
						for ( var i = 0; i < x.digits.length; ++i) {
							n = x.digits[i] - y.digits[i] + c;
							result.digits[i] = n % biRadix;
							// Stupid non-conforming modulus operation.
							if (result.digits[i] < 0)
								result.digits[i] += biRadix;
							c = 0 - Number(n < 0);
						}
						// Fix up the negative sign, if any.
						if (c == -1) {
							c = 0;
							for ( var i = 0; i < x.digits.length; ++i) {
								n = 0 - result.digits[i] + c;
								result.digits[i] = n % biRadix;
								// Stupid non-conforming modulus operation.
								if (result.digits[i] < 0)
									result.digits[i] += biRadix;
								c = 0 - Number(n < 0);
							}
							// Result is opposite sign of arguments.
							result.isNeg = !x.isNeg;
						} else {
							// Result is same sign.
							result.isNeg = x.isNeg;
						}
					}
					return result;
				};

				RSAUtils.biHighIndex = function(x) {
					var result = x.digits.length - 1;
					while (result > 0 && x.digits[result] == 0)
						--result;
					return result;
				};

				RSAUtils.biNumBits = function(x) {
					var n = RSAUtils.biHighIndex(x);
					var d = x.digits[n];
					var m = (n + 1) * bitsPerDigit;
					var result;
					for (result = m; result > m - bitsPerDigit; --result) {
						if ((d & 0x8000) != 0)
							break;
						d <<= 1;
					}
					return result;
				};

				RSAUtils.biMultiply = function(x, y) {
					var result = new BigInt();
					var c;
					var n = RSAUtils.biHighIndex(x);
					var t = RSAUtils.biHighIndex(y);
					var u, uv, k;

					for ( var i = 0; i <= t; ++i) {
						c = 0;
						k = i;
						for (var j = 0; j <= n; ++j, ++k) {
							uv = result.digits[k] + x.digits[j] * y.digits[i] + c;
							result.digits[k] = uv & maxDigitVal;
							c = uv >>> biRadixBits;
							//c = Math.floor(uv / biRadix);
						}
						result.digits[i + n + 1] = c;
					}
					// Someone give me a logical xor, please.
					result.isNeg = x.isNeg != y.isNeg;
					return result;
				};

				RSAUtils.biMultiplyDigit = function(x, y) {
					var n, c, uv;

					var result = new BigInt();
					n = RSAUtils.biHighIndex(x);
					c = 0;
					for ( var j = 0; j <= n; ++j) {
						uv = result.digits[j] + x.digits[j] * y + c;
						result.digits[j] = uv & maxDigitVal;
						c = uv >>> biRadixBits;
						//c = Math.floor(uv / biRadix);
					}
					result.digits[1 + n] = c;
					return result;
				};

				RSAUtils.arrayCopy = function(src, srcStart, dest, destStart, n) {
					var m = Math.min(srcStart + n, src.length);
					for ( var i = srcStart, j = destStart; i < m; ++i, ++j) {
						dest[j] = src[i];
					}
				};

				var highBitMasks = [ 0x0000, 0x8000, 0xC000, 0xE000, 0xF000, 0xF800,
						0xFC00, 0xFE00, 0xFF00, 0xFF80, 0xFFC0, 0xFFE0, 0xFFF0, 0xFFF8,
						0xFFFC, 0xFFFE, 0xFFFF ];

				RSAUtils.biShiftLeft = function(x, n) {
					var digitCount = Math.floor(n / bitsPerDigit);
					var result = new BigInt();
					RSAUtils.arrayCopy(x.digits, 0, result.digits, digitCount,
							result.digits.length - digitCount);
					var bits = n % bitsPerDigit;
					var rightBits = bitsPerDigit - bits;
					for ( var i = result.digits.length - 1, i1 = i - 1; i > 0; --i, --i1) {
						result.digits[i] = ((result.digits[i] << bits) & maxDigitVal)
								| ((result.digits[i1] & highBitMasks[bits]) >>> (rightBits));
					}
					result.digits[0] = ((result.digits[i] << bits) & maxDigitVal);
					result.isNeg = x.isNeg;
					return result;
				};

				var lowBitMasks = [ 0x0000, 0x0001, 0x0003, 0x0007, 0x000F, 0x001F,
						0x003F, 0x007F, 0x00FF, 0x01FF, 0x03FF, 0x07FF, 0x0FFF, 0x1FFF,
						0x3FFF, 0x7FFF, 0xFFFF ];

				RSAUtils.biShiftRight = function(x, n) {
					var digitCount = Math.floor(n / bitsPerDigit);
					var result = new BigInt();
					RSAUtils.arrayCopy(x.digits, digitCount, result.digits, 0,
							x.digits.length - digitCount);
					var bits = n % bitsPerDigit;
					var leftBits = bitsPerDigit - bits;
					for ( var i = 0, i1 = i + 1; i < result.digits.length - 1; ++i, ++i1) {
						result.digits[i] = (result.digits[i] >>> bits)
								| ((result.digits[i1] & lowBitMasks[bits]) << leftBits);
					}
					result.digits[result.digits.length - 1] >>>= bits;
					result.isNeg = x.isNeg;
					return result;
				};

				RSAUtils.biMultiplyByRadixPower = function(x, n) {
					var result = new BigInt();
					RSAUtils.arrayCopy(x.digits, 0, result.digits, n,
							result.digits.length - n);
					return result;
				};

				RSAUtils.biDivideByRadixPower = function(x, n) {
					var result = new BigInt();
					RSAUtils.arrayCopy(x.digits, n, result.digits, 0,
							result.digits.length - n);
					return result;
				};

				RSAUtils.biModuloByRadixPower = function(x, n) {
					var result = new BigInt();
					RSAUtils.arrayCopy(x.digits, 0, result.digits, 0, n);
					return result;
				};

				RSAUtils.biCompare = function(x, y) {
					if (x.isNeg != y.isNeg) {
						return 1 - 2 * Number(x.isNeg);
					}
					for ( var i = x.digits.length - 1; i >= 0; --i) {
						if (x.digits[i] != y.digits[i]) {
							if (x.isNeg) {
								return 1 - 2 * Number(x.digits[i] > y.digits[i]);
							} else {
								return 1 - 2 * Number(x.digits[i] < y.digits[i]);
							}
						}
					}
					return 0;
				};

				RSAUtils.biDivideModulo = function(x, y) {
					var nb = RSAUtils.biNumBits(x);
					var tb = RSAUtils.biNumBits(y);
					var origYIsNeg = y.isNeg;
					var q, r;
					if (nb < tb) {
						// |x| < |y|
						if (x.isNeg) {
							q = RSAUtils.biCopy(bigOne);
							q.isNeg = !y.isNeg;
							x.isNeg = false;
							y.isNeg = false;
							r = biSubtract(y, x);
							// Restore signs, 'cause they're references.
							x.isNeg = true;
							y.isNeg = origYIsNeg;
						} else {
							q = new BigInt();
							r = RSAUtils.biCopy(x);
						}
						return [ q, r ];
					}

					q = new BigInt();
					r = x;

					// Normalize Y.
					var t = Math.ceil(tb / bitsPerDigit) - 1;
					var lambda = 0;
					while (y.digits[t] < biHalfRadix) {
						y = RSAUtils.biShiftLeft(y, 1);
						++lambda;
						++tb;
						t = Math.ceil(tb / bitsPerDigit) - 1;
					}
					// Shift r over to keep the quotient constant. We'll shift the
					// remainder back at the end.
					r = RSAUtils.biShiftLeft(r, lambda);
					nb += lambda; // Update the bit count for x.
					var n = Math.ceil(nb / bitsPerDigit) - 1;

					var b = RSAUtils.biMultiplyByRadixPower(y, n - t);
					while (RSAUtils.biCompare(r, b) != -1) {
						++q.digits[n - t];
						r = RSAUtils.biSubtract(r, b);
					}
					for ( var i = n; i > t; --i) {
						var ri = (i >= r.digits.length) ? 0 : r.digits[i];
						var ri1 = (i - 1 >= r.digits.length) ? 0 : r.digits[i - 1];
						var ri2 = (i - 2 >= r.digits.length) ? 0 : r.digits[i - 2];
						var yt = (t >= y.digits.length) ? 0 : y.digits[t];
						var yt1 = (t - 1 >= y.digits.length) ? 0 : y.digits[t - 1];
						if (ri == yt) {
							q.digits[i - t - 1] = maxDigitVal;
						} else {
							q.digits[i - t - 1] = Math.floor((ri * biRadix + ri1) / yt);
						}

						var c1 = q.digits[i - t - 1] * ((yt * biRadix) + yt1);
						var c2 = (ri * biRadixSquared) + ((ri1 * biRadix) + ri2);
						while (c1 > c2) {
							--q.digits[i - t - 1];
							c1 = q.digits[i - t - 1] * ((yt * biRadix) | yt1);
							c2 = (ri * biRadix * biRadix) + ((ri1 * biRadix) + ri2);
						}

						b = RSAUtils.biMultiplyByRadixPower(y, i - t - 1);
						r = RSAUtils.biSubtract(r, RSAUtils.biMultiplyDigit(b,
								q.digits[i - t - 1]));
						if (r.isNeg) {
							r = RSAUtils.biAdd(r, b);
							--q.digits[i - t - 1];
						}
					}
					r = RSAUtils.biShiftRight(r, lambda);
					// Fiddle with the signs and stuff to make sure that 0 <= r < y.
					q.isNeg = x.isNeg != origYIsNeg;
					if (x.isNeg) {
						if (origYIsNeg) {
							q = RSAUtils.biAdd(q, bigOne);
						} else {
							q = RSAUtils.biSubtract(q, bigOne);
						}
						y = RSAUtils.biShiftRight(y, lambda);
						r = RSAUtils.biSubtract(y, r);
					}
					// Check for the unbelievably stupid degenerate case of r == -0.
					if (r.digits[0] == 0 && RSAUtils.biHighIndex(r) == 0)
						r.isNeg = false;

					return [ q, r ];
				};

				RSAUtils.biDivide = function(x, y) {
					return RSAUtils.biDivideModulo(x, y)[0];
				};

				RSAUtils.biModulo = function(x, y) {
					return RSAUtils.biDivideModulo(x, y)[1];
				};

				RSAUtils.biMultiplyMod = function(x, y, m) {
					return RSAUtils.biModulo(RSAUtils.biMultiply(x, y), m);
				};

				RSAUtils.biPow = function(x, y) {
					var result = bigOne;
					var a = x;
					while (true) {
						if ((y & 1) != 0)
							result = RSAUtils.biMultiply(result, a);
						y >>= 1;
						if (y == 0)
							break;
						a = RSAUtils.biMultiply(a, a);
					}
					return result;
				};

				RSAUtils.biPowMod = function(x, y, m) {
					var result = bigOne;
					var a = x;
					var k = y;
					while (true) {
						if ((k.digits[0] & 1) != 0)
							result = RSAUtils.biMultiplyMod(result, a, m);
						k = RSAUtils.biShiftRight(k, 1);
						if (k.digits[0] == 0 && RSAUtils.biHighIndex(k) == 0)
							break;
						a = RSAUtils.biMultiplyMod(a, a, m);
					}
					return result;
				};

				$w.BarrettMu = function(m) {
					this.modulus = RSAUtils.biCopy(m);
					this.k = RSAUtils.biHighIndex(this.modulus) + 1;
					var b2k = new BigInt();
					b2k.digits[2 * this.k] = 1; // b2k = b^(2k)
					this.mu = RSAUtils.biDivide(b2k, this.modulus);
					this.bkplus1 = new BigInt();
					this.bkplus1.digits[this.k + 1] = 1; // bkplus1 = b^(k+1)
					this.modulo = BarrettMu_modulo;
					this.multiplyMod = BarrettMu_multiplyMod;
					this.powMod = BarrettMu_powMod;
				};

				function BarrettMu_modulo(x) {
					var $dmath = RSAUtils;
					var q1 = $dmath.biDivideByRadixPower(x, this.k - 1);
					var q2 = $dmath.biMultiply(q1, this.mu);
					var q3 = $dmath.biDivideByRadixPower(q2, this.k + 1);
					var r1 = $dmath.biModuloByRadixPower(x, this.k + 1);
					var r2term = $dmath.biMultiply(q3, this.modulus);
					var r2 = $dmath.biModuloByRadixPower(r2term, this.k + 1);
					var r = $dmath.biSubtract(r1, r2);
					if (r.isNeg) {
						r = $dmath.biAdd(r, this.bkplus1);
					}
					var rgtem = $dmath.biCompare(r, this.modulus) >= 0;
					while (rgtem) {
						r = $dmath.biSubtract(r, this.modulus);
						rgtem = $dmath.biCompare(r, this.modulus) >= 0;
					}
					return r;
				}

				function BarrettMu_multiplyMod(x, y) {
					/*
					x = this.modulo(x);
					y = this.modulo(y);
					 */
					var xy = RSAUtils.biMultiply(x, y);
					return this.modulo(xy);
				}

				function BarrettMu_powMod(x, y) {
					var result = new BigInt();
					result.digits[0] = 1;
					var a = x;
					var k = y;
					while (true) {
						if ((k.digits[0] & 1) != 0)
							result = this.multiplyMod(result, a);
						k = RSAUtils.biShiftRight(k, 1);
						if (k.digits[0] == 0 && RSAUtils.biHighIndex(k) == 0)
							break;
						a = this.multiplyMod(a, a);
					}
					return result;
				}

				var RSAKeyPair = function(encryptionExponent, decryptionExponent,
						modulus) {
					var $dmath = RSAUtils;
					this.e = $dmath.biFromHex(encryptionExponent);
					this.d = $dmath.biFromHex(decryptionExponent);
					this.m = $dmath.biFromHex(modulus);
					
					// We can do two bytes per digit, so
					// chunkSize = 2 * (number of digits in modulus - 1).
					// Since biHighIndex returns the high index, not the number of digits, 1 has
					// already been subtracted.
					this.chunkSize = 2 * $dmath.biHighIndex(this.m);
					this.radix = 16;
					this.barrett = new $w.BarrettMu(this.m);
				};

				RSAUtils.getKeyPair = function(encryptionExponent, decryptionExponent,
						modulus) {
					return new RSAKeyPair(encryptionExponent, decryptionExponent,
							modulus);
				};

				if (typeof $w.twoDigit === 'undefined') {
					$w.twoDigit = function(n) {
						return (n < 10 ? "0" : "") + String(n);
					};
				}

				// Altered by Rob Saunders (rob@robsaunders.net). New routine pads the
				// string after it has been converted to an array. This fixes an
				// incompatibility with Flash MX's ActionScript.
				RSAUtils.encryptedString = function(key, s) {
					var a = [];
					var sl = s.length;
					var i = 0;
					while (i < sl) {
						a[i] = s.charCodeAt(i);
						i++;
					}

					while (a.length % key.chunkSize != 0) {
						a[i++] = 0;
					}

					var al = a.length;
					var result = "";
					var j, k, block;
					for (i = 0; i < al; i += key.chunkSize) {
						block = new BigInt();
						j = 0;
						for (k = i; k < i + key.chunkSize; ++j) {
							block.digits[j] = a[k++];
							block.digits[j] += a[k++] << 8;
						}
						var crypt = key.barrett.powMod(block, key.e);
						var text = key.radix == 16 ? RSAUtils.biToHex(crypt) : RSAUtils
								.biToString(crypt, key.radix);
						result += text + " ";
					}
					return result.substring(0, result.length - 1); // Remove last space.
				};

				RSAUtils.decryptedString = function(key, s) {
					var blocks = s.split(" ");
					var result = "";
					var i, j, block;
					for (i = 0; i < blocks.length; ++i) {
						var bi;
						if (key.radix == 16) {
							bi = RSAUtils.biFromHex(blocks[i]);
						} else {
							bi = RSAUtils.biFromString(blocks[i], key.radix);
						}
						block = key.barrett.powMod(bi, key.d);
						for (j = 0; j <= RSAUtils.biHighIndex(block); ++j) {
							result += String.fromCharCode(block.digits[j] & 255,
									block.digits[j] >> 8);
						}
					}
					// Remove trailing null, if any.
					if (result.charCodeAt(result.length - 1) == 0) {
						result = result.substring(0, result.length - 1);
					}
					return result;
				};

				RSAUtils.setMaxDigits(130);

			})(window);


			function createSecurityData(m, e, random, md5, pwd) {
			    var key = window.RSAUtils.getKeyPair(e, '', m);
			    var envilope = md5 + '!,!' + random + '!,!' + pwd;
			    return window.RSAUtils.encryptedString(key, envilope)
			}


			function createSecurityData3(m, e, envilope) {
			    var key = window.RSAUtils.getKeyPair(e, '', m);
			    return window.RSAUtils.encryptedString(key, envilope)
			}

			1+1`

const smJs = `
/********************************************/
/*                                          */
/*        Copyright (c) 2020 Aostar         */
/*        @author bgu                       */
/*        @version 2.9  		            */
/*                                          */
/********************************************/

var window = {};
var dB = {
	log: print,
	error: print
};

function K(src,ei,ed,eo,dm){var gX=dm;if(ei+dm>src.length&&eo+dm<=ed.length){gX=src.length-ei;}else if(eo+dm>ed.length&&ei+dm<=src.length){gX=ed.length-eo;}else if(ei+dm<=src.length&&eo+dm<=ed.length){gX=dm;}else if(ed.length<src.length){gX=ed.length-eo;}else{gX=src.length-eo;}for(var i=0;i<gX;i++){ed[i+eo]=src[i+ei];}};function co(jr,ei,ij,eo,dm){if(jr.length-ei<dm){return-1;}if(ij.length-eo<dm){return-1;}for(var i=0;i<dm;i++){if(jr[ei++]!=ij[eo++]){return-1;}}return 0;};function F(hA){return new Array((hA>>24)&0x000000FF,(hA>>16)&0x000000FF,(hA>>8)&0x000000FF,(hA)&0x000000FF);};function ap(gL){var b=new Array(gL.length*4);for(var i=0;i<gL.length;i++){K(F(gL[i]),0,b,i*4,4);}return b;};function bB(b,pos){if(pos+3<b.length){return((b[pos])<<24)|((b[pos+1])<<16)|((b[pos+2])<<8)|((b[pos+3]));}else if(pos+2<b.length){return((b[pos+1])<<16)|((b[pos+2])<<8)|((b[pos+3]));}else if(pos+1<b.length){return((b[pos])<<8)|((b[pos+1]));}else{return((b[pos]));}};function aI(b){var kF=Math.ceil(b.length/4);var out=new Array(kF);for(var i=0;i<b.length;i++){b[i]=b[i]&0xFF;}for(var i=0;i<out.length;i++){out[i]=bB(b,i*4);}return out;};var CommonUtils={bytesToHex:function(ev){var dm=ev.length;var dg=new Array(dm*2);var ju=new Array('0','1','2','3','4','5','6','7','8','9','A','B','C','D','E','F');for(var i=0,j=0;i<dm;i++,j++){dg[j]=ju[(ev[i]&0xFF)>>4];dg[++j]=ju[(ev[i]&0x0F)];}return dg.join('');},hexToBytes:function(hC){if(hC==null||hC==''){return null;}if(hC.length%2!=0){return null;}var il=hC.length/2;var dg=this.jY(hC);var eG=new Array(il);for(var i=0;i<il;i++){if(dg[2*i]>=0x30&&dg[2*i]<=0x39){eG[i]=((dg[2*i]-0x30)<<4);}else if(dg[2*i]>=0x41&&dg[2*i]<=0x46){eG[i]=((dg[2*i]-0x41+10)<<4);}else if(dg[2*i]>=0x61&&dg[2*i]<=0x66){eG[i]=((dg[2*i]-0x61+10)<<4);}else{return null;}if(dg[2*i+1]>=0x30&&dg[2*i+1]<=0x39){eG[i]=(eG[i]|(dg[2*i+1]-0x30));}else if(dg[2*i+1]>=0x41&&dg[2*i+1]<=0x46){eG[i]=(eG[i]|(dg[2*i+1]-0x41+10));}else if(dg[2*i+1]>=0x61&&dg[2*i+1]<=0x66){eG[i]=(eG[i]|(dg[2*i+1]-0x61+10));}else{return null;}}return eG;},utf8StrToHex:function(fT){var iG=encodeURIComponent(fT);var gA=unescape(iG);var iY=gA.length;var eN=[];for(var i=0;i<iY;i++){eN[i]=(gA.charCodeAt(i).toString(16)).toUpperCase();}return eN.join('');},utf8StrToBytes:function(fT){var iG=encodeURIComponent(fT);var gA=unescape(iG);var iY=gA.length;var eN=[];for(var i=0;i<iY;i++){eN[i]=gA.charCodeAt(i);}return eN;},hexToUtf8Str:function(fT){var gn=CommonUtils.hexToBytes(fT);var gE=[];for(var i=0;i<gn.length;i++){gE.push(String.fromCharCode(gn[i]));}return decodeURIComponent(escape(gE.join('')));},bytesToUtf8Str:function(kt){var gn=kt;var gE=[];for(var i=0;i<gn.length;i++){gE.push(String.fromCharCode(gn[i]));}return decodeURIComponent(escape(gE.join('')));},jY:function(hN){var iW=new Array(hN.length);for(var i=0;i<hN.length;i++){iW[i]=hN.charCodeAt(i);}return iW;},isHexString:function(fL){var pattern=new RegExp(/^[A-Za-z0-9]+$/);return fL!=null&&pattern.test(fL);},randomWord:function(dm){var fL="";var hF=['0','1','2','3','4','5','6','7','8','9','a','b','c','d','e','f','g','h','i','j','k','l','m','n','o','p','q','r','s','t','u','v','w','x','y','z','A','B','C','D','E','F','G','H','I','J','K','L','M','N','O','P','Q','R','S','T','U','V','W','X','Y','Z','~','!','@','#','$','%','^','&','*','(',')','_'];for(var i=0;i<dm;i++){pos=Math.round(Math.random()*(hF.length-1));fL+=hF[pos];}return fL;},randomHex:function(dm){var fL="";var hF=['0','1','2','3','4','5','6','7','8','9','a','b','c','d','e','f'];for(var i=0;i<dm;i++){pos=Math.round(Math.random()*(hF.length-1));fL+=hF[pos];}return fL;},equals:function(eH,fv){for(fK in eH){if(eH.hasOwnProperty(fK)!=fv.hasOwnProperty(fK)){return false;}else if(typeof eH[fK]!=typeof fv[fK]){return false;}}for(fK in fv){if(eH.hasOwnProperty(fK)!=fv.hasOwnProperty(fK)){return false;}else if(typeof eH[fK]!=typeof fv[fK]){return false;}if(!eH.hasOwnProperty(fK))continue;if(eH[fK]instanceof Array&&fv[fK]instanceof Array){if(!eH[fK].equals(fv[fK]))return false;}else if(eH[fK]instanceof Object&&fv[fK]instanceof Object){if(!eH[fK].equals(fv[fK]))return false;}else if(eH[fK]!=fv[fK]){return false;}}return true;},v_ltd:function(){var word=new jW.mG.mP.gD([-442786843,-1214389064,-1377332043,-443502618,-1735465320,-1612398687,-427708442,-1971263844,-1343841143,-375811867,-2052266609,-1207959552]);var text=word.toString(jW.ey.nj);return text;}};var ep;var kK=0xdeadbeefcafe;var iL=((kK&0xffffff)==0xefcafe);function L(a,b,c){if(a!=null)if("number"==typeof a)this.iM(a,b,c);else if(b==null&&"string"!=typeof a)this.hK(a,256);else this.hK(a,b);};function C(){return new L(null);};function al(i,x,w,j,c,n){while(--n>=0){var v=x*this[i++]+w[j]+c;c=Math.floor(v/0x4000000);w[j++]=v&0x3ffffff;}return c;};function bD(i,x,w,j,c,n){var gw=x&0x7fff,gl=x>>15;while(--n>=0){var l=this[i]&0x7fff;var h=this[i++]>>15;var m=gl*l+h*gw;l=gw*l+((m&0x7fff)<<15)+w[j]+(c&0x3fffffff);c=(l>>>30)+(m>>>15)+gl*h+(c>>>30);w[j++]=l&0x3fffffff;}return c;};function dK(i,x,w,j,c,n){var gw=x&0x3fff,gl=x>>14;while(--n>=0){var l=this[i]&0x3fff;var h=this[i++]>>14;var m=gl*l+h*gw;l=gw*l+((m&0x3fff)<<14)+w[j]+c;c=(l>>28)+(m>>14)+gl*h;w[j++]=l&0xfffffff;}return c;};if(iL&&(navigator.appName=="Microsoft Internet Explorer")){L.prototype.eZ=bD;ep=30;}else if(iL&&(navigator.appName!="Netscape")){L.prototype.eZ=al;ep=26;}else{L.prototype.eZ=dK;ep=28;}L.prototype.ct=ep;L.prototype.fi=((1<<ep)-1);L.prototype.ez=(1<<ep);var im=52;L.prototype.ku=Math.pow(2,im);L.prototype.iR=im-ep;L.prototype.io=2*ep-im;var jV="0123456789abcdefghijklmnopqrstuvwxyz";var gQ=new Array();var gr="0".charCodeAt(0);for(var ej=0;ej<=9;++ej)gQ[gr++]=ej;gr="a".charCodeAt(0);for(var ej=10;ej<36;++ej)gQ[gr++]=ej;gr="A".charCodeAt(0);for(var ej=10;ej<36;++ej)gQ[gr++]=ej;function bk(n){return jV.charAt(n);};function bw(s,i){var c=gQ[s.charCodeAt(i)];return(c==null)? -1:c;};function dQ(r){for(var i=this.t-1;i>=0;--i)r[i]=this[i];r.t=this.t;r.s=this.s;};function cS(x){this.t=1;this.s=(x<0)? -1:0;if(x>0)this[0]=x;else if(x< -1)this[0]=x+this.ez;else this.t=0;};function H(i){var r=C();r.eT(i);return r;};function dc(s,b){var k;if(b==16)k=4;else if(b==8)k=3;else if(b==256)k=8;else if(b==2)k=1;else if(b==32)k=5;else if(b==4)k=2;else{this.jP(s,b);return;}this.t=0;this.s=0;var i=s.length,gO=false,fr=0;while(--i>=0){var x=(k==8)?s[i]&0xff:bw(s,i);if(x<0){if(s.charAt(i)=="-")gO=true;continue;}gO=false;if(fr==0)this[this.t++]=x;else if(fr+k>this.ct){this[this.t-1]|=(x&((1<<(this.ct-fr))-1))<<fr;this[this.t++]=(x>>(this.ct-fr));}else this[this.t-1]|=x<<fr;fr+=k;if(fr>=this.ct)fr-=this.ct;}if(k==8&&(s[0]&0x80)!=0){this.s= -1;if(fr>0)this[this.t-1]|=((1<<(this.ct-fr))-1)<<fr;}this.ff();if(gO)L.ZERO.cp(this,this);};function cf(){var c=this.s&this.fi;while(this.t>0&&this[this.t-1]==c)--this.t;};function aH(b){if(this.s<0)return "-"+this.el().toString(b);var k;if(b==16)k=4;else if(b==8)k=3;else if(b==2)k=1;else if(b==32)k=5;else if(b==4)k=2;else return this.eD(b);var hx=(1<<k)-1,d,m=false,r="",i=this.t;var p=this.ct-(i*this.ct)%k;if(i-- >0){if(p<this.ct&&(d=this[i]>>p)>0){m=true;r=bk(d);}while(i>=0){if(p<k){d=(this[i]&((1<<p)-1))<<(k-p);d|=this[--i]>>(p+=this.ct-k);}else{d=(this[i]>>(p-=k))&hx;if(p<=0){p+=this.ct;--i;}}if(d>0)m=true;if(m)r+=bk(d);}}return m?r:"0";};function aL(){var r=C();L.ZERO.cp(this,r);return r;};function cn(){return(this.s<0)?this.el():this;};function aD(a){var r=this.s-a.s;if(r!=0)return r;var i=this.t;r=i-a.t;if(r!=0)return(this.s<0)? -r:r;while(--i>=0)if((r=this[i]-a[i])!=0)return r;return 0;};function bN(x){var r=1,t;if((t=x>>>16)!=0){x=t;r+=16;}if((t=x>>8)!=0){x=t;r+=8;}if((t=x>>4)!=0){x=t;r+=4;}if((t=x>>2)!=0){x=t;r+=2;}if((t=x>>1)!=0){x=t;r+=1;}return r;};function dI(){if(this.t<=0)return 0;return this.ct*(this.t-1)+bN(this[this.t-1]^(this.s&this.fi));};function da(n,r){var i;for(i=this.t-1;i>=0;--i)r[i+n]=this[i];for(i=n-1;i>=0;--i)r[i]=0;r.t=this.t+n;r.s=this.s;};function cP(n,r){for(var i=n;i<this.t;++i)r[i-n]=this[i];r.t=Math.max(this.t-n,0);r.s=this.s;};function cq(n,r){var fD=n%this.ct;var gC=this.ct-fD;var hD=(1<<gC)-1;var eL=Math.floor(n/this.ct),c=(this.s<<fD)&this.fi,i;for(i=this.t-1;i>=0;--i){r[i+eL+1]=(this[i]>>gC)|c;c=(this[i]&hD)<<fD;}for(i=eL-1;i>=0;--i)r[i]=0;r[eL]=c;r.t=this.t+eL+1;r.s=this.s;r.ff();};function dM(n,r){r.s=this.s;var eL=Math.floor(n/this.ct);if(eL>=this.t){r.t=0;return;}var fD=n%this.ct;var gC=this.ct-fD;var hD=(1<<fD)-1;r[0]=this[eL]>>fD;for(var i=eL+1;i<this.t;++i){r[i-eL-1]|=(this[i]&hD)<<gC;r[i-eL]=this[i]>>fD;}if(fD>0)r[this.t-eL-1]|=(this.s&hD)<<gC;r.t=this.t-eL;r.ff();};function de(a,r){var i=0,c=0,m=Math.min(a.t,this.t);while(i<m){c+=this[i]-a[i];r[i++]=c&this.fi;c>>=this.ct;}if(a.t<this.t){c-=a.s;while(i<this.t){c+=this[i];r[i++]=c&this.fi;c>>=this.ct;}c+=this.s;}else{c+=this.s;while(i<a.t){c-=a[i];r[i++]=c&this.fi;c>>=this.ct;}c-=a.s;}r.s=(c<0)? -1:0;if(c< -1)r[i++]=this.ez+c;else if(c>0)r[i++]=c;r.t=i;r.ff();};function bX(a,r){var x=this.abs(),y=a.abs();var i=x.t;r.t=i+y.t;while(--i>=0)r[i]=0;for(i=0;i<y.t;++i)r[i+x.t]=x.eZ(0,y[i],r,i,0,x.t);r.s=0;r.ff();if(this.s!=a.s)L.ZERO.cp(r,r);};function cr(r){var x=this.abs();var i=r.t=2*x.t;while(--i>=0)r[i]=0;for(i=0;i<x.t-1;++i){var c=x.eZ(i,x[i],r,2*i,0,1);if((r[i+x.t]+=x.eZ(i+1,2*x[i],r,2*i+1,c,x.t-i-1))>=x.ez){r[i+x.t]-=x.ez;r[i+x.t+1]=1;}}if(r.t>0)r[r.t-1]+=x.eZ(i,x[i],r,2*i,0,1);r.s=0;r.ff();};function dJ(m,q,r){var hz=m.abs();if(hz.t<=0)return;var pt=this.abs();if(pt.t<hz.t){if(q!=null)q.eT(0);if(r!=null)this.eh(r);return;}if(r==null)r=C();var y=C(),jb=this.s,ms=m.s;var gP=this.ct-bN(hz[hz.t-1]);if(gP>0){hz.gx(gP,y);pt.gx(gP,r);}else{hz.eh(y);pt.eh(r);}var fG=y.t;var gb=y[fG-1];if(gb==0)return;var jA=gb*(1<<this.iR)+((fG>1)?y[fG-2]>>this.io:0);var kA=this.ku/jA,ko=(1<<this.iR)/jA,e=1<<this.io;var i=r.t,j=i-fG,t=(q==null)?C():q;y.gS(j,t);if(r.dw(t)>=0){r[r.t++]=1;r.cp(t,r);}L.ONE.gS(fG,t);t.cp(y,y);while(y.t<fG)y[y.t++]=0;while(--j>=0){var gc=(r[--i]==gb)?this.fi:Math.floor(r[i]*kA+(r[i-1]+e)*ko);if((r[i]+=y.eZ(0,gc,r,j,0,fG))<gc){y.gS(j,t);r.cp(t,r);while(r[i]< --gc)r.cp(t,r);}}if(q!=null){r.gJ(fG,q);if(jb!=ms)L.ZERO.cp(q,q);}r.t=fG;r.ff();if(gP>0)r.dY(gP,r);if(jb<0)L.ZERO.cp(r,r);};function aZ(a){var r=C();this.abs().fw(a,null,r);if(this.s<0&&r.dw(L.ZERO)>0)a.cp(r,r);return r;};function af(m){this.m=m;};function ak(x){if(x.s<0||x.dw(this.m)>=0)return x.cl(this.m);else return x;};function aj(x){return x;};function aU(x){x.fw(this.m,null,x);};function bM(x,y,r){x.gG(y,r);this.reduce(r);};function bE(x,r){x.hc(r);this.reduce(r);};af.prototype.convert=ak;af.prototype.revert=aj;af.prototype.reduce=aU;af.prototype.fO=bM;af.prototype.eO=bE;function dh(){if(this.t<1)return 0;var x=this[0];if((x&1)==0)return 0;var y=x&3;y=(y*(2-(x&0xf)*y))&0xf;y=(y*(2-(x&0xff)*y))&0xff;y=(y*(2-(((x&0xffff)*y)&0xffff)))&0xffff;y=(y*(2-x*y%this.ez))%this.ez;return(y>0)?this.ez-y: -y;};function O(m){this.m=m;this.ik=m.kC();this.jB=this.ik&0x7fff;this.kl=this.ik>>15;this.lf=(1<<(m.ct-15))-1;this.kV=2*m.t;};function bC(x){var r=C();x.abs().gS(this.m.t,r);r.fw(this.m,null,r);if(x.s<0&&r.dw(L.ZERO)>0)this.m.cp(r,r);return r;};function ax(x){var r=C();x.eh(r);this.reduce(r);return r;};function au(x){while(x.t<=this.kV)x[x.t++]=0;for(var i=0;i<this.m.t;++i){var j=x[i]&0x7fff;var kh=(j*this.jB+(((j*this.kl+(x[i]>>15)*this.jB)&this.lf)<<15))&x.fi;j=i+this.m.t;x[j]+=this.m.eZ(0,kh,x,i,0,this.m.t);while(x[j]>=x.ez){x[j]-=x.ez;x[++j]++;}}x.ff();x.gJ(this.m.t,x);if(x.dw(this.m)>=0)x.cp(this.m,x);};function aC(x,r){x.hc(r);this.reduce(r);};function bi(x,y,r){x.gG(y,r);this.reduce(r);};O.prototype.convert=bC;O.prototype.revert=ax;O.prototype.reduce=au;O.prototype.fO=bi;O.prototype.eO=aC;function cz(){return((this.t>0)?(this[0]&1):this.s)==0;};function df(e,z){if(e>0xffffffff||e<1)return L.ONE;var r=C(),r2=C(),g=z.convert(this),i=bN(e)-1;g.eh(r);while(--i>=0){z.eO(r,r2);if((e&(1<<i))>0)z.fO(r2,g,r);else{var t=r;r=r2;r2=t;}}return z.revert(r);};function bH(e,m){var z;if(e<256||m.ef())z=new af(m);else z=new O(m);return this.exp(e,z);};L.prototype.eh=dQ;L.prototype.eT=cS;L.prototype.hK=dc;L.prototype.ff=cf;L.prototype.gS=da;L.prototype.gJ=cP;L.prototype.gx=cq;L.prototype.dY=dM;L.prototype.cp=de;L.prototype.gG=bX;L.prototype.hc=cr;L.prototype.fw=dJ;L.prototype.kC=dh;L.prototype.ef=cz;L.prototype.exp=df;L.prototype.toString=aH;L.prototype.el=aL;L.prototype.abs=cn;L.prototype.dw=aD;L.prototype.fg=dI;L.prototype.cl=aZ;L.prototype.kr=bH;L.ZERO=H(0);L.ONE=H(1);function an(){var r=C();this.eh(r);return r;};function bK(){if(this.s<0){if(this.t==1)return this[0]-this.ez;else if(this.t==0)return-1;}else if(this.t==1)return this[0];else if(this.t==0)return 0;return((this[1]&((1<<(32-this.ct))-1))<<this.ct)|this[0];};function cI(){return(this.t==0)?this.s:(this[0]<<24)>>24;};function dn(){return(this.t==0)?this.s:(this[0]<<16)>>16;};function ca(r){return Math.floor(Math.LN2*this.ct/Math.log(r));};function aF(){if(this.s<0)return-1;else if(this.t<=0||(this.t==1&&this[0]<=0))return 0;else return 1;};function dk(b){if(b==null)b=10;if(this.eJ()==0||b<2||b>36)return "0";var gf=this.iw(b);var a=Math.pow(b,gf);var d=H(a),y=C(),z=C(),r="";this.fw(d,y,z);while(y.eJ()>0){r=(a+z.hs()).toString(b).substr(1)+r;y.fw(d,y,z);}return z.hs().toString(b)+r;};function bY(s,b){this.eT(0);if(b==null)b=10;var gf=this.iw(b);var d=Math.pow(b,gf),gO=false,j=0,w=0;for(var i=0;i<s.length;++i){var x=bw(s,i);if(x<0){if(s.charAt(i)=="-"&&this.eJ()==0)gO=true;continue;}w=b*w+x;if(++j>=gf){this.ja(d);this.hJ(w,0);j=0;w=0;}}if(j>0){this.ja(Math.pow(b,j));this.hJ(w,0);}if(gO)L.ZERO.cp(this,this);};function cX(a,b,c){if("number"==typeof b){if(a<2)this.eT(1);else{this.iM(a,c);if(!this.en(a-1))this.gj(L.ONE.shiftLeft(a-1),ay,this);if(this.ef())this.hJ(1,0);while(!this.kv(b)){this.hJ(2,0);if(this.fg()>a)this.cp(L.ONE.shiftLeft(a-1),this);}}}else{var x=new Array(),t=a&7;x.length=(a>>3)+1;b.kj(x);if(t>0)x[0]&=((1<<t)-1);else x[0]=0;this.hK(x,256);}};function bv(){var i=this.t,r=new Array();r[0]=this.s;var p=this.ct-(i*this.ct)%8,d,k=0;if(i-- >0){if(p<this.ct&&(d=this[i]>>p)!=(this.s&this.fi)>>p)r[k++]=d|(this.s<<(this.ct-p));while(i>=0){if(p<8){d=(this[i]&((1<<p)-1))<<(8-p);d|=this[--i]>>(p+=this.ct-8);}else{d=(this[i]>>(p-=8))&0xff;if(p<=0){p+=this.ct;--i;}}if((d&0x80)!=0)d|= -256;if(k==0&&(this.s&0x80)!=(d&0x80))++k;if(k>0||d!=this.s)r[k++]=d;}}return r;};function ab(a){return(this.dw(a)==0);};function aw(a){return(this.dw(a)<0)?this:a;};function ao(a){return(this.dw(a)>0)?this:a;};function cC(a,hv,r){var i,f,m=Math.min(a.t,this.t);for(i=0;i<m;++i)r[i]=hv(this[i],a[i]);if(a.t<this.t){f=a.s&this.fi;for(i=m;i<this.t;++i)r[i]=hv(this[i],f);r.t=this.t;}else{f=this.s&this.fi;for(i=m;i<a.t;++i)r[i]=hv(f,a[i]);r.t=a.t;}r.s=hv(this.s,a.s);r.ff();};function aR(x,y){return x&y;};function dN(a){var r=C();this.gj(a,aR,r);return r;};function ay(x,y){return x|y;};function bW(a){var r=C();this.gj(a,ay,r);return r;};function aT(x,y){return x^y;};function aQ(a){var r=C();this.gj(a,aT,r);return r;};function ag(x,y){return x& ~y;};function cg(a){var r=C();this.gj(a,ag,r);return r;};function db(){var r=C();for(var i=0;i<this.t;++i)r[i]=this.fi& ~this[i];r.t=this.t;r.s= ~this.s;return r;};function cA(n){var r=C();if(n<0)this.dY(-n,r);else this.gx(n,r);return r;};function dv(n){var r=C();if(n<0)this.gx(-n,r);else this.dY(n,r);return r;};function aA(x){if(x==0)return-1;var r=0;if((x&0xffff)==0){x>>=16;r+=16;}if((x&0xff)==0){x>>=8;r+=8;}if((x&0xf)==0){x>>=4;r+=4;}if((x&3)==0){x>>=2;r+=2;}if((x&1)==0)++r;return r;};function aq(){for(var i=0;i<this.t;++i)if(this[i]!=0)return i*this.ct+aA(this[i]);if(this.s<0)return this.t*this.ct;return-1;};function ad(x){var r=0;while(x!=0){x&=x-1;++r;}return r;};function cj(){var r=0,x=this.s&this.fi;for(var i=0;i<this.t;++i)r+=ad(this[i]^x);return r;};function aM(n){var j=Math.floor(n/this.ct);if(j>=this.t)return(this.s!=0);return((this[j]&(1<<(n%this.ct)))!=0);};function dE(n,hv){var r=L.ONE.shiftLeft(n);this.gj(r,hv,r);return r;};function cc(n){return this.hY(n,ay);};function cT(n){return this.hY(n,ag);};function bb(n){return this.hY(n,aT);};function ce(a,r){var i=0,c=0,m=Math.min(a.t,this.t);while(i<m){c+=this[i]+a[i];r[i++]=c&this.fi;c>>=this.ct;}if(a.t<this.t){c+=a.s;while(i<this.t){c+=this[i];r[i++]=c&this.fi;c>>=this.ct;}c+=this.s;}else{c+=this.s;while(i<a.t){c+=a[i];r[i++]=c&this.fi;c>>=this.ct;}c+=a.s;}r.s=(c<0)? -1:0;if(c>0)r[i++]=c;else if(c< -1)r[i++]=this.ez+c;r.t=i;r.ff();};function cO(a){var r=C();this.hG(a,r);return r;};function bQ(a){var r=C();this.cp(a,r);return r;};function bF(a){var r=C();this.gG(a,r);return r;};function aa(){var r=C();this.hc(r);return r;};function bd(a){var r=C();this.fw(a,r,null);return r;};function cQ(a){var r=C();this.fw(a,null,r);return r;};function bI(a){var q=C(),r=C();this.fw(a,q,r);return new Array(q,r);};function dp(n){this[this.t]=this.eZ(0,n-1,this,0,0,this.t);++this.t;this.ff();};function cJ(n,w){if(n==0)return;while(this.t<=w)this[this.t++]=0;this[w]+=n;while(this[w]>=this.ez){this[w]-=this.ez;if(++w>=this.t)this[this.t++]=0;++this[w];}};function bj(){};function aY(x){return x;};function aG(x,y,r){x.gG(y,r);};function bA(x,r){x.hc(r);};bj.prototype.convert=aY;bj.prototype.revert=aY;bj.prototype.fO=aG;bj.prototype.eO=bA;function cG(e){return this.exp(e,new bj());};function cW(a,n,r){var i=Math.min(this.t+a.t,n);r.s=0;r.t=i;while(i>0)r[--i]=0;var j;for(j=r.t-this.t;i<j;++i)r[i+this.t]=this.eZ(0,a[i],r,i,0,this.t);for(j=Math.min(a.t,n);i<j;++i)this.eZ(0,a[i],r,i,0,n-i);r.ff();};function ci(a,n,r){--n;var i=r.t=this.t+a.t-n;r.s=0;while(--i>=0)r[i]=0;for(i=Math.max(n-this.t,0);i<a.t;++i)r[this.t+i-n]=this.eZ(n-i,a[i],r,0,0,this.t+i-n);r.ff();r.gJ(1,r);};function bh(m){this.r2=C();this.jx=C();L.ONE.gS(2*m.t,this.r2);this.jQ=this.r2.divide(m);this.m=m;};function bu(x){if(x.s<0||x.t>2*this.m.t)return x.cl(this.m);else if(x.dw(this.m)<0)return x;else{var r=C();x.eh(r);this.reduce(r);return r;}};function aN(x){return x;};function bS(x){x.gJ(this.m.t-1,this.r2);if(x.t>this.m.t+1){x.t=this.m.t+1;x.ff();}this.jQ.ld(this.r2,this.m.t+1,this.jx);this.m.lx(this.jx,this.m.t+1,this.r2);while(x.dw(this.r2)<0)x.hJ(1,this.m.t+1);x.cp(this.r2,x);while(x.dw(this.m)>=0)x.cp(this.m,x);};function dq(x,r){x.hc(r);this.reduce(r);};function bU(x,y,r){x.gG(y,r);this.reduce(r);};bh.prototype.convert=bu;bh.prototype.revert=aN;bh.prototype.reduce=bS;bh.prototype.fO=bU;bh.prototype.eO=dq;function aE(e,m){var i=e.fg(),k,r=H(1),z;if(i<=0)return r;else if(i<18)k=1;else if(i<48)k=3;else if(i<144)k=4;else if(i<768)k=5;else k=6;if(i<8)z=new af(m);else if(m.ef())z=new bh(m);else z=new O(m);var g=new Array(),n=3,hh=k-1,hx=(1<<k)-1;g[1]=z.convert(this);if(k>1){var jg=C();z.eO(g[1],jg);while(n<=hx){g[n]=C();z.fO(jg,g[n-2],g[n]);n+=2;}}var j=e.t-1,w,iZ=true,r2=C(),t;i=bN(e[j])-1;while(j>=0){if(i>=hh)w=(e[j]>>(i-hh))&hx;else{w=(e[j]&((1<<(i+1))-1))<<(hh-i);if(j>0)w|=e[j-1]>>(this.ct+i-hh);}n=k;while((w&1)==0){w>>=1;--n;}if((i-=n)<0){i+=this.ct;--j;}if(iZ){g[w].eh(r);iZ=false;}else{while(n>1){z.eO(r,r2);z.eO(r2,r);n-=2;}if(n>0)z.eO(r,r2);else{t=r;r=r2;r2=t;}z.fO(r2,g[w],r);}while(j>=0&&(e[j]&(1<<i))==0){z.eO(r,r2);t=r;r=r2;r2=t;if(--i<0){i=this.ct-1;--j;}}}return z.revert(r);};function aV(a){var x=(this.s<0)?this.el():this.clone();var y=(a.s<0)?a.el():a.clone();if(x.dw(y)<0){var t=x;x=y;y=t;}var i=x.gg(),g=y.gg();if(g<0)return x;if(i<g)g=i;if(g>0){x.dY(g,x);y.dY(g,y);}while(x.eJ()>0){if((i=x.gg())>0)x.dY(i,x);if((i=y.gg())>0)y.dY(i,y);if(x.dw(y)>=0){x.cp(y,x);x.dY(1,x);}else{y.cp(x,y);y.dY(1,y);}}if(g>0)y.gx(g,y);return y;};function di(n){if(n<=0)return 0;var d=this.ez%n,r=(this.s<0)?n-1:0;if(this.t>0)if(d==0)r=this[0]%n;else for(var i=this.t-1;i>=0;--i)r=(d*r+this[i])%n;return r;};function ah(m){var fY=m.ef();if((this.ef()&&fY)||m.eJ()==0)return L.ZERO;var u=m.clone(),v=this.clone();var a=H(1),b=H(0),c=H(0),d=H(1);while(u.eJ()!=0){while(u.ef()){u.dY(1,u);if(fY){if(!a.ef()|| !b.ef()){a.hG(this,a);b.cp(m,b);}a.dY(1,a);}else if(!b.ef())b.cp(m,b);b.dY(1,b);}while(v.ef()){v.dY(1,v);if(fY){if(!c.ef()|| !d.ef()){c.hG(this,c);d.cp(m,d);}c.dY(1,c);}else if(!d.ef())d.cp(m,d);d.dY(1,d);}if(u.dw(v)>=0){u.cp(v,u);if(fY)a.cp(c,a);b.cp(d,b);}else{v.cp(u,v);if(fY)c.cp(a,c);d.cp(b,d);}}if(v.dw(L.ONE)!=0)return L.ZERO;if(d.dw(m)>=0)return d.cv(m);if(d.eJ()<0)d.hG(m,d);else return d;if(d.eJ()<0)return d.add(m);else return d;};var eb=[2,3,5,7,11,13,17,19,23,29,31,37,41,43,47,53,59,61,67,71,73,79,83,89,97,101,103,107,109,113,127,131,137,139,149,151,157,163,167,173,179,181,191,193,197,199,211,223,227,229,233,239,241,251,257,263,269,271,277,281,283,293,307,311,313,317,331,337,347,349,353,359,367,373,379,383,389,397,401,409,419,421,431,433,439,443,449,457,461,463,467,479,487,491,499,503,509,521,523,541,547,557,563,569,571,577,587,593,599,601,607,613,617,619,631,641,643,647,653,659,661,673,677,683,691,701,709,719,727,733,739,743,751,757,761,769,773,787,797,809,811,821,823,827,829,839,853,857,859,863,877,881,883,887,907,911,919,929,937,941,947,953,967,971,977,983,991,997];var kE=(1<<26)/eb[eb.length-1];function as(t){var i,x=this.abs();if(x.t==1&&x[0]<=eb[eb.length-1]){for(i=0;i<eb.length;++i)if(x[0]==eb[i])return true;return false;}if(x.ef())return false;i=1;while(i<eb.length){var m=eb[i],j=i+1;while(j<eb.length&&m<kE)m*=eb[j++];m=x.kq(m);while(i<j)if(m%eb[i++]==0)return false;}return x.kJ(t);};function dR(t){var gV=this.cv(L.ONE);var k=gV.gg();if(k<=0)return false;var r=gV.ic(k);t=(t+1)>>1;if(t>eb.length)t=eb.length;var a=C();for(var i=0;i<t;++i){a.eT(eb[Math.floor(Math.random()*eb.length)]);var y=a.lb(r,this);if(y.dw(L.ONE)!=0&&y.dw(gV)!=0){var j=1;while(j++ <k&&y.dw(gV)!=0){y=y.kr(2,this);if(y.dw(L.ONE)==0)return false;}if(y.dw(gV)!=0)return false;}}return true;};L.prototype.iw=ca;L.prototype.eD=dk;L.prototype.jP=bY;L.prototype.iM=cX;L.prototype.gj=cC;L.prototype.hY=dE;L.prototype.hG=ce;L.prototype.ja=dp;L.prototype.hJ=cJ;L.prototype.lx=cW;L.prototype.ld=ci;L.prototype.kq=di;L.prototype.kJ=dR;L.prototype.clone=an;L.prototype.hs=bK;L.prototype.lQ=cI;L.prototype.lZ=dn;L.prototype.eJ=aF;L.prototype.gv=bv;L.prototype.equals=ab;L.prototype.min=aw;L.prototype.max=ao;L.prototype.and=dN;L.prototype.jw=bW;L.prototype.xor=aQ;L.prototype.mg=cg;L.prototype.not=db;L.prototype.shiftLeft=cA;L.prototype.ic=dv;L.prototype.gg=aq;L.prototype.nx=cj;L.prototype.en=aM;L.prototype.nu=cc;L.prototype.mR=cT;L.prototype.nc=bb;L.prototype.add=cO;L.prototype.cv=bQ;L.prototype.multiply=bF;L.prototype.divide=bd;L.prototype.mU=cQ;L.prototype.lH=bI;L.prototype.lb=aE;L.prototype.fZ=ah;L.prototype.pow=cG;L.prototype.mN=aV;L.prototype.kv=as;L.prototype.square=aa;function J(q,x){this.x=x;this.q=q;};function bx(other){if(other==this)return true;return(this.q.equals(other.q)&&this.x.equals(other.x));};function av(){return this.x;};function aO(){return new J(this.q,this.x.el().cl(this.q));};function bz(b){return new J(this.q,this.x.add(b.cR()).cl(this.q));};function be(b){return new J(this.q,this.x.cv(b.cR()).cl(this.q));};function bf(b){return new J(this.q,this.x.multiply(b.cR()).cl(this.q));};function aJ(){return new J(this.q,this.x.square().cl(this.q));};function ar(b){return new J(this.q,this.x.multiply(b.cR().fZ(this.q)).cl(this.q));};J.prototype.equals=bx;J.prototype.cR=av;J.prototype.el=aO;J.prototype.add=bz;J.prototype.cv=be;J.prototype.multiply=bf;J.prototype.square=aJ;J.prototype.divide=ar;function G(dP,x,y,z){this.dP=dP;this.x=x;this.y=y;if(z==null){this.z=L.ONE;}else{this.z=z;}this.gH=null;};function az(){if(this.gH==null){this.gH=this.z.fZ(this.dP.q);}return this.dP.fm(this.x.cR().multiply(this.gH).cl(this.dP.q));};function bP(){if(this.gH==null){this.gH=this.z.fZ(this.dP.q);}return this.dP.fm(this.y.cR().multiply(this.gH).cl(this.dP.q));};function bV(other){if(other==this)return true;if(this.eg())return other.eg();if(other.eg())return this.eg();var u,v;u=other.y.cR().multiply(this.z).cv(this.y.cR().multiply(other.z)).cl(this.dP.q);if(!u.equals(L.ZERO))return false;v=other.x.cR().multiply(this.z).cv(this.x.cR().multiply(other.z)).cl(this.dP.q);return v.equals(L.ZERO);};function bp(){if((this.x==null)&&(this.y==null))return true;return this.z.equals(L.ZERO)&& !this.y.cR().equals(L.ZERO);};function aX(){return new G(this.dP,this.x,this.y.el(),this.z);};function aB(b){if(this.eg())return b;if(b.eg())return this;var u=b.y.cR().multiply(this.z).cv(this.y.cR().multiply(b.z)).cl(this.dP.q);var v=b.x.cR().multiply(this.z).cv(this.x.cR().multiply(b.z)).cl(this.dP.q);if(L.ZERO.equals(v)){if(L.ZERO.equals(u)){return this.gI();}return this.dP.gy();}var hm=new L("3");var x1=this.x.cR();var y1=this.y.cR();var iT=v.square();var hq=iT.multiply(v);var iy=x1.multiply(iT);var iq=u.square().multiply(this.z);var eE=iq.cv(iy.shiftLeft(1)).multiply(b.z).cv(hq).multiply(v).cl(this.dP.q);var fE=iy.multiply(hm).multiply(u).cv(y1.multiply(hq)).cv(iq.multiply(u)).multiply(b.z).add(u.multiply(hq)).cl(this.dP.q);var js=hq.multiply(this.z).multiply(b.z).cl(this.dP.q);return new G(this.dP,this.dP.fm(eE),this.dP.fm(fE),js);};function bq(){if(this.eg())return this;if(this.y.cR().eJ()==0)return this.dP.gy();var hm=new L("3");var x1=this.x.cR();var y1=this.y.cR();var gz=y1.multiply(this.z);var jp=gz.multiply(y1).cl(this.dP.q);var a=this.dP.a.cR();var w=x1.square().multiply(hm);if(!L.ZERO.equals(a)){w=w.add(this.z.square().multiply(a));}w=w.cl(this.dP.q);var eE=w.square().cv(x1.shiftLeft(3).multiply(jp)).shiftLeft(1).multiply(gz).cl(this.dP.q);var fE=w.multiply(hm).multiply(x1).cv(jp.shiftLeft(1)).shiftLeft(2).multiply(jp).cv(w.square().multiply(w)).cl(this.dP.q);var js=gz.square().multiply(gz).shiftLeft(3).cl(this.dP.q);return new G(this.dP,this.dP.fm(eE),this.dP.fm(fE),js);};function ai(k){if(this.eg())return this;if(k.eJ()==0)return this.dP.gy();var e=k;var h=e.multiply(new L("3"));var hw=this.el();var R=this;var i;for(i=h.fg()-2;i>0;--i){R=R.gI();var gm=h.en(i);var hV=e.en(i);if(gm!=hV){R=R.add(gm?this:hw);}}return R;};function aS(j,x,k){var i;if(j.fg()>k.fg())i=j.fg()-1;else i=k.fg()-1;var R=this.dP.gy();var both=this.add(x);while(i>=0){R=R.gI();if(j.en(i)){if(k.en(i)){R=R.add(both);}else{R=R.add(this);}}else{if(k.en(i)){R=R.add(x);}}--i;}return R;};G.prototype.fI=az;G.prototype.ec=bP;G.prototype.equals=bV;G.prototype.eg=bp;G.prototype.el=aX;G.prototype.add=aB;G.prototype.gI=bq;G.prototype.multiply=ai;G.prototype.nb=aS;function V(q,a,b){this.q=q;this.a=this.fm(a);this.b=this.fm(b);this.jz=new G(this,null,null);};function bO(){return this.q;};function aW(){return this.a;};function bc(){return this.b;};function bJ(other){if(other==this)return true;return(this.q.equals(other.q)&&this.a.equals(other.a)&&this.b.equals(other.b));};function bn(){return this.jz;};function ae(x){return new J(this.q,x);};function by(s){switch(parseInt(s.substr(0,2),16)){case 0:return this.jz;case 2:case 3:return null;case 4:case 6:case 7:var dm=(s.length-2)/2;var hf=s.substr(2,dm);var ih=s.substr(dm+2,dm);return new G(this,this.fm(new L(hf,16)),this.fm(new L(ih,16)));default:return null;}};function bG(x,y){return new G(this,this.fm(x),this.fm(y));};V.prototype.jh=bO;V.prototype.kH=aW;V.prototype.kB=bc;V.prototype.equals=bJ;V.prototype.gy=bn;V.prototype.fm=ae;V.prototype.kR=by;V.prototype.kU=bG;J.prototype.lO=function(){return Math.floor((this.cR().fg()+7)/8);};G.prototype.mj=function(jG){var hM=function(i,dm){var ev=i.lW();if(dm<ev.length){ev=ev.slice(ev.length-dm);}else while(dm>ev.length){ev.unshift(0);}return ev;};var x=this.fI().cR();var y=this.ec().cR();var ey=hM(x,32);if(jG){if(y.ef()){ey.unshift(0x02);}else{ey.unshift(0x03);}}else{ey.unshift(0x04);ey=ey.concat(hM(y,32));}return ey;};G.jZ=function(dP,ey){var ea=ey.length-1;var iP=ey.slice(1,1+ea/2);var je=ey.slice(1+ea/2,1+ea);iP.unshift(0);je.unshift(0);var x=new L(iP);var y=new L(je);return new G(dP,dP.fm(x),dP.fm(y));};G.hQ=function(dP,hU){var ea=hU.length-2;var hf=hU.substr(2,ea/2);var ih=hU.substr(2+ea/2,ea/2);var x=new L(hf,16);var y=new L(ih,16);return new G(dP,dP.fm(x),dP.fm(y));};G.prototype.jK=function(b){if(this.eg())return b;if(b.eg())return this;if(this.x.equals(b.x)){if(this.y.equals(b.y)){return this.gI();}return this.dP.gy();}var kW=b.x.cv(this.x);var la=b.y.cv(this.y);var hl=la.divide(kW);var eE=hl.square().cv(this.x).cv(b.x);var fE=hl.multiply(this.x.cv(eE)).cv(this.y);return new G(this.dP,eE,fE);};G.prototype.mV=function(){if(this.eg())return this;if(this.y.cR().eJ()==0){return this.dP.gy();}var jF=this.dP.fm(L.valueOf(2));var hm=this.dP.fm(L.valueOf(3));var hl=this.x.square().multiply(hm).add(this.dP.a).divide(this.y.multiply(jF));var eE=hl.square().cv(this.x.multiply(jF));var fE=hl.multiply(this.x.cv(eE)).cv(this.y);return new G(this.dP,eE,fE);};G.prototype.nl=function(k){if(this.eg())return this;if(k.eJ()==0)return this.dP.gy();var e=k;var h=e.multiply(new L("3"));var hw=this.el();var R=this;var i;for(i=h.fg()-2;i>0;--i){R=R.gI();var gm=h.en(i);var hV=e.en(i);if(gm!=hV){R=R.jK(gm?this:hw);}}return R;};G.prototype.kM=function(){var x=this.fI().cR();var y=this.ec().cR();var a=this.dP.kH().cR();var b=this.dP.kB().cR();var n=this.dP.jh();var lhs=y.multiply(y).cl(n);var rhs=x.multiply(x).multiply(x).add(a.multiply(x)).add(b).cl(n);return lhs.equals(rhs);};G.prototype.toString=function(){return '('+this.fI().cR().toString()+','+this.ec().cR().toString()+')';};G.prototype.mY=function(){var n=this.dP.jh();if(this.eg()){throw new Error("Point is at jz.");}var x=this.fI().cR();var y=this.ec().cR();if(x.dw(L.ONE)<0||x.dw(n.cv(L.ONE))>0){throw new Error('x coordinate out of bounds');}if(y.dw(L.ONE)<0||y.dw(n.cv(L.ONE))>0){throw new Error('y coordinate out of bounds');}if(!this.kM()){throw new Error("Point is not on the dP.");}if(this.multiply(n).eg()){throw new Error("Point is not a scalar multiple of G.");}return true;};function bo(){this.fB=null;this.et=null;this.iA=null;this.hb=256;this.ho=0;this.gR=0;this.eu=new Array();if(this.fB==null){this.fB=new Array();this.et=0;if(navigator.appName=="Netscape"&&navigator.appVersion<"5"&&window.crypto){var z=window.crypto.random(32);for(var t=0;t<z.length;++t){this.fB[this.et++]=z.charCodeAt(t)&255;}}while(this.et<this.hb){t=Math.floor(65536*Math.random());this.fB[this.et++]=t>>>8;this.fB[this.et++]=t&255;}this.et=0;this.jj();}};bo.prototype={ka:function(x){this.fB[this.et++]^=x&255;this.fB[this.et++]^=(x>>8)&255;this.fB[this.et++]^=(x>>16)&255;this.fB[this.et++]^=(x>>24)&255;if(this.et>=this.hb)this.et-=this.hb;},jj:function(){this.ka(new Date().getTime());},kg:function(){if(this.iA==null){this.jj();this.iA=1;this.gD(this.fB);for(var et=0;et<this.fB.length;++et){this.fB[et]=0;}this.et=0;}return this.next();},kj:function(iQ){var i;for(i=0;i<iQ.length;++i)iQ[i]=this.kg();},gD:function(key){var i,j,t;for(i=0;i<256;++i)this.eu[i]=i;j=0;for(i=0;i<256;++i){j=(j+this.eu[i]+key[i%key.length])&255;t=this.eu[i];this.eu[i]=this.eu[j];this.eu[j]=t;}this.ho=0;this.gR=0;},next:function(){var t;this.ho=(this.ho+1)&255;this.gR=(this.gR+this.eu[this.ho])&255;t=this.eu[this.ho];this.eu[this.ho]=this.eu[this.gR];this.eu[this.gR]=t;return this.eu[(t+this.eu[this.ho])&255];}};function aP(){this.dX=new Array(0xd6,0x90,0xe9,0xfe,0xcc,0xe1,0x3d,0xb7,0x16,0xb6,0x14,0xc2,0x28,0xfb,0x2c,0x05,0x2b,0x67,0x9a,0x76,0x2a,0xbe,0x04,0xc3,0xaa,0x44,0x13,0x26,0x49,0x86,0x06,0x99,0x9c,0x42,0x50,0xf4,0x91,0xef,0x98,0x7a,0x33,0x54,0x0b,0x43,0xed,0xcf,0xac,0x62,0xe4,0xb3,0x1c,0xa9,0xc9,0x08,0xe8,0x95,0x80,0xdf,0x94,0xfa,0x75,0x8f,0x3f,0xa6,0x47,0x07,0xa7,0xfc,0xf3,0x73,0x17,0xba,0x83,0x59,0x3c,0x19,0xe6,0x85,0x4f,0xa8,0x68,0x6b,0x81,0xb2,0x71,0x64,0xda,0x8b,0xf8,0xeb,0x0f,0x4b,0x70,0x56,0x9d,0x35,0x1e,0x24,0x0e,0x5e,0x63,0x58,0xd1,0xa2,0x25,0x22,0x7c,0x3b,0x01,0x21,0x78,0x87,0xd4,0x00,0x46,0x57,0x9f,0xd3,0x27,0x52,0x4c,0x36,0x02,0xe7,0xa0,0xc4,0xc8,0x9e,0xea,0xbf,0x8a,0xd2,0x40,0xc7,0x38,0xb5,0xa3,0xf7,0xf2,0xce,0xf9,0x61,0x15,0xa1,0xe0,0xae,0x5d,0xa4,0x9b,0x34,0x1a,0x55,0xad,0x93,0x32,0x30,0xf5,0x8c,0xb1,0xe3,0x1d,0xf6,0xe2,0x2e,0x82,0x66,0xca,0x60,0xc0,0x29,0x23,0xab,0x0d,0x53,0x4e,0x6f,0xd5,0xdb,0x37,0x45,0xde,0xfd,0x8e,0x2f,0x03,0xff,0x6a,0x72,0x6d,0x6c,0x5b,0x51,0x8d,0x1b,0xaf,0x92,0xbb,0xdd,0xbc,0x7f,0x11,0xd9,0x5c,0x41,0x1f,0x10,0x5a,0xd8,0x0a,0xc1,0x31,0x88,0xa5,0xcd,0x7b,0xbd,0x2d,0x74,0xd0,0x12,0xb8,0xe5,0xb4,0xb0,0x89,0x69,0x97,0x4a,0x0c,0x96,0x77,0x7e,0x65,0xb9,0xf1,0x09,0xc5,0x6e,0xc6,0x84,0x18,0xf0,0x7d,0xec,0x3a,0xdc,0x4d,0x20,0x79,0xee,0x5f,0x3e,0xd7,0xcb,0x39,0x48);this.gi=new Array(0xa3b1bac6,0x56aa3350,0x677d9197,0xb27022dc);this.kw=new Array(0x00070e15,0x1c232a31,0x383f464d,0x545b6269,0x70777e85,0x8c939aa1,0xa8afb6bd,0xc4cbd2d9,0xe0e7eef5,0xfc030a11,0x181f262d,0x343b4249,0x50575e65,0x6c737a81,0x888f969d,0xa4abb2b9,0xc0c7ced5,0xdce3eaf1,0xf8ff060d,0x141b2229,0x30373e45,0x4c535a61,0x686f767d,0x848b9299,0xa0a7aeb5,0xbcc3cad1,0xd8dfe6ed,0xf4fb0209,0x10171e25,0x2c333a41,0x484f565d,0x646b7279);};aP.prototype={hy:function(key){var k=new Array(36);var gM=aI(key);k[0]=gM[0]^this.gi[0];k[1]=gM[1]^this.gi[1];k[2]=gM[2]^this.gi[2];k[3]=gM[3]^this.gi[3];var fH=new Array(32);for(var i=0;i<32;i++){k[(i+4)]=(k[i]^this.jO(k[(i+1)]^k[(i+2)]^k[(i+3)]^this.kw[i]));fH[i]=k[(i+4)];}return fH;},jO:function(fS){var b=new Array(4);var a=F(fS);b[0]=this.dX[a[0]&0xFF];b[1]=this.dX[a[1]&0xFF];b[2]=this.dX[a[2]&0xFF];b[3]=this.dX[a[3]&0xFF];var ek=bB(b,0);var fH=ek^(ek<<13|(ek>>>(32-13)))^(ek<<23|(ek>>>(32-23)));return fH;},eA:function(fH,data){var x=new Array(36);x[0]=bB(data,0);x[1]=bB(data,4);x[2]=bB(data,8);x[3]=bB(data,12);for(var i=0;i<32;i++){x[(i+4)]=x[i]^this.kN(x[(i+1)]^x[(i+2)]^x[(i+3)]^fH[i]);}var iI=new Array(4);for(var i=35;i>=32;i--){iI[35-i]=x[i];}var kS=ap(iI);return kS;},kN:function(fS){var a=F(fS);var b=new Array(4);b[0]=this.dX[a[0]&0xFF];b[1]=this.dX[a[1]&0xFF];b[2]=this.dX[a[2]&0xFF];b[3]=this.dX[a[3]&0xFF];var ek=bB(b,0);var c=ek^(ek<<2|(ek>>>(32-2)))^(ek<<10|(ek>>>(32-10)))^(ek<<18|(ek>>>(32-18)))^(ek<<24|(ek>>>(32-24)));return c;},gY:function(input,mode){if(input==null){return null;}var gT=null;if(mode==1){var p=16-input.length%16;gT=new Array(input.length+p);K(input,0,gT,0,input.length);for(var i=0;i<p;i++){gT[input.length+i]=p;}}else{var p=input[input.length-1];gT=new Array(input.length-p);K(input,0,gT,0,input.length-p);}return gT;},lk:function(data,key){if(data==undefined||data==null||data.length<=0){dB.log("data is error!");return null;}if(key==undefined||key==null||key.length%16!=0){dB.log("dV key is error!");return null;}var fH=this.hy(key);var cU=16;var loop=parseInt(data.length/cU);var cw=new Array((loop+1)*cU);var dF=new Array(cU);var hk=null;for(var i=0;i<loop;i++){K(data,i*cU,dF,0,cU);hk=this.eA(fH,dF);K(hk,0,cw,i*cU,cU);}var eR=new Array(data.length%cU);if(eR.length>0){K(data,loop*cU,eR,0,data.length%cU);}var padding=this.gY(eR,1);hk=this.eA(fH,padding);K(hk,0,cw,loop*cU,cU);return cw;},kY:function(data,key){if(data==undefined||data==null||data.length%16!=0){dB.log("data is error!");return null;}if(key==undefined||key==null||key.length%16!=0){dB.log("dV key is error!");return null;}var fH=this.hy(key);var gW=new Array(32);for(var i=0;i<fH.length;i++){gW[i]=fH[32-i-1];}var cU=16;var loop=data.length/cU-1;var dF=new Array(cU);var fh=null;var fz=null;K(data,loop*cU,dF,0,cU);fh=this.eA(gW,dF);var eY=this.gY(fh,0);fz=new Array(loop*cU+eY.length);K(eY,0,fz,loop*cU,eY.length);for(var i=0;i<loop;i++){K(data,i*cU,dF,0,cU);fh=this.eA(gW,dF);K(fh,0,fz,i*cU,cU);}return fz;},ll:function(data,key,cd){if(cd==undefined||cd==null||cd.length%16!=0){dB.log("cd is error!");return null;}if(key==undefined||key==null||key.length%16!=0){dB.log("dV key is error!");return null;}if(data==undefined||data==null||data.length<=0){dB.log("data is error!");return null;}var fH=this.hy(key);var cU=16;var loop=parseInt(data.length/cU);var cw=new Array((loop+1)*cU);var dF=new Array(cU);for(var i=0;i<loop;i++){K(data,i*cU,dF,0,cU);for(var j=0;j<cU;j++){dF[j]=dF[j]^cd[j];}cd=this.eA(fH,dF);K(cd,0,cw,i*cU,cU);}var eR=new Array(data.length%cU);if(eR.length>0){K(data,loop*cU,eR,0,data.length%cU);}var padding=this.gY(eR,1);for(var i=0;i<cU;i++){padding[i]=padding[i]^cd[i];}cd=this.eA(fH,padding);K(cd,0,cw,loop*cU,cU);return cw;},kD:function(data,key,cd){if(cd==undefined||cd==null||cd.length%16!=0){dB.log("cd is error!");return null;}if(key==undefined||key==null||key.length%16!=0){dB.log("dV key is error!");return null;}if(data==undefined||data==null||data.length%16!=0){dB.log("data is error!");return null;}var fH=this.hy(key);var gW=new Array(32);for(var i=0;i<fH.length;i++){gW[i]=fH[32-i-1];}var cU=16;var loop=data.length/cU;var dF=new Array(cU);var fh=null;var fz=null;fz=new Array(data.length);for(var i=0;i<loop;i++){K(data,i*cU,dF,0,cU);fh=this.eA(gW,dF);for(var j=0;j<cU;j++){fh[j]=fh[j]^cd[j];}K(dF,0,cd,0,cU);K(fh,0,fz,i*cU,cU);}var eY=this.gY(fh,0);var gq=new Array(fz.length-cU+eY.length);K(fz,0,gq,0,fz.length-cU);K(eY,0,gq,fz.length-cU,eY.length);return gq;}};function T(){this.ft=new Array();this.eW=0;this.hB=0;this.hZ=32;this.gU=[0x7380166f,0x4914b2b9,0x172442d7,0xda8a0600,0xa96f30bc,0x163138aa,0xe38dee4d,0xb0fb0e4e];this.gU=[0x7380166f,0x4914b2b9,0x172442d7,-628488704,-1452330820,0x163138aa,-477237683,-1325724082];this.v=new Array(8);this.jJ=new Array(8);this.fU=[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0];this.X=new Array(68);this.eI=0;this.lw=0x79cc4519;this.lz=0x7a879d8a;if(arguments.length>0){this.lc(arguments[0]);}else{this.gD();}};T.prototype={dj:{ki: -parseInt('10000000000000000000000000000000',2),ir:parseInt('1111111111111111111111111111111',2),parse:function(n){if(n<this.ki){var fa=new Number(-n);var fe=fa.toString(2);var dU=fe.substr(fe.length-31,31);var fC='';for(var i=0;i<dU.length;i++){var hE=dU.substr(i,1);fC+=hE=='0'?'1':'0';}var result=parseInt(fC,2);return(result+1);}else if(n>this.ir){var fa=Number(n);var fe=fa.toString(2);var dU=fe.substr(fe.length-31,31);var fC='';for(var i=0;i<dU.length;i++){var hE=dU.substr(i,1);fC+=hE=='0'?'1':'0';}var result=parseInt(fC,2);return-(result+1);}else{return n;}},gK:function(n){if(n<0){var fa=new Number(-n);var fe=fa.toString(2);var dU=fe.substr(fe.length-8,8);var fC='';for(var i=0;i<dU.length;i++){var hE=dU.substr(i,1);fC+=hE=='0'?'1':'0';}var result=parseInt(fC,2);return(result+1);}else if(n>255){var fa=Number(n);var fe=fa.toString(2);return parseInt(fe.substr(fe.length-8,8),2);}else{return n;}}},gD:function(){this.ft=new Array(4);this.reset();},lc:function(t){this.ft=[].concat(t.ft);this.eW=t.eW;this.hB=t.hB;K(t.X,0,this.X,0,t.X.length);this.eI=t.eI;K(t.v,0,this.v,0,t.v.length);},hL:function(){return this.hZ;},reset:function(){this.hB=0;this.eW=0;for(var kc in this.ft)this.ft[kc]=null;K(this.gU,0,this.v,0,this.gU.length);this.eI=0;K(this.fU,0,this.X,0,this.fU.length);},hS:function(){var i;var fs=this.X;var gu=new Array(64);for(i=16;i<68;i++){fs[i]=this.kT(fs[i-16]^fs[i-9]^(this.rotate(fs[i-3],15)))^(this.rotate(fs[i-13],7))^fs[i-6];}for(i=0;i<64;i++){gu[i]=fs[i]^fs[i+4];}var ej=this.v;var dr=this.jJ;K(ej,0,dr,0,this.gU.length);var fj,fV,hI,fQ,fW;for(i=0;i<16;i++){fW=this.rotate(dr[0],12);fj=this.dj.parse(this.dj.parse(fW+dr[4])+this.rotate(this.lw,i));fj=this.rotate(fj,7);fV=fj^fW;hI=this.dj.parse(this.dj.parse(this.lu(dr[0],dr[1],dr[2])+dr[3])+fV)+gu[i];fQ=this.dj.parse(this.dj.parse(this.ky(dr[4],dr[5],dr[6])+dr[7])+fj)+fs[i];dr[3]=dr[2];dr[2]=this.rotate(dr[1],9);dr[1]=dr[0];dr[0]=hI;dr[7]=dr[6];dr[6]=this.rotate(dr[5],19);dr[5]=dr[4];dr[4]=this.ii(fQ);}for(i=16;i<64;i++){fW=this.rotate(dr[0],12);fj=this.dj.parse(this.dj.parse(fW+dr[4])+this.rotate(this.lz,i));fj=this.rotate(fj,7);fV=fj^fW;hI=this.dj.parse(this.dj.parse(this.kk(dr[0],dr[1],dr[2])+dr[3])+fV)+gu[i];fQ=this.dj.parse(this.dj.parse(this.jL(dr[4],dr[5],dr[6])+dr[7])+fj)+fs[i];dr[3]=dr[2];dr[2]=this.rotate(dr[1],9);dr[1]=dr[0];dr[0]=hI;dr[7]=dr[6];dr[6]=this.rotate(dr[5],19);dr[5]=dr[4];dr[4]=this.ii(fQ);}for(i=0;i<8;i++){ej[i]^=this.dj.parse(dr[i]);}this.eI=0;K(this.fU,0,this.X,0,this.fU.length);},iC:function(hg,ee){var n=hg[ee]<<24;n|=(hg[++ee]&0xff)<<16;n|=(hg[++ee]&0xff)<<8;n|=(hg[++ee]&0xff);this.X[this.eI]=n;if(++this.eI==16){this.hS();}},ly:function(fg){if(this.eI>14){this.hS();}this.X[14]=(this.ke(fg,32));this.X[15]=(fg&(0xffffffff));},kP:function(n,fD,off){fD[off]=this.dj.gK(this.hu(n,24))&0x000000FF;fD[++off]=this.dj.gK(this.hu(n,16))&0x000000FF;fD[++off]=this.dj.gK(this.hu(n,8))&0x000000FF;fD[++off]=this.dj.gK(n)&0x000000FF;},gN:function(lm,lv){this.finish();for(var i=0;i<8;i++){this.kP(this.v[i],lm,lv+i*4);}this.reset();return this.hZ;},update:function(input){this.ft[this.eW++]=input;if(this.eW==this.ft.length){this.iC(this.ft,0);this.eW=0;}this.hB++;},fc:function(input,ee,length){while((this.eW!=0)&&(length>0)){this.update(input[ee]);ee++;length--;}while(length>this.ft.length){this.iC(input,ee);ee+=this.ft.length;length-=this.ft.length;this.hB+=this.ft.length;}while(length>0){this.update(input[ee]);ee++;length--;}},finish:function(){var fg=(this.hB<<3);this.update((128));while(this.eW!=0)this.update((0));this.ly(fg);this.hS();},rotate:function(x,n){return(x<<n)|(this.hu(x,(32-n)));},ii:function(X){return((X)^this.rotate((X),9)^this.rotate((X),17));},kT:function(X){return((X)^this.rotate((X),15)^this.rotate((X),23));},lu:function(X,Y,Z){return(X^Y^Z);},kk:function(X,Y,Z){return((X&Y)|(X&Z)|(Y&Z));},ky:function(X,Y,Z){return(X^Y^Z);},jL:function(X,Y,Z){return((X&Y)|(~X&Z));},hu:function(number,fu){if(number>this.dj.ir||number<this.dj.ki){number=this.dj.parse(number);}if(number>=0){return number>>fu;}else{return(number>>fu)+(2<< ~fu);}},ke:function(number,fu){var fR;var big=new L();big.eT(number);if(big.eJ()>=0){fR=big.ic(fu).hs();}else{var jm=new L();jm.eT(2);var iU= ~fu;var gh='';if(iU<0){var jS=64+iU;for(var i=0;i<jS;i++){gh+='0';}var jE=new L();jE.eT(number>>fu);var jC=new L("10"+gh,2);gh=jC.eD(10);var r=jC.add(jE);fR=r.eD(10);}else{gh=jm.shiftLeft((~fu)).hs();fR=(number>>fu)+gh;}}return fR;},jN:function(g,kz){var jf=new Array(32);var iz=new Array(32);K(kz,1,jf,0,32);K(kz,33,iz,0,32);var iB=CommonUtils.utf8StrToHex('1234567812345678');var dm=iB.length*4;this.update((dm>>8&0x00ff));this.update((dm&0x00ff));var ia=this.gB(iB);this.fc(ia,0,ia.length);var it=this.gB(g.dP.a.cR().eD(16));var ip=this.gB(g.dP.b.cR().eD(16));var jl=this.gB(g.fI().cR().eD(16));var ie=this.gB(g.ec().cR().eD(16));var ig=jf;var iu=iz;this.fc(it,0,it.length);this.fc(ip,0,ip.length);this.fc(jl,0,jl.length);this.fc(ie,0,ie.length);this.fc(ig,0,ig.length);this.fc(iu,0,iu.length);var hX=new Array(this.hL());this.gN(hX,0);return hX;},gB:function(jq){var eN=[];var lp=jq.length;for(var i=0;i<lp;i+=2){eN.push(parseInt(jq.substr(i,2),16));}return eN;}};CipherMode={C1C2C3:0,C1C3C2:1,c1c2c3:0,c1c3c2:1};function aK(eV,eC){if(typeof(eV)!='undefined'&&eV===1){this.eV=CipherMode.C1C3C2;}else{this.eV=CipherMode.C1C2C3;}if(typeof(eC)!='undefined'&&CommonUtils.isHexString(eC)){this.eC=eC;}else{this.eC=null;}this.kQ=new L("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFF",16);this.ks=new L("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFC",16);this.kG=new L("28E9FA9E9D9F5E344D5A9E4BCF6509A7F39789F515AB8F92DDBCBD414D940E93",16);this.eB=new L("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFF7203DF6B21C6052B53BBF40939D54123",16);this.jX=new L("32C4AE2C1F1981195F9904466A39C9948FE30BBFF2660BE1715A4589334C74C7",16);this.kI=new L("BC3736A2F4F6779C59BDCEE36B692153D0A9877CC62A474002DF32E52139F0A0",16);this.lK=256;this.jM=new bo();this.gZ=new V(this.kQ,this.ks,this.kG);this.ha=G.hQ(this.gZ,"04"+this.jX.toString(16)+this.kI.toString(16));};aK.prototype={jU:function(jo){return new L(jo.fg(),this.jM).cl(jo.cv(L.ONE)).add(L.ONE);},generateKeyPair:function(){var eF=this.jt();var cZ=eF['gk'].gv();var jd=eF['fb'].fI().cR();var jn=eF['fb'].ec().cR();var jH=this.fN(jd,32);var lj=this.fN(jn,32);var dC=new Array(65);dC[0]=0x04;K(jH,0,dC,1,32);K(lj,0,dC,33,32);return{'gk':cZ,'fb':dC};},kX:function(){var eF=this.generateKeyPair();var lo=eF.fb;var lA=eF.gk;return{'prvkeyhex':CommonUtils.bytesToHex(lA),'pubkeyhex':CommonUtils.bytesToHex(lo)};},jt:function(eC){var iJ=this.eB;var gF=null;var iV=null;if(CommonUtils.isHexString(eC)){gF=new L(eC,16);}else{do{gF=this.jU(iJ);}while(gF.equals(L.ZERO)||gF.dw(iJ)>=0);}iV=this.ha.multiply(gF);return{'gk':gF,"fb":iV};},jR:function(cZ){var hn=this.ha.multiply(new L(CommonUtils.bytesToHex(cZ),16));var x=this.fN(hn.fI().cR(),32);var y=this.fN(hn.ec().cR(),32);var dC=new Array(65);dC[0]=0x04;K(x,0,dC,1,32);K(y,0,dC,33,32);return dC;},jT:function(){var eF=this.jt();var jd=eF['fb'].fI().cR();var jn=eF['fb'].ec().cR();var hn=this.gZ.kU(jd,jn);eF.x1=hn.fI().cR();return eF;},fN:function(kb,hH){var dF=new Array(hH);for(var i=0;i<dF.length;i++){dF[i]=0;}var eK=kb.gv();if(eK==null){return null;}if(eK.length>hH){K(eK,eK.length-hH,dF,0,hH);}else if(eK.length==hH){dF=eK;}else{K(eK,0,dF,hH-eK.length,eK.length);}return dF;},iH:function(eU,data){var loop=Math.ceil(data.length/32);var eP;var hash=new Array(32);for(var i=0;i<loop;i++){eP=new T();eP.fc(eU,0,eU.length);eP.fc(F(i+1),0,4);eP.gN(hash,0);for(var j=0;j<hash.length&&(i*32+j)<data.length;j++){data[i*32+j]^=hash[j];}}return 0;},iO:function(data,dC){var hi=new T();var iK=new T().jN(this.ha,dC);var hO=data;var eQ=new Array(hi.hL());hi.fc(iK,0,iK.length);hi.fc(hO,0,hO.length);hi.gN(eQ,0);return eQ;},er:function(data,cZ,bZ){var dC=this.jR(cZ);var eQ=this.iO(data,dC);var iE=new L(CommonUtils.bytesToHex(cZ),16);var e=new L(CommonUtils.bytesToHex(eQ),16);var k=null;var r=null;var s=null;do{do{var ib=this.jT();k=ib.gk;r=e.add(ib.x1).cl(this.eB);}while(r.equals(L.ZERO)||r.add(k).equals(this.eB));s=iE.add(L.ONE).fZ(this.eB).multiply(k.cv(r.multiply(iE))).cl(this.eB);}while(s.equals(L.ZERO));if(typeof(bZ)!='undefined'&&bZ==true){var kx=this.kd(r.gv(),s.gv());return kx;}else{var kO=this.fN(r,32);var jI=this.fN(s,32);return CommonUtils.hexToBytes(CommonUtils.bytesToHex(kO)+CommonUtils.bytesToHex(jI));}},ga:function(data,cB,dC,bZ){var eQ=this.iO(data,dC);var r=null;var s=null;if(typeof(bZ)!='undefined'&&bZ==true){var cN=this.le(cB);var hW=new Array(cN["r"].length);var iF=new Array(cN["s"].length);K(cN["r"],0,hW,0,cN["r"].length);K(cN["s"],0,iF,0,cN["s"].length);r=new L(CommonUtils.bytesToHex(hW),16);s=new L(CommonUtils.bytesToHex(iF),16);}else{var dW=CommonUtils.bytesToHex(cB);r=new L(dW.substring(0,64),16);s=new L(dW.substring(64),16);}var hn=this.gZ.kR(CommonUtils.bytesToHex(dC));var e=new L(CommonUtils.bytesToHex(eQ),16);var t=r.add(s).cl(this.eB);if(t.equals(L.ZERO))return false;var lB=this.ha.multiply(s).add(hn.multiply(t));var R=e.add(lB.fI().cR()).cl(this.eB);return r.equals(R);},gd:function(dC,data,bZ){if(dC==null||dC.length==0||data==null||data.length==0){return null;}var lq=G.jZ(this.gZ,dC);var cL=null;var fA=null;var x2=null;var y2=null;var ji=this.jt(this.eC);fA=ji['fb'];var iS=lq.multiply(ji['gk']);x2=this.fN(iS.fI().cR(),32);y2=this.fN(iS.ec().cR(),32);cL=new Array(data.length);K(data,0,cL,0,data.length);var eU=new Array(x2.length+y2.length);K(x2,0,eU,0,x2.length);K(y2,0,eU,x2.length,y2.length);this.iH(eU,cL);var cD=new Array(32);var eP=new T();eP.fc(x2,0,x2.length);eP.fc(data,0,data.length);eP.fc(y2,0,y2.length);eP.gN(cD,0);if(typeof(bZ)!='undefined'&&bZ==true){var ln=fA.fI().cR().gv();var kf=fA.ec().cR().gv();var kn=this.kZ(ln,kf,cL,cD);return kn;}else{var dG=this.fN(fA.fI().cR(),32);var cF=this.fN(fA.ec().cR(),32);var cw=new Array(97+cL.length);cw[0]=0x04;K(dG,0,cw,1,dG.length);K(cF,0,cw,1+dG.length,cF.length);if(this.eV===CipherMode.C1C2C3){K(cL,0,cw,1+dG.length+cF.length,cL.length);K(cD,0,cw,1+dG.length+cF.length+cL.length,cD.length);}else{K(cD,0,cw,1+dG.length+cF.length,cD.length);K(cL,0,cw,1+dG.length+cF.length+cD.length,cL.length);}return cw;}},ht:function(cZ,dH,bZ){if(cZ==null||cZ.length==0||dH==null||dH.length==0){return null;}var fA=new Array(64+1);var cL=null;var cD=new Array(32);if(typeof(bZ)!='undefined'&&bZ==true){var cN=this.lh(dH);cL=new Array(cN["cL"].length);for(var i=0;i<fA.length;i++){fA[i]=0;}fA[0]=0x04;if(cN["dG"].length<=32){K(cN["dG"],0,fA,1+(32-cN["dG"].length),cN["dG"].length);}else{K(cN["dG"],cN["dG"].length-32,fA,1,32);}if(cN["cF"].length<=32){K(cN["cF"],0,fA,1+32+(32-cN["cF"].length),cN["cF"].length);}else{K(cN["cF"],cN["cF"].length-32,fA,1+32,32);}K(cN["cL"],0,cL,0,cL.length);K(cN["cD"],0,cD,0,32);}else{cL=new Array(dH.length-97);K(dH,0,fA,0,65);if(this.eV===CipherMode.C1C2C3){K(dH,65,cL,0,dH.length-97);K(dH,dH.length-32,cD,0,32);}else{K(dH,65,cD,0,32);K(dH,97,cL,0,dH.length-97);}}var lg=G.hQ(this.gZ,CommonUtils.bytesToHex(fA));var jc=lg.multiply(new L(CommonUtils.bytesToHex(cZ),16));var x2=this.fN(jc.fI().cR(),32);var y2=this.fN(jc.ec().cR(),32);var eU=new Array(x2.length+y2.length);K(x2,0,eU,0,x2.length);K(y2,0,eU,x2.length,y2.length);var jv=new Array(cL.length);K(cL,0,jv,0,cL.length);this.iH(eU,cL);if(co(jv,0,cL,0,cL.length)==0){return null;}var eP=new T();var hash=new Array(32);eP.fc(x2,0,x2.length);eP.fc(cL,0,cL.length);eP.fc(y2,0,y2.length);eP.gN(hash,0);var hT=CommonUtils.bytesToHex(cD);var ix=CommonUtils.bytesToHex(hash);if(!hT|| !ix||hT!=ix){return null;}return cL;},kd:function(r,s){var fl=r.length+s.length+4;var ew=[];ew[0]=0x30;if(fl<0x80){ew[1]=fl;}else{var fp=F(fl);var i=0;while(fp[i]==0&&i<fp.length){i++;}ew[1]=0x80|(fp.length-i);for(var j=2;i<fp.length;i++,j++){ew[j]=fp[i];}}var bZ=new Array(fl+ew.length);var cY=0;K(ew,0,bZ,0,ew.length);cY+=ew.length;bZ[cY++]=0x02;bZ[cY++]=r.length;K(r,0,bZ,cY,r.length);cY+=r.length;bZ[cY++]=0x02;bZ[cY++]=s.length;K(s,0,bZ,cY,s.length);cY+=s.length;return bZ;},kZ:function(dG,cF,cL,cD){var fJ=cL.length;var dS=[];if(fJ<0x80){dS[0]=0x04;dS[1]=fJ;}else{dS[0]=0x04;var gs=F(fJ);var i=0;while(gs[i]==0&&i<gs.length){i++;}dS[1]=0x80|(gs.length-i);for(var j=2;i<gs.length;i++,j++){dS[j]=gs[i];}}var fl=dG.length+cF.length+cL.length+cD.length+6+dS.length;var ew=[];ew[0]=0x30;if(fl<0x80){ew[1]=fl;}else{var fp=F(fl);var i=0;while(fp[i]==0&&i<fp.length){i++;}ew[1]=0x80|(fp.length-i);for(var j=2;i<fp.length;i++,j++){ew[j]=fp[i];}}var bZ=new Array(fl+ew.length);var cY=0;K(ew,0,bZ,0,ew.length);cY+=ew.length;bZ[cY++]=0x02;bZ[cY++]=dG.length;K(dG,0,bZ,cY,dG.length);cY+=dG.length;bZ[cY++]=0x02;bZ[cY++]=cF.length;K(cF,0,bZ,cY,cF.length);cY+=cF.length;bZ[cY++]=0x04;bZ[cY++]=cD.length;K(cD,0,bZ,cY,cD.length);cY+=cD.length;K(dS,0,bZ,cY,dS.length);cY+=dS.length;K(cL,0,bZ,cY,cL.length);return bZ;},le:function(bZ){var pos=0;if(bZ[pos++]!=0x30){return null;}if((bZ[pos]&0xFF)<=0x7F){pos++;}else{pos+=(bZ[pos]&0x7F)+1;}pos++;var fX=bZ[pos];var r=new Array(fX);K(bZ,++pos,r,0,fX);pos+=fX;pos++;var he=bZ[pos];var s=new Array(he);K(bZ,++pos,s,0,he);pos+=he;return{'r':r,"s":s};},lh:function(bZ){var pos=0;if(bZ[pos++]!=0x30){return null;}if((bZ[pos]&0xFF)<=0x7F){pos++;}else{pos+=(bZ[pos]&0x7F)+1;}pos++;var jk=bZ[pos];var dG=new Array(jk);K(bZ,++pos,dG,0,jk);pos+=jk;pos++;var iX=bZ[pos];var cF=new Array(iX);K(bZ,++pos,cF,0,iX);pos+=iX;pos++;var jD=bZ[pos];var cD=new Array(jD);K(bZ,++pos,cD,0,jD);pos+=jD;pos++;var fJ=0;if((bZ[pos]&0xFF)<=0x7F){fJ=bZ[pos]&0xFF;}else{for(var i=0,j=(bZ[pos]&0x7F)-1;i<(bZ[pos]&0x7F);i++,j--){fJ=fJ|((bZ[pos+i+1]&0xFF)<<(j*8));}pos+=(bZ[pos]&0x7F);}var cL=new Array(fJ);K(bZ,++pos,cL,0,fJ);pos+=fJ;return{'dG':dG,"cF":cF,"cL":cL,"cD":cD};}};function bL(){};bL.prototype={source:function(source){if(!source){dB.error("source is empty!");return false;}else{return true;}},fd:function(eS){if(typeof(eS)=='undefined'){dB.error("eS type undefined!");}else if(eS instanceof Array){eS=CommonUtils.bytesToHex(eS);}if(this.ge(eS,"04")&&eS.length==130){return true;}else{dB.error("eS length error!");return false;}},fo:function(fM){if(typeof(fM)=='undefined'){dB.error("fM type undefined!");}else if(fM instanceof Array){fM=CommonUtils.bytesToHex(fM);}if(fM.length%2==0){return true;}else{dB.error("fM length error!");return false;}},gp:function(cw){if(typeof(cw)=='undefined'){dB.error("cw type undefined!");}else if(cw instanceof Array){cw=CommonUtils.bytesToHex(cw);}if(this.ge(cw,"04")&&cw.length%2==0&&cw.length>194){return true;}else if(this.ge(cw,"30")&&cw.length%2==0&&cw.length>216){return true;}else{dB.error("cw format jw length error!");return false;}},fP:function(er){if(typeof(er)=='undefined'){dB.error("er type undefined!");}else if(er instanceof Array){er=CommonUtils.bytesToHex(er);}if(er.length==128){return true;}else if(this.ge(er,"30")&&er.length%2==0){return true;}else{dB.error("er format jw length error!");return false;}},fF:function(key){if(typeof(key)=='undefined'){dB.error("key type undefined!");}else if(key instanceof Array){key=CommonUtils.bytesToHex(key);}if(key.length==32){return true;}else{dB.error("key length error!");return false;}},hj:function(cd){if(typeof(cd)=='undefined'){dB.error("cd type undefined!");}else if(cd instanceof Array){cd=CommonUtils.bytesToHex(cd);}if(cd.length==32){return true;}else{dB.error("cd length error!");return false;}},hd:function(cw){if(typeof(cw)=='undefined'){dB.error("cw type undefined!");}else if(cw instanceof Array){cw=CommonUtils.bytesToHex(cw);}if(cw&&cw.length%16==0){return true;}else{dB.error("cw format jw length error!");return false;}},ge:function(fL,start){var reg=new RegExp("^"+start);return reg.test(fL);}};function Sm2Utils(eV,eC){if(typeof(eV)!='undefined'){this.eV=eV;}else{this.eV=CipherMode.C1C2C3;}if(typeof(eC)!='undefined'&&CommonUtils.isHexString(eC)){this.eC=eC;}else{this.eC=null;}this.cw=new aK(this.eV,this.eC);this.smValidate=new bL();};Sm2Utils.prototype={generateKeyPair:function(){return this.cw.kX();},getPublicKey:function(kL){var cZ=CommonUtils.hexToBytes(kL);var dC=this.cw.jR(cZ);return CommonUtils.bytesToHex(dC);},encryptFromText:function(cV,cb){if(!this.smValidate.fd(cV)){return null;}if(!this.smValidate.source(cb)){return null;}var dC=CommonUtils.hexToBytes(cV);var cu=CommonUtils.utf8StrToBytes(cb);var dH=this.cw.gd(dC,cu,false);return CommonUtils.bytesToHex(dH);},decryptToText:function(du,cE){if(!this.smValidate.fo(du)){return null;}if(!this.smValidate.gp(cE)){return null;}var cZ=CommonUtils.hexToBytes(du);var dH=CommonUtils.hexToBytes(cE);var cu=this.cw.ht(cZ,dH,false);return CommonUtils.bytesToUtf8Str(cu);},encryptFromHex:function(cV,dO){if(!this.smValidate.fd(cV)){return null;}if(!this.smValidate.source(dO)){return null;}var dC=CommonUtils.hexToBytes(cV);var cu=CommonUtils.hexToBytes(dO);var dH=this.cw.gd(dC,cu,false);return CommonUtils.bytesToHex(dH);},decryptToHex:function(du,cE){if(!this.smValidate.fo(du)){return null;}if(!this.smValidate.gp(cE)){return null;}var cZ=CommonUtils.hexToBytes(du);var dH=CommonUtils.hexToBytes(cE);var cu=this.cw.ht(cZ,dH,false);return CommonUtils.bytesToHex(cu);},encryptFromData:function(dC,cu){if(!this.smValidate.fd(dC)){return null;}if(!this.smValidate.source(cu)){return null;}return this.cw.gd(dC,cu,false);},decryptToData:function(cZ,dH){if(!this.smValidate.fo(cZ)){return null;}if(!this.smValidate.gp(dH)){return null;}return this.cw.ht(cZ,dH,false);},encryptASN1FromText:function(cV,cb){if(!this.smValidate.fd(cV)){return null;}if(!this.smValidate.source(cb)){return null;}var dC=CommonUtils.hexToBytes(cV);var cu=CommonUtils.utf8StrToBytes(cb);var dH=this.cw.gd(dC,cu,true);return CommonUtils.bytesToHex(dH);},decryptASN1ToText:function(du,cE){if(!this.smValidate.fo(du)){return null;}if(!this.smValidate.gp(cE)){return null;}var cZ=CommonUtils.hexToBytes(du);var dH=CommonUtils.hexToBytes(cE);var cu=this.cw.ht(cZ,dH,true);return CommonUtils.bytesToUtf8Str(cu);},encryptASN1FromHex:function(cV,dO){if(!this.smValidate.fd(cV)){return null;}if(!this.smValidate.source(dO)){return null;}var dC=CommonUtils.hexToBytes(cV);var cu=CommonUtils.hexToBytes(dO);var dH=this.cw.gd(dC,cu,true);return CommonUtils.bytesToHex(dH);},decryptASN1ToHex:function(du,cE){if(!this.smValidate.fo(du)){return null;}if(!this.smValidate.gp(cE)){return null;}var cZ=CommonUtils.hexToBytes(du);var dH=CommonUtils.hexToBytes(cE);var cu=this.cw.ht(cZ,dH,true);return CommonUtils.bytesToHex(cu);},encryptASN1FromData:function(dC,cu){if(!this.smValidate.fd(dC)){return null;}if(!this.smValidate.source(cu)){return null;}return this.cw.gd(dC,cu,true);},decryptASN1ToData:function(cZ,dH){if(!this.smValidate.fo(cZ)){return null;}if(!this.smValidate.gp(dH)){return null;}return this.cw.ht(cZ,dH,true);},signFromText:function(du,cb){if(!this.smValidate.fo(du)){return null;}if(!this.smValidate.source(cb)){return null;}var cZ=CommonUtils.hexToBytes(du);var cu=CommonUtils.utf8StrToBytes(cb);var cB=this.cw.er(cu,cZ,false);return CommonUtils.bytesToHex(cB);},verifySignFromText:function(cV,cb,dW){if(!this.smValidate.fd(cV)){return null;}if(!this.smValidate.source(cb)){return null;}if(!this.smValidate.fP(dW)){return null;}var dC=CommonUtils.hexToBytes(cV);var cu=CommonUtils.utf8StrToBytes(cb);var cB=CommonUtils.hexToBytes(dW);return this.cw.ga(cu,cB,dC,false);},signFromHex:function(du,dO){if(!this.smValidate.fo(du)){return null;}if(!this.smValidate.source(dO)){return null;}var cZ=CommonUtils.hexToBytes(du);var cu=CommonUtils.hexToBytes(dO);var cB=this.cw.er(cu,cZ,false);return CommonUtils.bytesToHex(cB);},verifySignFromHex:function(cV,dO,dW){if(!this.smValidate.fd(cV)){return null;}if(!this.smValidate.source(dO)){return null;}if(!this.smValidate.fP(dW)){return null;}var dC=CommonUtils.hexToBytes(cV);var cu=CommonUtils.hexToBytes(dO);var cB=CommonUtils.hexToBytes(dW);return this.cw.ga(cu,cB,dC,false);},signFromData:function(cZ,cu){if(!this.smValidate.fo(cZ)){return null;}if(!this.smValidate.source(cu)){return null;}return this.cw.er(cu,cZ,false);},verifySignFromData:function(dC,cu,cB){if(!this.smValidate.fd(dC)){return null;}if(!this.smValidate.source(cu)){return null;}if(!this.smValidate.fP(cB)){return null;}return this.cw.ga(cu,cB,dC,false);},signASN1FromText:function(du,cb){if(!this.smValidate.fo(du)){return null;}if(!this.smValidate.source(cb)){return null;}var cZ=CommonUtils.hexToBytes(du);var cu=CommonUtils.utf8StrToBytes(cb);var cB=this.cw.er(cu,cZ,true);return CommonUtils.bytesToHex(cB);},verifySignASN1FromText:function(cV,cb,dW){if(!this.smValidate.fd(cV)){return null;}if(!this.smValidate.source(cb)){return null;}if(!this.smValidate.fP(dW)){return null;}var dC=CommonUtils.hexToBytes(cV);var cu=CommonUtils.utf8StrToBytes(cb);var cB=CommonUtils.hexToBytes(dW);return this.cw.ga(cu,cB,dC,true);},signASN1FromHex:function(du,dO){if(!this.smValidate.fo(du)){return null;}if(!this.smValidate.source(dO)){return null;}var cZ=CommonUtils.hexToBytes(du);var cu=CommonUtils.hexToBytes(dO);var cB=this.cw.er(cu,cZ,true);return CommonUtils.bytesToHex(cB);},verifySignASN1FromHex:function(cV,dO,dW){if(!this.smValidate.fd(cV)){return null;}if(!this.smValidate.source(dO)){return null;}if(!this.smValidate.fP(dW)){return null;}var dC=CommonUtils.hexToBytes(cV);var cu=CommonUtils.hexToBytes(dO);var cB=CommonUtils.hexToBytes(dW);return this.cw.ga(cu,cB,dC,true);},signASN1FromData:function(cZ,cu){if(!this.smValidate.fo(cZ)){return null;}if(!this.smValidate.source(cu)){return null;}return this.cw.er(cu,cZ,true);},verifySignASN1FromData:function(dC,cu,cB){if(!this.smValidate.fd(dC)){return null;}if(!this.smValidate.source(cu)){return null;}if(!this.smValidate.fP(cB)){return null;}return this.cw.ga(cu,cB,dC,true);}};var Sm3Utils={smValidate:new bL(),encryptFromText:function(cb){if(!this.smValidate.source(cb)){return null;}var cu=CommonUtils.utf8StrToBytes(cb);var eQ=this.encryptFromData(cu);return CommonUtils.bytesToHex(eQ);},encryptFromHex:function(dO){if(!this.smValidate.source(dO)){return null;}var cu=CommonUtils.hexToBytes(dO);var eQ=this.encryptFromData(cu);return CommonUtils.bytesToHex(eQ);},encryptFromData:function(cu){if(!this.smValidate.source(cu)){return null;}var eQ=new Array(32);var jy=new T();jy.fc(cu,0,cu.length);jy.gN(eQ,0);return eQ;}};var Sm4Utils={ECB:{smValidate:new bL(),encryptFromText:function(cb,dz){if(!this.smValidate.source(cb)){return null;}if(!this.smValidate.fF(dz)){return null;}var cK=CommonUtils.hexToBytes(dz);var cu=CommonUtils.utf8StrToBytes(cb);var dH=this.encryptFromData(cu,cK);return CommonUtils.bytesToHex(dH);},encryptFromHex:function(dO,dz){if(!this.smValidate.source(dO)){return null;}if(!this.smValidate.fF(dz)){return null;}var cK=CommonUtils.hexToBytes(dz);var cu=CommonUtils.hexToBytes(dO);var dH=this.encryptFromData(cu,cK);return CommonUtils.bytesToHex(dH);},encryptFromData:function(cu,cK){if(!this.smValidate.source(cu)){return null;}if(!this.smValidate.fF(cK)){return null;}try{var dV=new aP();return dV.lk(cu,cK);}catch(e){dB.error(e);return null;}},decryptToText:function(cE,dz){if(!this.smValidate.hd(cE)){return null;}if(!this.smValidate.fF(dz)){return null;}var cK=CommonUtils.hexToBytes(dz);var dH=CommonUtils.hexToBytes(cE);var cu=this.decryptToData(dH,cK);return CommonUtils.bytesToUtf8Str(cu);},decryptToHex:function(cE,dz){if(!this.smValidate.hd(cE)){return null;}if(!this.smValidate.fF(dz)){return null;}var cK=CommonUtils.hexToBytes(dz);var dH=CommonUtils.hexToBytes(cE);var cu=this.decryptToData(dH,cK);return CommonUtils.bytesToHex(cu);},decryptToData:function(dH,cK){if(!this.smValidate.hd(dH)){return null;}if(!this.smValidate.fF(cK)){return null;}try{var dV=new aP();return dV.kY(dH,cK);}catch(e){dB.error(e);return null;}}},CBC:{smValidate:new bL(),encryptFromText:function(cb,dz,dZ){if(!this.smValidate.source(cb)){return null;}if(!this.smValidate.fF(dz)){return null;}if(!this.smValidate.hj(dZ)){return null;}var cK=CommonUtils.hexToBytes(dz);var fq=CommonUtils.hexToBytes(dZ);var cu=CommonUtils.utf8StrToBytes(cb);var dH=this.encryptFromData(cu,cK,fq);return CommonUtils.bytesToHex(dH);},encryptFromHex:function(dO,dz,dZ){if(!this.smValidate.source(dO)){return null;}if(!this.smValidate.fF(dz)){return null;}if(!this.smValidate.hj(dZ)){return null;}var cK=CommonUtils.hexToBytes(dz);var fq=CommonUtils.hexToBytes(dZ);var cu=CommonUtils.hexToBytes(dO);var dH=this.encryptFromData(cu,cK,fq);return CommonUtils.bytesToHex(dH);},encryptFromData:function(cu,cK,fq){if(!this.smValidate.source(cu)){return null;}if(!this.smValidate.fF(cK)){return null;}if(!this.smValidate.hj(fq)){return null;}try{var dV=new aP();return dV.ll(cu,cK,fq);}catch(e){dB.error(e);return null;}},decryptToText:function(cE,dz,dZ){if(!this.smValidate.hd(cE)){return null;}if(!this.smValidate.fF(dz)){return null;}if(!this.smValidate.hj(dZ)){return null;}var cK=CommonUtils.hexToBytes(dz);var fq=CommonUtils.hexToBytes(dZ);var dH=CommonUtils.hexToBytes(cE);var cu=this.decryptToData(dH,cK,fq);return CommonUtils.bytesToUtf8Str(cu);},decryptToHex:function(cE,dz,dZ){if(!this.smValidate.hd(cE)){return null;}if(!this.smValidate.fF(dz)){return null;}if(!this.smValidate.hj(dZ)){return null;}var cK=CommonUtils.hexToBytes(dz);var fq=CommonUtils.hexToBytes(dZ);var dH=CommonUtils.hexToBytes(cE);var cu=this.decryptToData(dH,cK,fq);return CommonUtils.bytesToHex(cu);},decryptToData:function(dH,cK,fq){if(!this.smValidate.hd(dH)){return null;}if(!this.smValidate.fF(cK)){return null;}if(!this.smValidate.hj(fq)){return null;}try{var dV=new aP();return dV.kD(dH,cK,fq);}catch(e){dB.error(e);return null;}}}};

1+1`
