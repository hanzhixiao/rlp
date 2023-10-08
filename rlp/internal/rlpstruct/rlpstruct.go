package rlpstruct

import (
	"fmt"
	"reflect"
	"strings"
)

type NilKind uint8

type Field struct {
	Name     string
	Index    int
	Exported bool
	Type     Type
	Tag      string
}

type TagError struct {
	StructType string

	// These are set by this package.
	Field string
	Tag   string
	Err   string
}

func (e TagError) Error() string {
	field := "field " + e.Field
	if e.StructType != "" {
		field = e.StructType + "." + e.Field
	}
	return fmt.Sprintf("rlp: invalid struct tag %q for %s (%s)", e.Tag, field, e.Err)
}

const (
	NilKindString NilKind = 0x80
	NilKindList   NilKind = 0xC0
)

type Type struct {
	Name      string
	Kind      reflect.Kind
	IsEncoder bool  // whether type implements rlp.Encoder
	IsDecoder bool  // whether type implements rlp.Decoder
	Elem      *Type // non-nil for Kind values of Ptr, Slice, Array
}

func (t Type) DeaultNilValue() NilKind {
	k := t.Kind
	if isUint(k) || k == reflect.String || k == reflect.Bool || isByteArray(t) {
		return NilKindString
	}
	return NilKindList
}

// Tags represents struct tags.
type Tags struct {
	// rlp:"nil" controls whether empty input results in a nil pointer.
	// nilKind is the kind of empty value allowed for the field.
	NilKind NilKind
	NilOK   bool

	// rlp:"optional" allows for a field to be missing in the input list.
	// If this is set, all subsequent fields must also be optional.
	Optional bool

	// rlp:"tail" controls whether this field swallows additional list elements. It can
	// only be set for the last field, which must be of slice type.
	Tail bool

	// rlp:"-" ignores fields.
	Ignored bool
}

// 不懂为什么要去除
func ProcessFields(allStructFields []Field) ([]Field, []Tags, error) {
	lastPublic := lastPublicField(allStructFields)

	var fields []Field
	var tags []Tags
	for _, field := range allStructFields { //为什么只要未命名数据类型
		if !field.Exported {
			continue
		}
		ts, err := parseTag(field, lastPublic)
		if err != nil {
			return nil, nil, err
		}
		if ts.Ignored {
			continue
		}
		fields = append(fields, field)
		tags = append(tags, ts)
	}
	//为什么optional的field后面必须都是optional
	var anyOptional bool
	var firstOptionalName string
	for i, ts := range tags {
		name := fields[i].Name
		if ts.Optional || ts.Tail {
			if !anyOptional {
				firstOptionalName = name
			}
			anyOptional = true
		} else {
			if anyOptional {
				msg := fmt.Sprintf("must be optional because preceding field %q is optional", firstOptionalName)
				return nil, nil, TagError{Field: name, Err: msg}
			}
		}
	}

	return fields, tags, nil
}

func parseTag(field Field, lastPublic int) (Tags, error) {
	name := field.Name
	tag := reflect.StructTag(field.Tag)
	var ts Tags
	for _, t := range strings.Split(tag.Get("rlp"), ",") {
		switch strings.TrimSpace(t); t {
		case "":
		case "-":
			ts.Ignored = true
		case "nil", "nilString", "nilList":
			ts.NilOK = true
			if field.Type.Kind != reflect.Ptr {
				return ts, TagError{Field: name, Tag: t, Err: "field is not a pointer"}
			}
			switch t {
			case "nil":
				ts.NilKind = field.Type.DeaultNilValue()
			case "nilString":
				ts.NilKind = NilKindString
			case "nilList":
				ts.NilKind = NilKindList
			}
		case "optional":
			ts.Optional = true
			if ts.Tail {
				return ts, TagError{Field: name, Tag: t, Err: `also has "tail" tag`}
			}
		case "tail":
			ts.Tail = true
			if field.Index != lastPublic {
				return ts, TagError{Field: name, Tag: t, Err: "must be on last field"}
			}
			if ts.Optional {
				return ts, TagError{Field: name, Tag: t, Err: `also has "optional" tag`}
			}
			//为什么tag为tail的字段必须是切片
			if field.Type.Kind != reflect.Slice {
				return ts, TagError{Field: name, Tag: t, Err: "field type is not slice"}
			}
		default:
			return ts, TagError{Field: name, Tag: t, Err: "unknown tag"}
		}
	}
	return ts, nil
}

func lastPublicField(allStructFields []Field) int {
	lastPublic := 0
	for _, Field := range allStructFields {
		if Field.Exported {
			lastPublic = Field.Index
		}
	}
	return lastPublic
}

func isUint(k reflect.Kind) bool {
	return k >= reflect.Uint && k <= reflect.Uintptr
}

func isByte(typ Type) bool {
	return typ.Kind == reflect.Uint8 && !typ.IsEncoder
}

func isByteArray(typ Type) bool {
	return (typ.Kind == reflect.Slice || typ.Kind == reflect.Array) && isByte(*typ.Elem)
}
