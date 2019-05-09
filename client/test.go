package main

import (
	"fmt"
	"log"
	"net/rpc"
	"strings"

	"direction"
)

func main() {
// 42.2278667,-88.1211408/42.2380348,-88.1760725
	test_input := [2]string{"42.2278667 -88.1211408", "41.2380348 -87.1760725"}

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
		query.Key = "api"
		err = client.Call("Driver.Drivable", query, &reply)
		fmt.Println(reply.Drivability)
		done <- 1
	}()
	fmt.Println("Done")
	//fmt.Println(reply.GetLocation())
	//fmt.Println(reply.NetInfo)
	<-done
}
