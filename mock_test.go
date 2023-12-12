package gomock

import (
	"context"
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
	Pros []string  `json:"pros" mock:"gte=1,lte=5,into=1,key=string,gte=3,lte=6"`
}
type Book struct {
	Id   int64  `json:"id" mock:"key=integer,eq=5"`
	Name string `json:"name"  mock:"key=string,gte=4,lte=23"`
}
type Man struct {
	Id          int64    `json:"id" mock:"key=integer,eq=5"`
	Ids         []int64  `json:"ids" mock:"key=ids"`
	Name        string   `json:"name" mock:"key=chinese,chinese_tag=李明"`
	Age         *int8    `json:"age" mock:"key=integer,gte=23"`
	Hobby       *Hobby   `json:"hobby,omitempty" mock:"into=1"`
	Hobbies     []*Hobby `json:"hobbies,omitempty"  mock:"eq=1,into=1"`
	Books       []Book   `json:"books,omitempty" mock:"eq=1,into=1,key=book"`
	Option      int32    `json:"option,omitempty"  mock:"key=integer,options=2 3 4 5,weights=10 5 2 2"`
	Decimal     float64  `json:"decimal,omitempty"  mock:"key=decimal,gte=-23.235,lte=5.580"`
	MobilePhone *string  `json:"mobile_phone,omitempty"  mock:"key=mobile_phone"`
	Email       *string  `json:"email,omitempty"  mock:"key=email"`
	Address     *string  `json:"address,omitempty"  mock:"key=addr,addr=city county"`
	CreateTime  *int64   `json:"create_time,omitempty"  mock:"key=time,time=ts_ms"`
	UpdateTime  *string  `json:"update_time,omitempty"  mock:"key=time,time=2006-01-02 15:04:05"`
	RegDecimal  float64  `json:"reg_decimal,omitempty"  mock:"key=decimal,reg=[1-9]{3}\\.\\d{1,5}"`
	RegName     string   `json:"reg_name,omitempty"  mock:"key=string,reg=[\u4e00-\u9fa5]{6,}"`
}

func TestMock(t *testing.T) {
	mock := New()
	mock.RegisterMock("chinese", func(_ context.Context, fl FieldLevel) (reflect.Value, error) {
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

	mock.RegisterMock("book", func(_ context.Context, fl FieldLevel) (reflect.Value, error) {
		if fl.GetKind() != reflect.Struct {
			return reflect.New(fl.GetType()), errors.New("only support the type struct")
		}
		book := Book{Id: 555, Name: "test book"}
		if fl.IsPtr() {
			return reflect.ValueOf(&book), nil
		}
		return reflect.ValueOf(book), nil
	})

	mock.RegisterMock("ids", func(_ context.Context, fl FieldLevel) (reflect.Value, error) {
		if fl.GetKind() != reflect.Slice {
			return reflect.Value{}, errors.New("only support the type slice")
		}
		ids := []int64{101, 201, 301, 999}
		if fl.IsPtr() {
			return reflect.ValueOf(&ids), nil
		}
		return reflect.ValueOf(ids), nil
	})

	man := &Man{}
	err := mock.Struct(man)
	if err != nil {
		t.Error(err)
	}
	b, _ := json.Marshal(man)
	t.Logf("success: %s", string(b))
}

type Human struct {
	Books []Book `json:"books,omitempty" mock:"eq=1,into=1,key=book"`
}

func TestMockContext(t *testing.T) {
	mockCtx := context.WithValue(context.Background(), "book", Hobby{Id: 555, Name: "test book"})
	mock := New()
	mock.RegisterMock("book", func(ctx context.Context, fl FieldLevel) (reflect.Value, error) {
		if fl.GetKind() != reflect.Struct {
			return reflect.Value{}, errors.New("only support the type struct")
		}
		book, _ := ctx.Value("book").(Book)
		if fl.IsPtr() {
			return reflect.ValueOf(&book), nil
		}
		return reflect.ValueOf(book), nil
	})

	human := &Human{}
	err := mock.StructCtx(mockCtx, human)
	if err != nil {
		t.Error(err)
	}

	b, _ := json.Marshal(human)
	t.Logf("success: %s", string(b))
}
