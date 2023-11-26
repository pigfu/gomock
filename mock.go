package gomock

import (
	"context"
	"errors"
	"reflect"
	"sync"
)

const (
	defaultTag          = "mock"
	defaultSeparator    = ","
	mockTagSeparator    = "="
	mockTagValSeparator = " "
	mockTagKeyPattern   = ",[a-z_]+="
	maxTagKey           = 99
)

type Mock struct {
	*sync.Mutex
	tag          string //mock tag mark
	separator    string
	tagSeparator string
	cache        *cache
	mockFactory  map[string]MockFunc
	tagFactory   map[string]TagFunc
}

func New() *Mock {
	mock := &Mock{
		Mutex:        &sync.Mutex{},
		tag:          defaultTag,
		separator:    defaultSeparator,
		tagSeparator: mockTagSeparator,
		cache:        newCache(),
		mockFactory:  make(map[string]MockFunc),
		tagFactory:   make(map[string]TagFunc),
	}
	for key, val := range mockFactory {
		mock.mockFactory[key] = val
	}
	for key, val := range tagFuncMap {
		mock.tagFactory[key] = val
	}
	return mock
}

// RegisterMock register mock function by yourself
func (m *Mock) RegisterMock(key string, mf MockFunc) {
	if mf == nil {
		return
	}
	m.Lock()
	defer m.Unlock()
	if _, ok := m.mockFactory[key]; ok {
		return
	}
	m.mockFactory[key] = mf
}

// RegisterTag register tag parse function by yourself
func (m *Mock) RegisterTag(key string, mt TagFunc) {
	if mt == nil {
		return
	}
	m.Lock()
	defer m.Unlock()
	if _, ok := m.tagFactory[key]; ok {
		return
	}
	m.tagFactory[key] = mt
}
func (m *Mock) Struct(s any) error {
	return m.StructCtx(context.Background(), s)
}
func (m *Mock) StructCtx(ctx context.Context, s any) (err error) {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Pointer || val.Elem().Kind() != reflect.Struct {
		return errors.New("not a initialize struct ptr")
	}
	return m.mockStruct(ctx, val, nil)
}

func (m *Mock) mockStruct(ctx context.Context, val reflect.Value, fl FieldLevel) (err error) {
	if fl == nil {
		fl, err = m.genCache(ctx, val)
	}
	if err != nil {
		return
	}
	return m.mockStructValue(ctx, val, fl)
}
func (m *Mock) mockStructValue(ctx context.Context, val reflect.Value, fl FieldLevel) (err error) {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if fl.IsPtr() {
		_ = m.mockValue(val, fl)
		val = val.Elem()
	}
	for _, field := range fl.GetChildren() {
		switch field.GetKind() {
		case reflect.Slice:
			err = m.mockSliceValue(ctx, val.Field(field.GetIndex()), field)
		case reflect.Struct:
			err = m.mockStructValue(ctx, val.Field(field.GetIndex()), field)
		default:
			err = m.mockValue(val.Field(field.GetIndex()), field)
		}
		if err != nil {
			return
		}
	}
	return
}
func (m *Mock) mockSliceValue(ctx context.Context, val reflect.Value, fl FieldLevel) error {
	_ = m.mockValue(val, fl)
	if len(fl.GetChildren()) == 0 { //not into
		return nil
	}
	var (
		err error
		rt  reflect.Type
	)
	for i := 0; i < val.Len(); i++ {
		rt, _ = m.Indirect(val.Index(i).Type())
		if rt.Kind() == reflect.Struct {
			err = m.mockStructValue(ctx, val.Index(i), fl.GetChildren()[0])
		} else {
			err = m.mockValue(val.Index(i), fl.GetChildren()[0])
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Mock) mockValue(val reflect.Value, fl FieldLevel) error {
	if fl.GetMockFunc() == nil {
		return nil
	}

	rv, err := fl.GetMockFunc()(fl)
	if err != nil {
		return err
	}
	if baseTypes[fl.GetKind()] && fl.GetKind().String() != fl.GetType().String() {
		rv = rv.Convert(fl.GetType())
	}
	val.Set(rv)
	return nil
}
