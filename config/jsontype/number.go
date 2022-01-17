package jsontype

import (
  "encoding/json"
  "reflect"
  "strconv"
)

func (n Number) Kind() Kind {
  return NumberK
}

func NumberWithStr(str string) Number {
  return Number{str: &str}
}

func NumberWithInt(i int64) Number {
  return Number{i: &i}
}

func NumberWithFloat(f float64) Number {
  return Number{f: &f}
}

func (n Number) Int() (v int64, err error) {
  if n.i != nil {
    return *n.i, nil
  }
  if n.f != nil {
    return int64(*n.f), nil
  }
  if n.str != nil {
    r, err := strconv.ParseInt(*n.str, 10, 64)
    if err != nil {
      return 0, err
    }
    return r, nil
  }

  return 0, nil
}

func (n Number) Uint() (v uint64, err error) {
  if n.str != nil {
    r, err := strconv.ParseUint(*n.str, 10, 64)
    if err != nil {
      return 0, err
    }
    return r, nil
  }
  if n.i != nil {
    return uint64(*n.i), nil
  }
  if n.f != nil {
    return uint64(*n.f), nil
  }

  return 0, nil
}

func (n Number) Float() (v float64, err error) {
  if n.f != nil {
    return *n.f, nil
  }
  if n.i != nil {
    return float64(*n.i), nil
  }

  if n.str != nil {
    r, err := strconv.ParseFloat(*n.str, 64)
    if err != nil {
      return 0, err
    }
    return r, nil
  }

  return 0, nil
}

func (n Number) String() string {
  if n.str != nil {
    return *n.str
  }
  if n.i != nil {
    return strconv.FormatInt(*n.i, 10)
  }
  if n.f != nil {
    return strconv.FormatFloat(*n.f, 'f', -1, 64)
  }

  // zero value
  return "0"
}

var jsonNumberType = reflect.TypeOf(json.Number(""))
var thisNumberType = reflect.TypeOf(NumberWithInt(0))

func (n Number) Value(i interface{}, name func(tag reflect.StructTag) (name string)) error {
  value := reflect.ValueOf(i)
  if !value.IsValid() {
    // skip
    return nil
  }
  value = indirect(value, false)

  if value.Type() == reflect.TypeOf(n) {
    value.Set(reflect.ValueOf(n))
    return nil
  }

  switch value.Kind() {
  default:
    if value.Kind() == reflect.String && value.Type() == jsonNumberType {
      // s must be a valid number, because it's
      // already been tokenized.
      value.SetString(n.String())
      return nil
    }
    return &json.UnmarshalTypeError{Value: "number", Type: value.Type()}

  case reflect.Interface:
    if value.NumMethod() != 0 {
      return &json.UnmarshalTypeError{Value: "number", Type: value.Type()}
    }
    value.Set(reflect.ValueOf(n))

  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    v, err := n.Int()
    if err != nil || value.OverflowInt(v) {
      return &json.UnmarshalTypeError{Value: "number " + n.String(), Type: value.Type()}
    }
    value.SetInt(v)

  case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
    v, err := n.Uint()
    if err != nil || value.OverflowUint(v) {
      return &json.UnmarshalTypeError{Value: "number " + n.String(), Type: value.Type()}
    }
    value.SetUint(v)

  case reflect.Float32, reflect.Float64:
    v, err := n.Float()
    if err != nil || value.OverflowFloat(v) {
      return &json.UnmarshalTypeError{Value: "number " + n.String(), Type: value.Type()}
    }
    value.SetFloat(v)
  }

  return nil
}

func (n Number) MarshalJSON() ([]byte, error) {
  return []byte(n.String()), nil
}
