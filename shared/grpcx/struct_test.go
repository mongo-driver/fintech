package grpcx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStructConversionAndGetters(t *testing.T) {
	src := map[string]any{
		"name":  "x",
		"count": 3,
	}
	st, err := ToStruct(src)
	require.NoError(t, err)
	m := ToMap(st)
	require.Equal(t, "x", GetString(m, "name"))
	require.Equal(t, 3.0, GetFloat64(m, "count"))
	require.NotPanics(t, func() {
		_ = MustStruct(src)
	})
}

func TestGettersFallback(t *testing.T) {
	m := map[string]any{
		"f1": float32(1.5),
		"i1": int(2),
		"i2": int64(3),
	}
	require.Equal(t, "", GetString(m, "missing"))
	require.Equal(t, 1.5, GetFloat64(m, "f1"))
	require.Equal(t, 2.0, GetFloat64(m, "i1"))
	require.Equal(t, 3.0, GetFloat64(m, "i2"))
	require.Equal(t, 0.0, GetFloat64(m, "missing"))
	require.Equal(t, map[string]any{}, ToMap(nil))
}
