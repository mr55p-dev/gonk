package gonk

import (
	"fmt"
	"reflect"
)

type (
	ValueNotPresent   string
	ValueNotSupported string
	InvalidKey        string
	InvalidValue      string
)

func (msg ValueNotPresent) Error() string {
	return string(msg)
}

func (msg ValueNotSupported) Error() string {
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

func errValueNotPresent(key Tag, ldr Loader) ValueNotPresent {
	return ValueNotPresent(formatError(key, ldr, "Key not found"))
}

func errValueNotSupported(key Tag, ldr Loader) ValueNotSupported {
	return ValueNotSupported(formatError(key, ldr, "Expected value of this key is not supported by this loader"))
}

func errInvalidKey(key Tag, ldr Loader) InvalidKey {
	return InvalidKey(formatError(key, ldr, "Attempted to read an invalid key"))
}

func errInvalidValue(key Tag, ldr Loader) InvalidValue {
	return InvalidValue(formatError(key, ldr, "Attempted to set using an invalid value"))
}
