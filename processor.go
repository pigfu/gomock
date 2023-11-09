package gomock

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	makeSlice       = "slice"
	makeStruct      = "struct"
	makeString      = "string"
	makeInteger     = "integer"
	makeDecimal     = "decimal"
	makeMobilePhone = "mobile_phone"
	makeEmail       = "email"
	makeAddress     = "addr"
	makeTime        = "time"
)
const (
	province = "province"
	city     = "city"
	county   = "county"

	//
	timestampMs     = "ts_ms"
	timestampSecond = "ts_s"
)

var (
	mockFactory = map[string]MockFunc{
		makeSlice:       mockSlice,
		makeStruct:      mockStruct,
		makeString:      mockString,
		makeInteger:     mockInteger,
		makeDecimal:     mockDecimal,
		makeMobilePhone: mockMobilePhone,
		makeEmail:       mockEmail,
		makeAddress:     mockAddress,
		makeTime:        mockTime,
	}
)

type MockFunc func(FieldLevel) (reflect.Value, error)

// make slice
func mockSlice(fl FieldLevel) (reflect.Value, error) {
	rt, tm := fl.GetType(), fl.GetTags()
	eq := tm.Key(MockEqual).GetInt()
	if eq > 0 {
		return reflect.MakeSlice(rt, eq, eq), nil
	}
	gte, gteExists := makeGteVal(reflect.Uint8, tm.Key(MockGt).GetInt(), tm.Key(MockGte).GetInt(),
		tm.Key(MockGt).Exists(), tm.Key(MockGte).Exists())
	lt, ltExists := makeLtVal(reflect.Uint8, tm.Key(MockLt).GetInt(), tm.Key(MockLte).GetInt(),
		tm.Key(MockLt).Exists(), tm.Key(MockLte).Exists())
	if !gteExists && !ltExists || gte < 0 || lt <= 0 || gte >= lt {
		return reflect.MakeSlice(rt, 0, 0), nil
	}
	eq = gte
	if gte < lt {
		eq += rand.Intn(lt - gte)
	}
	return reflect.MakeSlice(rt, eq, eq), nil
}

// make struct
func mockStruct(fl FieldLevel) (reflect.Value, error) {
	return reflect.New(fl.GetType()), nil
}

// mock random string
func mockString(fl FieldLevel) (reflect.Value, error) {
	if fl.GetKind() != reflect.String {
		return reflect.New(fl.GetType()), errors.New("only support the type string")
	}
	val := generateString(fl)
	if fl.IsPtr() {
		return reflect.ValueOf(&val), nil
	}
	return reflect.ValueOf(val), nil
}

func generateString(fl FieldLevel) string {
	tm := fl.GetTags()
	eqVal := tm.Key(MockEqual).GetStr()
	if eqVal != "" {
		return eqVal
	}
	if tm.Key(MockOptions).Exists() {
		return selectOne(fl, tm.Key(MockOptions).GetStrSet())
	}
	return rangeString(fl)
}

func rangeString(fl FieldLevel) string {
	tm := fl.GetTags()
	gte, gteExists := makeGteVal(reflect.Uint8, tm.Key(MockGt).GetInt(), tm.Key(MockGte).GetInt(),
		tm.Key(MockGt).Exists(), tm.Key(MockGte).Exists())
	lt, ltExists := makeLtVal(reflect.Uint8, tm.Key(MockLt).GetInt(), tm.Key(MockLte).GetInt(),
		tm.Key(MockLt).Exists(), tm.Key(MockLte).Exists())
	if !gteExists && !ltExists || gte < 0 || lt <= 0 || gte >= lt {
		return ""
	}
	n := gte
	if gte < lt {
		n += rand.Intn(lt - gte)
	}
	str := &strings.Builder{}
	str.Grow(n)
	for i := 0; i < n; i++ {
		str.WriteByte(letters[rand.Int63()%int64(len(letters))])
	}
	return str.String()
}

