package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"

	"github.com/alexthemonk/direction"
)

// Server bootstraping /////////////////////////////////////////////////////////

func main() {
	fmt.Println("Loading Cache")
	direction.LoadCache()
	sigs := make(chan os.Signal)
	done := make(chan bool)
	signal.Notify(sigs, os.Interrupt, os.Kill)
	go direction.SaveCache(sigs, done)
	fmt.Println("Starting Direction")
	driver := new(direction.Driver)
	rpc.Register(driver)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1279")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	fmt.Println("Starting server..")
	go http.Serve(l, nil)
	<-done
}
