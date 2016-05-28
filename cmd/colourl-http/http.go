package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	chttpd "github.com/nochso/colourl/http"
	"net/http"
)

var port int
var (
	Version   string
	BuildDate string
)

func init() {
	flag.IntVar(&port, "port", 9191, "HTTP listening port")
	flag.IntVar(&port, "p", 9191, "HTTP listening port (shorthand)")
}

func main() {
	flag.Parse()
	log.WithFields(log.Fields{
		"version":    Version,
		"build_date": BuildDate,
	}).Info("colourl-http")
	log.WithFields(log.Fields{
		"port":    port,
		"verbose": verbose,
	}).Info("Starting HTTP server")
	http.HandleFunc("/", chttpd.IndexHandler)
	http.HandleFunc("/svg", chttpd.SVGHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
