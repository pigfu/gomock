package gomock

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

type HobbyType int32

var (
	SportHobbyType   HobbyType = 1
	ArtHobbyType     HobbyType = 2
	OutdoorHobbyType HobbyType = 3
)

type Hobby struct {
	Id   int64     `json:"id" mock:"key=integer,eq=5"`
	HT   HobbyType `json:"ht"  mock:"key=integer,options=1 2 3"`
	Name string    `json:"name"  mock:"key=string,gte=4,lte=23"`
	Pros []string  `json:"pros" mock:"gte=1,lte=5,into,key=string,gte=3,lte=6"`
}
type Man struct {
	Id          int64    `json:"id" mock:"key=integer,eq=5"`
	Ids         []int64  `json:"ids" mock:"eq=2,into,key=integer,gte=23,lte=55"`
	Name        string   `json:"name" mock:"key=chinese,chinese_tag=李明"`
	Age         *int8    `json:"age" mock:"key=integer,gte=23"`
	Hobby       *Hobby   `json:"hobby,omitempty" mock:"into"`
	Hobbies     []*Hobby `json:"hobbies,omitempty"  mock:"eq=1,into"`
	Option      int32    `json:"option,omitempty"  mock:"key=integer,options=2 3 4 5,weights=10 5 2 2"`
	Decimal     float64  `json:"decimal,omitempty"  mock:"key=decimal,gte=-23.235,lte=5.580"`
	MobilePhone string   `json:"mobile_phone,omitempty"  mock:"key=mobile_phone"`
	Email       string   `json:"email,omitempty"  mock:"key=email"`
	Address     string   `json:"address,omitempty"  mock:"key=addr,addr=city county"`
	CreateTime  int64    `json:"create_time,omitempty"  mock:"key=time,time=ts_ms"`
	UpdateTime  string   `json:"update_time,omitempty"  mock:"key=time,time=2006-01-02 15:04:05"`
}

func TestMock(t *testing.T) {
	mock := New()
	mock.RegisterMock("chinese", func(fl FieldLevel) (reflect.Value, error) {
		if fl.GetKind() != reflect.String {
			return reflect.New(fl.GetType()), errors.New("only support the type string")
		}
		value := fl.GetTags().Key("chinese_tag").GetStr()
		if value != "" {
			return reflect.ValueOf(value), nil
		}
		return reflect.ValueOf("你好世界！！！"), nil
	})

	mock.RegisterTag("chinese_tag", func(_ reflect.Type, key, value string) (TagLevel, error) {
		return &MockTag{
			Key:    key,
			Value:  value,
			StrVal: value,
		}, nil
	})
	//var man *Man
	man := &Hobby{}
	err := mock.Struct(man)
	if err != nil {
		t.Error(err)
	}
	b, _ := json.Marshal(man)
	t.Logf("success: %s", string(b))
}
