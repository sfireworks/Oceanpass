package prometh

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type PromeMetrics struct {
	ReqCounter *prometheus.CounterVec
	ReqSummary *prometheus.SummaryVec

	ErrCounter *prometheus.CounterVec
	ErrSummary *prometheus.SummaryVec
}

type PromeArgs struct {
	Method      string
	Code        string
	Api         string
	AccessKeyId string
	Bucket      string
	Object      string
	StartAt     time.Time
	Unique      map[string]interface{}
	Error       error // only ErrRecord will use this field
}

func (Prome *PromeMetrics) Init() {
	Prome.ReqCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total",
			Help: "The total number of current requests.",
		},
		[]string{"method", "code", "api", "access_key_id", "bucket", "object", "unique"},
	)
	Prome.ReqSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_duration_milliseconds",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"method", "code", "api", "access_key_id", "bucket", "object", "unique"},
	)

	Prome.ErrCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "error_total",
			Help: "The total number of current request error.",
		},
		[]string{"method", "code", "api", "access_key_id", "bucket", "object", "start_at", "unique", "error"},
	)
	Prome.ErrSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_error_request_duration_milliseconds",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"method", "code", "api", "access_key_id", "bucket", "object", "start_at", "unique", "error"},
	)

	prometheus.MustRegister(Prome.ReqCounter)
	prometheus.MustRegister(Prome.ReqSummary)

	prometheus.MustRegister(Prome.ErrCounter)
	prometheus.MustRegister(Prome.ErrSummary)
}

func (Prome PromeMetrics) do(pa PromeArgs) {
	if pa.Error == nil {
		Prome.ReqCounter.WithLabelValues(pa.Method, pa.Code, pa.Api, pa.AccessKeyId, pa.Bucket, pa.Object, fmt.Sprintf("%v", pa.Unique)).Inc()
		Prome.ReqSummary.WithLabelValues(pa.Method, pa.Code, pa.Api, pa.AccessKeyId, pa.Bucket, pa.Object, fmt.Sprintf("%v", pa.Unique)).Observe(float64(time.Since(pa.StartAt).Milliseconds()))
	} else {
		Prome.ErrCounter.WithLabelValues(pa.Method, pa.Code, pa.Api, pa.AccessKeyId, pa.Bucket, pa.Object, pa.StartAt.String(), fmt.Sprintf("%v", pa.Unique), pa.Error.Error()).Inc()
		Prome.ErrSummary.WithLabelValues(pa.Method, pa.Code, pa.Api, pa.AccessKeyId, pa.Bucket, pa.Object, pa.StartAt.String(), fmt.Sprintf("%v", pa.Unique), pa.Error.Error()).Observe(float64(time.Since(pa.StartAt).Milliseconds()))
	}
}

func (Prome PromeMetrics) Do(pa PromeArgs, params ...interface{}) {
	if len(params) > 0 {
		for _, v := range params {
			switch v.(type) {
			case string:
				pa.Code = v.(string)
			case error:
				pa.Error = v.(error)
			default:
				continue
			}
		}
	}
	Prome.do(pa)
}

var Prometheus PromeMetrics

func InitPrometheus() {
	Prometheus = PromeMetrics{}
	Prometheus.Init()
}
