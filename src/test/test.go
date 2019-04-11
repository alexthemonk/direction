package main

import (
	"fmt"
	"log"
	"net/rpc"
	"strings"

	"github.com/alexthemonk/drivability"
)


func main() {

	test_input := [2]string{ "51.2463426 2.9617203", "51.3585961 1.4392685" }

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
		query.Key = "AIzaSyAXUo6I_JuyD4FHFFZfDji5E_20dl2G5tY"
		err = client.Call("Driver.Drivable", , &reply)
		fmt.Println(reply.Drivability)
		done <- 1
	}()
	fmt.Println("Done")
	//fmt.Println(reply.GetLocation())
	//fmt.Println(reply.NetInfo)
	<-done
}
