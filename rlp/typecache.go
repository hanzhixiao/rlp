package rlp

import (
	"awesomeProject/rlp/internal/rlpstruct"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
)

// 为什么要继承Type
type typekey struct {
	reflect.Type
	rlpstruct.Tags
}

type field struct {
	index    int
	info     *typeinfo
	optional bool //不知道干嘛的
}

type decoder func()

type writer func(reflect.Value, *encBuffer) error

var theTC = newTypeCache()

func newTypeCache() *typeCache {
	tempTC := new(typeCache)
	tempTC.cur.Store(make(map[typekey]*typeinfo))
	return tempTC
}

type typeinfo struct {
	decoder    decoder
	decoderErr error
	writer     writer
	writerErr  error
}

type structFieldError struct {
	typ   reflect.Type
	field int
	err   error
}

func (e structFieldError) Error() string {
	return fmt.Sprintf("%v (struct field %v.%s)", e.err, e.typ, e.typ.Field(e.field).Name)
}

type typeCache struct {
	//cur是干什么的
	cur atomic.Value

	// This lock synchronizes writers.
	mu   sync.Mutex
	next map[typekey]*typeinfo
}

func (c *typeCache) info(typ reflect.Type) *typeinfo {
	key := typekey{Type: typ}
	if info := c.cur.Load().(map[typekey]*typeinfo)[key]; info != nil {
		return info
	}

	// Not in the cache, need to generate info for this type.
	return c.generate(typ, rlpstruct.Tags{})
}
func (c *typeCache) generate(typ reflect.Type, tags rlpstruct.Tags) *typeinfo {
	c.mu.Lock()
	defer c.mu.Unlock()

	cur := c.cur.Load().(map[typekey]*typeinfo)
	if info := cur[typekey{typ, tags}]; info != nil {
		return info
	}

	// Copy cur to next.
	c.next = make(map[typekey]*typeinfo, len(cur)+1)
	for k, v := range cur {
		c.next[k] = v
	}

	// Generate.
	info := c.infoWhileGenerating(typ, tags)

	// next -> cur
	c.cur.Store(c.next)
	c.next = nil
	return info
}

func (c *typeCache) infoWhileGenerating(typ reflect.Type, tags rlpstruct.Tags) *typeinfo {
	key := typekey{typ, tags}
	if info := c.next[key]; info != nil {
		return info
	}
	// Put a dummy value into the cache before generating.
	// If the generator tries to lookup itself, it will get
	// the dummy value and won't call itself recursively.
	info := new(typeinfo)
	c.next[key] = info
	info.generate(typ, tags)
	return info
}

func (i *typeinfo) generate(typ reflect.Type, tags rlpstruct.Tags) {
	//i.decoder, i.decoderErr = makeDecoder(typ, tags)
	i.writer, i.writerErr = makeWriter(typ, tags)
}

func cachedWriter(typ reflect.Type) (writer, error) {
	info := theTC.info(typ)
	return info.writer, info.writerErr
}

func structFlieds(p reflect.Type) (fields []*field, err error) {
	//为什么要先转为rlpstruct.Field类型，Field类型有什么用
	var allStructFields []rlpstruct.Field
	for i := 0; i < p.NumField(); i++ {
		rf := p.Field(i)
		allStructFields = append(allStructFields, rlpstruct.Field{
			Name:     rf.Name,
			Index:    i,
			Exported: rf.PkgPath == "",
			Type:     *rtypeToStructType(rf.Type, nil),
			Tag:      string(rf.Tag),
		})
	}
	structFields, structTags, err := rlpstruct.ProcessFields(allStructFields)
	if err != nil {
		if tagErr, ok := err.(rlpstruct.TagError); ok {
			tagErr.StructType = p.String()
			return nil, tagErr
		}
		return nil, err
	}

	for i, sf := range structFields {
		typ := p.Field(sf.Index).Type
		tag := structTags[i]
		info := theTC.infoWhileGenerating(typ, tag)
		fields = append(fields, &field{
			index:    sf.Index,
			info:     info,
			optional: tag.Optional,
		})
	}
	return fields, nil
}

func firstOptionalField(fields []*field) int {
	for i, f := range fields {
		if f.optional {
			return i
		}
	}
	return len(fields)
}

// 可以优化
func typeNilKind(p reflect.Type, ts rlpstruct.Tags) Kind {
	//为什么要转为rlpstruct
	styp := rtypeToStructType(p, nil)

	var nk rlpstruct.NilKind
	if ts.NilOK {
		//什么时候Nilok是true，什么时候是fasle
		nk = ts.NilKind
	} else {
		nk = styp.DeaultNilValue()
	}
	switch nk {
	case rlpstruct.NilKindList:
		return List
	case rlpstruct.NilKindString:
		return String
	default: //什么时候既不是nilList也不是 nilString
		panic("invalid nil kind value")
	}
}

func rtypeToStructType(p reflect.Type, rec map[reflect.Type]*rlpstruct.Type) *rlpstruct.Type {
	k := p.Kind()
	if k == reflect.Invalid {
		panic("invalid kind")
	}
	if rec[p] != nil {
		return rec[p]
	}
	if rec == nil {
		rec = make(map[reflect.Type]*rlpstruct.Type)
	}
	t := &rlpstruct.Type{
		Name: p.String(),
		Kind: k,
		//这两个属性什么时候用
		IsEncoder: p.Implements(encoderInterface),
		IsDecoder: p.Implements(decoderInterface),
	}
	rec[p] = t
	if k == reflect.Array || k == reflect.Slice || k == reflect.Ptr {
		t.Elem = rtypeToStructType(p.Elem(), rec)
	}
	return t
}

func isUint(kind reflect.Kind) bool {
	return kind >= reflect.Uint && kind <= reflect.Uintptr
}

func isByte(typ reflect.Type) bool {
	//为什么不能是encoder
	return typ.Kind() == reflect.Uint8 && !typ.Implements(encoderInterface)
}
