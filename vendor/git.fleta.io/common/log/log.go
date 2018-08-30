package log

import (
	"context"
	stdlog "log"
	"time"

	plog "github.com/go-playground/log"
	"github.com/go-playground/log/handlers/console"
)

const (
	// DefaultTimeFormat is the default time format when parsing Time values.
	// it is exposed to allow handlers to use and not have to redefine
	DefaultTimeFormat = "2006-01-02T15:04:05.000000000Z07:00"
)

func init() {
	cLog := console.New(true)
	AddHandler(cLog, plog.AllLevels...)

	//TODO contact config system and load file config
	WithDefaultFields(plog.Fields{
		{"program", "FLETA"},
		{"version", "0.0.1"},
	}...)

	stdlog.Println("init")
}

// SetExitFunc sets the provided function as the exit function used in Fatal(),
// Fatalf(), Panic() and Panicf(). This is primarily used when wrapping this library,
// you can set this to to enable testing (with coverage) of your Fatal() and Fatalf()
// methods.
func SetExitFunc(fn func(code int)) {
	plog.ExitFunc = fn
}

// SetWithErrorFn sets a custom WithError function handlers
func SetWithErrorFn(fn func(plog.Entry, error) plog.Entry) {
	plog.WithErrFn = fn
}

// SetContext sets a log entry into the provided context
func SetContext(ctx context.Context, e plog.Entry) context.Context {
	return context.WithValue(ctx, plog.CtxIdent, e)
}

// GetContext returns the log Entry found in the context,
// or a new Default log Entry if none is found
func GetContext(ctx context.Context) plog.Entry {
	v := ctx.Value(plog.CtxIdent)
	if v == nil {
		return plog.NewEntryWithFields(nil)
	}
	return v.(plog.Entry)
}

// HandleEntry handles the log entry and fans out to all handlers with the proper log level
// This is exposed to allow for centralized logging whereby the log entry is marshalled, passed
// to a central logging server, unmarshalled and finally fanned out from there.
func HandleEntry(e plog.Entry) {
	if !e.Start.IsZero() {
		e = e.WithField("duration", time.Since(e.Start))
	}
	e.Timestamp = time.Now()

	for _, h := range plog.LogHandlers[e.Level] {
		h.Log(e)
	}
}

// F creates a new Field using the supplied key + value.
// it is shorthand for defining field manually
func F(key string, value interface{}) plog.Field {
	return plog.Field{Key: key, Value: value}
}

// AddHandler adds a new log handler and accepts which log levels that
// handler will be triggered for
func AddHandler(h Handler, levels ...plog.Level) {
	for _, level := range levels {
		handler := append(plog.LogHandlers[level], h)
		plog.LogHandlers[level] = handler
	}
}

// WithDefaultFields adds fields to the underlying logger instance
func WithDefaultFields(fields ...plog.Field) {
	plog.LogFields = append(plog.LogFields, fields...)
}

// WithField returns a new log entry with the supplied field.
func WithField(key string, value interface{}) plog.Entry {
	ne := plog.NewEntryWithFields(plog.LogFields)
	ne.Fields = append(ne.Fields, plog.Field{Key: key, Value: value})
	return ne
}

// WithFields returns a new log entry with the supplied fields appended
func WithFields(fields ...plog.Field) plog.Entry {
	ne := plog.NewEntryWithFields(plog.LogFields)
	ne.Fields = append(ne.Fields, fields...)
	return ne
}

// WithTrace withh add duration of how long the between this function call and
// the susequent log
func WithTrace() plog.Entry {
	ne := plog.NewEntryWithFields(plog.LogFields)
	ne.Start = time.Now()
	return ne
}

// WithError add a minimal stack trace to the log Entry
func WithError(err error) plog.Entry {
	ne := plog.NewEntryWithFields(plog.LogFields)
	return plog.WithErrFn(ne, err)
}

// Debug logs a debug entry
func Debug(v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Debug(v...)
}

// Debugf logs a debug entry with formatting
func Debugf(s string, v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Debugf(s, v...)
}

// Info logs a normal. information, entry
func Info(v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Info(v...)
}

// Infof logs a normal. information, entry with formatiing
func Infof(s string, v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Infof(s, v...)
}

// Notice logs a notice log entry
func Notice(v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Notice(v...)
}

// Noticef logs a notice log entry with formatting
func Noticef(s string, v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Noticef(s, v...)
}

// Warn logs a warn log entry
func Warn(v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Warn(v...)
}

// Warnf logs a warn log entry with formatting
func Warnf(s string, v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Warnf(s, v...)
}

// Panic logs a panic log entry
func Panic(v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Panic(v...)
}

// Panicf logs a panic log entry with formatting
func Panicf(s string, v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Panicf(s, v...)
}

// Alert logs an alert log entry
func Alert(v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Alert(v...)
}

// Alertf logs an alert log entry with formatting
func Alertf(s string, v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Alertf(s, v...)
}

// Fatal logs a fatal log entry
func Fatal(v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Fatal(v...)
}

// Fatalf logs a fatal log entry with formatting
func Fatalf(s string, v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Fatalf(s, v...)
}

// Error logs an error log entry
func Error(v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Error(v...)
}

// Errorf logs an error log entry with formatting
func Errorf(s string, v ...interface{}) {
	e := plog.NewEntryWithFields(plog.LogFields)
	e.Errorf(s, v...)
}

// Handler is an interface that log handlers need to comply with
type Handler interface {
	Log(plog.Entry)
}
