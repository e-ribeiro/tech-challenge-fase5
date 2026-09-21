package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	UploadsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "fiapx_video_uploads_total",
		Help: "Total de vídeos enviados para processamento",
	})

	JobsProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fiapx_video_jobs_total",
		Help: "Total de jobs de vídeo processados por status",
	}, []string{"status"})

	ProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "fiapx_video_processing_duration_seconds",
		Help:    "Tempo de processamento de vídeos pelo FFmpeg em segundos",
		Buckets: prometheus.DefBuckets,
	})

	ActiveProcessing = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "fiapx_video_active_processing",
		Help: "Quantidade de vídeos sendo processados no momento",
	})
)
