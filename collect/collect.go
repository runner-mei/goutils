package collect

import (
	"encoding/json"
	"fmt"

	"github.com/runner-mei/check"
	"github.com/runner-mei/errors"
)

type NotExistStrategy int

const (
	SKIP_IF_NOT_EXIST NotExistStrategy = iota
	NIL_IF_NOT_EXIST
	ERROR_IF_NOT_EXIST
)

var (
	ErrStopped = errors.New("stopped")
	// 	//RecordNotMatch = errors.New("value is not match.")
	ErrValueNotObject = errors.New("value isn't a map[string]interface{}.")

	ErrValueIsEmpty = errors.New("value is empty.")
)

// type CollectFactory interface {
// 	Create(capacity int) Collector
// 	CollectOne(rows []map[string]interface{}, row map[string]interface{}, v interface{}) error
// }

// type Collector interface {
// 	Collect(row map[string]interface{}, v interface{}) error
// 	End() error
// }

// type fieldResolver struct {
// 	name string
// }

// func (self *fieldResolver) resolveOne(v map[string]interface{}, collector CollectFactory, strategy NotExistStrategy) error {
// 	field_value, ok := v[self.name]
// 	if !ok {
// 		switch strategy {
// 		case SKIP_IF_NOT_EXIST:
// 			return nil
// 		case NIL_IF_NOT_EXIST:
// 			collector.CollectOne(nil, v, nil)
// 			return nil
// 			//case ERROR_IF_NOT_EXIST:
// 			//	return errors.FieldNotExists(self.name)
// 		}
// 		return errors.FieldNotExists(self.name)
// 	}

// 	return collector.CollectOne(nil, v, field_value)
// }

// func (self *fieldResolver) resolve(index int, v map[string]interface{}, collector Collector, strategy NotExistStrategy) error {
// 	field_value, ok := v[self.name]
// 	if !ok {
// 		switch strategy {
// 		case SKIP_IF_NOT_EXIST:
// 			return nil
// 		case NIL_IF_NOT_EXIST:
// 			collector.Collect(v, nil)
// 			return nil
// 			//case ERROR_IF_NOT_EXIST:
// 			//	return errors.FieldNotExists(self.name)
// 		}
// 		return errors.FieldNotExists(self.name)
// 	}

// 	return collector.Collect(v, field_value)
// }

// type Filter struct {
// 	expressions []ExprSpec
// 	matchers    []Matcher
// 	index       int
// }

// func removeIndex(expressions []ExprSpec) (int, []ExprSpec, error) {
// 	index := -1
// 	offset := 0

// 	for i := 0; i < len(expressions); i++ {
// 		exp := expressions[i]
// 		if "_index" == exp.FieldName {
// 			if !IsEqualOp(exp.Operator) {
// 				return -1, nil, errors.New("create '_index " + exp.Operator + " " + fmt.Sprint(exp.Value) + "' failed, _index is only support '=='.")
// 			}

// 			idx, e := as.Int(exp.Value)
// 			if nil != e || idx < 0 {
// 				return -1, nil, errors.New("create '_index " + exp.Operator + " " + fmt.Sprint(exp.Value) + "' failed, operant isn't a integer and >= zero.")
// 			}

// 			if index >= 0 {
// 				return -1, nil, errors.New("create '_index " + exp.Operator + " " + fmt.Sprint(exp.Value) + "' failed, _index is repeated.")
// 			}
// 			index = int(idx)
// 			continue
// 		}

// 		if i != offset {
// 			expressions[offset] = expressions[i]
// 		}
// 		offset++
// 	}

// 	return index, expressions[:offset], nil
// }

// func (self *Filter) Init(expressions []ExprSpec) error {
// 	var e error
// 	self.index, expressions, e = removeIndex(expressions)
// 	if nil != e {
// 		return e
// 	}

