package main

import (
	"fmt"
	"log"
	"net/rpc"
	"strings"

	"github.com/alexthemonk/drivability"
)

func main() {

	test_input := [2]string{"48.6803404 1.9421686", "52.7602109 -0.8134228"}

	client, err := rpc.DialHTTP("tcp", "127.0.0.1:1279")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	done := make(chan int)
	go func() {
		var reply direction.DirectionInfo
		var query direction.DirectionQuery
		query.Lat1 = strings.Fields(test_input[0])[0]
		query.Lon1 = strings.Fields(test_input[0])[1]
		query.Lat2 = strings.Fields(test_input[1])[0]
		query.Lon2 = strings.Fields(test_input[1])[1]
		query.Key = "API"
		err = client.Call("Driver.Drivable", query, &reply)
		fmt.Println(reply.Drivability)
		done <- 1
	}()
	fmt.Println("Done")
	//fmt.Println(reply.GetLocation())
	//fmt.Println(reply.NetInfo)
	<-done
}
