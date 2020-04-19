package util

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	hjson "github.com/hjson/hjson-go"
)

func fixJSON(data []byte) []byte {
	data = bytes.Replace(data, []byte("\\u003c"), []byte("<"), -1)
	data = bytes.Replace(data, []byte("\\u003e"), []byte(">"), -1)
	data = bytes.Replace(data, []byte("\\u0026"), []byte("&"), -1)
	data = bytes.Replace(data, []byte("\\u0008"), []byte("\\b"), -1)
	data = bytes.Replace(data, []byte("\\u000c"), []byte("\\f"), -1)
	return data
}

func HjsonToJSON(bs []byte) ([]byte, error) {
	var value interface{}
	if err := hjson.Unmarshal(bs, &value); err != nil {
		return nil, err
	}

	out, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return fixJSON(out), nil
}

func UnmarshalFromHjson(in []byte, value interface{}) error {
	bs, err := HjsonToJSON(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, value)
}

func FromHjsonFile(filename string, target interface{}, contentFuncs ...func([]byte) []byte) error {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if bytes.HasPrefix(bs, []byte{0xEF, 0xBB, 0xBF}) {
		bs = bs[3:]
	}

	for _, cb := range contentFuncs {
		bs = cb(bs)
	}
	bs, err = HjsonToJSON(bs)
	if err != nil {
		return err
	}

	return json.Unmarshal(bs, target)
}

func FromHjsonBytes(in []byte, target interface{}) error {
	bs, err := HjsonToJSON(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(bs, target)
}

func WriteToFile(filename string, value interface{}, indent ...bool) error {
	var out []byte
	var err error

	if len(indent) > 0 && indent[0] {
		out, err = json.MarshalIndent(value, "", "  ")
	} else {
		out, err = json.Marshal(value)
	}
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, out, 0666)
}
