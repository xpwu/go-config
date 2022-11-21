package configs

import (
  "errors"
  "fmt"
  "github.com/xpwu/go-x/jsontype"
  "reflect"
  "strings"
)

var (
  confer     Configurator = &JsonConfig{}
  allConfigs              = make(map[string]interface{})

  isInit  = false
  hasRead = false

  validTypeSet = map[reflect.Kind]bool{
    reflect.Bool: true,

    reflect.Int:   true,
    reflect.Int8:  true,
    reflect.Int16: true,
    reflect.Int32: true,
    reflect.Int64: true,

    reflect.Float32: true,
    reflect.Float64: true,

    reflect.String: true,

    reflect.Uint:   true,
    reflect.Uint8:  true,
    reflect.Uint16: true,
    reflect.Uint32: true,
    reflect.Uint64: true,

    reflect.Ptr:    true,
    reflect.Array:  true,
    reflect.Slice:  true,
    reflect.Struct: true,
  }
  validTypeString = ""
)

func init() {
  for tp := range validTypeSet {
    validTypeString += tp.String() + ","
  }
  validTypeString = strings.TrimRight(validTypeString, ",")
}

func validateType(value reflect.Value, itsKeyPath string, depth int) (err error) {
  if !value.IsValid() {
    return errors.New("value of " + itsKeyPath + " is invalid")
  }

  const maxDepth = 50
  if depth > maxDepth {
    return errors.New("nested depth of " + itsKeyPath +
      "is more than 5. The depth must be less than or equal 5")
  }

  kind := value.Kind()
  if _, ok := validTypeSet[kind]; !ok {
    return errors.New("type of config must be one of " + validTypeString +
      ", but type of " + itsKeyPath + " is " + kind.String())
  }

  switch kind {
  case reflect.Ptr:
    if value.IsNil() {
      return errors.New("value of " + itsKeyPath + " is nil. Can NOT be nil pointer")
    }
    return validateType(value.Elem(), itsKeyPath, depth)
  case reflect.Slice:
    if value.IsNil() {
      return errors.New("value of " + itsKeyPath + " is nil. Can NOT be nil slice")
    }
    fallthrough
  case reflect.Array:
    if value.Len() == 0 {
      // 如果elem不能为zero value, 则slice/array就不能为empty
      return validateType(reflect.Zero(value.Type().Elem()), itsKeyPath+".[]", depth+1)
    }
    for i := 0; i < value.Len(); i++ {
      err = validateType(value.Index(i), fmt.Sprintf("%s.[%d]", itsKeyPath, i), depth+1)
      if err != nil {
        return
      }
    }
  case reflect.Struct:
    ty := value.Type()
    for i := 0; i < value.NumField(); i++ {
      if ty.Field(i).PkgPath != "" {
        continue
      }
      tag, _ := parseTage(ty.Field(i).Tag)
      if tag == "-" {
        continue
      }

      err = validateType(value.Field(i), itsKeyPath+"."+value.Type().Field(i).Name, depth+1)
      if err != nil {
        return
      }
    }
  }

  return
}

// 初始化之前才能解析配置，初始化之后再解析存在print时信息不全，造成配置遗漏
// struct 中的field支持conf tag修改输出的名字；tips tag添加本字段的帮助信息
type StructPtr = interface{}

