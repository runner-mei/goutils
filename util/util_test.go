package util

import "testing"


func TestPassword(t *testing.T) {
	tests := []string{"abcssss", "aaaa", "a", "12324345456456dgddrfgdrdg", "XJ5QJSVMKZGBOQO7HMSIJO5BERW2OYWDVNPM3BH32NLSWUCNJ4FIP3BML7EKUBNO"}

	t.Log(len("XJ5QJSVMKZGBOQO7HMSIJO5BERW2OYWDVNPM3BH32NLSWUCNJ4FIP3BML7EKUBNO"))

	s := BlockString("[1]77a247da9fa842bee28bba8ac93858f695d05d4e6e216abadfa4f2a1e8f676c3a885d7c06dcda8a81a702d724b4d5d65dedfffcd73697cdbaf9a79889281731a3dc6d0900ef3ffb6b33623f2a8ce5ff5ca82567e7657f1f7510d982ddde0974d")
	t.Log(len(s))
	t.Log(s)

	//t.Log(BlockString("[1]434944f7a004c28362c2e4f64a714874b5ac6bfdb510976028278cbdead62e1c74e79ccf75fc37658e4e5b2ff3a2dadef7c9c623b77ed44a00f21d22a86a520f00000000000000000000000000000000"))
	t.Log(CopyToBlock(tests[4]))
	for _, abc1 := range tests {
		res := CopyToBlock(abc1)
		abc2 := BlockString(res)
		if abc1 != abc2 || abc1 == res {
			t.Error("abc1=", abc1)
			t.Error("abc2=", abc2)
			t.Error(res)
		}
	}

	for _, abc1 := range tests {
		abc2 := BlockString(abc1)
		if abc1 != abc2 {
			t.Error("abc1=", abc1)
			t.Error("abc2=", abc2)
		}
	}

	encrypt_texts := []string{
		"[1]2c53ec26b6e8e4057a8233be5e126bc349825446e6cd733ea54d1e57fdb30ff0",
		"[1]a4826ba6d68a487012178e28697ec91d57cc51c0a82fdd92a61ea6584c956098",
		"[1]23bf99f1900cf407e2dfc4aad0db94065a7e771a36ff4366b79aacdca9c69d85",
		"[1]5653d890189daf37fdd595c62eeb3595f3ee725889dce8f7770d6693ffc493273b4fbc2c7d8a320f365bbb087b65b0bb",
		"[1]7e75ce4c67654d56957b6e574a684a9c4c2227d35a2afb7a53106ce4e9a9a33722e4d2216aeab1b37416b6405bd7ea7fdae4980b7192f4f75ede46ba9c023963720810091ba9828e161e4b173a1620d9c8b757d566fba1002a7bc62ade35b8d0",
	}

	for idx, abc1 := range encrypt_texts {
		abc2 := BlockString(abc1)
		if abc2 != tests[idx] {
			t.Error("abc1=", abc1)
			t.Error("abc2=", abc2)
		}
	}
}
