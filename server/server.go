package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

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
	time.Sleep(5 * time.Second)
	fmt.Println("QQQ " + sqlCfg.FormatDSN())
	sqlDB, err := sql.Open("mysql", sqlCfg.FormatDSN())
	if err != nil {
		panic(err)
	}
	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}

	dbStore := db.NewSQLStore(sqlDB)
	cacheStore := cache.NewRedisStore(*config.RedisAddr, *config.RedisAdminPassword)

	shortenSrv := shortener.NewService(dbStore, cacheStore)
	redirectSrv := redirect.NewService(dbStore, cacheStore)

	server := api.NewServer(
		net.JoinHostPort(*config.ServerHost, *config.ServerPort), shortenSrv, redirectSrv, converter.NewConverter())

	log.Fatal(server.Serve(":" + *config.ServerPort))
}
