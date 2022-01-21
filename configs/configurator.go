package configs

import "github.com/xpwu/go-x/jsontype"

type Configurator interface {
  Read(allDefaultConfigs jsontype.Type) (allConfigs jsontype.Type)
  Print(allDefaultConfigs jsontype.Type)
}

