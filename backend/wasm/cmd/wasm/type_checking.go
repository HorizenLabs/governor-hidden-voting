package main

import (
	"fmt"
	"syscall/js"
)

func isType(v js.Value, t js.Type) error {
	if v.Type() != t {
		return fmt.Errorf("expected %s, got %s", t.String(), v.Type().String())
	}
	return nil
}

func isObject(obj js.Value, keys []string, types []js.Type) error {
	if len(keys) != len(types) {
		return fmt.Errorf("keys and types should have the same length")
	}
	if err := isType(obj, js.TypeObject); err != nil {
		return err
	}
	for i, key := range keys {
		if err := hasField(obj, key, types[i]); err != nil {
			return err
		}
	}
	return nil
}

func hasField(obj js.Value, key string, t js.Type) error {
	val := obj.Get(key)
	if val.Type() == js.TypeUndefined {
		return fmt.Errorf("key \"%s\" undefined", key)
	}
	if err := isType(val, t); err != nil {
		return fmt.Errorf("type of field \"%s\": %v", key, err)
	}
	return nil
}

type fieldParsingError struct {
	key string
	err error
}

func NewFieldParsingError(key string, err error) *fieldParsingError {
	return &fieldParsingError{
		key: key,
		err: err,
	}
}

func (e *fieldParsingError) Error() string {
	return fmt.Sprintf("error parsing field %s: %v", e.key, e.err)
}
