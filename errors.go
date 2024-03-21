package gonk

import (
	"fmt"
	"reflect"
)

type (
	KeyNotPresent string
	InvalidKey    string
	InvalidValue  string
)

func (msg KeyNotPresent) Error() string {
	return string(msg)
}

func (msg InvalidKey) Error() string {
	return string(msg)
}

func (msg InvalidValue) Error() string {
	return string(msg)
}

func formatError(key Tag, loader Loader, msg string) string {
	name := reflect.TypeOf(loader).Name()
	return fmt.Sprintf("Error: %s; Loader %s; Key: %s\n", msg, name, key)
}

func errKeyNotPresent(key Tag, ldr Loader) KeyNotPresent {
	return KeyNotPresent(formatError(key, ldr, "Key not found"))
}

func errInvalidKey(key Tag, ldr Loader) InvalidKey {
	return InvalidKey(formatError(key, ldr, "Attempted to read an invalid key"))
}

func errInvalidValue(key Tag, ldr Loader) InvalidValue {
	return InvalidValue(formatError(key, ldr, "Attempted to set using an invalid value"))
}
