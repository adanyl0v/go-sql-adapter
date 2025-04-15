package trace

type Level int

const (
	TraceLevel Level = iota
	ErrorLevel
)

const (
	ErrorKey    = "error"
	QueryKey    = "query"
	ResultKey   = "result"
	DurationKey = "duration"
)

type Logger interface {
	Log(level Level, message string, fields map[string]any)
	With(fields map[string]any) Logger
	WithCallerSkip(skip int) Logger
}
