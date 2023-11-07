package gomock

import (
	"reflect"
	"strconv"
	"strings"
)

type MockTag struct {
	Key    string
	Value  any      //parsed value
	StrVal string   //original value
	StrSet []string //original value of set
	isNil  bool
}
type TagLevel interface {
	GetKey() string
	Exists() bool
	GetVal() any
	GetInt() int
	GetInt64() int64
	GetInt64Set() []int64
	GetStr() string
	GetStrSet() []string
	GetFloat64() float64
	GetFloat64Set() []float64
}

type TagLevelMap map[string]TagLevel

func (tm TagLevelMap) Key(key string) TagLevel {
	tf, ok := tm[key]
	if !ok {
		return &MockTag{isNil: true}
	}
	return tf
}

func (mt *MockTag) GetKey() string {
	if mt.isNil {
		return ""
	}
	if v, ok := mt.Value.(string); ok {
		return v
	}
	return ""
}
func (mt *MockTag) Exists() bool {
	if mt.isNil {
		return false
	}
	return true
}

func (mt *MockTag) GetVal() any {
	if mt.isNil {
		return nil
	}
	return mt.Value
}
func (mt *MockTag) GetInt64() int64 {
	if mt.isNil {
		return 0
	}
	if v, ok := mt.Value.(int64); ok {
		return v
	}
	return 0
}
func (mt *MockTag) GetInt64Set() []int64 {
	if mt.isNil {
		return nil
	}
	if v, ok := mt.Value.([]int64); ok {
		return v
	}
	return nil
}
func (mt *MockTag) GetInt() int {
	if mt.isNil {
		return 0
	}
	if v, ok := mt.Value.(int64); ok {
		return int(v)
	}
	return 0
}
func (mt *MockTag) GetStr() string {
	if mt.isNil {
		return ""
	}
	return mt.StrVal
}
func (mt *MockTag) GetStrSet() []string {
	if mt.isNil {
		return nil
	}
	return mt.StrSet
}

func (mt *MockTag) GetFloat64() float64 {
	if mt.isNil {
		return 0.0
	}
	if v, ok := mt.Value.(float64); ok {
		return v
	}
	return 0.0
}
func (mt *MockTag) GetFloat64Set() []float64 {
	if mt.isNil {
		return nil
	}
	if v, ok := mt.Value.([]float64); ok {
		return v
	}
	return nil
}

//////////////////the tag parse func/////////////////////////

type TagFunc func(rt reflect.Type, key, value string) (TagLevel, error)

const (
	MockKey     = "key"
	MockEqual   = "eq"
	MockLt      = "lt"
	MockLte     = "lte"
	MockGt      = "gt"
	MockGte     = "gte"
	MockOptions = "options"
	MockWeights = "weights"
	MockInto    = "into"
	MockSkip    = "-"
	MockAddress = "addr"
	MockTime    = "time"
)

var (
	tagFuncMap = map[string]TagFunc{
		MockKey:     SimpleFunc,
		MockInto:    SimpleFunc,
		MockSkip:    SimpleFunc,
		MockEqual:   EqualFunc,
		MockLt:      NumberFunc,
		MockLte:     NumberFunc,
		MockGt:      NumberFunc,
		MockGte:     NumberFunc,
		MockOptions: OptionsFunc,
		MockWeights: WeightsFunc,
		MockAddress: AddressFunc,
		MockTime:    SimpleFunc,
	}
)

func SimpleFunc(_ reflect.Type, key, value string) (TagLevel, error) {
	return &MockTag{
		Key:    key,
		Value:  value,
		StrVal: value,
	}, nil
}
func EqualFunc(rt reflect.Type, key, value string) (TagLevel, error) {

	var (
		err error
		mt  = &MockTag{Key: key, StrVal: value}
	)
	switch rt.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		mt.Value, err = strconv.ParseInt(value, 10, 64)
	case reflect.Float32, reflect.Float64:
		mt.Value, err = strconv.ParseFloat(value, 64)
	case reflect.String:
		mt.Value = value
	case reflect.Slice:
		mt.Value, err = strconv.ParseInt(value, 10, 64)
	}
	return mt, err
}
func NumberFunc(rt reflect.Type, key, value string) (TagLevel, error) {
	var (
		err error
		mt  = &MockTag{Key: key, StrVal: value}
	)
	switch rt.Kind() {
	case reflect.Float32, reflect.Float64:
		mt.Value, err = strconv.ParseFloat(value, 64)
	default:
		mt.Value, err = strconv.ParseInt(value, 10, 64)
	}
	return mt, err
}
func OptionsFunc(rt reflect.Type, key, value string) (TagLevel, error) {
	var (
		err    error
		values = strings.Split(value, mockTagValSeparator)
		mt     = &MockTag{Key: key, StrVal: value, StrSet: values}
	)
	switch rt.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		mt.Value, err = OptionsIntFunc(values)
	case reflect.Float32, reflect.Float64:
		mt.Value, err = OptionsFloatFunc(values)
	case reflect.String:
		mt.Value = values
	}
	return mt, err
}
func OptionsIntFunc(values []string) ([]int64, error) {
	var (
		err     error
		nv      int64
		options = make([]int64, 0, len(values))
	)
	for _, value := range values {
		nv, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		options = append(options, nv)
	}
	return options, nil
}
func OptionsFloatFunc(values []string) ([]float64, error) {
	var (
		err     error
		nv      float64
		options = make([]float64, 0, len(values))
	)
	for _, value := range values {
		nv, err = strconv.ParseFloat(value, 10)
		if err != nil {
			return nil, err
		}
		options = append(options, nv)
	}
	return options, nil
}

func WeightsFunc(_ reflect.Type, key, value string) (TagLevel, error) {
	var (
		err     error
		values  = strings.Split(value, mockTagValSeparator)
		mt      = &MockTag{Key: key, StrVal: value, StrSet: values}
		weights = make([]int64, 0, len(values))
		nv      int64
	)
	for _, v := range values {
		nv, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		weights = append(weights, nv)
	}
	mt.Value = weights
	return mt, err
}
func AddressFunc(_ reflect.Type, key, value string) (TagLevel, error) {
	var (
		values = strings.Split(value, mockTagValSeparator)
		mt     = &MockTag{Key: key, Value: values, StrVal: value, StrSet: values}
	)
	return mt, nil
}