// 	matchers := make([]Matcher, len(expressions))
// 	for i := 0; i < len(expressions); i++ {
// 		exp := expressions[i]
// 		f, e := MakeMatcher(exp.FieldType, exp.Operator, exp.Value)
// 		if nil != e {
// 			return errors.New("create '" + exp.FieldName + " " + exp.Operator + " " + fmt.Sprint(exp.Value) + "' failed, " + e.Error())
// 		}
// 		matchers[i] = f
// 	}
// 	self.expressions = expressions
// 	self.matchers = matchers
// 	return nil
// }

// func (self *Filter) Call(index int, res map[string]interface{}) (bool, error) {
// 	if self.index >= 0 && index != self.index {
// 		return false, nil
// 	}

// 	for i := 0; i < len(self.expressions); i++ {
// 		v, ok := res[self.expressions[i].FieldName]
// 		if !ok {
// 			return false, errors.FieldNotExists(self.expressions[i].FieldName)
// 		}
// 		if ok, e := self.matchers[i].Call(ok, v); !ok || nil != e {
// 			return ok, e
// 		}
// 	}
// 	return true, nil
// }

// type filterResolver struct {
// 	resolver resolverHelper
// 	filter   Filter
// }

// func (self *filterResolver) init(expressions []ExprSpec) error {
// 	return self.filter.Init(expressions)
// }

// func (self *filterResolver) resolveOne(v map[string]interface{}, collectFactory CollectFactory, strategy NotExistStrategy) error {
// 	matched, e := self.filter.Call(0, v)
// 	if nil != e {
// 		return e
// 	}
// 	if matched {
// 		return self.resolver.resolveOne(v, collectFactory, strategy)
// 	}
// 	return nil
// }

// func (self *filterResolver) resolve(index int, v map[string]interface{}, collector Collector, strategy NotExistStrategy) error {
// 	matched, e := self.filter.Call(index, v)
// 	if nil != e {
// 		return e
// 	}
// 	if matched {
// 		return self.resolver.resolve(index, v, collector, strategy)
// 	}
// 	return nil
// }

// type nullResolver struct {
// }

// func (self *nullResolver) resolveOne(v map[string]interface{}, collectFactory CollectFactory, strategy NotExistStrategy) error {
// 	return collectFactory.CollectOne(nil, v, v)
// }

// func (self *nullResolver) resolve(index int, v map[string]interface{}, collector Collector, strategy NotExistStrategy) error {
// 	return collector.Collect(v, v)
// }

// type resolverHelper interface {
// 	resolveOne(v map[string]interface{}, collectFactory CollectFactory, fieldNotExists NotExistStrategy) error
// 	resolve(index int, v map[string]interface{}, collector Collector, fieldNotExists NotExistStrategy) error
// }

// type ResolverImplWithNotAggregate struct {
// 	null       nullResolver
// 	field      fieldResolver
// 	filter     filterResolver
// 	resolver   resolverHelper
// 	resultType TypeMeta
// }

// func (self *ResolverImplWithNotAggregate) ResultType() TypeMeta {
// 	return self.resultType
// }

// func (self *ResolverImplWithNotAggregate) init(expressions []ExprSpec, fieldType, fieldName string) error {
// 	self.field.name = fieldName
// 	if 0 == len(expressions) {
// 		if "" == fieldName {
// 			self.resolver = &self.null
// 		} else {
// 			self.resolver = &self.field
// 		}
// 	} else {
// 		if "" == fieldName {
// 			self.filter.resolver = &self.null
// 		} else {
// 			self.filter.resolver = &self.field
// 		}
// 		if e := self.filter.init(expressions); nil != e {
// 			return e
// 		}
// 		self.resolver = &self.filter
// 	}

// 	return nil
// }

// func (self *ResolverImplWithNotAggregate) ResolveResult(v interface{}, collectFactory CollectFactory, skipIfTypeNotMatch bool, fieldNotExists NotExistStrategy) error {
// 	if m, ok := v.(map[string]interface{}); ok {
// 		return self.resolver.resolveOne(m, collectFactory, fieldNotExists)
// 	}
// 	var msg_bytes []byte

