package jsongo

import (
	"fmt"
	"strconv"
)

type JsonValueType uint8

const (
	JVT_NONE JsonValueType = iota
	JVT_STRING
	JVT_INTEGER
	JVT_FLOAT
	JVT_BOOLEAN
	// JVT_NULL
	// JVT_OBJECT
	// JVT_ARRAY
)

func (jvt JsonValueType) String() string {
	switch jvt {
	case JVT_NONE:
		return "JVT_NONE"
	case JVT_STRING:
		return "JVT_STRING"
	case JVT_INTEGER:
		return "JVT_INTEGER"
	case JVT_FLOAT:
		return "JVT_FLOAT"
	case JVT_BOOLEAN:
		return "JVT_BOOLEAN"
	// case JVT_NULL:
	// 	return "JVT_NULL"
	// case JVT_OBJECT:
	// 	return "JVT_OBJECT"
	// case JVT_ARRAY:
	// 	return "JVT_ARRAY"
	default:
		return "UNKNOWN"
	}
}

type Json struct {
	Data map[string]interface{}
}

func ParseFile(fileName string) (*Json, error) {
	json := &Json{
		Data: make(map[string]interface{}),
	}

	tokenizer, err := NewJsonTokenizer(fileName)
	if err != nil {
		return nil, err
	}

	_, err = tokenizer.NextToken()
	if err != nil {
		return nil, err
	}

	for tokenizer.CurrentToken.Type != JTT_EOF {
		err = parseObject(tokenizer, json)
		if err != nil {
			return nil, err
		}
	}

	return json, err
}

func parseObject(jt *JsonTokenizer, json *Json) error {
	if jt.CurrentToken.Type != JTT_LBRACE {
		return fmt.Errorf("expected '{' got '%s' (%s) at %s", jt.CurrentToken.Value, jt.CurrentToken.Type.String(), jt.CurrentToken.Location.String())
	}

	token, err := jt.NextToken()
	if err != nil {
		return err
	}

	key := ""
	needKey := true
	needValue := false
	needColon := false
	needComma := false
	needRBrace := false

	for token.Type != JTT_RBRACE {
		if needKey {
			if token.Type != JTT_STRING {
				return fmt.Errorf("expected STRING got '%s' (%s) at %s", token.Value, token.Type.String(), token.Location.String())
			}
			key = token.Value
			needKey = false
			needValue = false
			needColon = true
			needComma = false
			needRBrace = false
		} else if needColon {
			if token.Type != JTT_COLON {
				return fmt.Errorf("expected ':' got '%s' (%s) at %s", token.Value, token.Type.String(), token.Location.String())
			}
			needKey = false
			needValue = true
			needColon = false
			needComma = false
			needRBrace = false
		} else if needValue {
			switch token.Type {
			case JTT_STRING:
				json.Data[key] = token.Value
			case JTT_INTEGER:
				json.Data[key], err = strconv.ParseInt(token.Value, 10, 64)
				if err != nil {
					return err
				}
			case JTT_FLOAT:
				json.Data[key], err = strconv.ParseFloat(token.Value, 64)
				if err != nil {
					return err
				}
			case JTT_BOOLEAN:
				json.Data[key], err = strconv.ParseBool(token.Value)
				if err != nil {
					return err
				}
			case JTT_NULL:
				json.Data[key] = nil
			default:
				return fmt.Errorf("expected STRING got '%s' (%s) at %s", token.Value, token.Type.String(), token.Location.String())
			}
			needKey = false
			needValue = false
			needColon = false
			needComma = true
			needRBrace = true
		} else if needComma {
			if token.Type != JTT_COMMA {
				return fmt.Errorf("expected ',' got '%s' (%s) at %s", token.Value, token.Type.String(), token.Location.String())
			}
			needKey = true
			needValue = false
			needColon = false
			needComma = false
			needRBrace = false
		} else if needRBrace {
			return fmt.Errorf("expected '}' got '%s' (%s) at %s", token.Value, token.Type.String(), token.Location.String())
		}

		token, err = jt.NextToken()
		if err != nil {
			return err
		}
	}

	if !needRBrace {
		return fmt.Errorf("did not expected '%s' (%s) at %s", token.Value, token.Type.String(), token.Location.String())
	} else {
		_, err = jt.NextToken()
		if err != nil {
			return err
		}
	}

	return nil
}
