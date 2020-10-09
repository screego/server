package ws

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	roomGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "screego_room_total",
		Help: "The total number of rooms",
	})
	userGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "screego_user_total",
		Help: "The total number of users",
	})
	sessionGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "screego_session_total",
		Help: "The total number of sessions",
	})
)
