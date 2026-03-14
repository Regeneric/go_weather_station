package lora

import (
	"log/slog"

	"github.com/Regeneric/iot-drivers/libs/sx126x"
)

type SlogAdapter struct {
	Log *slog.Logger
}

func (l SlogAdapter) With(kv ...any) sx126x.Logger {
	if l.Log == nil {
		return SlogAdapter{Log: slog.Default().With(kv...)}
	}
	return SlogAdapter{Log: l.Log.With(kv...)}
}
func (l SlogAdapter) Debug(msg string, kv ...any) { l.Log.Debug(msg, kv...) }
func (l SlogAdapter) Info(msg string, kv ...any)  { l.Log.Info(msg, kv...) }
func (l SlogAdapter) Warn(msg string, kv ...any)  { l.Log.Warn(msg, kv...) }
func (l SlogAdapter) Error(msg string, kv ...any) { l.Log.Error(msg, kv...) }
