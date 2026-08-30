package modbus

import "strings"

// FC16 responses always return 4 payload bytes: register address and quantity.
const writeMultipleRegistersResponseSizeErrorSuffix = "does not match count '4'"

func IsWriteMultipleRegistersResponseSizeError(err error) bool {
	return err != nil &&
		strings.Contains(err.Error(), "modbus: response data size") &&
		strings.Contains(err.Error(), writeMultipleRegistersResponseSizeErrorSuffix)
}
