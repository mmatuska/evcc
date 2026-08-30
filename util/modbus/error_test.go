package modbus

import (
	"errors"
	"fmt"
	"testing"

	gridx "github.com/grid-x/modbus"
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
			err: &gridx.DataSizeError{
				ExpectedBytes: 4,
				ActualBytes:   1,
			},
			want: true,
		},
		{
			name: "wrapped matching error",
			err: fmt.Errorf("wrapped: %w", &gridx.DataSizeError{
				ExpectedBytes: 4,
				ActualBytes:   1,
			}),
			want: true,
		},
		{
			name: "different count",
			err: &gridx.DataSizeError{
				ExpectedBytes: 6,
				ActualBytes:   1,
			},
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
