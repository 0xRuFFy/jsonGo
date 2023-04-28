package jsongo

import (
	"fmt"
	"strconv"
)

type Json struct {
	Data map[string]interface{}
}

func baseParse(jt *jsonTokenizer) (*Json, error) {
	json := &Json{
		Data: make(map[string]interface{}),
	}

	_, err := jt.NextToken()
	if err != nil {
		return nil, err
	}

	for jt.CurrentToken.Type != JTT_EOF {
		err = parseObject(jt, &json.Data)
		if err != nil {
			return nil, err
		}
	}

	return json, nil
}

func Parse(str string) (*Json, error) {
	tokenizer, err := newJsonTokenizerContent(str)
	if err != nil {
		return nil, err
	}

	return baseParse(tokenizer)
}

func ParseFile(fileName string) (*Json, error) {
	tokenizer, err := newJsonTokenizer(fileName)
	if err != nil {
		return nil, err
	}

	return baseParse(tokenizer)
}

func parseArray(jt *jsonTokenizer, data *[]interface{}) error {

	if jt.CurrentToken.Type != _JTT_LBRACKET {
		return fmt.Errorf("expected '[' got '%s' (%s) at %s", jt.CurrentToken.Value, jt.CurrentToken.Type.String(), jt.CurrentToken.Location.String())
	}

	token, err := jt.NextToken()
	if err != nil {
		return err
	}

	needValue := true
	needComma := false
	needRBracket := false

	for token.Type != _JTT_RBRACKET && token.Type != JTT_EOF {
		if needValue {
			switch token.Type {
			case _JTT_STRING:
				*data = append(*data, token.Value)
				needValue = false
				needComma = true
				needRBracket = true
			case _JTT_INTEGER:
				value, err := strconv.ParseInt(token.Value, 10, 64)
				if err != nil {
					return err
				}
				*data = append(*data, value)
				needValue = false
				needComma = true
				needRBracket = true
			case _JTT_FLOAT:
				value, err := strconv.ParseFloat(token.Value, 64)
				if err != nil {
					return err
				}
				*data = append(*data, value)
				needValue = false
				needComma = true
				needRBracket = true
			case _JTT_BOOLEAN:
				value, err := strconv.ParseBool(token.Value)
				if err != nil {
					return err
				}
				*data = append(*data, value)
				needValue = false
				needComma = true
				needRBracket = true
			case _JTT_NULL:
				*data = append(*data, nil)
				needValue = false
				needComma = true
				needRBracket = true
			case _JTT_LBRACE:
				value := make(map[string]interface{})
				err = parseObject(jt, &value)
				if err != nil {
					return err
				}
				*data = append(*data, value)
				needValue = false
				needComma = true
				needRBracket = true
				token = jt.CurrentToken
				continue
			case _JTT_LBRACKET:
				value := make([]interface{}, 0)
				err = parseArray(jt, &value)
				if err != nil {
					return err
				}
				*data = append(*data, value)
				needValue = false
				needComma = true
				needRBracket = true
			default:
				return fmt.Errorf("expected VALUE got '%s' (%s) at %s", token.Value, token.Type.String(), token.Location.String())
			}
		} else if needComma {
			if token.Type != _JTT_COMMA {
				return fmt.Errorf("expected ',' got '%s' (%s) at %s", token.Value, token.Type.String(), token.Location.String())
			}
			needValue = true
			needComma = false
			needRBracket = false
		} else if needRBracket {
			if token.Type != _JTT_RBRACKET {
				return fmt.Errorf("expected ']' got '%s' (%s) at %s", token.Value, token.Type.String(), token.Location.String())
			}
			needValue = false
			needComma = false
			needRBracket = false
		}

		token, err = jt.NextToken()
		if err != nil {
			return err
		}
	}

	return nil

}

func parseObject(jt *jsonTokenizer, data *map[string]interface{}) error {
	if jt.CurrentToken.Type != _JTT_LBRACE {
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
	needRBrace := true

	for token.Type != _JTT_RBRACE && token.Type != JTT_EOF {
		if needKey {
			if token.Type != _JTT_STRING {
				return fmt.Errorf("expected STRING got '%s' (%s) at %s", token.Value, token.Type.String(), token.Location.String())
			}
			key = token.Value
			needKey = false
			needValue = false
			needColon = true
			needComma = false
			needRBrace = false
		} else if needColon {
			if token.Type != _JTT_COLON {
				return fmt.Errorf("expected ':' got '%s' (%s) at %s", token.Value, token.Type.String(), token.Location.String())
			}
			needKey = false
			needValue = true
			needColon = false
			needComma = false
			needRBrace = false
		} else if needValue {
			switch token.Type {
			case _JTT_STRING:
				(*data)[key] = token.Value
			case _JTT_INTEGER:
				(*data)[key], err = strconv.ParseInt(token.Value, 10, 64)
				if err != nil {
					return err
				}
			case _JTT_FLOAT:
				(*data)[key], err = strconv.ParseFloat(token.Value, 64)
				if err != nil {
					return err
				}
			case _JTT_BOOLEAN:
				(*data)[key], err = strconv.ParseBool(token.Value)
				if err != nil {
					return err
				}
			case _JTT_NULL:
				(*data)[key] = nil
			case _JTT_LBRACE:
				subdata := make(map[string]interface{})
				err = parseObject(jt, &subdata)
				if err != nil {
					return err
				}
				(*data)[key] = subdata
				token = jt.CurrentToken
				needKey = false
				needValue = false
				needColon = false
				needComma = true
				needRBrace = true
				continue
			case _JTT_LBRACKET:
				subdata := make([]interface{}, 0)
				err = parseArray(jt, &subdata)
				if err != nil {
					return err
				}
				(*data)[key] = subdata
			default:
				return fmt.Errorf("expected Value got '%s' (%s) at %s", token.Value, token.Type.String(), token.Location.String())
			}
			needKey = false
			needValue = false
			needColon = false
			needComma = true
			needRBrace = true
		} else if needComma {
			if token.Type != _JTT_COMMA {
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
