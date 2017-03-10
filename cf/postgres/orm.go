// Copyright 2017 The Goboot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//http://gobook.io/read/github.com/go-xorm/manual-en-US/
package postgres

import (
	_ "github.com/lib/pq"
	"github.com/go-xorm/xorm"
)

var engine *xorm.Engine

func InitORM(env PostgresEnv) *xorm.Engine {
	uri := settings.PostgresUri(env.Name)
	log.Infof("Postgres Init ORM uri: %s", maskedUrl(uri))

	eng, err := xorm.NewEngine("postgres", uri)
	if err != nil {
		log.Error(err)
		return nil
	}

	eng.ShowSQL(env.ORMShowSQL)

	eng.SetMaxOpenConns(env.MaxOpenConns)
	eng.SetMaxIdleConns(env.MaxIdleConns)

	return eng
}

func ORM() *xorm.Engine {
	return engine
}