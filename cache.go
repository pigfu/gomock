package gomock

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type cache struct {
	lock  *sync.Mutex
	cache *sync.Map // map[reflect.Type]*structParse
}

var (
	baseTypes = map[reflect.Kind]bool{
		reflect.Bool:    true,
		reflect.Int:     true,
		reflect.Int8:    true,
		reflect.Int16:   true,
		reflect.Int32:   true,
		reflect.Int64:   true,
		reflect.Uint:    true,
		reflect.Uint8:   true,
		reflect.Uint16:  true,
		reflect.Uint32:  true,
		reflect.Uint64:  true,
		reflect.Float32: true,
		reflect.Float64: true,
		reflect.String:  true,
	}
	notSupportTypes = map[reflect.Kind]struct{}{
		reflect.Array:         {},
		reflect.Map:           {},
		reflect.Chan:          {},
		reflect.Func:          {},
		reflect.Interface:     {},
		reflect.Uintptr:       {},
		reflect.UnsafePointer: {},
		reflect.Complex64:     {},
		reflect.Complex128:    {},
	}
)

func newCache() *cache {
	return &cache{
		lock:  &sync.Mutex{},
		cache: &sync.Map{},
	}
}

func (c *cache) get(rt reflect.Type) *mockField {
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	value, ok := c.cache.Load(rt.String())
	if !ok {
		return nil
	}
	return value.(*mockField)
}
func (c *cache) set(ty reflect.Type, sp *mockField) {
	if ty.Kind() == reflect.Pointer {
		ty = ty.Elem()
	}
	c.cache.Store(ty.String(), sp)
}

type mockField struct {
	index    int
	rt       reflect.Type
	rk       reflect.Kind
	isPtr    bool
	tempTags []string //temporarily used during parsing
	intoTags []string //into slice

	name     string       //the field name
	tags     TagLevelMap  //map[string]*mockTag
	parent   FieldLevel   //*mockField
	children []FieldLevel //[]*mockField

	mf MockFunc
}
type FieldLevel interface {
	GetIndex() int
	GetType() reflect.Type
	GetKind() reflect.Kind
	IsPtr() bool
	GetName() string
	GetParent() FieldLevel
	GetChildren() []FieldLevel
	GetTags() TagLevelMap
	GetMockFunc() MockFunc
}

func (mf *mockField) GetIndex() int {
	return mf.index
}
func (mf *mockField) GetType() reflect.Type {
	return mf.rt
}
func (mf *mockField) GetKind() reflect.Kind {
	return mf.rk
}
func (mf *mockField) IsPtr() bool {
	return mf.isPtr
}
func (mf *mockField) GetName() string {
	return mf.name
}
func (mf *mockField) GetParent() FieldLevel {
	return mf.parent
}
func (mf *mockField) GetChildren() []FieldLevel {
	return mf.children
}
func (mf *mockField) GetTags() TagLevelMap {
	return mf.tags
}
func (mf *mockField) GetMockFunc() MockFunc {
	return mf.mf
}
func (m *Mock) Indirect(rt reflect.Type) (reflect.Type, bool) {
	if rt.Kind() == reflect.Pointer {
		return rt.Elem(), true
	}
	return rt, false
}
func (m *Mock) genCache(ctx context.Context, rv reflect.Value) (FieldLevel, error) {
	defer m.cache.lock.Unlock()
	m.cache.lock.Lock()
	mf := m.cache.get(rv.Type())
	if mf != nil {
		return mf, nil
	}
	rt, isPtr := m.Indirect(rv.Type())
	mf = &mockField{
		rt:    rt,
		rk:    rt.Kind(),
		isPtr: isPtr,
	}
	err := m.parseStruct(ctx, mf, rt)
	if err != nil {
		return nil, err
	}
	m.cache.set(rt, mf)
	return mf, nil
}
func (m *Mock) contactName(mf *mockField, name string) {
	filler, separator := "", ""
	if mf.parent.GetKind() == reflect.Slice {
		filler = ".0"
	}
	if mf.parent.GetName() != "" && name != "" {
		separator = "."
	}
	mf.name = mf.parent.GetName() + filler + separator + name
}
func (m *Mock) parseStruct(ctx context.Context, mf *mockField, rt reflect.Type) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	var err error
	for i := 0; i < rt.NumField(); i++ {
		if rt.Field(i).Tag.Get(m.tag) == "" { //no mock tag,skip the field
			continue
		}
		err = m.parseStructField(ctx, mf, i, rt.Field(i))
		if err != nil {
			return err
		}
	}
	return nil
}
func (m *Mock) parseStructField(ctx context.Context, parent *mockField, index int, rs reflect.StructField) error {
	mf := &mockField{
		index:  index,
		tags:   make(TagLevelMap),
		parent: parent,
	}
	mf.rt, mf.isPtr = m.Indirect(rs.Type)
	mf.rk = mf.rt.Kind()
	if _, ok := notSupportTypes[mf.rk]; ok {
		return fmt.Errorf("not support the kind:%s", mf.rk.String())
	}
	m.contactName(mf, rs.Name)
	err := m.parseTag(ctx, mf, rs.Tag)
	if err == nil {
		parent.children = append(parent.children, mf)
	}
	return err
}
func (m *Mock) parseIntoBase(parent *mockField, rt reflect.Type) error {
	if len(parent.intoTags) == 0 { //ignore
		return nil
	}
	mf := &mockField{
		tags:     make(map[string]TagLevel),
		parent:   parent,
		tempTags: parent.intoTags,
	}
	mf.rt, mf.isPtr = m.Indirect(rt)
	mf.rk = mf.rt.Kind()
	if _, ok := notSupportTypes[mf.rk]; ok {
		return fmt.Errorf("not support the kind:%s", mf.rk.String())
	}
	m.contactName(mf, "")
	err := m.parseBaseTag(mf)
	if err == nil {
		parent.children = append(parent.children, mf)
	}
	return err
}
func (m *Mock) parseIntoStruct(ctx context.Context, parent *mockField, rt reflect.Type) error {
	mf := &mockField{
		tags:     make(map[string]TagLevel),
		parent:   parent,
		tempTags: parent.intoTags,
	}
	mf.rt, mf.isPtr = m.Indirect(rt)
	mf.rk = mf.rt.Kind()
	if _, ok := notSupportTypes[mf.rk]; ok {
		return fmt.Errorf("not support the kind:%s", mf.rk.String())
	}
	if mf.isPtr { //init struct ptr
		mf.mf = m.mockFactory[makeStruct]
	}
	m.contactName(mf, rt.Name())
	err := m.parseStruct(ctx, mf, rt)
	if err == nil {
		parent.children = append(parent.children, mf)
	}
	return err
}
func (m *Mock) genMockFunc(mf *mockField) error {
	key := mf.tags.Key(MockKey).GetKey()
	if fn, ok := m.mockFactory[key]; ok {
		mf.mf = fn
		return nil
	}
	return fmt.Errorf("not found mock key type:%s", key)
}

