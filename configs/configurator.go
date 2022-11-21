package configs

import "github.com/xpwu/go-x/jsontype"

type Configurator interface {
  Read(allDefaultConfigs jsontype.Type) (allConfigs jsontype.Type, err error)
  Print(allDefaultConfigs jsontype.Type) error
}