// mock integer. for int,int8,int64...
func mockInteger(fl FieldLevel) (reflect.Value, error) {
	fmt.Println(fl.GetKind(), fl.GetType())
	val := generateInteger(fl)
	switch fl.GetKind() {
	case reflect.Int:
		return int64ToInt(fl, val), nil
	case reflect.Int8:
		return int64ToInt8(fl, val), nil
	case reflect.Int16:
		return int64ToInt16(fl, val), nil
	case reflect.Int32:
		return int64ToInt32(fl, val), nil
	case reflect.Int64:
		return int64ToInt64(fl, val), nil
	case reflect.Uint:
		return int64ToUint(fl, val), nil
	case reflect.Uint8:
		return int64ToUint8(fl, val), nil
	case reflect.Uint16:
		return int64ToUint16(fl, val), nil
	case reflect.Uint32:
		return int64ToUint32(fl, val), nil
	case reflect.Uint64:
		return int64ToUint64(fl, val), nil
	}
	return reflect.New(fl.GetType()), fmt.Errorf("not support the type %s", fl.GetKind())
}

func generateInteger(fl FieldLevel) int64 {
	tm := fl.GetTags()
	if tm.Key(MockEqual).Exists() {
		return tm.Key(MockEqual).GetInt64()
	}
	if tm.Key(MockOptions).Exists() {
		return selectOne(fl, tm.Key(MockOptions).GetInt64Set())
	}
	return rangeInteger(fl)
}

func rangeInteger(fl FieldLevel) int64 {
	tm := fl.GetTags()
	gte, gteExists := makeGteVal(fl.GetKind(), tm.Key(MockGt).GetInt64(), tm.Key(MockGte).GetInt64(),
		tm.Key(MockGt).Exists(), tm.Key(MockGte).Exists())
	lt, ltExists := makeLtVal(fl.GetKind(), tm.Key(MockLt).GetInt64(), tm.Key(MockLte).GetInt64(),
		tm.Key(MockLt).Exists(), tm.Key(MockLte).Exists())
	if !gteExists && !ltExists || gte >= lt {
		return 0
	}
	return randRangeInt64(gte, lt)
}

// return value one of [gte,lt)
func randRangeInt64(gte, lt int64) int64 {
	if gte >= 0 {
		return gte + rand.Int63n(lt-gte)
	}
	if gte < 0 && lt <= 0 {
		return -(1 - lt + rand.Int63n(lt-gte))
	}
	point := rand.Int63n(lt - gte)
	if point < lt {
		return randRangeInt64(0, lt)
	}
	return randRangeInt64(gte, 0)
}

func int64ToInt(fl FieldLevel, val int64) reflect.Value {
	nv := int(val)
	if fl.IsPtr() {
		return reflect.ValueOf(&nv)
	}
	return reflect.ValueOf(nv)
}

func int64ToInt8(fl FieldLevel, val int64) reflect.Value {
	nv := int8(val)
	if fl.IsPtr() {
		return reflect.ValueOf(&nv)
	}
	return reflect.ValueOf(nv)
}

func int64ToInt16(fl FieldLevel, val int64) reflect.Value {
	nv := int16(val)
	if fl.IsPtr() {
		return reflect.ValueOf(&nv)
	}
	return reflect.ValueOf(nv)
}

func int64ToInt32(fl FieldLevel, val int64) reflect.Value {
	nv := int32(val)
	if fl.IsPtr() {
		return reflect.ValueOf(&nv)
	}
	return reflect.ValueOf(nv)
}

func int64ToInt64(fl FieldLevel, val int64) reflect.Value {
	if fl.IsPtr() {
		return reflect.ValueOf(&val)
	}
	return reflect.ValueOf(val)
}

func int64ToUint(fl FieldLevel, val int64) reflect.Value {
	nv := uint(val)
	if fl.IsPtr() {
		return reflect.ValueOf(&nv)
	}
	return reflect.ValueOf(nv)
}

func int64ToUint8(fl FieldLevel, val int64) reflect.Value {
	nv := uint8(val)
	if fl.IsPtr() {
		return reflect.ValueOf(&nv)
	}
	return reflect.ValueOf(nv)
}

