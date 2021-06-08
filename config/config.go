package config

import "flag"

var (
	// ServerPort is the port that url_shortener server serves at.
	ServerPort = flag.String("server_port", "80", "host port of URL Shortener")
	// RedirectServeURL is the url that the redirect API serves at.
	RedirectServeURL = flag.String("redirect_serve_url", "http://localhost", "url to serve redirect api")

	// MySQLServerAddr is the address for the mysql server.
	MySQLServerAddr = flag.String("mysql_server_addr", "localhost:3306", "mysql server addr")
	// MySQLRootPassword is the password for root user on the mysql server.
	MySQLRootPassword = flag.String("mysql_server_root_password", "", "root password for connecting mysql server")

	// RedisServerAddr is the address for the redis server
	RedisServerAddr = flag.String("redis_server_addr", "localhost:6379", "redis server addr")
	// RedisAdminPassword is the password for admin user on the redis server.
	RedisAdminPassword = flag.String("redis_server_admin_password", "", "redis server admin password")
)
