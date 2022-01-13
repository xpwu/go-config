package config

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "path"
)

type JsonConfig struct {
  ReadFile string
  PrintDir string
  ExeAbsDir   string
}

func defaultConfigFile()string  {
  return "config.json"
}

func Tips()string  {
  return "配置文件，如果没有设置，则为'<exeDir>/'" + defaultConfigFile()
}

func (j *JsonConfig) Read() (allValues map[string]Json) {
  if j.ReadFile == "" {
    j.ReadFile = path.Join(j.ExeAbsDir, defaultConfigFile())
  }

  if !path.IsAbs(j.ReadFile) {
    j.ReadFile = path.Join(j.ExeAbsDir, j.ReadFile)
  }

  if path.Ext(j.ReadFile) != ".json" {
    panic("config file name must be end '.json'")
  }

  data,err := ioutil.ReadFile(j.ReadFile)
  if err != nil {
    panic("cant read file: " + j.ReadFile)
  }

  readM := make(map[string]map[string]interface{})
  if err = json.Unmarshal(data, &readM); err != nil {
    panic("cant json.unmarshal from file: " + j.ReadFile + ". " + err.Error())
  }

  allValues = make(map[string]Json)
  for key,m := range readM {
    js,_ := json.Marshal(m)
    allValues[key] = js
  }

  return
}

func (j *JsonConfig) Print(allValues map[string]Json) {
  printM := make(map[string]map[string]interface{})
  for key,js := range allValues {
    m := make(map[string]interface{})
    _ = json.Unmarshal(js, &m)
    printM[key] = m
  }

  data,err := json.Marshal(printM)
  if err != nil {
    panic("cant json.marshal for config. " + err.Error())
  }

  buffer := bytes.NewBuffer([]byte{})
  if err = json.Indent(buffer, data, "", "\t"); err != nil {
    panic(err.Error())
  }

  if err = ioutil.WriteFile(j.printFileName(), buffer.Bytes(),
    0644); err != nil {
    panic(err.Error())
  }

  fmt.Printf("print config ok! file: %s\n", j.printFileName())
}

func (j *JsonConfig) printFileName() string  {
  return path.Join(j.PrintDir, "config.json.default")
}

