package config

type Configurator interface {
  Read() (allValues map[string]Json)
  Print(allValues map[string]Json)
}