// 	switch values := v.(type) {
// 	case []map[string]interface{}:
// 		if 0 == len(values) {
// 			return nil
// 		}
// 		if 1 == len(values) {
// 			return self.resolver.resolveOne(values[0], collectFactory, fieldNotExists)
// 		}
// 		collector := collectFactory.Create(len(values))
// 		for idx, value := range values {
// 			if e := self.resolver.resolve(idx, value, collector, fieldNotExists); nil != e {
// 				return e
// 			}
// 		}
// 		return collector.End()
// 	case []interface{}:
// 		if 0 == len(values) {
// 			return nil
// 		}
// 		if 1 == len(values) {
// 			value, ok := values[0].(map[string]interface{})
// 			if !ok {
// 				if skipIfTypeNotMatch {
// 					return nil
// 				}
// 				return ErrValueNotObject
// 			}

// 			return self.resolver.resolveOne(value, collectFactory, fieldNotExists)
// 		}

// 		collector := collectFactory.Create(len(values))
// 		for idx, vv := range values {
// 			value, ok := vv.(map[string]interface{})
// 			if !ok {
// 				if skipIfTypeNotMatch {
// 					continue
// 				}
// 				return ErrValueNotObject
// 			}

// 			if e := self.resolver.resolve(idx, value, collector, fieldNotExists); nil != e {
// 				return e
// 			}
// 		}
// 		return collector.End()
// 	case *json.RawMessage:
// 		msg_bytes = *values
// 		goto to_bytes
// 	case json.RawMessage:
// 		msg_bytes = values
// 		goto to_bytes
// 	case RawMessages:
// 		msg_bytes = values.Bytes()
// 		goto to_bytes
// 	case *RawMessages:
// 		msg_bytes = values.Bytes()
// 		goto to_bytes
// 	}
// 	return fmt.Errorf("value must is map[string]interface{} or []map[string]interface{}, actual is `%T` %#v", v, v)
// to_bytes:
// 	if msg_bytes[0] == '[' {
// 		var results []map[string]interface{}
// 		if e := json.Unmarshal(msg_bytes, &results); nil != e {
// 			return errors.New("`" + string(msg_bytes) + "` is error - " + e.Error())
// 		}

// 		if 1 == len(results) {
// 			return self.resolver.resolveOne(results[0], collectFactory, fieldNotExists)
// 		}

// 		collector := collectFactory.Create(len(results))
// 		for idx, value := range results {
// 			if e := self.resolver.resolve(idx, value, collector, fieldNotExists); nil != e {
// 				return e
// 			}
// 		}
// 		return collector.End()
// 	} else if msg_bytes[0] == '{' {
// 		var result map[string]interface{}
// 		if e := json.Unmarshal(msg_bytes, &result); nil != e {
// 			return errors.New("`" + string(msg_bytes) + "` is error - " + e.Error())
// 		}
// 		return self.resolver.resolveOne(result, collectFactory, fieldNotExists)
// 	}
// 	return fmt.Errorf("value must is map[string]interface{} or []map[string]interface{} -- " + string(msg_bytes))
// }

// type ResolverImpl struct {
// 	notAggregate ResolverImplWithNotAggregate
// 	aggregation  AggregateFactory
// }

// func (self *ResolverImpl) ResultType() TypeMeta {
// 	return self.notAggregate.resultType
// }

// func (self *ResolverImpl) ResolveResult(v interface{}, collectFactory CollectFactory, skipIfTypeNotMatch bool, fieldNotExists NotExistStrategy) error {
// 	aggregation := AggregateCollectFactory{aggregation: self.aggregation, collectFactory: collectFactory}
// 	return self.notAggregate.ResolveResult(v, &aggregation, skipIfTypeNotMatch, fieldNotExists)
// }

// func (self *ResolverImpl) init(expressions []ExprSpec, fieldType, fieldName, function string) error {
// 	if "" == function {
// 		return errors.New("'func' is empty.")
// 	}

// 	e := self.notAggregate.init(expressions, fieldType, fieldName)
// 	if nil != e {
// 		return e
// 	}

