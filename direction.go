package direction


import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
	"sync"
	"os"
  "errors"
  "strconv"
  "path"

	"github.com/patrickmn/go-cache"
	"googlemaps.github.io/maps"
)

type Driver struct{}

type DirectionInfo struct {
	Drivability bool
}

type DirectionQuery struct {
  Lat1 string
  Lon1 string
  Lat2 string
  Lon2 string
  Key string
}

func (d *Driver) Drivable(locs DirectionQuery, reply *DirectionInfo) error {
	if reply == nil {
		return errors.New("Cannot be given nil")
	}
	lat1 := locs.Lat1 
	lon1 := locs.Lon1
	lat2 := locs.Lat2
	lon2 := locs.Lon2
  api := locs.Key
  fmt.Println(reply.Key)
	reply.Drivability = Drivable(lat1, lon1, lat2, lon2, api)
	return nil
}

var data map[string]cache.Item
var cacheLock sync.RWMutex
var c *cache.Cache

func LoadCache() {
  cacheLock.Lock()
	// caching
	// read the saved cache
	data_json, err := ioutil.ReadFile(path.Join(os.Getenv("GOPATH"), "data/route_cache.json"))
	if err != nil {
		// previously no cache
		// create new file
		fmt.Println("Creating new cache file")
	} else {
		// load cache
		fmt.Println("Loading cache")
		json.Unmarshal(data_json, &data)
	}
	cacheLock.Unlock()
	// create a cache
	c = cache.NewFrom(cache.NoExpiration, 10*time.Minute, data)
  return
}


func SaveCache(sigs chan os.Signal, done chan bool) {
	<-sigs
	cacheLock.RLock()

	data_json, _ := json.Marshal(data)
	err := ioutil.WriteFile(path.Join(os.Getenv("GOPATH"), "data/route_cache.json"), data_json, 0644)
	if err != nil {
		fmt.Printf("Unable to write file: %s", err)
	}

	cacheLock.RUnlock()
	done <- true
	return
}

func Query_to_Key(c *maps.Client, req1 *maps.GeocodingRequest, req2 *maps.GeocodingRequest) string {
	result1, err1 := c.ReverseGeocode(context.Background(), req1)
	result2, err2 := c.ReverseGeocode(context.Background(), req2)
	if err1 != nil || err2 != nil {
		fmt.Println("Error during reverse geocoding")
	}
	var area1 string
	var country1 string
	for _, component := range result1[0].AddressComponents {
		if component.Types[0] == "administrative_area_level_1" {
			area1 = component.LongName
		} else if component.Types[0] == "country" {
			country1 += component.LongName
		}
	}
	name1 := area1 + country1

	var area2 string
	var country2 string
	for _, component := range result2[0].AddressComponents {
		if component.Types[0] == "administrative_area_level_1" {
			area2 = component.LongName
		} else if component.Types[0] == "country" {
			country2 += component.LongName
		}
	}
	name2 := area2 + " " + country2
	// THOUGHT: postal code instead of city name?
	return name1 + " - " + name2
}

func Drivable(lat1 string, lon1 string, lat2 string, lon2 string, api string) bool {
	loc1 := lat1 + ", " + lon1
	loc2 := lat2 + ", " + lon2
	fmt.Println(loc1)
	fmt.Println(loc2)

	// initialize the client for querying google api
	client, err := maps.NewClient(maps.WithAPIKey(api))
	if err != nil {
		fmt.Println("Error initializing client: %s", err)
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
	// request for reverse geocoding
	geo1 := &maps.LatLng{
		Lat: lat_g1,
		Lng: lon_g1,
	}
	geo2 := &maps.LatLng{
		Lat: lat_g2,
		Lng: lon_g2,
	}
	geo_request1 := &maps.GeocodingRequest{
		LatLng: geo1,
	}
	geo_request2 := &maps.GeocodingRequest{
		LatLng: geo2,
	}


	// search for query in cache
	var drivable bool
	var search_result maps.Route

	cached_result, found := c.Get(Query_to_Key(client, geo_request1, geo_request2))
	if found {
		// already cached
		fmt.Println("Found")
		return cached_result.(bool)
	} else {
		// not in cache
		// spend some money and search
		fmt.Println("Search and add to cache")

		route, _, err := client.Directions(context.Background(), query)
		if err != nil {
			fmt.Println("Error during get direction: %s", err)
		} else {
			if len(route) > 0 {
				search_result = route[0]
				drivable = true
			} else {
				fmt.Println("Not drivable")
				drivable = false
			}
		}
	}
  // the following only happens when not found in cache and got result from googlemaps

	// now in route, it stores a map with all details from the direction api search
	// result has the first direction from route
	// result['legs'] has all the dirving
	// not sure why it is an array
	// for now just index the first element of legs
	if drivable {
		// result from search
		var temp_s string = strings.ToLower(fmt.Sprintf("%s", search_result))
		if strings.Contains(temp_s, "ferry") || strings.Contains(temp_s, "ferries") || strings.Contains(temp_s, "tunnel") {
			drivable = false
		} else {
			fmt.Println(search_result.Legs[0].Duration.String())
		}
	}
  c.Set(Query_to_Key(client, geo_request1, geo_request2), drivable, cache.NoExpiration)
	// fmt.Println(drivable)
	return drivable
}
