package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
)

func Normalize(raw []byte) ([]byte, error) {
	value, err := parseJSONObject(raw)
	if err != nil {
		return nil, err
	}

	delete(value, "last_refresh")

	normalized, err := marshalCanonical(value)
	if err != nil {
		return nil, fmt.Errorf("%w: canonicalize auth JSON: %v", domain.ErrInvalidAuth, err)
	}

	return normalized, nil
}

func parseJSONObject(raw []byte) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var decoded any
	if err := decoder.Decode(&decoded); err != nil {
		return nil, fmt.Errorf("%w: parse auth JSON: %v", domain.ErrInvalidAuth, err)
	}

	object, ok := decoded.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: top-level auth JSON value must be an object", domain.ErrInvalidAuth)
	}

	return object, nil
}

func marshalCanonical(value any) ([]byte, error) {
	switch typed := value.(type) {
	case nil:
		return []byte("null"), nil
	case bool, string, json.Number:
		return json.Marshal(typed)
	case float64:
		return json.Marshal(typed)
	case []any:
		return marshalCanonicalArray(typed)
	case map[string]any:
		return marshalCanonicalObject(typed)
	default:
		return nil, fmt.Errorf("unsupported JSON value type %T", value)
	}
}

func marshalCanonicalArray(values []any) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteByte('[')

	for i, value := range values {
		if i > 0 {
			buffer.WriteByte(',')
		}

		item, err := marshalCanonical(value)
		if err != nil {
			return nil, err
		}

		buffer.Write(item)
	}

	buffer.WriteByte(']')
	return buffer.Bytes(), nil
}

func marshalCanonicalObject(object map[string]any) ([]byte, error) {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var buffer bytes.Buffer
	buffer.WriteByte('{')

	for i, key := range keys {
		if i > 0 {
			buffer.WriteByte(',')
		}

		keyJSON, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}

		valueJSON, err := marshalCanonical(object[key])
		if err != nil {
			return nil, err
		}

		buffer.Write(keyJSON)
		buffer.WriteByte(':')
		buffer.Write(valueJSON)
	}

	buffer.WriteByte('}')
	return buffer.Bytes(), nil
}
