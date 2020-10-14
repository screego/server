package ws

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	roomsCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "screego_room_created_total",
		Help: "The total number of rooms created",
	})
	roomsClosedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "screego_room_closed_total",
		Help: "The total number of rooms closed",
	})
	usersJoinedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "screego_user_joined_total",
		Help: "The total number of users joined",
	})
	usersLeftTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "screego_user_left_total",
		Help: "The total number of users left",
	})
	sessionCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "screego_session_created_total",
		Help: "The total number of sessions created",
	})
	sessionClosedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "screego_session_closed_total",
		Help: "The total number of sessions closed",
	})
)
