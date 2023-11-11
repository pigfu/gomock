# gomock
a mock tool is based on struct tag
## install
```sh
go get github.com/pigfu/gomock
```
## recommend
1. if the struct generated by the proto file, you can use [protoc-go-inject-tag](https://github.com/favadi/protoc-go-inject-tag) library.
2. for testers,you can do API fuzzy testing by this library.
3. for front-end developers,you can quickly build mock services for front-end API debugging, interrupting the development mode of front-end waiting for back-end API.
4. for back-end developers, you can quickly build parameters for API debugging.
## mock function

| Function     | Description                           |
|--------------|---------------------------------------|
| string       | mock string data, only english string |
| integer      | mock integer data                     |
| decimal      | mock decimal data                     |
| mobile_phone | mock mobile phone                     |
| email        | mock email                            |
| addr         | mock addr, only china                 |
| time         | mock time                             |

## mock tag

| Tag     | Description                                                                                                                                         |
|---------|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| key     | appoint mock function                                                                                                                               |
| eq      | for integer, decimal, string, eq will ensure that the value is equal to the parameter given. for slice, it will ensure the length.                  |
| lt      | for integer, decimal, lt will ensure the maximum value.  for string, slice, it will ensure the maximum length.                                      |
| lte     | for integer, decimal, lt will ensure the maximum value.  for string, slice, it will ensure the maximum length.                                      |
| gt      | for integer, decimal, lt will ensure the minimum value.  for string, slice, it will ensure the minimum length.                                      |
| gte     | for integer, decimal, lt will ensure the minimum value.  for string, slice, it will ensure the minimum length.                                      |
| options | specify optional data, like options=2 5 8                                                                                                           |
| weights | specify the weight of optional data, default weights=1 1 1                                                                                          |
| time    |                                                                                                                                                     |
| into    | appoint to mock struct or slice field,previous modifications to slice, subsequent modifications to internal fields of slice                         |
| -       | skip the field                                                                                                                                      |
| addr    | only support addr mock function, any one or any combination of optional province city county                                                        |
| time    | only support time mock function, supports timestamps at the second (ts_s) and millisecond (ts_ms) level or any time format (eg:2006/01/02 15:04:05) |

## example

```go
package main

import (
	. "github.com/pigfu/gomock"
	"encoding/json"
	"fmt"
)

type HobbyType int32

type Hobby struct {
	Id   int64     `json:"id" mock:"key=integer,eq=5"`
	HT   HobbyType `json:"ht"  mock:"key=integer,options=1 2 3"`
	Name string    `json:"name"  mock:"key=string,gte=4,lte=23"`
	Pros []string  `json:"pros" mock:"gte=1,lte=5,into,key=string,gte=3,lte=6"`
}

type Man struct {
	Id          int64    `json:"id" mock:"key=integer,eq=5"`
	Ids         []int64  `json:"ids" mock:"eq=2,into,key=integer,gte=23,lte=55"`
	Name        string   `json:"name"  mock:"key=string,gte=6,lte=10"`
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

func main() {
	mock := New()
	man := &Man{}
	err := mock.Struct(man)
	if err != nil {
		fmt.Println(err)
		return
	}
	b, _ := json.Marshal(man)
	fmt.Println("success: ", string(b))
}
```
You will get the following:
```json
{"id":5,"ids":[33,25],"name":"X5K8LD3F","age":49,"hobby":{"id":5,"ht":1,"name":"8fzH9yJQ","pros":["5XDTkS","mQLJ","G9T0"]},"hobbies":[{"id":5,"ht":2,"name":"q97NuPswO0I6VZ","pros":["MGbx","ZEi7L4","xOwM67","zpl","LYzBo0"]}],"decimal":0.159,"mobile_phone":"18909537318","email":"08A7z7PMZS@hotmail.com","address":"天津市 西青区","create_time":1699695881544,"update_time":"2023-11-11 17:44:41"}
```
Of course, you can customize the mock method what you need
```go
package main

import (
	. "github.com/pigfu/gomock"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type Hobby struct {
	Name string `json:"name" mock:"key=chinese,chinese_tag=李明"`
	Nickname string `json:"nickname" mock:"key=chinese"`
}

func main() {
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
	hobby := &Hobby{}
	err := mock.Struct(hobby)
	if err != nil {
		fmt.Println(err)
		return
	}
	b, _ := json.Marshal(hobby)
	fmt.Println("success: ", string(b))
}
```
You will get the following:
```json
{"name":"李明","nickname":"你好世界！！！"}
```

