package config

import "flag"

var (
	HostIp = flag.String("host_ip", "localhost", "host ip of URL Shortener")
	HostPort = flag.Int("host_port", 8787, "host port of URL Shortener")
	HostName = flag.String("hostname", "http://localhost:8787", "hostname of URL Shortener")
)