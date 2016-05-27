package main

import (
	"flag"
	"fmt"
	chttpd "github.com/nochso/colourl/http"
	"log"
	"net/http"
)

var port int

func init() {
	flag.IntVar(&port, "port", 9191, "HTTP listening port")
	flag.IntVar(&port, "p", 9191, "HTTP listening port (shorthand)")
}

func main() {
	flag.Parse()
	log.Printf("colourl-http port=%d", port)
	http.HandleFunc("/", chttpd.IndexHandler)
	http.HandleFunc("/svg", chttpd.SVGHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