func (m *Mock) genMockTag(mf *mockField) error {
	var (
		values []string
		key    string
		value  string
		tl     TagLevel
		err    error
	)
	for i, tag := range mf.tempTags {
		values = strings.SplitN(tag, mockTagSeparator, 2)
		key, value = values[0], ""
		if len(values) > 1 {
			value = values[1]
		}
		fn, ok := m.tagFactory[values[0]]
		if !ok {
			return fmt.Errorf("not support the mock tag:%s", values[0])
		}
		tl, err = fn(mf.rt, key, value)
		if err != nil {
			return err
		}
		mf.tags[key] = tl
		if key == MockInto {
			mf.intoTags = mf.tempTags[i+1:]
			break
		}
	}
	return nil
}
func (m *Mock) parseSliceTag(ctx context.Context, mf *mockField) error {
	if err := m.genMockTag(mf); err != nil {
		return err
	}
	if mf.tags.Key(MockSkip).Exists() {
		return nil
	}
	mf.mf = m.mockFactory[makeSlice]
	//for slice element
	if !mf.tags.Key(MockInto).Exists() {
		return nil
	}
	rt, _ := m.Indirect(mf.rt.Elem())
	switch rt.Kind() {
	case reflect.Struct:
		return m.parseIntoStruct(ctx, mf, mf.rt.Elem())
	default:
		return m.parseIntoBase(mf, mf.rt.Elem())
	}
}
func (m *Mock) parseStructTag(ctx context.Context, mf *mockField) error {
	if err := m.genMockTag(mf); err != nil {
		return err
	}
	if mf.tags.Key(MockSkip).Exists() {
		return nil
	}
	if mf.isPtr { //init struct ptr
		mf.mf = m.mockFactory[makeStruct]
	}
	//for struct element
	if !mf.tags.Key(MockInto).Exists() {
		return nil
	}
	return m.parseStruct(ctx, mf, mf.rt)
}
func (m *Mock) parseBaseTag(mf *mockField) error {
	if err := m.genMockTag(mf); err != nil {
		return err
	}
	if mf.tags.Key(MockSkip).Exists() {
		return nil
	}
	return m.genMockFunc(mf)
}
func (m *Mock) parseTag(ctx context.Context, mf *mockField, tag reflect.StructTag) error {
	mf.tempTags = strings.Split(tag.Get(m.tag), m.separator)
	switch mf.rk {
	case reflect.Slice:
		return m.parseSliceTag(ctx, mf)
	case reflect.Struct:
		return m.parseStructTag(ctx, mf)
	default:
		return m.parseBaseTag(mf)
	}
}
