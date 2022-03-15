package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIntVersion(t *testing.T) {
	testCases := []struct {
		name      string
		version   string
		expResult int64
		expError  error
	}{
		{
			name:      "first_test_5.4.0-89-generic",
			version:   "5.4.0-89-generic",
			expResult: 500040000,
			expError:  nil,
		},
		{
			name:      "second_test_5.4.0-89",
			version:   "5.4.0-89",
			expResult: 500040000,
			expError:  nil,
		},
		{
			name:      "third_test_5.4.0-generic",
			version:   "5.4.0-generic",
			expResult: 500040000,
			expError:  nil,
		},
		{
			name:      "fourth_test_5.4.0",
			version:   "5.4.0",
			expResult: 500040000,
			expError:  nil,
		},
		{
			name:      "fifth_test_5.4.0--generic",
			version:   "5.4.0--generic",
			expResult: 500040000,
			expError:  nil,
		},
		{
			name:      "sixth_test_5.4.0-43-.generic",
			version:   "5.4.0-43-.generic",
			expResult: 500040000,
			expError:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			actualReturn, err := intVersion(tc.version)

			assert.Equal(t, tc.expResult, actualReturn, "Returned value does not match expected value")

			if err != nil {
				require.Error(t, tc.expError, err, "Error was expected")
				assert.Contains(t, err.Error(), tc.expError.Error(), "Unexpected error returned")
			}
		})
	}

}
