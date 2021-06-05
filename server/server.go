package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/thegodmouse/url-shortener/api"
	"github.com/thegodmouse/url-shortener/config"
	"github.com/thegodmouse/url-shortener/services/redirect"
	"github.com/thegodmouse/url-shortener/services/shortener"
)

func main() {
	flag.Parse()

	server := api.NewServer(*config.HostName, shortener.NewService(), redirect.NewService())

	log.Fatal(server.Serve(fmt.Sprintf("%v:%v", *config.HostIp, *config.HostPort)))
}