// 	self.aggregation, e = CreateAggregateFactory(fieldType, function)
// 	return e
// }

// type AggregateCollector struct {
// 	aggregation    Aggregation
// 	collectFactory CollectFactory
// }

// func (self *AggregateCollector) Collect(row map[string]interface{}, v interface{}) error {
// 	return self.aggregation.Aggregate(row, v)
// }

// func (self *AggregateCollector) End() error {
// 	rows, v, e := self.aggregation.Result()
// 	if nil != e {
// 		return e
// 	}
// 	return self.collectFactory.CollectOne(rows, nil, v)
// }

// type AggregateCollectFactory struct {
// 	aggregation    AggregateFactory
// 	collectFactory CollectFactory
// }

// func (self *AggregateCollectFactory) Create(capacity int) Collector {
// 	return &AggregateCollector{aggregation: self.aggregation.Create(capacity),
// 		collectFactory: self.collectFactory}
// }

// func (self *AggregateCollectFactory) CollectOne(rows []map[string]interface{}, row map[string]interface{}, v interface{}) error {
// 	_, value, e := self.aggregation.AggregateOne(row, v)
// 	if nil != e {
// 		return e
// 	}
// 	return self.collectFactory.CollectOne(rows, row, value)
// }

// func InitResolver(impl *ResolverImpl, classType string, isArrayAfterFilterd bool, expressions []ExprSpec, fieldType, fieldName, function string) (Resolver, error) {
// 	if "" == function {
// 		if e := impl.notAggregate.init(expressions, fieldType, fieldName); nil != e {
// 			return nil, e
// 		}
// 		impl.notAggregate.resultType = typeForResolve(classType, isArrayAfterFilterd, fieldType, fieldName, function)
// 		return &impl.notAggregate, nil
// 	}
// 	if "" == fieldName && "count" != strings.ToLower(function) {
// 		return nil, errors.New("'Func' must is empty while 'FieldName' is empty.")
// 	}
// 	if e := impl.init(expressions, fieldType, fieldName, function); nil != e {
// 		return nil, e
// 	}
// 	impl.notAggregate.resultType = typeForResolve(classType, isArrayAfterFilterd, fieldType, fieldName, function)
// 	return impl, nil
// }

// func IsPrimary(classType string) bool {
// 	classType = strings.ToLower(classType)
// 	return "string" == classType ||
// 		"decimal" == classType ||
// 		"datetime" == classType ||
// 		"integer" == classType ||
// 		"ipaddress" == classType ||
// 		"physicaladdress" == classType ||
// 		"password" == classType ||
// 		"objectid" == classType ||
// 		"biginteger" == classType
// }

// func CreateResolver(classType string, isArrayAfterFilterd bool, expressions []ExprSpec, fieldType, fieldName, function string) (Resolver, error) {
// 	if 0 == len(expressions) && "" == fieldName && "" == function {
// 		return &NullResolver{typeMeta: TypeMeta{Type: strings.ToLower(classType)}}, nil
// 	}

// 	if IsPrimary(classType) && "value" == fieldName && 0 == len(expressions) && "" == function {
// 		return &NullResolver{typeMeta: TypeMeta{Type: strings.ToLower(classType)}}, nil
// 	}

// 	if "" == function {
// 		impl := &ResolverImplWithNotAggregate{}
// 		if e := impl.init(expressions, fieldType, fieldName); nil != e {
// 			return nil, e
// 		}
// 		impl.resultType = typeForResolve(classType, isArrayAfterFilterd, fieldType, fieldName, function)
// 		return impl, nil
// 	}
// 	if "" == fieldName && "count" != strings.ToLower(function) {
// 		return nil, errors.New("'Func' must is empty while 'FieldName' is empty.")
// 	}
// 	impl := &ResolverImpl{}
// 	if e := impl.init(expressions, fieldType, fieldName, function); nil != e {
// 		return nil, e
// 	}
// 	impl.notAggregate.resultType = typeForResolve(classType, isArrayAfterFilterd, fieldType, fieldName, function)
// 	return impl, nil
// }

