package collect

import (
	"fmt"
	"reflect"
	"testing"
)

type ExprSpec struct {
	FieldName string      `json:"attribute"`
	FieldType string      `json:"attribute_type,omitempty"`
	Operator  string      `json:"operator"`
	Value     interface{} `json:"value"`
}

type SubscribeSpecTest struct {
	Name string `json:"name"`
	// Paths               []ds_PathValue     `json:"paths"`
	// Arguments           []ArgumentValue `json:"arguments,omitempty"`
	Expressions         []ExprSpec `json:"expressions,omitempty"`
	FieldName           string     `json:"attribute_name"`
	FieldType           string     `json:"attribute_type,omitempty"`
	IsArrayAfterFilterd bool       `json:"is_array_after_filter"`
	Func                string     `json:"function"`
	// Class               TypeMeta        `json:"class"`
}

// type TestCollectFactory interface {
// 	CollectFactory
// 	Result() interface{}
// 	CallCount() uint
// }

func M(nm string, v interface{}) map[string]interface{} {
	return map[string]interface{}{nm: v}
}

func TestCollectOk(t *testing.T) {
	for idx, test := range []struct {
		spec      SubscribeSpecTest
		arguments interface{}
		// collectFactory      Aggregation
		excepted_result     interface{}
		excepted_call_count uint
	}{{spec: SubscribeSpecTest{FieldName: "a", FieldType: "integer"},
		arguments: map[string]interface{}{"a": 12},
		// collectFactory:      &CollectOne{},
		excepted_result:     12,
		excepted_call_count: 1},

		{spec: SubscribeSpecTest{FieldName: "a", FieldType: "integer"},
			arguments: []interface{}{M("a", 12)},
			// collectFactory:      &CollectOne{},
			excepted_result:     12,
			excepted_call_count: 1},

		{spec: SubscribeSpecTest{FieldName: "a", FieldType: "integer"},
			arguments: []map[string]interface{}{M("a", 12)},
			// collectFactory:      &CollectOne{},
			excepted_result:     12,
			excepted_call_count: 1},

		{spec: SubscribeSpecTest{FieldName: "a", FieldType: "integer", Func: "last"},
			arguments: map[string]interface{}{"a": 12},
			// collectFactory:      &CollectOne{},
			excepted_result:     12,
			excepted_call_count: 1},

		{spec: SubscribeSpecTest{FieldName: "a", FieldType: "string"},
			arguments: map[string]interface{}{"a": "abc"},
			// collectFactory:      &CollectOne{},
			excepted_result:     "abc",
			excepted_call_count: 1},

		{spec: SubscribeSpecTest{FieldName: "a", FieldType: "integer", Func: "first"},
			arguments: []interface{}{M("a", 12), 12, M("a", 1), M("a", 1)},
			// collectFactory:      &CollectOne{},
			excepted_result:     12,
			excepted_call_count: 1},

		{spec: SubscribeSpecTest{FieldName: "a", FieldType: "integer", Func: "first"},
			arguments: []map[string]interface{}{M("a", 12), M("a", 1), M("a", 1)},
			// collectFactory:      &CollectOne{},
			excepted_result:     12,
			excepted_call_count: 1},

		{spec: SubscribeSpecTest{FieldName: "a", FieldType: "string", Func: "count"},
			arguments: []interface{}{M("a", 12), 12, M("a", 1), M("a", 1)},
			// collectFactory:      &CollectOne{},
			excepted_result:     uint(3),
			excepted_call_count: 1},

		{spec: SubscribeSpecTest{FieldName: "a", FieldType: "string", Func: "count"},
			arguments: []map[string]interface{}{M("a", 12), M("a", 1), M("a", 1)},
			// collectFactory:      &CollectOne{},
			excepted_result:     uint(3),
			excepted_call_count: 1},

		{spec: SubscribeSpecTest{FieldName: "a", FieldType: "integer"},
			arguments: []interface{}{M("a", 12), 12, M("a", 1), M("a", 1)},
			// collectFactory:      &CollectAll{},
			excepted_result:     []interface{}{12, 1, 1},
			excepted_call_count: 3},

		{spec: SubscribeSpecTest{FieldName: "a", FieldType: "integer"},
			arguments: []map[string]interface{}{M("a", 12), M("a", 1), M("a", 1)},
			// collectFactory:      &CollectAll{},
			excepted_result:     []interface{}{12, 1, 1},
			excepted_call_count: 3},

		{spec: SubscribeSpecTest{FieldName: "a",
			FieldType: "integer",
			Expressions: []ExprSpec{
				ExprSpec{FieldName: "a",
					FieldType: "integer",
					Operator:  ">=",
					Value:     "9"}}},
			arguments: []map[string]interface{}{M("a", 12), M("a", 1), M("a", 10)},
			// collectFactory:      &CollectAll{},
			excepted_result:     []interface{}{12, 10},
			excepted_call_count: 2},

		{spec: SubscribeSpecTest{FieldName: "a",
			FieldType: "integer",
			Expressions: []ExprSpec{
				ExprSpec{FieldName: "a",
					FieldType: "integer",
					Operator:  ">=",
					Value:     "9"},
				ExprSpec{FieldName: "a",
					FieldType: "integer",
					Operator:  "==",
					Value:     "12"}}},
			arguments: []map[string]interface{}{M("a", 12), M("a", 1), M("a", 10)},
			// collectFactory:      &CollectAll{},
			excepted_result:     []interface{}{12},
			excepted_call_count: 1},

		{spec: SubscribeSpecTest{FieldName: "a",
			FieldType: "integer",
			Expressions: []ExprSpec{
				ExprSpec{FieldName: "a",
					FieldType: "integer",
					Operator:  ">=",
					Value:     "9"}}},
			arguments: []interface{}{M("a", 12), 12, M("a", 1), M("a", 10)},
			// collectFactory:      &CollectAll{},
			excepted_result:     []interface{}{12, 10},
			excepted_call_count: 2},

		{spec: SubscribeSpecTest{FieldName: "a",
			FieldType: "integer",
			Expressions: []ExprSpec{
				ExprSpec{FieldName: "a",
					FieldType: "integer",
					Operator:  ">=",
					Value:     "9"},
				ExprSpec{FieldName: "a",
					FieldType: "integer",
					Operator:  "==",
					Value:     "12"}}},
			arguments: []interface{}{M("a", 12), 12, M("a", 1), M("a", 10)},
			// collectFactory:      &CollectAll{},
			excepted_result:     []interface{}{12},
			excepted_call_count: 1},

		{spec: SubscribeSpecTest{FieldName: "a",
			FieldType: "integer",
			Func:      "count",
			Expressions: []ExprSpec{
				ExprSpec{FieldName: "a",
					FieldType: "integer",
					Operator:  ">=",
					Value:     "9"}}},
			arguments: []interface{}{M("a", 12), 12, M("a", 1), M("a", 10)},
			// collectFactory:      &CollectOne{},
			excepted_result:     uint(2),
			excepted_call_count: 1},

		{spec: SubscribeSpecTest{FieldName: "a",
			FieldType: "integer",
			Func:      "count",
			Expressions: []ExprSpec{
				ExprSpec{FieldName: "a",
					FieldType: "integer",
					Operator:  ">=",
					Value:     "9"},
				ExprSpec{FieldName: "a",
					FieldType: "integer",
					Operator:  "==",
					Value:     "12"}}},
			arguments:           []interface{}{M("a", 12), 12, M("a", 1), M("a", 10)},
			excepted_result:     uint(1),
			excepted_call_count: 1},

		{spec: SubscribeSpecTest{FieldName: "a",
			FieldType: "integer",
			Func:      "count",
			Expressions: []ExprSpec{
				ExprSpec{FieldName: "a",
					Operator: "nin",
					Value:    []interface{}{"12"}}}},
			arguments:           []interface{}{M("a", 12), 12, M("a", 1), M("a", 10)},
			excepted_result:     uint(2),
			excepted_call_count: 1},

		{spec: SubscribeSpecTest{FieldName: "a",
			FieldType: "integer",
			Func:      "count",
			Expressions: []ExprSpec{
				ExprSpec{FieldName: "a",
					Operator: "nin",
					Value:    []interface{}{"12", 1}}}},
			arguments:           []interface{}{M("a", 12), 12, M("a", 1), M("a", 10)},
			excepted_result:     uint(1),
			excepted_call_count: 1},
	} {
		//t.Log("[", idx, fmt.Sprintf("%#v", test.spec), "]")

		filter := &Filter{}
		for _, expr := range test.spec.Expressions {
			filter.Add(expr.FieldName, expr.FieldType, expr.Operator, expr.Value)
		}

		var agg Aggregation
		if test.spec.Func == "" {
			agg = &CollectAll{}
		} else {
			f, e := CreateAggregateFactory(test.spec.FieldType, test.spec.Func)
			if nil != e {
				t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "]", e)
				continue
			}
			agg = f.Create(4)
		}
		e := Select(test.arguments, filter, true, SKIP_IF_NOT_EXIST, Map(Field(test.spec.FieldName, SKIP_IF_NOT_EXIST), Aggregate(agg)))
		if nil != e {
			t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "]", e)
			continue
		}

		//actual_call_count := test. collectFactory.CallCount()
		_, actual_result, e := agg.Result()
		if nil != e {
			t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "]", e)
			continue
		}

		// if test.excepted_call_count != uint(actual_call_count) {
		// 	t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "] excepted call is", test.excepted_call_count, ", actual call is", actual_call_count)
		// }

		_, ok := test.excepted_result.([]interface{})
		if !ok {
			a, ok := actual_result.([]interface{})
			if ok && len(a) == 1 {
				actual_result = a[0]
			}
		}

		if !reflect.DeepEqual(test.excepted_result, actual_result) {
			t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "] excepted is", test.excepted_result, ", actual is", actual_result)
		}

		// var impl ResolverImpl
		// resolver, e := InitResolver(&impl, "", true, test.spec.Expressions, test.spec.FieldType, test.spec.FieldName, test.spec.Func)
		// if nil != e {
		// 	t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "]", e)
		// 	continue
		// }

		// if e = resolver.ResolveResult(test.arguments, test.collectFactory, true, SKIP_IF_NOT_EXIST); nil != e {
		// 	t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "]", e)
		// 	continue
		// }

		// actual_call_count := test.collectFactory.CallCount()
		// actual_result := test.collectFactory.Result()

		// if test.excepted_call_count != actual_call_count {
		// 	t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "] excepted call is", test.excepted_call_count, ", actual call is", actual_call_count)
		// }

		// if !reflect.DeepEqual(test.excepted_result, actual_result) {
		// 	t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "] excepted is", test.excepted_result, ", actual is", actual_result)
		// }
	}
}

