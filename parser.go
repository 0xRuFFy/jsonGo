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

	for jt.CurrentToken.Type != _JTT_EOF {
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
	needRBrace := false

	for token.Type != _JTT_RBRACE && token.Type != _JTT_EOF {
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
			default:
				return fmt.Errorf("expected STRING got '%s' (%s) at %s", token.Value, token.Type.String(), token.Location.String())
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
