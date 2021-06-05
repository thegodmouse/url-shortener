package config

import "flag"

var (
	HostIp   = flag.String("host_ip", "localhost", "host ip of URL Shortener")
	HostPort = flag.Int("host_port", 8787, "host port of URL Shortener")
	HostName = flag.String("hostname", "http://localhost:8787", "hostname of URL Shortener")

	MySQLRootPassword = flag.String("mysql_server_root_password", "", "root password for connecting mysql server")
	MySQLServerAddr   = flag.String("mysql_server_addr", "localhost:3306", "mysql server addr")

	RedisAddr     = flag.String("redis_server_addr", "localhost:6379", "redis server addr")
	RedisPassword = flag.String("redis_server_password", "", "redis server password")
)
