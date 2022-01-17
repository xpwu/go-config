package jsontype

import (
	"reflect"
)

type Kind uint

const (
  NullK Kind = iota
	NumberK
	StringK
	BoolK
	SliceK
	ObjectK
)

type Type interface {
	Kind() Kind
	Value(i interface{}, name func(tag reflect.StructTag) (name string)) error
}

type Null struct {}
type Number struct {str *string; i *int64; f *float64}
type String string
type Bool bool
type Slice []Type
// 为了有序性，所以没有使用map
type Object []objectElem

type objectElem struct {Key string; Value Type; Tips string}

