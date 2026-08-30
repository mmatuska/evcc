package modbus

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWriteMultipleRegistersResponseSizeError(t *testing.T) {
	tc := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "matching error",
			err:  errors.New("modbus: response data size '1' does not match count '4'"),
			want: true,
		},
		{
			name: "different count",
			err:  errors.New("modbus: response data size '1' does not match count '6'"),
			want: false,
		},
		{
			name: "different error",
			err:  errors.New("modbus: exception '4' (server device failure), function '16'"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsWriteMultipleRegistersResponseSizeError(tc.err))
		})
	}
}
