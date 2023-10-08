package rlp

import (
	"awesomeProject/rlp/internal/rlpstruct"
	"errors"
	"fmt"
	"io"
	"math/big"
	"reflect"
)

type listhead struct {
	offset int // index of this header in string data
	size   int // total size of encoded data (including list headers)
}

func (l *listhead) encode(dst []byte) []byte {
	return dst[:puthead(dst, 0xC0, 0xF7, uint64(l.size))]
}

type Encoder interface {
	EncodeRLP(io.Writer) error
}

func Encode(w io.Writer, val interface{}) error {
	if buf := encBufferFromWriter(w); buf != nil {
		return buf.encode(val)
	}
	buf := getEncBuffer()
	defer encBufferPool.Put(buf)
	if err := buf.encode(val); err != nil {
		return nil
	}
	return buf.writeTo(w)
}

func puthead(dst []byte, smalltag, largetag byte, size uint64) int {
	if size < 56 {
		dst[0] = smalltag + byte(size)
		return 1
	}
	sizesize := putint(dst, size)
	dst = append(dst[1:], largetag+byte(sizesize))
	return sizesize + 1
}

var ErrNegativeBigInt = errors.New("rlp: cannot encode negative big.Int")

// 不知道干嘛的
var encoderInterface = reflect.TypeOf(new(Encoder)).Elem()

func makeWriter(p reflect.Type, tags rlpstruct.Tags) (writer, error) {
	kind := p.Kind()
	switch {
	//什么时候是rawvalue
	case p == rawValueType:
		return writeRawValue, nil
	case p.AssignableTo(reflect.PtrTo(bigInt)): //同下
		return writeBigIntPtr, nil
	case p.AssignableTo(bigInt): //为什么是assignableto大整数，不是等于大整数
		return writeBigIntNoPtr, nil
	case kind == reflect.Ptr:
		return makePtrWriter(p, tags)
	case reflect.PtrTo(p).Implements(encoderInterface):
		return makeEncoderWriter(p), nil
	case isUint(kind):
		return writeUint, nil
	case kind == reflect.Bool:
		return writeBool, nil
	case kind == reflect.String:
		return writeString, nil
	case kind == reflect.Slice && isByte(p.Elem()):
		return writeBytes, nil
	case kind == reflect.Array && isByte(p.Elem()):
		return makeByteArrayWriter(p), nil
	case kind == reflect.Slice || kind == reflect.Array:
		return makeSliceWriter(p, tags)
	case kind == reflect.Struct:
		return makeStructWriter(p)
	case kind == reflect.Interface:
		return writeInterface, nil
	default:
		return nil, fmt.Errorf("rlp: type %v is not RLP-serializable", p)
	}
}

func makeEncoderWriter(typ reflect.Type) writer {
	if typ.Implements(encoderInterface) {
		return func(val reflect.Value, w *encBuffer) error {
			return val.Interface().(Encoder).EncodeRLP(w)
		}
	}
	w := func(val reflect.Value, w *encBuffer) error {
		if !val.CanAddr() {
			// package json simply doesn't call MarshalJSON for this case, but encodes the
			// value as if it didn't implement the interface. We don't want to handle it that
			// way.
			return fmt.Errorf("rlp: unadressable value of type %v, EncodeRLP is pointer method", val.Type())
		}
		return val.Addr().Interface().(Encoder).EncodeRLP(w)
	}
	return w
}

func writeBigIntPtr(val reflect.Value, w *encBuffer) error {
	ptr := val.Interface().(*big.Int)
	if ptr == nil {
		w.str = append(w.str, 0x80)
		return nil
	}
	if ptr.Sign() == -1 {
		return ErrNegativeBigInt
	}
	w.writeBigInt(ptr)
	return nil
}

func writeBigIntNoPtr(val reflect.Value, w *encBuffer) error {
	i := val.Interface().(big.Int)
	if i.Sign() == -1 {
		return ErrNegativeBigInt
	}
	w.writeBigInt(&i)
	return nil
}

