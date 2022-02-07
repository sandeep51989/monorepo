package logger

import (
	"time"

	meta "github.com/nationaloilwellvarco/max-edge/lib-meta-go/meta"
)

//error constants
const (
	ErrServiceNotFoundf  string = meta.ErrServiceNotFoundf
	ErrCastMetaDebug     string = "unable to cast into meta debug"
	ErrMetaDebugNotFound string = "meta debug not found"
	ErrCastRawMessage    string = "unable to cast into json raw message"
)

// LogEntry : Message format for API call
type LogEntry struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// LogEntry : Message format for API call
type LogEntryDocker struct {
	Level     string `json:"level"`
	Timestamp int64  `json:"ts"`
	Name      string `json:"name"`
	Content   string `json:"content"`
}

//Debug env var
const (
	EnvNameDebugTimer string = "debugtimer"
)

//default configuration constants
const (
	DefaultDebugTimer time.Duration = 10 * time.Minute
)

//configuration variables
var (
	ConfigDebugTimer time.Duration = DefaultDebugTimer
)

//Defines the severity (level) strings
const (
	DEBUG = "Debug"
	INFO  = "Info"
	WARN  = "Warn"
	FATAL = "Fatal"
	ERROR = "Error"
)

//Debug errors
const (
	InfoErrRetrieveDebug string = "Error encountered while retrieving debug \"%s\""
	InfoErrUpdateDebug   string = "Error encountered while updating debug \"%s\""
)

//DebugJSON defines the payload that must be sent when enabling or disabling debug mode
type DebugJSON struct {
	DebugEnabled bool `json:"DebugEnabled"`
}
