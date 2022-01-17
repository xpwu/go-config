package config

import "github.com/xpwu/go-config/config/jsontype"

type Configurator interface {
  Read(allDefaultConfigs jsontype.Type) (allConfigs jsontype.Type)
  Print(allDefaultConfigs jsontype.Type)
}

