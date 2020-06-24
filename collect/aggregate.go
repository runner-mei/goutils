package collect

import (
	"errors"
	"math"
	"time"

	"github.com/runner-mei/goutils/as"
)

type AggregateFactory interface {
	Create(capacity int) Aggregation
	AggregateOne(row map[string]interface{}, v interface{}) (map[string]interface{}, interface{}, error)
}

type Aggregation interface {
	Aggregate(row map[string]interface{}, v interface{}) error
	Result() ([]map[string]interface{}, interface{}, error)
}

func CreateAggregateFactory(t, operator string) (AggregateFactory, error) {
	if "first" == operator {
		return &firstAggregation{}, nil
	} else if "last" == operator {
		return &lastAggregation{}, nil
	} else if "count" == operator {
		return &countAggregation{}, nil
	}

	switch t {
	case "boolean":
		return nil, errors.New("'" + operator + "' is not supported for the boolean.")
	case "integer", "biginteger":
		switch operator {
		case "max":
			return &maxAggregation{}, nil
		case "min":
			return &minAggregation{}, nil
		case "sum":
			return &sumIntAggregation{}, nil
		case "avg":
			return &avgIntAggregation{}, nil
		default:
			return nil, errors.New("'" + operator + "' is not supported for the " + t + ".")
		}
	case "datetime":
		switch operator {
		case "max":
			return &maxDateAggregation{}, nil
		case "min":
			return &minDateAggregation{}, nil
		default:
			return nil, errors.New("'" + operator + "' is not supported for the " + t + ".")
		}
	case "decimal", "dynamic", "":
		switch operator {
		case "max":
			return &maxAggregation{}, nil
		case "min":
			return &minAggregation{}, nil
		case "sum":
			return &sumFloatAggregation{}, nil
		case "avg":
			return &avgFloatAggregation{}, nil
		default:
			return nil, errors.New("'" + operator + "' is not supported for the " + t + ".")
		}
	case "ipAddress":
		return nil, errors.New("'" + operator + "' is not supported for the ipAddress.")
	case "physicalAddress":
		return nil, errors.New("'" + operator + "' is not supported for the physicalAddress.")
	case "password":
		return nil, errors.New("'" + operator + "' is not supported for the string.")
	case "objectId":
		return nil, errors.New("'" + operator + "' is not supported for the objectId.")
	case "string":
		return nil, errors.New("'" + operator + "' is not supported for the string.")
	default:
		return nil, errors.New("'" + t + "' is unknown type for aggregated.")
	}
}

type CollectOne struct {
	count uint
	value interface{}
	row   map[string]interface{}
}

func (co *CollectOne) Create(capacity int) Aggregation {
	return &CollectOne{}
}

func (co *CollectOne) AggregateOne(row map[string]interface{}, v interface{}) (map[string]interface{}, interface{}, error) {
	return row, v, nil
}

func (co *CollectOne) Aggregate(row map[string]interface{}, v interface{}) error {
	if 0 != co.count {
		return errors.New("method is already call.")
	}

	co.count++
	co.value = v
	co.row = row
	return nil
}

func (co *CollectOne) Result() ([]map[string]interface{}, interface{}, error) {
	return []map[string]interface{}{co.row}, co.value, nil
}

type CollectAll struct {
	Values []interface{}
	Rows   []map[string]interface{}
}

func (co *CollectAll) Create(capacity int) Aggregation {
	return &CollectAll{}
}

func (co *CollectAll) AggregateOne(row map[string]interface{}, v interface{}) (map[string]interface{}, interface{}, error) {
	return row, v, nil
}

func (co *CollectAll) Aggregate(row map[string]interface{}, v interface{}) error {
	co.Values = append(co.Values, v)
	co.Rows = append(co.Rows, row)
	return nil
}

func (co *CollectAll) Result() ([]map[string]interface{}, interface{}, error) {
	return co.Rows, co.Values, nil
}

type firstAggregation struct {
	firstRow map[string]interface{}
	first    interface{}
}

func (self *firstAggregation) Create(capacity int) Aggregation {
	return &firstAggregation{}
}

func (self *firstAggregation) AggregateOne(row map[string]interface{}, v interface{}) (map[string]interface{}, interface{}, error) {
	return row, v, nil
}

func (self *firstAggregation) Aggregate(row map[string]interface{}, v interface{}) error {
	if nil == self.first {
		self.first = v
		self.firstRow = row
	}
	return nil
}

func (self *firstAggregation) Result() ([]map[string]interface{}, interface{}, error) {
	return []map[string]interface{}{self.firstRow}, self.first, nil
}

type lastAggregation struct {
	lastRow map[string]interface{}
	last    interface{}
}

func (self *lastAggregation) Create(capacity int) Aggregation {
	return &lastAggregation{}
}

func (self *lastAggregation) AggregateOne(row map[string]interface{}, v interface{}) (map[string]interface{}, interface{}, error) {
	return row, v, nil
}

func (self *lastAggregation) Aggregate(row map[string]interface{}, v interface{}) error {
	self.last = v
	self.lastRow = row
	return nil
}

