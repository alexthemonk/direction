package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/rpc"
	"path"
	"strings"
	"os"

	"github.com/alexthemonk/direction"
)

var data [][]string
var result map[string]bool

func main() {
	limit := 100
	// test_data := "[[\"37.33939 -121.89496\",\"39.04372 -77.48749\"],[\"0 0\",\"0 0\"],[\"0 0\",\"50.11552 8.68417\"]]"
	result = make(map[string]bool)
	data_json, err := ioutil.ReadFile(path.Join(os.Getenv("GOPATH"), "data/"+os.Args[1]))
	if err != nil {
		fmt.Println("Error loading data")
		return
	} else {
		// load cache
		fmt.Println("Loading data")
		json.Unmarshal(data_json, &data)
	}
	// json.Unmarshal([]byte(test_data), &data)

	client, err := rpc.DialHTTP("tcp", "127.0.0.1:1279")
	if err != nil {
		log.Fatal("dialing:", err)
		return
	}
	res := make(chan map[string]bool, limit)
	count := 0
	q_c := make(chan []string, limit)

	for _, d := range data {
		// fmt.Println(i, detail[0], detail[1])
		if d[0] != "0 0" && d[1] != "0 0"{
			q_c <- d
			fmt.Println("Start query: ", d)
			count ++
			go func() {
				detail := <- q_c
				var reply direction.DirectionInfo
				var query direction.DirectionQuery
				query.Lat1 = strings.Fields(detail[0])[0]
				query.Lon1 = strings.Fields(detail[0])[1]
				query.Lat2 = strings.Fields(detail[1])[0]
				query.Lon2 = strings.Fields(detail[1])[1]
				query.Key = "InputAPIKeyHere"
				err = client.Call("Driver.Drivable", query, &reply)
				res <- map[string]bool{detail[0]+","+detail[1]: reply.Drivability}
				fmt.Println("Subroutine exiting: ", detail[0]+","+detail[1], ":", reply.Drivability)
			}()
		} else {
			result[d] = false
		}
	}

	fmt.Println("Waiting...")
	for i := 0; i < count; i ++ {
		fmt.Println("Appending...")
		for k, v := range <- res {
			result[k] = v
			fmt.Println("Subroutine Done Querying: ", k, v)
		}
	}

	fmt.Println("Saving...")

	dri_json, _ := json.Marshal(result)
	err = ioutil.WriteFile(path.Join(os.Getenv("GOPATH"), "data/"+"drivable_"+os.Args[1]), dri_json, 0644)
	if err != nil {
		fmt.Printf("Unable to write file: %s", err)
	}
	fmt.Println("Done")
}
