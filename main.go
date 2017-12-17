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

var addr = flag.String("addr", ":8080", "http service address, default to :8080")

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

	r.HandleFunc("/ws/{key}", func(w http.ResponseWriter, r *http.Request) {

		key := mux.Vars(r)["key"]
		if key == "" {
			return
		}
		hub := hubs[key]
		if hub == nil {
			hub = newHub()
			hub.id = key
			hubs[key] = hub
			go hub.run()
		}

		serveWs(hub, w, r)
	})

	// r.HandleFunc("/{key}", func(w http.ResponseWriter, r *http.Request) {

	// 	key := mux.Vars(r)["key"]
	// 	if key == "" {
	// 		return
	// 	}
	// 	hub := hubs[key]
	// 	if hub == nil {
	// 		hub = newHub()
	// 		hub.id = key
	// 		hubs[key] = hub
	// 		go hub.run()
	// 	}

	// 	serveHttp(hub, w, r)
	// })

	// r.HandleFunc("/{key}/{value}", func(w http.ResponseWriter, r *http.Request) {

	// 	vars := mux.Vars(r)
	// 	key := vars["key"]
	// 	value := vars["value"]
	// 	if key == "" {
	// 		return
	// 	}
	// 	hub := hubs[key]
	// 	if hub == nil {
	// 		hub = newHub()
	// 		hub.id = key
	// 		hubs[key] = hub
	// 		go hub.run()
	// 	}

	// 	serveHttpUpdate(hub, []byte(value), w, r)
	// })

	srv := &http.Server{
		Handler:      r,
		Addr:         *addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go func() {
		log.Fatal(srv.ListenAndServe())
		log.Println("started!")
	}()

	<-done
	fmt.Println("Bye!")
}
