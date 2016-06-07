package logging

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/formatters/logstash"
)

func init() {
	logrus.SetLevel(logrus.ErrorLevel) // default to printing only Error level and above
}

// GetLogger returns a logger where the module field is set to name
func GetLogger(module string) *logrus.Entry {
	if module == "" {
		logrus.Warnf("missing module name parameter")
		module = "undefined"
	}
	return logrus.WithField("module", module)
}

// GetLogFormatter returns a formatter according to the given format.
// Supported formats are 'text', 'json', and 'logstash'.
func GetLogFormatter(format string) (logrus.Formatter, error) {
	switch format {
	case "text":
		formatter := &logrus.TextFormatter{}
		formatter.DisableColors = true
		return formatter, nil
	case "json":
		return &logrus.JSONFormatter{}, nil
	case "logstash":
		return &logstash.LogstashFormatter{}, nil
	default:
		return nil, fmt.Errorf("unknown log format: %v\n", format)
	}
}
