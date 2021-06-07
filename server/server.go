package main

import (
	"context"
	"database/sql"
	"flag"

	"github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/thegodmouse/url-shortener/api"
	"github.com/thegodmouse/url-shortener/cache"
	"github.com/thegodmouse/url-shortener/config"
	"github.com/thegodmouse/url-shortener/converter"
	"github.com/thegodmouse/url-shortener/db"
	"github.com/thegodmouse/url-shortener/services/redirect"
	"github.com/thegodmouse/url-shortener/services/shortener"
	"github.com/thegodmouse/url-shortener/util"
)

func main() {
	flag.Parse()

	sqlCfg := mysql.Config{
		User:                 "root",
		Passwd:               *config.MySQLRootPassword,
		Addr:                 *config.MySQLServerAddr,
		Net:                  "tcp",
		DBName:               "url_shortener",
		AllowNativePasswords: true,
		ParseTime:            true,
	}
	sqlDB, err := sql.Open("mysql", sqlCfg.FormatDSN())
	if err != nil {
		panic(err)
	}

	dbStore := db.NewSQLStore(sqlDB)
	cacheStore := cache.NewRedisStore(*config.RedisAddr, *config.RedisAdminPassword)

	shortenSrv := shortener.NewService(dbStore, cacheStore)
	redirectSrv := redirect.NewService(dbStore, cacheStore)

	server := api.NewServer(*config.RedirectServeURL, shortenSrv, redirectSrv, converter.NewConverter())

	// start checking for expire short urls
	ctx, cancel := context.WithCancel(context.Background())

	done := util.DeleteExpiredURLs(ctx, dbStore, cacheStore)

	if err := server.Serve(":" + *config.ServerPort); err != nil {
		log.Errorf("Server: serve err: %v, at port: %v", err, *config.ServerPort)
	}
	cancel()
	// wait for done
	<-done
}