// func typeForResolve(classType string, isArrayAfterFilterd bool, fieldType, fieldName, function string) TypeMeta {
// 	if "count" == function {
// 		return IntegerType
// 	}
// 	if "" != function {
// 		return TypeMeta{Type: fieldType}
// 	}
// 	if "" != fieldName {
// 		if !isArrayAfterFilterd {
// 			return TypeMeta{Type: fieldType, IsArray: false}
// 		}
// 		return TypeMeta{Type: fieldType, IsArray: isArrayAfterFilterd}
// 	}
// 	return TypeMeta{Type: classType, IsArray: isArrayAfterFilterd}
// }

// var typeMeta = TypeMeta{Type: "dynamic"}

// type NullResolver struct {
// 	typeMeta TypeMeta
// }

// func (self *NullResolver) ResultType() TypeMeta {
// 	return self.typeMeta
// }

// func (self *NullResolver) ResolveResult(v interface{}, collectFactory CollectFactory, skipIfTypeNotMatch bool, fieldNotExists NotExistStrategy) error {
// 	return collectFactory.CollectOne(nil, nil, v)
// }

type Location struct {
	Total         int
	OriginCurrent int
	Current       int
}

func ForEach(v interface{}, skipIfTypeNotMatch bool, cb func(Location, map[string]interface{}) error) error {
	if m, ok := v.(map[string]interface{}); ok {
		return cb(Location{Total: 1, OriginCurrent: 0, Current: 0}, m)
	}
	var msg_bytes []byte

	switch values := v.(type) {
	case []map[string]interface{}:
		if 0 == len(values) {
			return nil
		}
		for idx, value := range values {
			if e := cb(Location{Total: len(values), OriginCurrent: idx, Current: idx}, value); e != nil {
				return e
			}
		}
		return nil
	case []interface{}:
		if 0 == len(values) {
			return nil
		}

		for idx, vv := range values {
			value, ok := vv.(map[string]interface{})
			if !ok {
				if skipIfTypeNotMatch {
					continue
				}

				return ErrValueNotObject
			}

			if e := cb(Location{Total: len(values), OriginCurrent: idx, Current: idx}, value); e != nil {
				return e
			}
		}
		return nil
	case *json.RawMessage:
		msg_bytes = *values
		goto to_bytes
	case json.RawMessage:
		msg_bytes = values
		goto to_bytes
	default:
		bs, ok := values.(interface {
			Bytes() []byte
		})
		if ok {
			msg_bytes = bs.Bytes()
			goto to_bytes
		}
		// case RawMessages:
		// 	msg_bytes = values.Bytes()
		// 	goto to_bytes
		// case *RawMessages:
		// 	msg_bytes = values.Bytes()
		// 	goto to_bytes
	}
	return fmt.Errorf("value must is map[string]interface{} or []map[string]interface{}, actual is `%T` %#v", v, v)
to_bytes:
	if msg_bytes[0] == '[' {
		var results []map[string]interface{}
		if e := json.Unmarshal(msg_bytes, &results); nil != e {
			return errors.New("`" + string(msg_bytes) + "` is error - " + e.Error())
		}

		for idx, value := range results {
			if e := cb(Location{Total: len(results), OriginCurrent: idx, Current: idx}, value); nil != e {
				return e
			}
		}
		return nil
	} else if msg_bytes[0] == '{' {
		var result map[string]interface{}
		if e := json.Unmarshal(msg_bytes, &result); nil != e {
			return errors.New("`" + string(msg_bytes) + "` is error - " + e.Error())
		}
		return cb(Location{Total: 1, OriginCurrent: 0, Current: 0}, result)
	}
	return fmt.Errorf("value must is map[string]interface{} or []map[string]interface{} -- " + string(msg_bytes))
}

type filterItem struct {
	field   string
	isIndex bool
	checker check.Checker
}