// type CollectFactoryOneError struct {
// }

// func (self *CollectFactoryOneError) Create(capacity int) Collector {
// 	panic("notImplented")
// }

// func (self *CollectFactoryOneError) CollectOne(rows []map[string]interface{}, row map[string]interface{}, v interface{}) error {
// 	return errors.New("CollectFactoryOneError")
// }

// type CollectFactoryAllError struct {
// 	collectError error
// 	endError     error
// }

// type CollectAllError struct {
// 	collectError error
// 	endError     error
// }

// func (self *CollectAllError) Collect(row map[string]interface{}, v interface{}) error {
// 	return self.collectError
// }

// func (self *CollectAllError) End() error {
// 	return self.endError
// }

// func (self *CollectFactoryAllError) Create(capacity int) Collector {
// 	return &CollectAllError{collectError: self.collectError, endError: self.endError}
// }

// func (self *CollectFactoryAllError) CollectOne(rows []map[string]interface{}, row map[string]interface{}, v interface{}) error {
// 	panic("notImplented")
// }

// func TestCollectError(t *testing.T) {
// 	for idx, test := range []struct {
// 		spec           SubscribeSpecTest
// 		arguments      interface{}
// 		excepted_error string
// 		// collectFactory CollectFactory
// 	}{{spec: SubscribeSpecTest{FieldName: "a", FieldType: "integer"},
// 		arguments: map[string]interface{}{"a": 12},
// 		// collectFactory: &CollectFactoryOneError{},
// 		excepted_error: "CollectFactoryOneError"},

