package rlp

import (
	"reflect"
	"unsafe"
)

func byteArrayBytes(value reflect.Value, length int) []byte {
	var v []byte
	hr := (*reflect.SliceHeader)(unsafe.Pointer(&v))
	hr.Data = value.UnsafeAddr()
	hr.Len = length
	hr.Cap = length
	return v
}
