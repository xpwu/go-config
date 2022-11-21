package configs

import (
  "bytes"
  "encoding/json"
  "errors"
  "fmt"
  "github.com/xpwu/go-x/jsontype"
  "io/ioutil"
  "path/filepath"
)

type JsonConfig struct {
  ReadFile string
  PrintFile string
}

func absFilePath(setValue, defaultValue string) string {
  filePath := setValue

  if filePath == "" {
    filePath = defaultValue
  }

  filePath1,err := filepath.Abs(filePath)
  if err != nil {
    return filePath
  }

  return filePath1
}

func (j *JsonConfig) Read(allDefaultConfigs jsontype.Type) (allValues jsontype.Type, err error) {
  filePath := absFilePath(j.ReadFile, "config.json")

  data,err := ioutil.ReadFile(filePath)
  if err != nil {
    return nil, errors.New("can't read config file: " + filePath)
  }

  allValues, err = jsontype.FromJson(data)

  if err != nil {
    return nil, errors.New("can't jsontype.FromJson() from file: " + filePath + ". " + err.Error())
  }

  return
}

func (j *JsonConfig) Print(allDefaultConfigs jsontype.Type) error {

  data,err := jsontype.ToJson(allDefaultConfigs)
  if err != nil {
    return errors.New("cant json.marshal for config. " + err.Error())
  }

  buffer := bytes.NewBuffer([]byte{})
  if err = json.Indent(buffer, data, "", "\t"); err != nil {
    return err
  }

  filePath := absFilePath(j.PrintFile, "config.json.default")

  if err = ioutil.WriteFile(filePath, buffer.Bytes(),
    0644); err != nil {
    return err
  }

  fmt.Printf("print config ok! file: %s\n", filePath)

  return nil
}
