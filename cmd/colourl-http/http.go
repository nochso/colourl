package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	chttpd "github.com/nochso/colourl/http"
	log "github.com/sirupsen/logrus"
)

var (
	port    int
	verbose bool
)

var (
	Version   string
	BuildDate string
)

func init() {
	flag.IntVar(&port, "p", 9191, "HTTP listening port")
	flag.BoolVar(&verbose, "v", false, "Enable verbose / debug output")
}

func main() {
	flag.Parse()
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	if verbose {
		log.SetLevel(log.DebugLevel)
	}
	log.WithFields(log.Fields{
		"version":    Version,
		"build_date": BuildDate,
	}).Info("colourl-http")
	log.WithFields(log.Fields{
		"port":    port,
		"verbose": verbose,
	}).Info("Starting HTTP server")
	http.Handle("/", logHandler(chttpd.IndexMux().ServeHTTP))
	http.HandleFunc("/svg", logHandler(chttpd.SVGHandler))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func logHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fn(w, r)
		log.WithFields(log.Fields{
			"duration": time.Now().Sub(start),
			"url":      r.URL,
			"method":   r.Method,
			"remote":   r.RemoteAddr,
		}).Debug("HTTP request")
	}
}