func int64ToUint16(fl FieldLevel, val int64) reflect.Value {
	nv := uint16(val)
	if fl.IsPtr() {
		return reflect.ValueOf(&nv)
	}
	return reflect.ValueOf(nv)
}

func int64ToUint32(fl FieldLevel, val int64) reflect.Value {
	nv := uint32(val)
	if fl.IsPtr() {
		return reflect.ValueOf(&nv)
	}
	return reflect.ValueOf(nv)
}

func int64ToUint64(fl FieldLevel, val int64) reflect.Value {
	nv := uint64(val)
	if fl.IsPtr() {
		return reflect.ValueOf(&nv)
	}
	return reflect.ValueOf(nv)
}

func selectOne[T string | int64 | float64](fl FieldLevel, options []T) T {
	var (
		weights = fl.GetTags().Key(MockWeights).GetInt64Set()
	)
	var value T
	placement, sum := rand.Int63n(sumWeights(weights, len(options))), int64(0)
	for i, option := range options {
		if len(weights) > 0 {
			sum = weights[i]
		} else {
			sum += 1
		}
		if placement < sum {
			value = option
			break
		}
	}
	return value
}

func sumWeights(weights []int64, number int) int64 {
	if len(weights) > number {
		weights = weights[:number]
	}
	sum := int64(number)
	if len(weights) > 0 {
		sum = sumSlice(weights)
	}
	return sum
}

// mock decimal. for float32,float64
func mockDecimal(fl FieldLevel) (reflect.Value, error) {
	val := generateDecimal(fl)
	switch fl.GetKind() {
	case reflect.Float32:
		return float64ToFloat32(fl, val), nil
	case reflect.Float64:
		return float64ToFloat64(fl, val), nil
	}
	return reflect.New(fl.GetType()), fmt.Errorf("not support the type %s", fl.GetKind())
}
func float64ToFloat64(fl FieldLevel, val float64) reflect.Value {
	if fl.IsPtr() {
		return reflect.ValueOf(&val)
	}
	return reflect.ValueOf(val)
}
func float64ToFloat32(fl FieldLevel, val float64) reflect.Value {
	nv := float32(val)
	if fl.IsPtr() {
		return reflect.ValueOf(&nv)
	}
	return reflect.ValueOf(nv)
}
func generateDecimal(fl FieldLevel) float64 {
	tm := fl.GetTags()
	if tm.Key(MockEqual).Exists() {
		return tm.Key(MockEqual).GetFloat64()
	}
	if tm.Key(MockOptions).Exists() {
		return selectOne(fl, tm.Key(MockOptions).GetFloat64Set())
	}
	return rangeDecimal(fl)
}

func rangeDecimal(fl FieldLevel) float64 {
	tm := fl.GetTags()
	conversion := decimalConversion(tm)
	gte, gteExists := makeGteVal(fl.GetKind(), int64(tm.Key(MockGt).GetFloat64()*conversion),
		int64(tm.Key(MockGte).GetFloat64()*conversion), tm.Key(MockGt).Exists(), tm.Key(MockGte).Exists())
	lt, ltExists := makeLtVal(fl.GetKind(), int64(tm.Key(MockLt).GetFloat64()*conversion),
		int64(tm.Key(MockLte).GetFloat64()*conversion), tm.Key(MockLt).Exists(), tm.Key(MockLte).Exists())
	if !gteExists && !ltExists || gte >= lt {
		return 0
	}
	return float64(randRangeInt64(gte, lt)) / conversion
}
func decimalConversion(tm TagLevelMap) float64 {
	conversion := maxFunc(numberOfDecimal(tm.Key(MockGt).GetStr()), numberOfDecimal(tm.Key(MockGte).GetStr()))
	conversion = maxFunc(conversion, numberOfDecimal(tm.Key(MockLt).GetStr()))
	return maxFunc(conversion, numberOfDecimal(tm.Key(MockLte).GetStr()))
}
func numberOfDecimal(value string) float64 {
	values := strings.SplitN(value, ".", 2)
	if len(values) == 1 {
		return 1
	}
	return math.Pow(10, float64(len(values[1])))
}

