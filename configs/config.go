package configs

import (
	"errors"
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
    return errors.New("value of " + itsKeyPath +" is invalid")
  }

	const maxDepth = 5
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
      return errors.New("value of " + itsKeyPath +" is nil. Can NOT be nil pointer")
    }
		return validateType(value.Elem(), itsKeyPath, depth)
  case reflect.Slice:
    if value.IsNil() {
      return errors.New("value of " + itsKeyPath +" is nil. Can NOT be nil slice")
    }
    fallthrough
	case reflect.Array:
		return validateType(value.Elem(), itsKeyPath+".[]", depth+1)
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
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

	key := value.Type().PkgPath()+":"+value.Type().Name()
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

	return jsontype.FromGoType(allConfigs, parseTage, func(name string)bool {
		return name == "-"
	})
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

	values := confer.Read(getAllDefaultConfigs())

	err := jsontype.ToGoType(values, &allConfigs, func(tag reflect.StructTag) (name string) {
		name,_ = parseTage(tag)
		return
	})
	if err != nil {
		panic(err)
	}
}

func Print() {
	isInit = true

	confer.Print(getAllDefaultConfigs())
}

// todo: detail error
func Valid() error {
	need := getAllDefaultConfigs()
	read := confer.Read(need)
	if !read.Include(need) {
		return errors.New("缺少字段或者类型不匹配")
	}
	return nil
}
