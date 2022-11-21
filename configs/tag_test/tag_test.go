package tag_test

import (
  "github.com/stretchr/testify/assert"
  "github.com/xpwu/go-config/configs"
  "testing"
)

type sub struct {
  Sub1 bool
  Sub2 int64 `conf:"-"`
}

type configTest struct {
  LogLevel int
  Num      int
  Name     string `conf:"-"`
  Ptr      *int `conf:"-"`
  Default  string `conf:"-"`
  Sub      *sub
  Subs     []*sub
}

var configValue = &configTest{
  LogLevel: 2,
  Num:      0,
  Name:     "xpwu-default",
  Default:  "this is default",
  Sub: &sub{
    Sub1: false,
    Sub2: 0,
  },
  Subs: []*sub{
    {
      Sub1: true,
      Sub2: 20,
    },
  },
}

func init() {
  configs.Unmarshal(configValue)
}

var jsonC = &configs.JsonConfig{
  ReadFile:  "",
  PrintFile: "",
}

func TestPrint(t *testing.T) {
  configs.SetConfigurator(jsonC)
  assert.NoError(t, configs.Print())
}

func TestRead(t *testing.T) {
  configs.SetConfigurator(jsonC)
  configs.Read()

  expectV := &configTest{
    LogLevel: 45,
    Num:      11,
    Name:     "xpwu-conf",
    Ptr:      nil,
    Default:  "this is default",
    Sub: &sub{
      Sub1: true,
      Sub2: 19,
    },
    Subs: []*sub{
      {
        Sub1: true,
        Sub2: 20,
      },
      {
        Sub1: true,
        Sub2: 21,
      },
      {
        Sub1: true,
        Sub2: 20,
      },
    },
  }

  a := assert.New(t)
  a.EqualValues(expectV, configValue)
}
