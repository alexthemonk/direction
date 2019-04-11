package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"

	"direction"
)
// not done
func main() {
	direction.LoadCache()
	sigs := make(chan os.Signal)
	done := make(chan bool)
	signal.Notify(sigs, os.Interrupt, os.Kill)
	go iptools.SaveCache(sigs, done)
	fmt.Println("Starting driver")
	drivability := new(iptools.IPMapper)
	rpc.Register(ipmapper)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1279")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	fmt.Println("Starting server..")
	go http.Serve(l, nil)
	<-done
}