func makeStructWriter(p reflect.Type) (writer, error) {
	flieds, err := structFlieds(p)
	if err != nil {
		return nil, err
	}

	for _, flied := range flieds {
		if flied.info.writerErr != nil {
			return nil, structFieldError{
				typ:   p,
				field: flied.index,
				err:   flied.info.writerErr,
			}
		}
	}
	var writer writer
	firstOptionalField := firstOptionalField(flieds)
	if firstOptionalField == len(flieds) {
		writer = func(value reflect.Value, buffer *encBuffer) error {
			offset := buffer.list()
			for _, flied := range flieds {
				if err := flied.info.writer(value.Field(flied.index), buffer); err != nil {
					return err
				}
			}
			buffer.endlist(offset)
			return nil
		}
		return writer, nil
	} else {
		writer = func(value reflect.Value, buffer *encBuffer) error {
			lastField := len(flieds) - 1
			for ; lastField >= firstOptionalField; lastField-- {
				//已经确定是optional的了，为什么还要判断是不是零值,好像意思是只有零值才能不编码
				if !value.Field(flieds[lastField].index).IsZero() {
					break
				}
			}
			for i := 0; i <= lastField; i++ {
				offset := buffer.list()
				if err := flieds[i].info.writer(value.Field(flieds[i].index), buffer); err != nil {
					return err
				}
				buffer.endlist(offset)
			}
			return nil
		}
	}
	return writer, nil
}

func makeSliceWriter(p reflect.Type, tags rlpstruct.Tags) (writer, error) {
	etpyeinfo := theTC.infoWhileGenerating(p.Elem(), tags)
	if etpyeinfo.writerErr != nil {
		return nil, etpyeinfo.writerErr
	}
	var wfn writer
	//没理解tail的作用
	if tags.Tail {
		wfn = func(value reflect.Value, buffer *encBuffer) error {
			for i := 0; i < value.Len(); i++ {
				if err := etpyeinfo.writer(value.Index(i), buffer); err != nil {
					return err
				}
			}
			return nil
		}
	} else {
		wfn = func(value reflect.Value, buffer *encBuffer) error {
			vlen := value.Len()
			if vlen == 0 {
				buffer.str = append(buffer.str, 0xC0)
				return nil
			}
			listOffset := buffer.list()
			for i := 0; i < vlen; i++ {
				if err := etpyeinfo.writer(value.Index(i), buffer); err != nil {
					return err
				}
			}
			buffer.endlist(listOffset)
			return nil
		}
	}
	return wfn, nil
}

func writeRawValue(value reflect.Value, buffer *encBuffer) error {
	buffer.str = append(buffer.str, value.Bytes()...)
	return nil
}

func writeUint(value reflect.Value, buffer *encBuffer) error {
	buffer.writeUint64(value.Uint())
	return nil
}

func writeBool(value reflect.Value, buffer *encBuffer) error {
	buffer.writeBool(value.Bool())
	return nil
}

func writeString(value reflect.Value, buffer *encBuffer) error {
	str := value.String()
	l := len(str)
	if l == 1 && str[0] <= 0x7f {
		buffer.str = append(buffer.str, str[0])
	} else {
		buffer.encodeStringHeader(l)
		buffer.str = append(buffer.str, str...)
	}
	return nil
}

func writeBytes(value reflect.Value, buffer *encBuffer) error {
	buffer.writeBytes(value.Bytes())
	return nil
}

func writeInterface(value reflect.Value, buffer *encBuffer) error {
	if value.IsNil() {
		buffer.str = append(buffer.str, 0xC0)
		return nil
	}
	eval := value.Elem()
	writer, err := cachedWriter(eval.Type())
	if err != nil {
		return err
	}
	return writer(eval, buffer)
}