// 		{spec: SubscribeSpecTest{FieldName: "a", FieldType: "integer"},
// 			arguments: []interface{}{map[string]interface{}{"a": 12}},
// 			// collectFactory: &CollectFactoryOneError{},
// 			excepted_error: "CollectFactoryOneError"},

// 		{spec: SubscribeSpecTest{FieldName: "a", FieldType: "integer"},
// 			arguments: []map[string]interface{}{map[string]interface{}{"a": 12}},
// 			// collectFactory: &CollectFactoryOneError{},
// 			excepted_error: "CollectFactoryOneError"},

// 		{spec: SubscribeSpecTest{FieldName: "a", FieldType: "integer", Func: "count"},
// 			arguments: map[string]interface{}{"a": 12},
// 			// collectFactory: &CollectFactoryOneError{},
// 			excepted_error: "CollectFactoryOneError"},

// 		{spec: SubscribeSpecTest{FieldName: "a", FieldType: "integer", Func: "count"},
// 			arguments: []interface{}{map[string]interface{}{"a": 12}},
// 			// collectFactory: &CollectFactoryOneError{},
// 			excepted_error: "CollectFactoryOneError"},

// 		{spec: SubscribeSpecTest{FieldName: "a", FieldType: "integer", Func: "count"},
// 			arguments: []map[string]interface{}{map[string]interface{}{"a": 12}},
// 			// collectFactory: &CollectFactoryOneError{},
// 			excepted_error: "CollectFactoryOneError"},

