package jsontype

import (
  "bytes"
  "encoding/json"
  "errors"
  "fmt"
  "github.com/xpwu/go-config/config/flatfield"
  "reflect"
)

// indirect walks down v allocating pointers as needed,
// until it gets to a non-pointer.
// If decodingNull is true, indirect stops at the first settable pointer so it
// can be set to nil.
func indirect(v reflect.Value, decodingNull bool) reflect.Value {
  // Issue #24153 indicates that it is generally not a guaranteed property
  // that you may round-trip a reflect.Value by calling Value.Addr().Elem()
  // and expect the value to still be settable for values derived from
  // unexported embedded struct fields.
  //
  // The logic below effectively does this when it first addresses the value
  // (to satisfy possible pointer methods) and continues to dereference
  // subsequent pointers as necessary.
  //
  // After the first round-trip, we set v back to the original value to
  // preserve the original RW flags contained in reflect.Value.
  v0 := v
  haveAddr := false

  // If v is a named type and is addressable,
  // start with its address, so that if the type has pointer methods,
  // we find them.
  if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
    haveAddr = true
    v = v.Addr()
  }
  for {
    // Load value from interface, but only if the result will be
    // usefully addressable.
    if v.Kind() == reflect.Interface && !v.IsNil() {
      e := v.Elem()
      if e.Kind() == reflect.Ptr && !e.IsNil() && (!decodingNull || e.Elem().Kind() == reflect.Ptr) {
        haveAddr = false
        v = e
        continue
      }
    }

    if v.Kind() != reflect.Ptr {
      break
    }

    if decodingNull && v.CanSet() {
      break
    }

    // Prevent infinite loop if v is an interface pointing to its own address:
    //     var v interface{}
    //     v = &v
    if v.Elem().Kind() == reflect.Interface && v.Elem().Elem() == v {
      v = v.Elem()
      break
    }
    if v.IsNil() {
      v.Set(reflect.New(v.Type().Elem()))
    }

    if haveAddr {
      v = v0 // restore original value after round-trip Value.Addr().Elem()
      haveAddr = false
    } else {
      v = v.Elem()
    }
  }

  return v
}

func toString(value reflect.Value) (str string, ok bool) {
  defer func() {
    if r := recover(); r != nil {
      ok = false
    }
  }()

  strT := reflect.TypeOf("")
  if value.Type().ConvertibleTo(strT) {
    return value.Convert(strT).String(), true
  }
  if value.Type().Implements(reflect.TypeOf(fmt.Stringer(nil))) && value.CanInterface() {
    return value.Interface().(fmt.Stringer).String(), true
  }
  return "", false
}

var end = errors.New("end")

func FromInterface(i interface{}, tryTag func(tag reflect.StructTag) (key, tips string)) Type {
  return FromValue(reflect.ValueOf(i), tryTag)
}

func FromValue(value reflect.Value, tryTag func(tag reflect.StructTag) (key, tips string)) Type {
  if !value.IsValid() {
    return nil
  }

  switch value.Kind() {
  case reflect.Ptr, reflect.Interface:
    if value.IsNil() {
      return Null{}
    }
    return FromValue(value.Elem(), tryTag)
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint,
    reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
    return NumberWithInt(value.Convert(reflect.TypeOf(int64(0))).Int())
  case reflect.Float32, reflect.Float64:
    return NumberWithFloat(value.Convert(reflect.TypeOf(float64(0))).Float())
  case reflect.String:
    nt := reflect.TypeOf(json.Number(""))
    // json.Number
    if nt == value.Type() {
      return NumberWithStr(value.String())
    }

    return String(value.String())
  case reflect.Bool:
    return Bool(value.Bool())
  case reflect.Slice:
    if value.IsNil() {
      return Null{}
    }
    fallthrough
  case reflect.Array:
    ret := make([]Type, 0, value.Len())
    for i := 0; i < value.Len(); i++ {
      if v := FromValue(value.Index(i), tryTag); v != nil {
        ret = append(ret, v)
      }
    }
    return Slice(ret)
  case reflect.Map:
    ret := make([]objectElem, 0, value.Len())
    it := value.MapRange()
    for it.Next() {
      if !it.Key().IsValid() || !it.Value().IsValid() {
        continue
      }
      key, ok := toString(it.Key())
      v := FromValue(it.Value(), tryTag)
      if !ok || v == nil {
        continue
      }

      ret = append(ret, objectElem{Key: key, Value: v})
    }
    return Object(ret)
  //case reflect.Interface:
  //  if value.IsNil() {
  //    return nil
  //  }
  //  return FromValue(value.Elem(), tryTag)
  //  //if value.Type().Implements(reflect.TypeOf(fmt.Stringer(nil))) && value.CanInterface() {
  //  //  return String(value.Interface().(fmt.Stringer).String())
  //  //}
  case reflect.Complex64, reflect.Complex128:
    return String(fmt.Sprint(value.Complex()))

  case reflect.Struct:
    if !value.CanInterface() {
      return nil
    }
    fields, err := flatfield.Flatten(value.Interface(),
      flatfield.Name(func(tag reflect.StructTag) (name string) {
        name, _ = tryTag(tag)
        return
      }))
    if err != nil {
      return nil
    }

    ret := make([]objectElem, 0, len(fields))
    for _, f := range fields {
      if !f.HasValue {
        continue
      }
      key, tips := tryTag(f.SField.Tag)
      if key == "" {
        key = f.SField.Name
      }
      v := FromValue(value.FieldByIndex(f.SField.Index), tryTag)
      if v == nil {
        continue
      }

      ret = append(ret, objectElem{
        Key:   key,
        Value: v,
        Tips:  tips,
      })
    }

    return Object(ret)
  }

  return nil
}

func FromJsonDecoder(decoder *json.Decoder) (value Type, err error) {
  token, err := decoder.Token()
  if err != nil {
    return nil, err
  }
  if token == nil {
    return Null{}, nil
  }

  switch t := token.(type) {
  case json.Number:
    return NumberWithStr(t.String()), nil
  case string:
    return String(t), nil
  case bool:
    return Bool(t), nil
  case float64:
    return NumberWithFloat(t), nil
  case json.Delim:
    switch t.String() {
    case "[":
      ret := make(Slice, 0)
      for {
        v, err := FromJsonDecoder(decoder)
        if err != nil {
          break
        }
        ret = append(ret, v)
      }
      if err == end {
        return ret, nil
      }
      return nil, err

    case "{":
      ret := make(Object, 0)
      for {
        key, err := FromJsonDecoder(decoder)
        // may be end
        if err != nil {
          break
        }
        v, err := FromJsonDecoder(decoder)
        // error
        if err != nil {
          return nil, err
        }
        ret = append(ret, objectElem{Key: string(key.(String)), Value: v})
      }
      if err == end {
        return ret, nil
      }
      return nil, err

    case "]", "}":
      return nil, end
    }
  }

  return nil, errors.New("not support type -- " + reflect.TypeOf(token).String())
}

func FromJson(jsn []byte) (value Type, err error) {
  decoder := json.NewDecoder(bytes.NewReader(jsn))
  // note: 为了延后解析int/float
  decoder.UseNumber()

  return FromJsonDecoder(decoder)
}
