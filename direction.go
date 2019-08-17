package direction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	// "github.com/codingsince1985/geo-golang"
	// "github.com/codingsince1985/geo-golang/openstreetmap"
	"googlemaps.github.io/maps"
)

type Driver struct{}

type DirectionInfo struct {
	Drivability float64
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

type DirectionQuery struct {
	Coord1 Coordinates
	Coord2 Coordinates
	Key    string
}

type Drivability struct {
	Drivable float64 `json:"drivable"`
	Text     string  `json:"text"`
}

func (d *Driver) Drivable(locs DirectionQuery, reply *DirectionInfo) error {
	if reply == nil {
		return errors.New("Cannot be given nil")
	}
	lat1 := fmt.Sprintf("%f", locs.Coord1.Latitude)
	lon1 := fmt.Sprintf("%f", locs.Coord1.Longitude)
	lat2 := fmt.Sprintf("%f", locs.Coord2.Latitude)
	lon2 := fmt.Sprintf("%f", locs.Coord2.Longitude)
	api := locs.Key
	// fmt.Println(api)
	reply.Drivability = Drivable(lat1, lon1, lat2, lon2, api)
	return nil
}

var cache map[string]Drivability = make(map[string]Drivability)
var cacheLock sync.RWMutex

func LoadCache() {
	cacheLock.Lock()
	// caching
	// read the saved cache
	data_json, err := ioutil.ReadFile(path.Join(os.Getenv("GOPATH"), "data/drivable_cache.json"))
	if err == nil {
		// load cache
		fmt.Println("Loading local cache file")
		json.Unmarshal(data_json, &cache)
	}
	cacheLock.Unlock()
	// create a cache

	return
}

func SaveCache(sigs chan os.Signal, done chan bool) {
	<-sigs
	cacheLock.RLock()

	data_json, _ := json.Marshal(cache)
	err := ioutil.WriteFile(path.Join(os.Getenv("GOPATH"), "data/drivable_cache.json"), data_json, 0644)
	if err != nil {
		fmt.Printf("Unable to write file: %s", err)
	}
	fmt.Println("Cache Saved")

	cacheLock.RUnlock()
	done <- true
	return
}

type Geo struct {
	Lat float64
	Lon float64
}

func Query_to_Key_Nonreverse(geo1 Geo, geo2 Geo) (string, string) {
	k1 := fmt.Sprintf("%.0f,%.0f - %.0f,%.0f",
		geo1.Lat, geo1.Lon,
		geo2.Lat, geo2.Lon)
	k2 := fmt.Sprintf("%.0f,%.0f - %.0f,%.0f",
		geo2.Lat, geo2.Lon,
		geo1.Lat, geo1.Lon)
	return k1, k2
}

func Drivable(lat1 string, lon1 string, lat2 string, lon2 string, api string) float64 {
	// return travel distance, -1 for not drivable
	loc1 := lat1 + ", " + lon1
	loc2 := lat2 + ", " + lon2
	fmt.Println(loc1)
	fmt.Println(loc2)

	var fail bool = false
	// initialize the client for querying google api
	client, err := maps.NewClient(maps.WithAPIKey(api))
	if err != nil {
		fmt.Println("Error initializing client: %s", err)
		fail = true
	}

	// query for direction
	query := &maps.DirectionsRequest{
		Origin:      loc1,
		Destination: loc2,
		Mode:        maps.TravelModeDriving,
		Avoid:       []maps.Avoid{"ferries"},
	}
	lat_g1, _ := strconv.ParseFloat(lat1, 64)
	lon_g1, _ := strconv.ParseFloat(lon1, 64)
	lat_g2, _ := strconv.ParseFloat(lat2, 64)
	lon_g2, _ := strconv.ParseFloat(lon2, 64)

	geo1 := Geo{Lat: lat_g1, Lon: lon_g1}
	geo2 := Geo{Lat: lat_g2, Lon: lon_g2}

	// search for query in cache
	var drivable float64
	var search_result []byte
	var cacheHit bool = false
	var temp_s string

	// g := openstreetmap.Geocoder()

	// key1, key2 := Query_to_Key(g, geo1, geo2)
	key1, key2 := Query_to_Key_Nonreverse(geo1, geo2)
	if key1 == "" {
		// reverse geolocation error
		fail = true
	}

	if !fail {
		cacheLock.RLock()
		temp, ok := cache[key1]
		if ok {
			drivable = temp.Drivable
		} else {
			temp, ok = cache[key2]
			if ok {
				drivable = temp.Drivable
			}
		}
		cacheLock.RUnlock()
		if ok {
			fmt.Println("Found")
			fmt.Println(drivable)
			if temp.Text == "" {
				return drivable
			} else {
				cacheHit = true
				temp_s = temp.Text
				if strings.Contains(temp_s, "ferry") || strings.Contains(temp_s, "ferries") {
					drivable = -1.0
				} else {
					var temp_json interface{}
					err := json.Unmarshal([]byte(temp_s), &temp_json)
					if err != nil {
						drivable = -1.0
					} else {
						drivable = temp_json.(map[string]interface{})["distance"].(map[string]interface{})["value"].(float64)
					}
				}
			}
		}
	}
	// not in cache
	// if start and end at same city
	// save true
	if fmt.Sprintf("%.0f,%.0f", lat_g1, lon_g1) == fmt.Sprintf("%.0f,%.0f", lat_g2, lon_g2) {
		cacheLock.Lock()
		cache[key1] = Drivability{Drivable: 0.0, Text: ""}
		cacheLock.Unlock()
		return 0.0
	}

	if !cacheHit {
		// spend some money and search
		fmt.Println("Search")
		route, _, err := client.Directions(context.Background(), query)
		if err != nil {
			fail = true
			time.Sleep(time.Second * 2)
			route, _, err = client.Directions(context.Background(), query)
			if err != nil {
				fmt.Println("Error during get direction: %s", err)
				fail = true
			} else {
				fail = false
			}
		} else {
			if len(route) > 0 {
				for _, r := range route {
					search_result, _ = r.Legs[0].MarshalJSON()
					distance := float64(r.Legs[0].Distance.Meters)

					var text_map map[string]interface{}
					json.Unmarshal(search_result, &text_map)
					for i, step := range text_map["steps"].([]interface{}) {
						for k, _ := range step.(map[string]interface{}) {
							if k != "html_instructions" {
								delete(text_map["steps"].([]interface{})[i].(map[string]interface{}), k)
							}
						}
					}
					tes, _ := json.Marshal(text_map)
					temp_s = strings.ToLower(fmt.Sprintf("%s", tes))

					if strings.Contains(temp_s, "ferry") || strings.Contains(temp_s, "ferries") {
						drivable = -1.0
					} else {
						drivable = distance
						break
					}
				}
				// fmt.Println(string(search_result))
			} else {
				fmt.Println("Not drivable")
				drivable = -1.0
			}
		}
	}
	if !fail {
		if !cacheHit {
			fmt.Println("Adding to Cache: ", key1, drivable)
		}
		cacheLock.Lock()
		cache[key1] = Drivability{Drivable: drivable, Text: temp_s}
		cacheLock.Unlock()
	}
	// fmt.Println(drivable)
	return drivable
}