func (fi *filterItem) filter(notExist NotExistStrategy, index int, value map[string]interface{}) (bool, error) {
	if fi.isIndex {
		return fi.checker.Check(index)
	}
	fieldValue, ok := value[fi.field]
	if !ok {
		switch notExist {
		case SKIP_IF_NOT_EXIST:
			return false, nil
		case NIL_IF_NOT_EXIST:
			return fi.checker.Check(nil)
			//case ERROR_IF_NOT_EXIST:
			//	return errors.FieldNotExists(self.name)
		}
		return false, errors.FieldNotExists(fi.field)
	}
	return fi.checker.Check(fieldValue)
}

type Filter struct {
	err   error
	items []filterItem
}

func (filter *Filter) Add(field, typ, op string, value interface{}) *Filter {
	if filter.err != nil {
		return filter
	}

	checker, err := check.MakeChecker(typ, op, value)
	if err != nil {
		filter.err = err
		return filter
	}
	filter.items = append(filter.items, filterItem{
		field:   field,
		isIndex: field == "_index",
		checker: checker,
	})
	return filter
}

func (filter *Filter) Filter(notExist NotExistStrategy, index int, value map[string]interface{}) (bool, error) {
	if filter.err != nil {
		return false, filter.err
	}

	for idx := range filter.items {
		ok, err := filter.items[idx].filter(notExist, index, value)
		if err != nil {
			return ok, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func Select(v interface{}, filter *Filter, skipIfTypeNotMatch bool, notExist NotExistStrategy, resultCb func(Location, map[string]interface{}) error) error {
	count := 0
	return ForEach(v, skipIfTypeNotMatch, func(idx Location, value map[string]interface{}) error {
		ok, err := filter.Filter(notExist, idx.Current, value)
		if err != nil {
			return err
		}
		if ok {
			idx.Current = count
			err = resultCb(idx, value)
			if err != nil {
				return err
			}
			count++
		}
		return nil
	})
}

func Field(name string, notExist NotExistStrategy) func(idx Location, value map[string]interface{}) (interface{}, error) {
	if name == "_index" {
		return func(idx Location, value map[string]interface{}) (interface{}, error) {
			return idx, nil
		}
	}
	return func(idx Location, value map[string]interface{}) (interface{}, error) {
		fieldValue, ok := value[name]
		if !ok {
			switch notExist {
			case SKIP_IF_NOT_EXIST:
				return false, nil
			case NIL_IF_NOT_EXIST:
				return nil, nil
				//case ERROR_IF_NOT_EXIST:
				//	return errors.FieldNotExists(self.name)
			}
			return nil, errors.FieldNotExists(name)
		}
		return fieldValue, nil
	}
}

func Map(mapping func(Location, map[string]interface{}) (interface{}, error), result func(Location, map[string]interface{}, interface{}) error) func(Location, map[string]interface{}) error {
	return func(idx Location, value map[string]interface{}) error {
		newValue, err := mapping(idx, value)
		if err != nil {
			return err
		}
		return result(idx, value, newValue)
	}
}

func CallCount(count *int, cb func(Location, map[string]interface{}) error) func(Location, map[string]interface{}) error {
	return func(idx Location, value map[string]interface{}) error {
		*count++
		return cb(idx, value)
	}
}

func Count(count *int) func(Location, map[string]interface{}) error {
	return func(idx Location, value map[string]interface{}) error {
		*count++
		return nil
	}
}

func FirstObject(result func(map[string]interface{})) func(Location, map[string]interface{}) error {
	isFirst := true
	return func(idx Location, value map[string]interface{}) error {
		if isFirst {
			result(value)
			isFirst = false
		}
		return nil
	}
}

func First(result func(value interface{})) func(idx Location, value map[string]interface{}, fieldValue interface{}) error {
	isFirst := true
	return func(idx Location, value map[string]interface{}, fieldValue interface{}) error {
		if isFirst {
			result(value)
			isFirst = false
		}
		return nil
	}
}

func Aggregate(agg Aggregation) func(idx Location, value map[string]interface{}, fieldValue interface{}) error {
	return func(idx Location, value map[string]interface{}, fieldValue interface{}) error {
		return agg.Aggregate(value, fieldValue)
	}
}
