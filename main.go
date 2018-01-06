package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

var hubs = make(map[string]*Hub)

var addr = flag.String("addr", ":8080", "http service address, defaults to :8080")
var addrs = flag.String("addrs", ":8443", "http service address, defaults to :8443")
var cert = flag.String("cert", "crt.crt", "certificate file path, defaults to crt.crt")
var key = flag.String("key", "key.key", "key file path, defaults to key.key")

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case sig := <-sigs:
				fmt.Println()
				fmt.Println(sig)
				// cleanup code here
				done <- true
			}
		}
	}()

	flag.Parse()

	r := mux.NewRouter()

	r.HandleFunc("/ws/{key}/{value}", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", r.Header.Get("Access-Control-Request-Method"))
		w.Header().Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))

		key := mux.Vars(r)["key"]
		value := mux.Vars(r)["value"]
		if key == "" {
			return
		}
		hub := hubs[key]
		if hub == nil {
			hub = newHub()
			hub.id = key
			hub.pin = value
			hubs[key] = hub
			go hub.run()
		} else if hub.pin != value {
			return
		}

		serveWs(hub, w, r)
	})

	srv := &http.Server{
		Handler:      r,
		Addr:         *addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	srvs := &http.Server{
		Handler:      r,
		Addr:         *addrs,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		log.Println("Listening on", *addr)
		log.Fatal(srv.ListenAndServe())
	}()

	go func() {
		log.Println("Listening on", *addrs)
		log.Fatal(srvs.ListenAndServeTLS(*cert, *key))
	}()

	<-done
	fmt.Println("Bye!")
}
