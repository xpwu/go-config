package config

import (
  "encoding/json"
  "errors"
  "reflect"
  "strings"
)

type Json []byte

func (js Json)sameField(other Json) error  {
  self := map[string]interface{}{}
  if err := json.Unmarshal(js, &self); err != nil {
    return err
  }

  otherM := map[string]interface{}{}
  if err := json.Unmarshal(other, &otherM); err != nil {
    return err
  }

  for oKey := range otherM {
    if _,ok := self[oKey]; !ok {
      return errors.New("json key<" + oKey + "> is not the same")
    }
  }

  for oKey := range self {
    if _,ok := otherM[oKey]; !ok {
      return errors.New("json key<" + oKey + "> is not the same")
    }
  }

  return nil
}

var (
  confer Configurator
  isInit = false
  hasRead = false
  allConfigs = make([]interface{}, 0)
)

// 初始化之前才能解析配置，初始化之后再解析存在print时信息不全，造成配置遗漏
func Unmarshal(conf interface{})  {
  if isInit {
    panic("must be called before Read()")
  }

  rv := reflect.ValueOf(conf)
  if rv.Kind() != reflect.Ptr || rv.IsNil() {
    panic("Unmarshal(args)---args must be struct pointer, but not pointer is given")
  }

  rt := reflect.TypeOf(conf).Elem()
  if rt.Kind() != reflect.Struct {
    panic("Unmarshal(args)---args must be struct pointer, but not struct is given")
  }

  validTypeSet := map[reflect.Kind]bool{
    reflect.Bool:    true,
    reflect.Int:     true,
    reflect.Int8:    true,
    reflect.Int16:   true,
    reflect.Int32:   true,
    reflect.Int64:   true,
    reflect.Float32: true,
    reflect.Float64: true,
    reflect.String:  true,
    reflect.Uint:    true,
    reflect.Uint8:   true,
    reflect.Uint16:  true,
    reflect.Uint32:  true,
    reflect.Uint64:  true,
    reflect.Array:   true,
    reflect.Slice:   true,
    reflect.Struct:  true,
  }
  validTypeString := ""
  for tp := range validTypeSet {
    validTypeString += tp.String() + ","
  }
  validTypeString = strings.TrimRight(validTypeString, ",")

  fieldNum := rt.NumField()
  for i := 0; i < fieldNum; i++ {
    kind := rt.Field(i).Type.Kind()
    if _, ok := validTypeSet[kind]; !ok {
      panic("config must be one of " + validTypeString + ", but " +
        rt.Field(i).Name + " of " + rt.PkgPath() + ":" + rt.Name() +
        " is " + kind.String())
    }
  }

  allConfigs = append(allConfigs, conf)
}

func SetConfigurator(cfer Configurator) {
  confer = cfer
}

func GetConfigurator() Configurator {
  return confer
}

func HasRead() bool {
  return hasRead
}

func Read() {
  isInit = true

  if hasRead {
    return
  }
  hasRead = true

  if err := Valid(); err != nil {
    panic(err.Error())
  }

  values := confer.Read()
  for _, conf := range allConfigs {
    tp := reflect.TypeOf(conf).Elem()
    confKey := tp.PkgPath() + ":" + tp.Name()
    value := values[confKey]

    if err := json.Unmarshal(value, conf); err != nil {
      panic("Unmarshal " + confKey + " error, while json is " + string(value))
    }
  }
}

func Print() {
  isInit = true

  values := map[string]Json{}
  for _, conf := range allConfigs {
    tp := reflect.TypeOf(conf).Elem()
    confKey := tp.PkgPath() + ":" + tp.Name()

    js,err := json.Marshal(conf)
    if err != nil {
      panic(err.Error())
    }

    values[confKey] = js
  }

  confer.Print(values)
}

func Valid() error {
  isInit = true

  values := confer.Read()
  for _, conf := range allConfigs {
    tp := reflect.TypeOf(conf).Elem()
    confKey := tp.PkgPath() + ":" + tp.Name()
    value,ok := values[confKey]
    if !ok {
      return errors.New("can't find config of " + confKey)
    }

    needJson,err := json.Marshal(conf)
    if err != nil {
      return err
    }

    if err := value.sameField(needJson); err != nil {
      return errors.New("field of " + confKey + " is not the same with config. " + err.Error())
    }
  }

  return nil
}