// 		{spec: SubscribeSpecTest{FieldName: "a",
// 			FieldType: "integer",
// 			Func:      "count",
// 			Expressions: []ExprSpec{
// 				ExprSpec{FieldName: "a",
// 					FieldType: "integer",
// 					Operator:  ">=",
// 					Value:     "9"}}},
// 			arguments: []interface{}{M("a", 12), M("a", 1), M("a", 10)},
// 			// collectFactory: &CollectFactoryOneError{},
// 			excepted_error: "CollectFactoryOneError"},

// 		{spec: SubscribeSpecTest{FieldName: "a",
// 			FieldType: "integer"},
// 			arguments: []interface{}{M("a", 12), M("a", 1), M("a", 10)},
// 			// collectFactory: &CollectFactoryAllError{collectError: errors.New("collectError")},
// 			excepted_error: "collectError"},

// 		{spec: SubscribeSpecTest{FieldName: "a",
// 			FieldType: "integer"},
// 			arguments: []map[string]interface{}{M("a", 12), M("a", 1), M("a", 10)},
// 			// collectFactory: &CollectFactoryAllError{collectError: errors.New("collectError")},
// 			excepted_error: "collectError"},

// 		{spec: SubscribeSpecTest{FieldName: "a",
// 			FieldType: "integer",
// 			Expressions: []ExprSpec{
// 				ExprSpec{FieldName: "a",
// 					FieldType: "integer",
// 					Operator:  ">=",
// 					Value:     "9"}}},
// 			arguments: []interface{}{M("a", 12), M("a", 1), M("a", 10)},
// 			// collectFactory: &CollectFactoryAllError{collectError: errors.New("collectError")},
// 			excepted_error: "collectError"},

