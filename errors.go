package gonk

import (
	"fmt"
	"reflect"
)

type (
	ValueNotPresentError   string // ValueNotPresentError occurs when a value is requested from a loader that does not exist in that source
	ValueNotSupportedError string // ValueNotSupportedError occurs when a value of a non-supported type is requested from a loader
	InvalidValueError      string // InvalidValueError occurs when a value loaded in a loader is not compatible with the type of value requested.
)

func (msg ValueNotPresentError) Error() string {
	return string(msg)
}

func (msg ValueNotSupportedError) Error() string {
	return string(msg)
}

func (msg InvalidValueError) Error() string {
	return string(msg)
}

func formatError(key tagData, loader Loader, msg string) string {
	name := reflect.TypeOf(loader).Name()
	return fmt.Sprintf("Error: %s; Loader %s; Key: %s\n", msg, name, key)
}

func errValueNotPresent(key tagData, ldr Loader) ValueNotPresentError {
	return ValueNotPresentError(formatError(key, ldr, "Key not found"))
}

func errValueNotSupported(key tagData, ldr Loader) ValueNotSupportedError {
	return ValueNotSupportedError(formatError(key, ldr, "Expected value of this key is not supported by this loader"))
}

func errInvalidValue(key tagData, ldr Loader) InvalidValueError {
	return InvalidValueError(formatError(key, ldr, "Attempted to set using an invalid value"))
}
