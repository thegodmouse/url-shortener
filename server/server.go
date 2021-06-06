package main

import (
	"database/sql"
	"flag"
	"log"

	"github.com/go-sql-driver/mysql"
	"github.com/thegodmouse/url-shortener/api"
	"github.com/thegodmouse/url-shortener/cache"
	"github.com/thegodmouse/url-shortener/config"
	"github.com/thegodmouse/url-shortener/converter"
	"github.com/thegodmouse/url-shortener/db"
	"github.com/thegodmouse/url-shortener/services/redirect"
	"github.com/thegodmouse/url-shortener/services/shortener"
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

	log.Fatal(server.Serve(":" + *config.ServerPort))
}
