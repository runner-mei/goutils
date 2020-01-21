package util

import (
	"github.com/runner-mei/goutils/as"
)

var ErrValueNotFound = as.ErrValueNotFound
var ErrValueNull = as.ErrValueNull

var (
	AsMap              = as.Map
	AsObject           = as.Object
	AsObjects          = as.Objects
	IsEmptyArray       = as.IsEmptyArray
	AsArray            = as.Array
	AsInts             = as.Ints
	AsInt64s           = as.Int64s
	AsArrayWithDefault = as.ArrayWithDefault
	ArrayWithDefault   = as.ArrayWithDefault

	AsBool                 = as.Bool
	AsBoolWithDefaultValue = as.BoolWithDefault
	AsInt                  = as.Int
	AsInt64                = as.Int64
	AsInt32                = as.Int32
	AsInt16                = as.Int16
	AsInt8                 = as.Int8
	AsIntWithDefault       = as.IntWithDefault
	AsInt64WithDefault     = as.Int64WithDefault
	AsInt32WithDefault     = as.Int32WithDefault
	// AsInt16WithDefault     = as.Int16WithDefault
	// AsInt8WithDefault      = as.Int8WithDefault

	AsUint   = as.Uint
	AsUint64 = as.Uint64
	AsUint32 = as.Uint32
	AsUint16 = as.Uint16
	AsUint8  = as.Uint8

	AsUintWithDefault   = as.UintWithDefault
	AsUint64WithDefault = as.Uint64WithDefault
	AsUint32WithDefault = as.Uint32WithDefault
	// AsUint16WithDefault = as.Uint16WithDefault
	// AsUint8WithDefault  = as.Uint8WithDefault

	AsFloat64 = as.Float64
	AsFloat32 = as.Float32

	AsFloat64WithDefault = as.Float64WithDefault
	AsFloat32WithDefault = as.Float32WithDefault

	AsString            = as.String
	AsStringWithDefault = as.StringWithDefault
	AsDuration          = as.Duration
	AsTime              = as.Time
)
