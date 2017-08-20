package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/justinas/alice"
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

func main() {
	flag.IntVar(&port, "p", 9191, "HTTP listening port")
	flag.BoolVar(&verbose, "v", false, "Enable verbose / debug output")
	flag.Parse()
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	if verbose {
		log.SetLevel(log.DebugLevel)
	}
	log.WithFields(log.Fields{
		"version":    Version,
		"build_date": BuildDate,
	}).Info("colourl-http")

	srv := newServer()
	log.WithFields(log.Fields{
		"port":    port,
		"verbose": verbose,
	}).Info("Starting HTTP server")
	log.Fatal(srv.ListenAndServe())
}

func newServer() *http.Server {
	return &http.Server{
		Handler:        newHandler(),
		Addr:           fmt.Sprintf(":%d", port),
		ReadTimeout:    time.Second * 5,
		WriteTimeout:   time.Second * 10,
		IdleTimeout:    time.Second * 60,
		MaxHeaderBytes: 1 << 17, // 128kB
	}
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", chttpd.IndexMux())
	mux.HandleFunc("/svg", chttpd.SVGHandler)
	return alice.New(
		logHandler,
		gziphandler.GzipHandler,
	).Then(mux)
}

func logHandler(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fn.ServeHTTP(w, r)
		log.WithFields(log.Fields{
			"duration": time.Now().Sub(start),
			"url":      r.URL,
			"method":   r.Method,
			"remote":   r.RemoteAddr,
		}).Debug("HTTP request")
	})
}
