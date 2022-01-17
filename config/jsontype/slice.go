package jsontype

import (
  "encoding/json"
  "reflect"
)

func (s Slice) Kind() Kind {
  return SliceK
}

func (s Slice) Value(i interface{}, name func(tag reflect.StructTag) (name string)) error {
  value := reflect.ValueOf(i)
  if !value.IsValid() {
    // skip
    return nil
  }

  value = indirect(value, false)

  // Check type of target.
  switch value.Kind() {
  default:
    return &json.UnmarshalTypeError{Value: "array", Type: value.Type()}

  case reflect.Interface:
    if value.NumMethod() != 0 {
      return &json.UnmarshalTypeError{Value: "array", Type: value.Type()}
    }
    value.Set(reflect.ValueOf(make([]interface{}, len(s), len(s))))
  case reflect.Slice:
    value.Set(reflect.MakeSlice(value.Type(), len(s), len(s)))
  case reflect.Array:
    for i := len(s); i < value.Len(); i++ {
      // len(value) > len(s), zero value
      value.Index(i).Set(reflect.Zero(value.Type().Elem()))
    }
  }

  for i := 0; i < value.Len() && i < len(s); i++ {
    if err := s[i].Value(value.Index(i), name); err != nil {
      return err
    }
  }

  return nil
}

