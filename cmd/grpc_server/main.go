package main

import (
	"errors"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/lovesosa420/CurrencyService/internal/interceptors/logging"
	"github.com/lovesosa420/CurrencyService/internal/interceptors/metrics"
	"github.com/lovesosa420/CurrencyService/internal/interceptors/tracing"
	"github.com/lovesosa420/CurrencyService/internal/repository/redis"
	"github.com/lovesosa420/CurrencyService/internal/scraper/currency"
	desc "github.com/lovesosa420/CurrencyService/pkg/currency_service_v1"
	opentrace "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"time"
)

type server struct {
	desc.UnimplementedCurrencyV1Server
}

func (s *server) GetCourse(ctx context.Context, req *desc.GetCourseRequest) (*desc.GetCourseResponse, error) {
	cache := redis.NewCache()

	course, err := redis.GetInfo(cache, req.Name)
	if err != nil {
		if errors.Is(err, redis.ErrNoCachedCurrency) {
			ctx, cancel := context.WithTimeout(context.Background(), 480*time.Millisecond)
			defer cancel()
			course, err = currency.GetCurrencyCourse(ctx, req.Name)
			if err != nil {
				return nil, err
			}
			go func() {
				err := redis.SaveInfo(cache, req.Name, course)
				if err != nil {
					log.Println(err)
				}
				redis.CloseCache(cache)
			}()
		} else {
			log.Println(err)
			ctx, cancel := context.WithTimeout(context.Background(), 480*time.Millisecond)
			defer cancel()
			course, err = currency.GetCurrencyCourse(ctx, req.Name)
			if err != nil {
				return nil, err
			}
			redis.CloseCache(cache)
		}
	}
	return &desc.GetCourseResponse{
		Value: course,
	}, nil
}

func main() {
	logger := logging.NewLogger()
	logger.Debug("starting logrus logger")

	tracer, closer, err := tracing.NewTracer()
	if err != nil {
		logger.Fatalf("could not initialize jaeger: %v", err)
	}
	logger.Debug("starting jaeger tracer")
	defer closer.Close()
	opentrace.SetGlobalTracer(tracer)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		logger.Debug("starting metrics server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			logger.Fatalf("failed to start metrics server: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			logging.LoggingInterceptor(logger),
			opentracing.UnaryServerInterceptor(),
			metrics.MetricsInterceptor(),
		)),
	)

	reflection.Register(s)
	desc.RegisterCurrencyV1Server(s, &server{})

	logger.Debugf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		logger.Fatalf("failed to serve: %v", err)
	}
}
