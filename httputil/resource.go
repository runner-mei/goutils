package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/runner-mei/errors"
	"github.com/runner-mei/goutils/as"
)

var (
	NotImplementedResult = ReturnError(501, "not implemented")
)

type SimpleResult interface {
	ErrorCode() int
	ErrorMessage() string
	HasError() bool
	Error() errors.RuntimeError
	Value() Any
	InterfaceValue() interface{}
	CreatedAt() time.Time
}

type Result interface {
	SimpleResult

	Return(value interface{}) Result
	Warnings() interface{}
	Effected() int64
	LastInsertId() interface{}
	HasOptions() bool
	Options() map[string]interface{}
	ToJson() string
	ToMap() map[string]interface{}
}

type Any interface {
	AsInterface() interface{}

	AsBool() (bool, error)

	AsInt() (int, error)

	AsInt32() (int32, error)

	AsInt64() (int64, error)

	AsUint() (uint, error)

	AsUint32() (uint32, error)

	AsUint64() (uint64, error)

	AsString() (string, error)

	AsStrings() ([]string, error)

	AsObject() (map[string]interface{}, error)

	AsArray() ([]interface{}, error)

	AsObjects() ([]map[string]interface{}, error)

	ToString() string
}

type ResultImpl struct {
	Eid             interface{}              `json:"request_id,omitempty"`
	Eerror          *errors.ApplicationError `json:"error,omitempty"`
	Ewarnings       interface{}              `json:"warnings,omitempty"`
	Evalue          interface{}              `json:"value,omitempty"`
	Eeffected       *int64                   `json:"effected,omitempty"`
	ElastInsertId   interface{}              `json:"lastInsertId,omitempty"`
	Eoptions        map[string]interface{}   `json:"options,omitempty"`
	Ecreated_at     time.Time                `json:"created_at,omitempty"`
	Erepresentation string                   `json:"representation,omitempty"`

	value    AnyData
	effected int64
}

func Return(value interface{}) *ResultImpl {
	return &ResultImpl{Evalue: value, Ecreated_at: time.Now(), effected: -1, ElastInsertId: nil}
}

func ReturnError(code int, msg string) *ResultImpl {
	return Return(nil).SetError(code, msg)
}

func (self *ResultImpl) SetValue(value interface{}) *ResultImpl {
	self.Evalue = value
	return self
}

func (self *ResultImpl) Return(value interface{}) Result {
	self.Evalue = value
	return self
}

func (self *ResultImpl) SetOptions(options map[string]interface{}) *ResultImpl {
	self.Eoptions = options
	return self
}

func (self *ResultImpl) SetOption(key string, value interface{}) Result {
	if nil == self.Eoptions {
		self.Eoptions = make(map[string]interface{})
	}
	self.Eoptions[key] = value
	return self
}

func (self *ResultImpl) SetError(code int, msg string) *ResultImpl {
	if 0 == code && 0 == len(msg) {
		return self
	}

	if nil == self.Eerror {
		self.Eerror = errors.NewError(code, msg)
	} else {
		self.Eerror.Code = code
		self.Eerror.Message = msg
	}
	return self
}

func (self *ResultImpl) SetWarnings(value interface{}) *ResultImpl {
	self.Ewarnings = value
	return self
}

func (self *ResultImpl) SetEffected(effected int64) *ResultImpl {
	self.effected = effected
	self.Eeffected = &self.effected
	return self
}

func (self *ResultImpl) SetLastInsertId(id interface{}) *ResultImpl {
	self.ElastInsertId = id
	return self
}

func (self *ResultImpl) ErrorCode() int {
	if nil != self.Eerror {
		return self.Eerror.Code
	}
	return -1
}

func (self *ResultImpl) ErrorMessage() string {
	if nil != self.Eerror {
		return self.Eerror.Message
	}
	return ""
}

func (self *ResultImpl) HasError() bool {
	return nil != self.Eerror && (0 != self.Eerror.Code || 0 != len(self.Eerror.Message))
}

func (self *ResultImpl) Error() errors.RuntimeError {
	if nil == self.Eerror {
		return nil
	}
	return self.Eerror
}

func (self *ResultImpl) Warnings() interface{} {
	return self.Ewarnings
}

func (self *ResultImpl) Value() Any {
	self.value.Value = self.Evalue
	return &self.value
}

func (self *ResultImpl) InterfaceValue() interface{} {
	return self.Evalue
}

func (self *ResultImpl) Effected() int64 {
	if nil != self.Eeffected {
		return *self.Eeffected
	}
	return -1
}

