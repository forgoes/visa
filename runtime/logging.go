package runtime

import (
	"github.com/forgoes/logging"
	"github.com/forgoes/logging/handler"
)

func initLogger(log *Logging) error {
	r := logging.GetRootLogger()

	switch log.Level {
	case "debug":
		r.SetLevel(logging.DEBUG)
	case "info":
		r.SetLevel(logging.INFO)
	case "warn":
		r.SetLevel(logging.WARN)
	case "error":
		r.SetLevel(logging.ERROR)
	case "panic":
		r.SetLevel(logging.PANIC)
	case "fatal":
		r.SetLevel(logging.FATAL)
	default:
		r.SetLevel(logging.INFO)
	}

	switch log.EnableCaller {
	case "debug":
		r.EnableCaller(logging.DEBUG)
	case "info":
		r.EnableCaller(logging.INFO)
	case "warn":
		r.EnableCaller(logging.WARN)
	case "error":
		r.EnableCaller(logging.ERROR)
	case "panic":
		r.EnableCaller(logging.PANIC)
	case "fatal":
		r.EnableCaller(logging.FATAL)
	}

	switch log.EnableStack {
	case "debug":
		r.EnableStack(logging.DEBUG)
	case "info":
		r.EnableStack(logging.INFO)
	case "warn":
		r.EnableStack(logging.WARN)
	case "error":
		r.EnableStack(logging.ERROR)
	case "panic":
		r.EnableStack(logging.PANIC)
	case "fatal":
		r.EnableStack(logging.FATAL)
	}

	switch log.Handler {
	case "stdout":
		r.AddHandler(handler.NewStdoutHandler(handler.StdFormatter))
	}

	return nil
}
