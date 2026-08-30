package modbus

import "strings"

const writeMultipleRegistersResponseSizeErrorCount = "does not match count '4'"

func IsWriteMultipleRegistersResponseSizeError(err error) bool {
	return err != nil &&
		strings.Contains(err.Error(), "modbus: response data size") &&
		strings.Contains(err.Error(), writeMultipleRegistersResponseSizeErrorCount)
}
