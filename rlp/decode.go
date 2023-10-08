package rlp

import (
	"math/big"
	"reflect"
)

type Kind int8

const (
	Byte Kind = iota
	String
	List
)

var (
	decoderInterface = reflect.TypeOf(new(Decoder)).Elem()
	bigInt           = reflect.TypeOf(big.Int{})
)

type Stream interface {
}

type Decoder interface {
	DecodeRLP(*Stream) error
}
