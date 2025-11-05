package grpc

import "github.com/forgoes/logging"

var l *logging.Logger

func init() {
	l = logging.GetLogger("grpc")
}
