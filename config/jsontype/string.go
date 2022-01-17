package jsontype

import (
  "encoding/base64"
  "encoding/json"
  "fmt"
  "reflect"
)

func (s String) Kind() Kind {
  return StringK
}

// isValidNumber reports whether s is a valid JSON number literal.
func isValidNumber(s string) bool {
  // This function implements the JSON numbers grammar.
  // See https://tools.ietf.org/html/rfc7159#section-6
  // and https://www.json.org/img/number.png

  if s == "" {
    return false
  }

  // Optional -
  if s[0] == '-' {
    s = s[1:]
    if s == "" {
      return false
    }
  }

  // Digits
  switch {
  default:
    return false

  case s[0] == '0':
    s = s[1:]

  case '1' <= s[0] && s[0] <= '9':
    s = s[1:]
    for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
      s = s[1:]
    }
  }

  // . followed by 1 or more digits.
  if len(s) >= 2 && s[0] == '.' && '0' <= s[1] && s[1] <= '9' {
    s = s[2:]
    for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
      s = s[1:]
    }
  }

  // e or E followed by an optional - or + and
  // 1 or more digits.
  if len(s) >= 2 && (s[0] == 'e' || s[0] == 'E') {
    s = s[1:]
    if s[0] == '+' || s[0] == '-' {
      s = s[1:]
      if s == "" {
        return false
      }
    }
    for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
      s = s[1:]
    }
  }

  // Make sure we are at the end.
  return s == ""
}

func (s String) Value(i interface{}, name func(tag reflect.StructTag) (name string)) error {
  value := reflect.ValueOf(i)
  if !value.IsValid() {
    // skip
    return nil
  }
  value = indirect(value, false)

  switch value.Kind() {
  default:
    return &json.UnmarshalTypeError{Value: "string", Type: value.Type()}
  case reflect.Slice:
    if value.Type().Elem().Kind() != reflect.Uint8 {
      return &json.UnmarshalTypeError{Value: "string", Type: value.Type()}
    }
    b := make([]byte, base64.StdEncoding.DecodedLen(len(s)))
    n, err := base64.StdEncoding.Decode(b, []byte(s))
    if err != nil {
      return err
    }
    value.SetBytes(b[:n])
  case reflect.String:
    if value.Type() == jsonNumberType && !isValidNumber(string(s)) {
      return fmt.Errorf("json: invalid number literal, trying to unmarshal %s into Number", s)
    }
    value.SetString(string(s))
  case reflect.Interface:
    if value.NumMethod() != 0 {
      return &json.UnmarshalTypeError{Value: "string", Type: value.Type()}
    }
    value.Set(reflect.ValueOf(string(s)))
  case reflect.Struct:
    if value.Type() == thisNumberType && !isValidNumber(string(s)) {
      return fmt.Errorf("json: invalid number literal, trying to unmarshal %s into Number", s)
    }
    value.Set(reflect.ValueOf(NumberWithStr(string(s))))
  }

  return nil
}

