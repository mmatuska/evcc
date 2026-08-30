package modbus

import (
	"errors"

	gridx "github.com/grid-x/modbus"
)

func IsWriteMultipleRegistersResponseSizeError(err error) bool {
	var sizeErr *gridx.DataSizeError
	return errors.As(err, &sizeErr) && sizeErr.ExpectedBytes == 4
}
