package logging

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func NewLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.DebugLevel)

	return logger
}

func LoggingInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger.Infof("received request: %v", req)
		resp, err := handler(ctx, req)
		if err != nil {
			logger.Errorf("error handling request: %v", err)
		} else {
			logger.Infof("successfully handled request: %v", resp)
		}
		return resp, err
	}
}
