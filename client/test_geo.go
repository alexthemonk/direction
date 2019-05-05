package main

import (
  "fmt"
  "github.com/codingsince1985/geo-golang/openstreetmap"
)

func main() {
  g := openstreetmap.Geocoder()
  res, err := g.ReverseGeocode(35.02954, 135.75666)
  if err != nil {
    fmt.Println("Error", err)
  } else {
    fmt.Println(res.State, res.Country)
  }
}
