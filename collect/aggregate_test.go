package collect

import (
	"fmt"
	"testing"
	"time"
)

func TestAggregate(t *testing.T) {

	t1 := time.Date(2018, time.May, 1, 1, 1, 1, 21, time.Local)
	t2 := time.Date(2018, time.May, 2, 1, 1, 1, 21, time.Local)
	t3 := time.Date(2018, time.May, 3, 1, 1, 1, 21, time.Local)
	t12 := time.Date(2018, time.May, 12, 1, 1, 1, 21, time.Local)
	t13 := time.Date(2018, time.May, 13, 1, 1, 1, 21, time.Local)
	t14 := time.Date(2018, time.May, 14, 1, 1, 1, 21, time.Local)
	t15 := time.Date(2018, time.May, 15, 1, 1, 1, 21, time.Local)

	for idx, test := range []struct {
		t        string
		operator string
		excepted interface{}
		value    interface{}
	}{{t: "integer", operator: "first", excepted: 12, value: 12},
		{t: "integer", operator: "first", excepted: 12, value: []interface{}{12}},
		{t: "integer", operator: "first", excepted: 12, value: []interface{}{12, 13}},
		{t: "integer", operator: "last", excepted: 12, value: 12},
		{t: "integer", operator: "last", excepted: 12, value: []interface{}{12}},
		{t: "integer", operator: "last", excepted: 12, value: []interface{}{13, 12}},
		{t: "integer", operator: "count", excepted: uint(1), value: 12},
		{t: "integer", operator: "count", excepted: uint(1), value: []interface{}{12}},
		{t: "integer", operator: "count", excepted: uint(2), value: []interface{}{12, 12}},
		{t: "integer", operator: "count", excepted: uint(3), value: []interface{}{12, 1, 12}},
		{t: "integer", operator: "max", excepted: 12, value: 12},
		{t: "integer", operator: "max", excepted: 12, value: []interface{}{12}},
		{t: "integer", operator: "max", excepted: 12, value: []interface{}{12, 1, 2, 3}},
		{t: "integer", operator: "min", excepted: 12, value: 12},
		{t: "integer", operator: "min", excepted: 12, value: []interface{}{12}},
		{t: "integer", operator: "min", excepted: 12, value: []interface{}{12, 13, 14, 15}},
		{t: "integer", operator: "sum", excepted: 12, value: 12},
		{t: "integer", operator: "sum", excepted: int64(12), value: []interface{}{12}},
		{t: "integer", operator: "sum", excepted: int64(12), value: []interface{}{1, 3, 5, 2, 1}},
		{t: "integer", operator: "avg", excepted: 12, value: 12},
		{t: "integer", operator: "avg", excepted: 12, value: []interface{}{12}},
		{t: "integer", operator: "avg", excepted: int64(12), value: []interface{}{13, 11, 14, 10}},

		{t: "biginteger", operator: "first", excepted: 12, value: 12},
		{t: "biginteger", operator: "first", excepted: 12, value: []interface{}{12}},
		{t: "biginteger", operator: "first", excepted: 12, value: []interface{}{12, 13}},
		{t: "biginteger", operator: "last", excepted: 12, value: 12},
		{t: "biginteger", operator: "last", excepted: 12, value: []interface{}{12}},
		{t: "biginteger", operator: "last", excepted: 12, value: []interface{}{13, 12}},
		{t: "biginteger", operator: "count", excepted: uint(1), value: 12},
		{t: "biginteger", operator: "count", excepted: uint(1), value: []interface{}{12}},
		{t: "biginteger", operator: "count", excepted: uint(2), value: []interface{}{12, 12}},
		{t: "biginteger", operator: "count", excepted: uint(3), value: []interface{}{12, 1, 12}},
		{t: "biginteger", operator: "max", excepted: 12, value: 12},
		{t: "biginteger", operator: "max", excepted: 12, value: []interface{}{12}},
		{t: "biginteger", operator: "max", excepted: 12, value: []interface{}{12, 1, 2, 3}},
		{t: "biginteger", operator: "min", excepted: 12, value: 12},
		{t: "biginteger", operator: "min", excepted: 12, value: []interface{}{12}},
		{t: "biginteger", operator: "min", excepted: 12, value: []interface{}{12, 13, 14, 15}},
		{t: "biginteger", operator: "sum", excepted: 12, value: 12},
		{t: "biginteger", operator: "sum", excepted: int64(12), value: []interface{}{12}},
		{t: "biginteger", operator: "sum", excepted: int64(12), value: []interface{}{1, 3, 5, 2, 1}},
		{t: "biginteger", operator: "avg", excepted: 12, value: 12},
		{t: "biginteger", operator: "avg", excepted: 12, value: []interface{}{12}},
		{t: "biginteger", operator: "avg", excepted: int64(12), value: []interface{}{13, 11, 14, 10}},

		{t: "decimal", operator: "first", excepted: 12, value: 12},
		{t: "decimal", operator: "first", excepted: 12, value: []interface{}{12}},
		{t: "decimal", operator: "first", excepted: 12, value: []interface{}{12, 13}},
		{t: "decimal", operator: "last", excepted: 12, value: 12},
		{t: "decimal", operator: "last", excepted: 12, value: []interface{}{12}},
		{t: "decimal", operator: "last", excepted: 12, value: []interface{}{13, 12}},
		{t: "decimal", operator: "count", excepted: uint(1), value: 12},
		{t: "decimal", operator: "count", excepted: uint(1), value: []interface{}{12}},
		{t: "decimal", operator: "count", excepted: uint(2), value: []interface{}{12, 12}},
		{t: "decimal", operator: "count", excepted: uint(3), value: []interface{}{12, 1, 12}},
		{t: "decimal", operator: "max", excepted: 12, value: 12},
		{t: "decimal", operator: "max", excepted: 12, value: []interface{}{12}},
		{t: "decimal", operator: "max", excepted: 12, value: []interface{}{12, 1, 2, 3}},
		{t: "decimal", operator: "min", excepted: 12, value: 12},
		{t: "decimal", operator: "min", excepted: 12, value: []interface{}{12}},
		{t: "decimal", operator: "min", excepted: 12, value: []interface{}{12, 13, 14, 15}},
		{t: "decimal", operator: "sum", excepted: 12, value: 12},
		{t: "decimal", operator: "sum", excepted: float64(12), value: []interface{}{12}},
		{t: "decimal", operator: "sum", excepted: float64(12), value: []interface{}{1, 3, 5, 2, 1}},
		{t: "decimal", operator: "avg", excepted: 12, value: 12},
		{t: "decimal", operator: "avg", excepted: 12, value: []interface{}{12}},
		{t: "decimal", operator: "avg", excepted: float64(12), value: []interface{}{13, 11, 14, 10}},

		{t: "datetime", operator: "max", excepted: t12, value: t12},
		{t: "datetime", operator: "max", excepted: t12, value: []interface{}{t12}},
		{t: "datetime", operator: "max", excepted: t12, value: []interface{}{t12, t1, t2, t3}},
		{t: "datetime", operator: "min", excepted: t12, value: t12},
		{t: "datetime", operator: "min", excepted: t12, value: []interface{}{t12}},
		{t: "datetime", operator: "min", excepted: t12, value: []interface{}{t12, t13, t14, t15}},
	} {

		aggregateFactory, e := CreateAggregateFactory(test.t, test.operator)
		if nil != e {
			t.Error("[", idx, test.t, test.operator, "]", e)
			continue
		}

		var actual interface{}

		if values, ok := test.value.([]interface{}); ok {
			aggregate := aggregateFactory.Create(len(values))
			for _, v := range values {
				e := aggregate.Aggregate(nil, v)
				if nil != e {
					t.Error("[", idx, test.t, test.operator, "]", e)
					goto failed
				}
			}
			_, actual, e = aggregate.Result()
		} else {
			_, actual, e = aggregateFactory.AggregateOne(nil, test.value)
		}
	failed:
		if nil != e {
			t.Error("[", idx, test.t, test.operator, "]", e)
			continue
		}

		if fmt.Sprint(test.excepted) != fmt.Sprint(actual) {
			t.Error("[", idx, test.t, test.operator, "] excepted is", test.excepted, ", actual is", actual)
		}
	}
}