func makePtrWriter(typ reflect.Type, ts rlpstruct.Tags) (writer, error) {
	nilEncoding := byte(0xC0)
	if typeNilKind(typ.Elem(), ts) == String {
		nilEncoding = 0x80
	}
	typeinfo := theTC.infoWhileGenerating(typ.Elem(), rlpstruct.Tags{})
	if typeinfo.writerErr != nil {
		return nil, typeinfo.writerErr
	}
	writer := func(value reflect.Value, buffer *encBuffer) error {
		if ev := value.Elem(); ev.IsValid() {
			return typeinfo.writer(ev, buffer)
		}
		buffer.str = append(buffer.str, nilEncoding)
		return nil
	}
	return writer, nil
}

func makeByteArrayWriter(p reflect.Type) writer {
	switch p.Len() {
	case 0:
		return writeLengthZeroByteArray
	case 1:
		return writeLengthOneByteArray
	default:
		return func(value reflect.Value, buffer *encBuffer) error {
			//什么情况下数组是可寻址的
			//var b = [3]int{1, 2, 3}
			//fmt.Print(reflect.ValueOf(&b).Elem().Kind(), reflect.ValueOf(&b).Elem())
			//a(reflect.ValueOf(&b).Elem())
			if !value.CanAddr() {
				copy := reflect.New(value.Type()).Elem()
				copy.Set(value)
				value = copy
			}
			//不是按源码写的
			slice := byteArrayBytes(value, value.Len())
			buffer.encodeStringHeader(len(slice))
			buffer.str = append(buffer.str, slice...)
			return nil
		}
	}
}

func writeLengthOneByteArray(value reflect.Value, buffer *encBuffer) error {
	b := byte(value.Index(0).Uint())
	if b <= 0x7f {
		buffer.str = append(buffer.str, b)
	} else {
		buffer.str = append(buffer.str, 0x81)
		buffer.str = append(buffer.str, b)
	}
	return nil
}

func writeLengthZeroByteArray(value reflect.Value, buffer *encBuffer) error {
	buffer.str = append(buffer.str, 0x80)
	return nil
}

func putint(b []byte, u uint64) int {
	switch {
	case u < (1 << 8):
		b[0] = byte(u)
		return 1
	case u < (1 << 16):
		b[0] = byte(u >> 8)
		b[1] = byte(u)
		return 2
	case u < (1 << 24):
		b[0] = byte(u >> 16)
		b[1] = byte(u >> 8)
		b[2] = byte(u)
		return 3
	case u < (1 << 32):
		b[0] = byte(u >> 24)
		b[1] = byte(u >> 16)
		b[2] = byte(u >> 8)
		b[3] = byte(u)
		return 4
	case u < (1 << 40):
		b[0] = byte(u >> 32)
		b[1] = byte(u >> 24)
		b[2] = byte(u >> 16)
		b[3] = byte(u >> 8)
		b[4] = byte(u)
		return 5
	case u < (1 << 48):
		b[0] = byte(u >> 40)
		b[1] = byte(u >> 32)
		b[2] = byte(u >> 24)
		b[3] = byte(u >> 16)
		b[4] = byte(u >> 8)
		b[5] = byte(u)
		return 6
	case u < (1 << 56):
		b[0] = byte(u >> 48)
		b[1] = byte(u >> 40)
		b[2] = byte(u >> 32)
		b[3] = byte(u >> 24)
		b[4] = byte(u >> 16)
		b[5] = byte(u >> 8)
		b[6] = byte(u)
		return 7
	default:
		b[0] = byte(u >> 56)
		b[1] = byte(u >> 48)
		b[2] = byte(u >> 40)
		b[3] = byte(u >> 32)
		b[4] = byte(u >> 24)
		b[5] = byte(u >> 16)
		b[6] = byte(u >> 8)
		b[7] = byte(u)
		return 8
	}
}

func intsize(uint642 uint64) int {
	for size := 1; ; size++ {
		if uint642 >>= 8; uint642 == 0 {
			return size
		}
	}
}