func Unmarshal(conf StructPtr) {
  if isInit {
    panic("must be called before Read()")
  }

  value := reflect.ValueOf(conf)
  if value.Kind() != reflect.Ptr || value.IsNil() {
    panic("Unmarshal(args)---args must be struct pointer, but not pointer is given or is nil")
  }

  value = value.Elem()
  if value.Kind() != reflect.Struct {
    panic("Unmarshal(args)---args must be struct pointer, but not struct is given")
  }

  key := value.Type().PkgPath() + ":" + value.Type().Name()
  err := validateType(value, key, 0)
  if err != nil {
    panic(err)
  }

  allConfigs[key] = conf
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

// tag: `conf:name,tips` => key = name; tips = tips
// tag: `conf:name,tips,tip2` => key = name; tips = tips,tip2
// tag: `conf:name` => key = name; tips = ""
// tag: `conf:name,` => key = name; tips = ""
// tag: `conf:,tips` => key = ""; tips = tips
// tag: `conf:,tips,tip2` => key = ""; tips = tips,tip2
// tag: `conf:,` => key = ""; tips = ""
// tag: `conf:` => key = ""; tips = ""
// tag: `` => key = ""; tips = ""
// tag: `-` ignore
func parseTage(tag reflect.StructTag) (key, tips string) {
  content := tag.Get("conf")
  if content == "" {
    return
  }
  splits := strings.SplitN(content, ",", 2)
  key = splits[0]
  if len(splits) == 2 {
    tips = splits[1]
  }

  return
}

func getAllDefaultConfigs() (allDefaultConfigs jsontype.Type) {

  return jsontype.FromGoType(allConfigs, parseTage, func(name string) bool {
    return name == "-"
  })
}

func getAllDefaultValue() jsontype.Type {
  return jsontype.FromGoType(

    allConfigs,
    // 在取默认值时，"-" 标记的field，也需要返回
    func(tag reflect.StructTag) (key, tips string) {
      key, tips = parseTage(tag)
      if key == "-" {
        key = ""
      }
      return
    },
    func(name string) bool {
      return false
    })
}

func mergeJsonType(target jsontype.Type, from jsontype.Type) jsontype.Type {
  switch target.Kind() {
  default:
    return target
  case jsontype.SliceK:
    return mergeSlice(target.(jsontype.Slice), from.(jsontype.Slice))
  case jsontype.ObjectK:
    return mergeObject(target.(jsontype.Object), from.(jsontype.Object))
  }
}

func mergeSlice(target jsontype.Slice, from jsontype.Slice) jsontype.Slice {
  if len(from) == 0 {
    return target
  }

  zeroValue := from[0]
  switch zeroValue.Kind() {
  case jsontype.NullK, jsontype.NumberK, jsontype.StringK, jsontype.BoolK:
    return target
  }

  result := make(jsontype.Slice, 0, len(target))
  for _, v := range target {
    result = append(result, mergeJsonType(v, zeroValue))
  }

  return result
}

func mergeObject(target jsontype.Object, from jsontype.Object) jsontype.Object {
  result := make(jsontype.Object, 0, len(from)+len(target))

  targetMap := make(map[string]jsontype.Type, len(target))
  for _, v := range target {
    targetMap[v.Key] = v.Value
  }

  for _, f := range from {
    if v, ok := targetMap[f.Key]; ok {
      t := make(jsontype.Object, 1)
      t[0].Key = f.Key
      t[0].Tips = f.Tips
      t[0].Value = mergeJsonType(v, f.Value)

      result = append(result, t...)
      continue
    }

    // 不存在，就直接添加
    result = append(result, f)
  }

  return result
}

func Read() {
  if err := ReadWithErr(); err != nil {
    panic(err.Error())
  }
}

func ReadWithErr() error {
  isInit = true

  if hasRead {
    return nil
  }
  hasRead = true

  if err := Valid(); err != nil {
    return err
  }

  values,err := confer.Read(getAllDefaultConfigs())
  if err != nil {
    return err
  }

  // 合并默认值
  values = mergeJsonType(values, getAllDefaultValue())

  return jsontype.ToGoType(values, &allConfigs, func(tag reflect.StructTag) (name string) {
    name, _ = parseTage(tag)
    // 在读的时候，忽略 '-' 配置，即使标记了'-'的域，如果配置文件中有对应的值，也可读取成功。
    if name == "-" {
      name = ""
    }
    return
  })
}

func Print() error {
  isInit = true

  return confer.Print(getAllDefaultConfigs())
}

func Valid() error {
  need := getAllDefaultConfigs()
  read,err := confer.Read(need)
  if err != nil {
    return err
  }
  return read.IncludeErr(need, "")
}