// 		{spec: SubscribeSpecTest{FieldName: "a",
// 			FieldType: "integer"},
// 			arguments: []interface{}{M("a", 12), M("a", 1), M("a", 10)},
// 			// collectFactory: &CollectFactoryAllError{endError: errors.New("endError")},
// 			excepted_error: "endError"},

// 		{spec: SubscribeSpecTest{FieldName: "a",
// 			FieldType: "integer"},
// 			arguments: []map[string]interface{}{M("a", 12), M("a", 1), M("a", 10)},
// 			// collectFactory: &CollectFactoryAllError{endError: errors.New("endError")},
// 			excepted_error: "endError"},

// 		{spec: SubscribeSpecTest{FieldName: "a",
// 			FieldType: "integer",
// 			Expressions: []ExprSpec{
// 				ExprSpec{FieldName: "a",
// 					FieldType: "integer",
// 					Operator:  ">=",
// 					Value:     "9"}}},
// 			arguments: []interface{}{M("a", 12), M("a", 1), M("a", 10)},
// 			// collectFactory: &CollectFactoryAllError{endError: errors.New("endError")},
// 			excepted_error: "endError"},
// 	} {
// 		//t.Log("[", idx, fmt.Sprintf("%#v", test.spec), "]")

// 		filter := &Filter{}
// 		for _, expr := range test.spec.Expressions {
// 			filter.Add(expr.FieldName, expr.FieldType, expr.Operator, expr.Value)
// 		}

// 		var agg Aggregation
// 		if test.spec.Func == "" {
// 			agg = &CollectAll{}
// 		} else {
// 			f, e := CreateAggregateFactory(test.spec.FieldType, test.spec.Func)
// 			if nil != e {
// 				t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "]", e)
// 				continue
// 			}
// 			agg = f.Create(4)
// 		}
// 		e := Select(test.arguments, filter, true, SKIP_IF_NOT_EXIST, Map(Field(test.spec.FieldName, SKIP_IF_NOT_EXIST), Aggregate(agg)))
// 		if nil != e {
// 			// 	t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "]", e)
// 			// 	continue
// 			// }

// 			if !strings.Contains(e.Error(), test.excepted_error) {
// 				t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "]", e)
// 			}
// 			continue
// 		}
// 		t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "] excepted error is", test.excepted_error, ", actual is ok.")

// 		//actual_call_count := test. collectFactory.CallCount()
// 		// _, actual_result, e := agg.Result()
// 		// if nil != e {
// 		// 	t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "]", e)
// 		// 	continue
// 		// }

// 		// if test.excepted_call_count != uint(actual_call_count) {
// 		// 	t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "] excepted call is", test.excepted_call_count, ", actual call is", actual_call_count)
// 		// }

// 		// _, ok := test.excepted_result.([]interface{})
// 		// if !ok {
// 		// 	a, ok := actual_result.([]interface{})
// 		// 	if ok && len(a) == 1 {
// 		// 		actual_result = a[0]
// 		// 	}
// 		// }

// 		// if !reflect.DeepEqual(test.excepted_result, actual_result) {
// 		// 	t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "] excepted is", test.excepted_result, ", actual is", actual_result)
// 		// }

// 		// filter := &Filter{}
// 		// for _, expr := range test.spec.Expressions {
// 		// 	filter.Add(expr.FieldName, expr.FieldType, expr.Operator, expr.Value)
// 		// }

// 		// var impl ResolverImpl
// 		// resolver, e := InitResolver(&impl, "", true, test.spec.Expressions, test.spec.FieldType, test.spec.FieldName, test.spec.Func)
// 		// if nil != e {
// 		// 	if !strings.Contains(e.Error(), test.excepted_error) {
// 		// 		t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "]", e)
// 		// 	}
// 		// 	continue
// 		// }