// make mobile phone
func mockMobilePhone(fl FieldLevel) (reflect.Value, error) {
	if fl.GetKind() != reflect.String {
		return reflect.New(fl.GetType()), errors.New("only support the type string")
	}
	prefix := mobilePhonePrefix[rand.Intn(len(mobilePhonePrefix))]
	phone := &strings.Builder{}
	phone.Grow(mobilePhoneLen)
	phone.WriteString(prefix)
	for i := 0; i < mobilePhoneLen-len(prefix); i++ {
		phone.WriteString(strconv.Itoa(rand.Intn(10)))
	}
	return reflect.ValueOf(phone.String()), nil
}
func mockEmail(fl FieldLevel) (reflect.Value, error) {
	if fl.GetKind() != reflect.String {
		return reflect.New(fl.GetType()), errors.New("only support the type string")
	}
	postfix := emailPostfix[rand.Intn(len(emailPostfix))]
	emailLen := 7 + rand.Intn(6)
	email := &strings.Builder{}
	email.Grow(emailLen)
	for i := 0; i < emailLen; i++ {
		email.WriteByte(letters[rand.Int63()%int64(len(letters))])
	}
	email.WriteString(postfix)
	return reflect.ValueOf(email.String()), nil
}

func mockAddress(fl FieldLevel) (reflect.Value, error) {
	if fl.GetKind() != reflect.String {
		return reflect.New(fl.GetType()), errors.New("only support the type string")
	}
	if !fl.GetTags().Key(MockAddress).Exists() {
		return reflect.ValueOf(""), nil
	}
	result := &strings.Builder{}
	provinceVal := randProvince(area)
	cityVal := randCity(area[provinceVal])
	countyVal := randCountry(area[provinceVal][cityVal])

	types := fl.GetTags().Key(MockAddress).GetStrSet()
	for _, ty := range types {
		if result.Len() > 0 {
			result.WriteString(" ")
		}
		switch ty {
		case province:
			result.WriteString(provinceVal)
		case city:
			result.WriteString(cityVal)
		case county:
			result.WriteString(countyVal)
		}
	}
	return reflect.ValueOf(result.String()), nil
}
func randProvince(provinceMap map[string]map[string]map[string]struct{}) string {
	provinces := mapToSlice(provinceMap, func(key string,
		_ map[string]map[string]struct{}) string {
		return key
	})
	return provinces[rand.Int63()%int64(len(provinces))]
}

func randCity(cityMap map[string]map[string]struct{}) string {
	cities := mapToSlice(cityMap, func(key string, _ map[string]struct{}) string {
		return key
	})
	return cities[rand.Int63()%int64(len(cities))]
}
func randCountry(countryMap map[string]struct{}) string {
	countries := mapToSlice(countryMap, func(key string, _ struct{}) string {
		return key
	})
	return countries[rand.Int63()%int64(len(countries))]
}

func mockTime(fl FieldLevel) (reflect.Value, error) {
	if !fl.GetTags().Key(MockTime).Exists() {
		return reflect.New(fl.GetType()), nil
	}
	switch fl.GetKind() {
	case reflect.Int64:
		return genTimestamp(fl)
	case reflect.String:
		return genTimeFormat(fl)
	}
	return reflect.ValueOf(fl.GetType()), fmt.Errorf("not support the type %s", fl.GetKind())
}
func genTimestamp(fl FieldLevel) (reflect.Value, error) {
	mt := fl.GetTags().Key(MockTime).GetStr()
	if mt == timestampSecond {
		return reflect.ValueOf(time.Now().Unix()), nil
	}
	if mt == timestampMs {
		return reflect.ValueOf(time.Now().UnixMilli()), nil
	}
	return reflect.ValueOf(int64(0)), fmt.Errorf("not support the type %s", fl.GetKind())
}
func genTimeFormat(fl FieldLevel) (reflect.Value, error) {
	mt := fl.GetTags().Key(MockTime).GetStr()
	return reflect.ValueOf(time.Now().Format(mt)), nil
}
