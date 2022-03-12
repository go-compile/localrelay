package localrelay

import (
	"io"
	"log"
)

// Logger is used for logging debug information such as
// connections being created, dropped etc
type Logger struct {
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

// NewLogger creates a new logging system
func NewLogger(w io.Writer, name string) *Logger {
	return &Logger{
		Info:    log.New(w, "[INFO] ["+name+"] ", log.Lshortfile|log.Lmicroseconds),
		Warning: log.New(w, "[WARNING] ["+name+"] ", log.Lshortfile|log.Lmicroseconds),
		Error:   log.New(w, "[ERROR] ["+name+"] ", log.Lshortfile|log.Lmicroseconds),
	}
}
