// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoStoreErrIs(t *testing.T) {
	err0 := &NoStoreErr{
		addr: "1.2.3.4",
	}
	// identical errors are equal
	require.True(t, errors.Is(err0, err0))
	require.True(t, errors.Is(ErrNoStore, ErrNoStore))
	require.True(t, errors.Is(ErrNoStore, &NoStoreErr{}))
	require.True(t, errors.Is(&NoStoreErr{}, ErrNoStore))
	// not equal for different error types
	require.False(t, errors.Is(err0, errors.New("")))
	// default Value matches any error
	require.True(t, errors.Is(err0, ErrNoStore))
	// error with values are not matching default ones
	require.False(t, errors.Is(ErrNoStore, err0))

	err1 := &NoStoreErr{
		addr: "2.3.4.5",
	}
	require.True(t, errors.Is(err1, ErrNoStore))
	// errors with different values are not equal
	require.False(t, errors.Is(err0, err1))
	require.False(t, errors.Is(err1, err0))
}
