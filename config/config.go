package config

import "flag"

var (
	ServerHost = flag.String("server_host", "localhost", "host ip of URL Shortener")
	ServerPort = flag.String("server_port", "16000", "host port of URL Shortener")

	MySQLServerAddr   = flag.String("mysql_server_addr", "localhost:3306", "mysql server addr")
	MySQLRootPassword = flag.String("mysql_server_root_password", "", "root password for connecting mysql server")

	RedisAddr          = flag.String("redis_server_addr", "localhost:6379", "redis server addr")
	RedisAdminPassword = flag.String("redis_server_admin_password", "", "redis server admin password")
)