func (self *lastAggregation) Result() ([]map[string]interface{}, interface{}, error) {
	return []map[string]interface{}{self.lastRow}, self.last, nil
}

type avgFloatAggregation struct {
	rows []map[string]interface{}
	last interface{}
	sum  float64
}

func (self *avgFloatAggregation) Create(capacity int) Aggregation {
	return &avgFloatAggregation{}
}

func (self *avgFloatAggregation) AggregateOne(row map[string]interface{}, v interface{}) (map[string]interface{}, interface{}, error) {
	return row, v, nil
}

func (self *avgFloatAggregation) Aggregate(row map[string]interface{}, v interface{}) error {
	value, e := as.Float64(v)
	if nil != e {
		return e
	}

	self.rows = append(self.rows, row)
	self.sum += value
	self.last = value
	return nil
}

func (self *avgFloatAggregation) Result() ([]map[string]interface{}, interface{}, error) {
	count := len(self.rows)
	if 0 == count {
		return nil, nil, ErrValueIsEmpty
	}
	if 1 == count {
		return self.rows, self.last, nil
	}
	return self.rows, self.sum / float64(count), nil
}

type avgIntAggregation struct {
	avgFloatAggregation
}

func (self *avgIntAggregation) Create(capacity int) Aggregation {
	return &avgIntAggregation{}
}

func (self *avgIntAggregation) Result() ([]map[string]interface{}, interface{}, error) {
	count := len(self.rows)
	if 0 == count {
		return nil, nil, ErrValueIsEmpty
	}
	if 1 == count {
		return self.rows, self.last, nil
	}

	avg := self.sum / float64(count)
	if avg > 0 {
		if avg <= float64(math.MaxInt64) {
			return self.rows, int64(avg), nil
		}

		if avg <= float64(math.MaxUint64) {
			return self.rows, uint64(avg), nil
		}
		//return sel new(big.Int).SetString(strconv.FormatFloat(avg, 'f', 0, 64), 10)
	} else {
		if avg >= float64(math.MinInt64) {
			return self.rows, int64(avg), nil
		}
	}
	return self.rows, avg, nil
}

type countAggregation struct {
	rows  []map[string]interface{}
	count uint
}

func (self *countAggregation) Create(capacity int) Aggregation {
	return &countAggregation{}
}

func (self *countAggregation) AggregateOne(row map[string]interface{}, v interface{}) (map[string]interface{}, interface{}, error) {
	return row, uint(1), nil
}

func (self *countAggregation) Aggregate(row map[string]interface{}, v interface{}) error {
	self.count++
	self.rows = append(self.rows, row)
	return nil
}

func (self *countAggregation) Result() ([]map[string]interface{}, interface{}, error) {
	return self.rows, self.count, nil
}

type maxAggregation struct {
	max_row   map[string]interface{}
	max_value interface{}
	max_float float64
}

func (self *maxAggregation) Create(capacity int) Aggregation {
	return &maxAggregation{max_float: float64(math.MinInt64)}
}

func (self *maxAggregation) AggregateOne(row map[string]interface{}, v interface{}) (map[string]interface{}, interface{}, error) {
	return row, v, nil
}

func (self *maxAggregation) Aggregate(row map[string]interface{}, v interface{}) error {
	value, e := as.Float64(v)
	if nil != e {
		return e
	}
	if value > self.max_float {
		self.max_row = row
		self.max_value = v
		self.max_float = value
	}
	return nil
}

func (self *maxAggregation) Result() ([]map[string]interface{}, interface{}, error) {
	return []map[string]interface{}{self.max_row}, self.max_value, nil
}

type minAggregation struct {
	min_row   map[string]interface{}
	min_value interface{}
	min_float float64
}

func (self *minAggregation) Create(capacity int) Aggregation {
	return &minAggregation{min_float: float64(math.MaxInt64)}
}

func (self *minAggregation) AggregateOne(row map[string]interface{}, v interface{}) (map[string]interface{}, interface{}, error) {
	return row, v, nil
}

func (self *minAggregation) Aggregate(row map[string]interface{}, v interface{}) error {
	value, e := as.Float64(v)
	if nil != e {
		return e
	}
	if value < self.min_float {
		self.min_row = row
		self.min_value = v
		self.min_float = value
	}
	return nil
}

func (self *minAggregation) Result() ([]map[string]interface{}, interface{}, error) {
	return []map[string]interface{}{self.min_row}, self.min_value, nil
}

type maxDateAggregation struct {
	maxRow   map[string]interface{}
	maxValue interface{}
	maxDate  time.Time
}

func (self *maxDateAggregation) Create(capacity int) Aggregation {
	return &maxDateAggregation{maxDate: time.Time{}}
}

func (self *maxDateAggregation) AggregateOne(row map[string]interface{}, v interface{}) (map[string]interface{}, interface{}, error) {
	return row, v, nil
}

func (self *maxDateAggregation) Aggregate(row map[string]interface{}, v interface{}) error {
	value, e := as.Time(v)
	if nil != e {
		return e
	}
	if self.maxValue == nil || value.After(self.maxDate) {
		self.maxRow = row
		self.maxValue = v
		self.maxDate = value
	}
	return nil
}

