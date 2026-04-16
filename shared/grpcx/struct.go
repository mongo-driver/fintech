package grpcx

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

func MustStruct(in map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(in)
	if err != nil {
		panic(err)
	}
	return s
}

func ToStruct(in map[string]any) (*structpb.Struct, error) {
	return structpb.NewStruct(in)
}

func ToMap(in *structpb.Struct) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	return in.AsMap()
}

func GetString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprint(v)
	}
}

func GetFloat64(m map[string]any, key string) float64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	default:
		return 0
	}
}
