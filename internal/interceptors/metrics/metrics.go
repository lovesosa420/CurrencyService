package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

var (
	// Создаем счетчик для количества запросов
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method"},
	)

	// Создаем гистограмму для времени обработки запросов
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Histogram of latencies for gRPC requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

func init() {
	// Регистрация метрик в Prometheus
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestDuration)
}

// MetricsInterceptor - интерсептор для сбора метрик
func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()

		// Обработка запроса
		resp, err = handler(ctx, req)

		// Сбор метрик
		requestCount.WithLabelValues(info.FullMethod).Inc()
		requestDuration.WithLabelValues(info.FullMethod).Observe(time.Since(start).Seconds())

		return resp, err
	}
}