func (self *maxDateAggregation) Result() ([]map[string]interface{}, interface{}, error) {
	return []map[string]interface{}{self.maxRow}, self.maxValue, nil
}

type minDateAggregation struct {
	minRow   map[string]interface{}
	minValue interface{}
	minDate  time.Time
}

func (self *minDateAggregation) Create(capacity int) Aggregation {
	return &minDateAggregation{minDate: time.Unix(math.MaxInt64, 0).Local()}
}

func (self *minDateAggregation) AggregateOne(row map[string]interface{}, v interface{}) (map[string]interface{}, interface{}, error) {
	return row, v, nil
}

func (self *minDateAggregation) Aggregate(row map[string]interface{}, v interface{}) error {
	value, e := as.Time(v)
	if nil != e {
		return e
	}
	if self.minValue == nil || value.Before(self.minDate) {
		self.minRow = row
		self.minValue = v
		self.minDate = value
	}
	return nil
}

func (self *minDateAggregation) Result() ([]map[string]interface{}, interface{}, error) {
	return []map[string]interface{}{self.minRow}, self.minValue, nil
}

type sumFloatAggregation struct {
	rows []map[string]interface{}
	sum  float64
}

func (self *sumFloatAggregation) Create(capacity int) Aggregation {
	return &sumFloatAggregation{}
}

func (self *sumFloatAggregation) AggregateOne(row map[string]interface{}, v interface{}) (map[string]interface{}, interface{}, error) {
	return row, v, nil
}

func (self *sumFloatAggregation) Aggregate(row map[string]interface{}, v interface{}) error {
	value, e := as.Float64(v)
	if nil != e {
		return e
	}

	self.rows = append(self.rows, row)
	self.sum += value
	return nil
}

func (self *sumFloatAggregation) Result() ([]map[string]interface{}, interface{}, error) {
	return self.rows, self.sum, nil
}

type sumIntAggregation struct {
	rows []map[string]interface{}
	sum  float64
}

func (self *sumIntAggregation) Create(capacity int) Aggregation {
	return &sumIntAggregation{}
}

func (self *sumIntAggregation) AggregateOne(row map[string]interface{}, v interface{}) (map[string]interface{}, interface{}, error) {
	return row, v, nil
}

func (self *sumIntAggregation) Aggregate(row map[string]interface{}, v interface{}) (e error) {
	value, e := as.Float64(v)
	if nil != e {
		return e
	}

	self.rows = append(self.rows, row)
	self.sum += value
	return nil
}

func (self *sumIntAggregation) Result() ([]map[string]interface{}, interface{}, error) {
	if self.sum > 0 {
		if self.sum <= float64(math.MaxInt64) {
			return self.rows, int64(self.sum), nil
		}

		if self.sum <= float64(math.MaxUint64) {
			return self.rows, uint64(self.sum), nil
		}
		//return sel new(big.Int).SetString(strconv.FormatFloat(self.sum, 'f', 0, 64), 10)
	} else {
		if self.sum >= float64(math.MinInt64) {
			return self.rows, int64(self.sum), nil
		}
	}
	return self.rows, self.sum, nil
}

// type Value interface {
// 	add(v interface{}) (Value, error)
// 	value() interface{}
// }

// type intValue struct {
// 	value int64
// }

// func (self *intValue) add_string(s string) error {
// 	i64, err := strconv.ParseInt(s, 64)
// 	if nil == err {
// 		return self.add_int64(i64)
// 	}

// 	u64, err := strconv.ParseUint(s, 64)
// 	if nil == err {
// 		return self.add_uint64(u64)
// 	}

// 	f64, err := strconv.ParseFloat(s, 64)
// 	if nil == err {
// 		return self.add_float64(f64)
// 	}

// 	v, ok := new(big.Int).SetString(s)
// 	if ok {
// 		return self.add_big(v)
// 	}
// 	return errors.New("'" + s + "' isn't a number.")
// }

// func (self *intValue) add_int64(s string) error {

// }

// func (self *intValue) add_uint64(s string) error {
// }

// func (self *intValue) add_float64(s string) error {
// }

// func (self *intValue) add(v interface{}) error {
// 	switch v := value.(type) {
// 	case json.Number:
// 		return self.add_string(v.String())
// 	case uint:
// 		return self.add_uint64(uint64(v))
// 	case uint64:
// 		return self.add_uint64(v)
// 	case int:
// 		return self.add_int64(int64(v))
// 	case int64:
// 		return self.add_int64(v)
// 	case float32:
// 		return self.add_float64(float64(v))
// 	case float64:
// 		return self.add_float64(v)
// 	case uint8:
// 		return self.add_uint64(uint64(v))
// 	case uint16:
// 		return self.add_uint64(uint64(v))
// 	case uint32:
// 		return self.add_uint64(uint64(v))
// 	case int8:
// 		return self.add_int64(int64(v))
// 	case int16:
// 		return self.add_int64(int64(v))
// 	case int32:
// 		return self.add_int64(int64(v))
// 	case string:
// 		return self.add_string(v)
// 	}

// 	return fmt.Errorf("'[%T]%#v' isn't a number.", v)
// }
