package grpc

import (
	"context"

	models "github.com/yapryntsev/go-musthave-metrics/internal/model"
	"github.com/yapryntsev/go-musthave-metrics/internal/proto"
	"github.com/yapryntsev/go-musthave-metrics/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	proto.UnimplementedMetricsServer
	svc    service.MetricService
	logger *zap.Logger
}

func NewServer(svc service.MetricService, logger *zap.Logger) *Server {
	return &Server{
		svc:    svc,
		logger: logger,
	}
}

func (s *Server) UpdateMetrics(
	ctx context.Context,
	in *proto.UpdateMetricsRequest,
) (*proto.UpdateMetricsResponse, error) {
	metrics := make([]*models.Metrics, len(in.GetMetrics()))

	for i, v := range in.GetMetrics() {
		switch v.GetType() {
		case proto.Metric_GAUGE:
			del := v.GetDelta()
			metrics[i] = &models.Metrics{
				ID:    v.GetId(),
				MType: models.Gauge,
				Delta: &del,
			}
		case proto.Metric_COUNTER:
			val := v.GetValue()
			metrics[i] = &models.Metrics{
				ID:    v.GetId(),
				MType: models.Counter,
				Value: &val,
			}
		}
	}

	if err := s.svc.UpdateBatch(ctx, metrics); err != nil {
		s.logger.Error("failed to save batch", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to save batch")
	}

	return &proto.UpdateMetricsResponse{}, nil
}
