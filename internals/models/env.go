package models

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// Env context for db, logger, etc. This is passed within a request
type Env struct {
	Db        *sqlx.DB
	Log       *log.Logger
	Cfg       *Config
	Router    *mux.Router
	RandState string
}
