package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"path"
	"strings"

	"github.com/alexthemonk/drivability"
)

var data [][]string
var result map[string]bool

func main() {
	result = make(map[string]bool)
	data_json, err := ioutil.ReadFile(path.Join("../../data/", "location.json"))
	if err != nil {
		fmt.Println("Error loading data")
		return
	} else {
		// load cache
		fmt.Println("Loading cache")
		json.Unmarshal(data_json, &data)
	}

	client, err := rpc.DialHTTP("tcp", "127.0.0.1:1279")
	if err != nil {
		log.Fatal("dialing:", err)
		return
	}
	done := make(chan int)
	res := make(chan map[string]bool)

	for _, detail := range data {
		// fmt.Println(i, detail[0], detail[1])
		if detail[0] != "0 0" && detail[1] != "0 0"{
			go func() {
				var reply direction.DirectionInfo
				var query direction.DirectionQuery
				query.Lat1 = strings.Fields(detail[0])[0]
				query.Lon1 = strings.Fields(detail[0])[1]
				query.Lat2 = strings.Fields(detail[1])[0]
				query.Lon2 = strings.Fields(detail[1])[1]
				query.Key = "AIzaSyAXUo6I_JuyD4FHFFZfDji5E_20dl2G5tY"
				err = client.Call("Driver.Drivable", query, &reply)
				done <- 1
				res <- map[string]bool{detail[0]+","+detail[1]: reply.Drivability}
				fmt.Println("Subroutine exiting: ", detail[0]+","+detail[1], ":", reply.Drivability)
			}()
		}
	}

	fmt.Println("Waiting...")
	for _, _ = range data {
		fmt.Println("Appending...")
		for k, v := <-pair{
			result[k] = v
			fmt.Println("Subroutine Done Querying: ", k, v)
		}
	}

	fmt.Println("Saving...")

	dri_json, _ := json.Marshal(result)
	err = ioutil.WriteFile(path.Join("../../data/", "drivibility.json"), dri_json, 0644)
	if err != nil {
		fmt.Printf("Unable to write file: %s", err)
	}
	fmt.Println("Done")
}
