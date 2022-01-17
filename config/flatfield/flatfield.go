package flatfield

import (
  "errors"
  "reflect"
)

type option struct {
  nameF func(tag reflect.StructTag) (name string)
}

type Option func(op *option)


// 通过tag获取名字
func Name(nf func(tag reflect.StructTag)(name string)) Option {
  return func(op *option) {
    op.nameF = nf
  }
}

// 1、处理嵌套的匿名struct field(embedded field) 到top-level
// 2、去重所有同名的field: 可能因为取了相同的tag而造成
// 3、去掉所有的非导出field
// 判断重名的名字来源：1、取tag中的名字；2、取FieldName
// 如果匿名struct通过tag能够取到名字，就不认为是匿名struct
// 返回的排序按照struct field的代码书写顺序, breadth-first search over the set of structs // todo breadth-first
// hasValue: 此域是否有合法的值

type FlatField struct {
  SField *reflect.StructField
  HasValue bool
}

func Flatten(st interface{}, opts ...Option) (fields []FlatField, err error) {
  op := &option{
    nameF: func(tag reflect.StructTag) string {
      return ""
    },
  }
  for _,of := range opts {
    of(op)
  }

  typ := reflect.TypeOf(st)
  if typ.Kind() == reflect.Ptr {
    typ = typ.Elem()
  }

  if typ.Kind() != reflect.Struct {
    return nil, errors.New("func Flatten(st, ...) --- st must be a struct or struct pointer")
  }

  nameSet := map[string]bool{}
  flds := make([]reflect.StructField, 0)

  for i := 0; i < typ.NumField(); i++ {
    flat(typ.Field(i), nameSet, flds, op)
  }

  fields = make([]FlatField, len(flds))
  for i := range flds {
    fields[i].SField = &flds[i]
    fields[i].HasValue = hasValue(i, flds[i].Index)
  }

  return
}

func hasValue(i interface{}, index []int) (ok bool) {
  defer func() {
    // FieldByIndex 取不到值时，会panic
    if r := recover(); r != nil {
      ok = false
    }
  }()

  value := reflect.ValueOf(i)
  if !value.IsValid() {
    return false
  }
  if value.Kind() == reflect.Ptr && value.IsNil() {
    return false
  }

  return value.FieldByIndex(index).IsValid()
}

func flat(input reflect.StructField, nameSet map[string]bool, flds []reflect.StructField, opt *option) {
  isUnexported := input.PkgPath != ""
  if input.Anonymous {
    t := input.Type
    if t.Kind() == reflect.Ptr {
      t = t.Elem()
    }
    if isUnexported && t.Kind() != reflect.Struct {
      // Ignore embedded fields of unexported non-struct types.
      return
    }
    // Do not ignore embedded fields of unexported struct types
    // since they may have exported fields.
  } else if isUnexported {
    // Ignore unexported non-embedded fields.
    return
  }

  name := opt.nameF(input.Tag)
  if !input.Anonymous || (input.Anonymous && name != "") {
    if !nameSet[name] {
      flds = append(flds, input)
      nameSet[name] = true
    }
    return
  }

  // follow: input.Anonymous && name == ""
  t := input.Type
  if t.Kind() == reflect.Ptr {
    t = t.Elem()
  }
  if t.Kind() != reflect.Struct {
    // Ignore embedded and name=="" fields  of non-struct types.
    return
  }
  // t is struct
  for i := 0; i < t.NumField(); i++ {
    flat(t.Field(i), nameSet, flds, opt)
  }
}

//// byIndex sorts field by index sequence.
//type byIndex []field
//
//func (x byIndex) Len() int { return len(x) }
//
//func (x byIndex) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
//
//func (x byIndex) Less(i, j int) bool {
//  for k, xik := range x[i].index {
//    if k >= len(x[j].index) {
//      return false
//    }
//    if xik != x[j].index[k] {
//      return xik < x[j].index[k]
//    }
//  }
//  return len(x[i].index) < len(x[j].index)
//}

