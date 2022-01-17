package jsontype

import (
  "bytes"
  "encoding/json"
  "github.com/xpwu/go-config/config/flatfield"
  "reflect"
  "strconv"
)

func (o Object) Kind() Kind {
  return ObjectK
}

func (o Object) valueInterface(value reflect.Value, name func(tag reflect.StructTag)(name string)) error {
  if value.NumMethod() != 0 {
    return &json.UnmarshalTypeError{Value: "object", Type: value.Type()}
  }
  m := make(map[string]interface{})
  for _, e := range o {
    var i interface{}
    if err := e.Value.Value(reflect.ValueOf(&i), name); err != nil {
      return err
    }
    m[e.Key] = i
  }
  value.Set(reflect.ValueOf(m))
  return nil
}

func (o Object) valueMap(value reflect.Value, name func(tag reflect.StructTag)(name string)) error {
  // Map key must either have string kind, have an integer kind
  t := value.Type()
  switch t.Key().Kind() {
  case reflect.String,
    reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
    reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
  default:
    return &json.UnmarshalTypeError{Value: "object", Type: t}
  }
  if value.IsNil() {
    value.Set(reflect.MakeMap(t))
  }

  for _,e := range o {
    kt := t.Key()
    var key reflect.Value
    switch kt.Kind() {
    case  reflect.String:
      key = reflect.ValueOf(e.Key)
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
      n, err := strconv.ParseInt(e.Key, 10, 64)
      if err != nil || reflect.Zero(kt).OverflowInt(n) {
        return &json.UnmarshalTypeError{Value: "number " + e.Key, Type: kt}
      }
      key = reflect.ValueOf(n).Convert(kt)
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
      n, err := strconv.ParseUint(e.Key, 10, 64)
      if err != nil || reflect.Zero(kt).OverflowUint(n) {
        return &json.UnmarshalTypeError{Value: "number " + e.Key, Type: kt}
      }
      key = reflect.ValueOf(n).Convert(kt)
    }

    v := reflect.New(t.Elem()).Elem()
    if err := e.Value.Value(v, name); err != nil {
      return err
    }
    if !key.IsValid() {
      continue
    }

    value.SetMapIndex(key, v)
  }

  return nil
}

func realName(f *reflect.StructField, name func(tag reflect.StructTag)(name string)) string {
  n := name(f.Tag)
  if n == "" {
    n = f.Name
  }
  return n
}

func (o Object) valueStruct(value reflect.Value, name func(tag reflect.StructTag)(name string)) error {
  fields,err := flatfield.Flatten(value, flatfield.Name(name))
  if err != nil {
    return err
  }
  nfMap := make(map[string]*reflect.StructField, len(fields))
  for _,f := range fields {
    nfMap[realName(f.SField, name)] = f.SField
  }

oLoop:
  for _,e := range o {
    f,ok := nfMap[e.Key]
    if !ok {
      continue
    }

    subv := value
    for _, i := range f.Index {
      if subv.Kind() == reflect.Ptr {
        if subv.IsNil() {
          // If a struct embeds a pointer to an unexported type,
          // it is not possible to set a newly allocated value
          // since the field is unexported.
          //
          // See https://golang.org/issue/21357
          if !subv.CanSet() {
            //return fmt.Errorf("json: cannot set embedded pointer to unexported struct: %v", subv.Type().Elem())

            break oLoop
          }
          subv.Set(reflect.New(subv.Type().Elem()))
        }
        subv = subv.Elem()
      }
      subv = subv.Field(i)
    }

    if err := e.Value.Value(subv, name); err != nil {
      return err
    }
  }

  return nil
}

func (o Object) Value(i interface{}, name func(tag reflect.StructTag) (name string)) error {
  value := reflect.ValueOf(i)
  if !value.IsValid() {
    // skip
    return nil
  }
  value = indirect(value, false)

  switch value.Kind() {
  case reflect.Interface:
    return o.valueInterface(value, name)
  case reflect.Map:
    return o.valueMap(value, name)
  case reflect.Struct:
    return o.valueStruct(value, name)
  }

  return &json.UnmarshalTypeError{Value: "object ", Type: value.Type()}
}

func (o *Object) MarshalJSON() ([]byte, error) {
  buffer := &bytes.Buffer{}
  buffer.WriteRune('{')

  for i, elem := range *o {
    if i != 0 {
      buffer.WriteRune(',')
    }
    if elem.Tips != "" {
      // key"-tip" 作为tips的key
      buffer.WriteString("\"" + elem.Key + "-tips\":\"" + elem.Tips + "\",")
    }
    buffer.WriteString("\"" + elem.Key + "\":")
    v, err := json.Marshal(elem.Value)
    if err != nil {
      return nil, err
    }
    buffer.Write(v)
  }

  buffer.WriteRune('}')

  return buffer.Bytes(), nil
}
