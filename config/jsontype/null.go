package jsontype

import "reflect"

func (n Null) Kind() Kind {
  return NullK
}

func (n Null) Value(i interface{}, name func(tag reflect.StructTag) (name string)) error {
  value := reflect.ValueOf(i)
  if !value.IsValid() {
    // skip
    return nil
  }
  value = indirect(value, true)
  switch value.Kind() {
  case reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice:
    value.Set(reflect.Zero(value.Type()))
    // otherwise, ignore null for primitives/string
  }
  return nil
}

func (n Null) MarshalJSON() ([]byte, error) {
  return []byte("null"), nil
}

