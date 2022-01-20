package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type configTest struct {
	LogLevel int
	Num      int
	Name     string
}

var configValue = &configTest{
	LogLevel: 2,
	Num:      0,
	Name:     "xpwu",
}

func init() {
	Unmarshal(configValue)
}

var jsonC = &JsonConfig{
	ReadFile:  "",
	PrintFile: "",
}

func TestPrint(t *testing.T) {
	SetConfigurator(jsonC)
	Print()
}

func TestRead(t *testing.T) {
	SetConfigurator(jsonC)
	Read()

	expectV := &configTest{
		LogLevel: 45,
		Num:      11,
		Name:     "xpwu-0",
	}

	a := assert.New(t)
	a.EqualValues(expectV, configValue)
}

func TestValid(t *testing.T) {
	assert.NoError(t, Valid())
}
