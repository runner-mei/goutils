package util

import (
	"errors"
	"io"
	"math"
	"math/big"
	"strconv"
)

type Uint128 struct {
	H uint64
	L uint64
}

func (self *Uint128) IsZero() bool {
	return self.H == 0 && self.L == 0
}

func (self *Uint128) Zero() *Uint128 {
	self.H = 0
	self.L = 0
	return self
}

func (self *Uint128) Set(v *Uint128) *Uint128 {
	self.H = v.H
	self.L = v.L
	return self
}

func (self *Uint128) SetUint64(v uint64) *Uint128 {
	self.H = 0
	self.L = v
	return self
}

func (self *Uint128) Add(v uint64) *Uint128 {
	self.L += v
	if self.L < v {
		self.H++
	}
	return self
}

func (self *Uint128) Add128(v *Uint128) *Uint128 {
	self.H += v.H
	return self.Add(v.L)
}

func (self *Uint128) Sub(v uint64) *Uint128 {
	self.L -= v
	if self.L > v {
		self.H--
	}
	return self
}

func (self *Uint128) Sub128(v *Uint128) *Uint128 {
	self.H -= v.H
	return self.Sub(v.L)
}

func (self *Uint128) Compare(v uint64) int {
	if self.H > 0 {
		return 1
	}
	if self.L > v {
		return 1
	}
	if self.L < v {
		return -1
	}
	return 0
}

func (self *Uint128) ToBigInt() *big.Int {
	r := big.NewInt(0).Mul(big.NewInt(0).SetUint64(self.H),
		big.NewInt(0).SetUint64(math.MaxUint64))
	r.Add(r, big.NewInt(0).SetUint64(self.L))
	return r
}

func (self *Uint128) ToBigFloat() *big.Float {
	r := big.NewFloat(0).Mul(big.NewFloat(0).SetUint64(self.H),
		big.NewFloat(0).SetUint64(math.MaxUint64))
	r.Add(r, big.NewFloat(0).SetUint64(self.L))
	return r
}

func (self *Uint128) String() string {
	if 0 == self.H {
		return strconv.FormatUint(self.L, 10)
	}

	return self.ToBigInt().String()
}

func (self *Uint128) AppendFormat(bs []byte, base int) []byte {
	if 0 == self.H {
		return strconv.AppendUint(bs, self.L, base)
	}
	//return append(bs, []byte(self.ToBigInt().String())...)
	return self.ToBigInt().Append(bs, base)
}

func (self *Uint128) WriteFormat(w io.Writer, base int) (int, error) {
	var barray [128]byte

	bs := self.AppendFormat(barray[:0], base)
	return w.Write(bs)
}

func (self *Uint128) MarshalJSON() ([]byte, error) {
	return []byte(self.String()), nil
}

func Uint128BuildFromFloat64(f64 float64) Uint128 {
	if f64 < 0 {
		panic(errors.New("Uint128BuildFromFloat64 fail"))
	}
	if f64 < math.MaxUint64 {
		return Uint128{L: uint64(f64)}
	}

	h := f64 / math.MaxUint64

	l := f64 - (float64(uint64(h)) * math.MaxUint64)

	return Uint128{H: uint64(h), L: uint64(l)}
}
