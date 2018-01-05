// Copyright 2017 The Goboot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Setup optional env JSON value:
// goboot_postgres={
//   "name": "Your_Postgres_Service_Name",
//   "connection": {
//       "max_open": 0,
//       "max_idle": 0
//   },
//   "orm": {
//       "enable": false,
//       "show_sql": false
//   }
// }
// See the following for connection settings:
// https://golang.org/pkg/database/sql/#DB.SetMaxIdleConns
// https://golang.org/pkg/database/sql/#SetMaxOpenConns
package postgres

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/gostones/goboot/config"
	"github.com/gostones/goboot/logging"
	"net/url"
	"fmt"
)

var settings = config.AppSettings()
var log = logging.Logger()

var database  *sql.DB

type PostgresEnv struct {
	Name         string     `env:"goboot_postgres.name"`
	MaxOpenConns int        `env:"goboot_postgres.connection.max_open"`
	MaxIdleConns int        `env:"goboot_postgres.connection.max_idle"`
	ORMEnable    bool       `env:"goboot_postgres.orm.enable"`
	ORMShowSQL   bool       `env:"goboot_postgres.orm.show_sql"`
}

func init() {
	env := PostgresEnv{}

	err := settings.Parse(&env)
	if err != nil {
		log.Errorf("Postgres init error: %v", err)
		return
	}
	log.Debugf("Postgres env: %v", env)

	if env.ORMEnable {
		engine = InitORM(env)
		database = engine.DB().DB
	} else {
		database = InitDB(env)
	}
}

// mask password
func maskedUrl(uri string) string {
	u, _ := url.Parse(uri)
	return fmt.Sprintf("%s://%s:***@%s%s?%s", u.Scheme, u.User.Username(), u.Host, u.Path, u.RawQuery)
}

func InitDB(env PostgresEnv) *sql.DB {
	uri := settings.PostgresUri(env.Name)
	log.Infof("Postgres init DB uri: %s", maskedUrl(uri))

	db, err := sql.Open("postgres", uri)

	if err != nil {
		log.Error(err)
		return nil
	}

	db.SetMaxOpenConns(env.MaxOpenConns)
	db.SetMaxIdleConns(env.MaxIdleConns)

	return db
}

func Status() bool {
	var version string
	err := DB().QueryRow("select version()").Scan(&version)
	if err != nil {
		log.Error(err)
		return false
	}
	log.Debugf("Postgres version: %s\n", version)
	return true
}

func DB() *sql.DB {
	return database
}