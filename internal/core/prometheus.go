package core

import (
	"go-shop-backend/pkg/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/redis/go-redis/extra/redisprometheus/v9"
)

func NewPrometheusRegistry(container *Container) *prometheus.Registry {
	sqlDB, err := sqlDBFromContainer(container)
	if err != nil {
		logger.Fatal(container.Logger(), "NewPrometheusRegistry: failed to get sqlDB", err)
	}

	reg := prometheus.NewRegistry()

	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewDBStatsCollector(sqlDB, container.Config().Database.Name))
	reg.MustRegister(redisprometheus.NewCollector("goshop", "redis", container.RedisClient().RDB()))

	return reg
}
