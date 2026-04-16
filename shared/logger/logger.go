package logger

import "go.uber.org/zap"

func New(serviceName string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}
	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return l.With(zap.String("service", serviceName)), nil
}
