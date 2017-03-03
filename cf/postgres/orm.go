package postgres

import (
	"github.com/go-xorm/xorm"
)

var engine *xorm.Engine

func InitORM(env PostgresEnv) *xorm.Engine {
	uri := settings.PostgresUri(env.Name)
	log.Infof("Postgres Init ORM uri: %s", maskedUrl(uri))

	eng, err := xorm.NewEngine("postgres", uri)
	if err != nil {
		panic(err)
	}

	eng.SetMaxOpenConns(env.MaxOpenConns)
	eng.SetMaxIdleConns(env.MaxIdleConns)

	return eng
}

func ORM() *xorm.Engine {
	return engine
}