func (self *ResultImpl) LastInsertId() interface{} {
	if nil == self.ElastInsertId {
		return -1
	}
	return self.ElastInsertId
}

func (self *ResultImpl) HasOptions() bool {
	return len(self.Eoptions) > 0
}

func (self *ResultImpl) Options() map[string]interface{} {
	if nil == self.Eoptions {
		self.Eoptions = make(map[string]interface{})
	}
	return self.Eoptions
}

func (self *ResultImpl) CreatedAt() time.Time {
	return self.Ecreated_at
}

func (self *ResultImpl) ToJson() string {
	bs, e := json.Marshal(self)
	if nil != e {
		panic(e.Error())
	}
	return string(bs)
}

func (self *ResultImpl) ToMap() map[string]interface{} {
	res := map[string]interface{}{}

	res["created_at"] = self.Ecreated_at
	if 0 != len(self.Erepresentation) {
		res["representation"] = self.Erepresentation
	}

	if nil != self.Eerror {
		res["error"] = map[string]interface{}{
			"code":    self.Eerror.Code,
			"message": self.Eerror.Message,
		}
	}
	if nil != self.Ewarnings {
		res["warnings"] = self.Ewarnings
	}
	if nil != self.Evalue {
		res["value"] = self.Evalue
	}
	if nil != self.Eeffected && -1 != *self.Eeffected {
		res["effected"] = *self.Eeffected
	}
	if nil != self.ElastInsertId {
		res["lastInsertId"] = self.ElastInsertId
	}
	if nil != self.Eoptions {
		res["options"] = self.Eoptions
	}
	return res
}

type AnyData struct {
	Value interface{}
}

func (self *AnyData) IsNil() bool {
	return nil == self.Value
}

func (self *AnyData) AsInterface() interface{} {
	return self.Value
}

func (self *AnyData) AsBool() (bool, error) {
	return as.Bool(self.Value)
}

func (self *AnyData) AsInt() (int, error) {
	return as.Int(self.Value)
}

func (self *AnyData) AsInt32() (int32, error) {
	return as.Int32(self.Value)
}

func (self *AnyData) AsInt64() (int64, error) {
	return as.Int64(self.Value)
}

func (self *AnyData) AsUint() (uint, error) {
	return as.Uint(self.Value)
}

func (self *AnyData) AsUint32() (uint32, error) {
	return as.Uint32(self.Value)
}

func (self *AnyData) AsUint64() (uint64, error) {
	return as.Uint64(self.Value)
}

func (self *AnyData) AsString() (string, error) {
	return as.String(self.Value)
}

func (self *AnyData) AsStrings() ([]string, error) {
	return as.Strings(self.Value)
}

func (self *AnyData) AsArray() ([]interface{}, error) {
	return as.Array(self.Value)
}

func (self *AnyData) AsObject() (map[string]interface{}, error) {
	return as.Object(self.Value)
}

func (self *AnyData) AsObjects() ([]map[string]interface{}, error) {
	return as.Objects(self.Value)
}

func (self *AnyData) ToString() string {
	if nil == self.Value {
		return ""
	}
	return fmt.Sprint(self.Value)
}

func ReturnWithInternalError(message string) Result {
	return ReturnError(http.StatusInternalServerError, message)
}

func ReturnWithBadRequest(message string) Result {
	return ReturnError(http.StatusBadRequest, message)
}

func ReturnWithNotAcceptable(message string) Result {
	return ReturnError(http.StatusNotAcceptable, message)
}

func ReturnWithIsRequired(name string) Result {
	return ReturnError(http.StatusBadRequest, "'"+name+"' is required.")
}

func ReturnWithNotFoundWithMessage(id, msg string) Result {
	if 0 == len(id) {
		return ReturnError(http.StatusNotFound, msg)
	}
	return ReturnError(http.StatusNotFound, "'"+id+"' is not found - "+msg)
}
func ReturnWithNotFound(id string) Result {
	return ReturnError(http.StatusNotFound, "'"+id+"' is not found.")
}

func ReturnWithRecordNotFound(t, id string) Result {
	return ReturnError(http.StatusNotFound, t+" with id was '"+id+"' is not found.")
}

func ReturnWithRecordAlreadyExists(id string) Result {
	return ReturnError(http.StatusNotAcceptable, "'"+id+"' is already exists.")
}

func ReturnWithNotImplemented() Result {
	return ReturnError(http.StatusInternalServerError, "not implemented")
}

func ReturnWithServiceUnavailable(msg string) Result {
	return ReturnError(http.StatusServiceUnavailable, msg)
}
