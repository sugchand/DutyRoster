package logging

import (
)

//Interface for logging, logger and loggerproxy implement this Interface.
type LoggingInterface interface {
    Trace(string, ...interface{})
    Info(string, ...interface{})
    Warning(string, ...interface{})
    Error(string, ...interface{})
}