// 		// if e = resolver.ResolveResult(test.arguments, test.collectFactory, true, SKIP_IF_NOT_EXIST); nil != e {
// 		// 	if !strings.Contains(e.Error(), test.excepted_error) {
// 		// 		t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "]", e)
// 		// 	}
// 		// 	continue
// 		// }
// 		// t.Error("[", idx, fmt.Sprintf("%#v", test.spec), "] excepted error is", test.excepted_error, ", actual is ok.")
// 	}
// }

// func TestCollectTypeError(t *testing.T) {
// 	var impl ResolverImpl
// 	collectFactory := &CollectFactoryAllError{collectError: errors.New("collectError")}
// 	resolver, e := InitResolver(&impl, "", true, nil, "integer", "a", "count")
// 	if nil != e {
// 		t.Error(e)
// 		return
// 	}

// 	if e = resolver.ResolveResult([]interface{}{12}, collectFactory, false, SKIP_IF_NOT_EXIST); nil != e {
// 		if !strings.Contains(e.Error(), ErrValueNotObject.Error()) {
// 			t.Error(e)
// 		}
// 		return
// 	}
// 	t.Error(" excepted error is", ErrValueNotObject, ", actual is ok.")
// }

// func TestCollectNotExistError(t *testing.T) {
// 	var impl ResolverImpl
// 	collectFactory := &CollectFactoryAllError{collectError: errors.New("collectError")}
// 	resolver, e := InitResolver(&impl, "", true, nil, "integer", "a", "count")
// 	if nil != e {
// 		t.Error(e)
// 		return
// 	}

// 	if e = resolver.ResolveResult([]interface{}{M("b", 12)}, collectFactory, false, ERROR_IF_NOT_EXIST); nil != e {
// 		if !errors.IsFieldNotExists(e) {
// 			t.Error(e)
// 		}
// 		return
// 	}
// 	t.Error(" excepted error contains 'is not exists', actual is ok.")
// }

func TestFilter(t *testing.T) {
	for _, test := range []struct {
		expressions     []ExprSpec
		input_index     int
		input_value     map[string]interface{}
		excepted_result bool
	}{
		{expressions: []ExprSpec{
			ExprSpec{FieldName: "_index",
				FieldType: "integer",
				Operator:  "==",
				Value:     "3"}},
			input_index:     3,
			input_value:     nil,
			excepted_result: true},

		{expressions: []ExprSpec{
			ExprSpec{FieldName: "_index",
				FieldType: "integer",
				Operator:  "==",
				Value:     "3"}},
			input_index:     2,
			input_value:     nil,
			excepted_result: false},

		{expressions: []ExprSpec{
			ExprSpec{FieldName: "a",
				FieldType: "integer",
				Operator:  "==",
				Value:     "3"}},
			input_index:     2,
			input_value:     M("a", 3),
			excepted_result: true},

		{expressions: []ExprSpec{
			ExprSpec{FieldName: "a",
				FieldType: "integer",
				Operator:  "==",
				Value:     "3"},
			ExprSpec{FieldName: "_index",
				FieldType: "integer",
				Operator:  "==",
				Value:     "3"}},
			input_index:     3,
			input_value:     M("a", 3),
			excepted_result: true},

		{expressions: []ExprSpec{
			ExprSpec{FieldName: "a",
				FieldType: "integer",
				Operator:  "==",
				Value:     "3"},
			ExprSpec{FieldName: "b",
				FieldType: "integer",
				Operator:  "==",
				Value:     "4"}},
			input_index:     2,
			input_value:     map[string]interface{}{"a": 3, "b": 4},
			excepted_result: true},
	} {

		filter := &Filter{}
		for _, expr := range test.expressions {
			filter.Add(expr.FieldName, expr.FieldType, expr.Operator, expr.Value)
		}

		actual, e := filter.Filter(ERROR_IF_NOT_EXIST, test.input_index, test.input_value)
		if nil != e {
			t.Error(e)
			continue
		}
		if actual != test.excepted_result {
			t.Errorf("[%#v] excepted is %v, actual is %v", test, test.excepted_result, actual)
		}
	}
}
