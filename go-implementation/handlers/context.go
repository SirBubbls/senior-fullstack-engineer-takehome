package handlers

import (
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"sirbubbls.io/challenge/broadcast"
)

// this struct defines the application context for request handlers
type AppContext struct {
	DB      *sqlx.DB
	Updates *broadcast.WeatherUpdateBroadcast
}
