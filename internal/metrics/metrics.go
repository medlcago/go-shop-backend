package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace = "goshop"
)

type Factory struct {
	reg *prometheus.Registry

	httpMetrics *httpMetrics
}

func New(reg *prometheus.Registry) *Factory {
	f := &Factory{
		reg: reg,
	}

	f.httpMetrics = NewHTTPMetrics(reg)
	return f
}

func (f *Factory) Registry() *prometheus.Registry {
	return f.reg
}

func (f *Factory) HTTPMetrics() *httpMetrics {
	return f.httpMetrics
}
