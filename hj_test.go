package Hackajob

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func TestClone(t *testing.T) {
	bfr := bytes.Buffer{}
	Clone[Film]("films", 7, &bfr)
	films := Process[Film](&bfr, []string{"Characters", "Planets", "Species", "Starships", "Vehicles"})
	WriteJSON("films.json", films)

	bfr.Reset()
	Clone[Character]("people", 88, &bfr)
	chars := Process[Character](&bfr, []string{"Films", "Species", "Starships", "Vehicles", "Homeworld"})
	WriteJSON("characters.json", chars)

	bfr.Reset()
	Clone[Planet]("planets", 60, &bfr)
	planets := Process[Planet](&bfr, []string{"Films", "Residents"})
	WriteJSON("planets.json", planets)

	bfr.Reset()
	Clone[Starship]("starships", 76, &bfr)
	starships := Process[Starship](&bfr, []string{"Films", "Pilots"})
	WriteJSON("starships.json", starships)

	bfr.Reset()
	Clone[Vehicle]("vehicles", 76, &bfr)
	vehicles := Process[Vehicle](&bfr, []string{"Films", "Pilots"})
	WriteJSON("vehicles.json", vehicles)

	bfr.Reset()
	Clone[Species]("species", 37, &bfr)
	species := Process[Species](&bfr, []string{"Films", "People", "Homeworld"})
	WriteJSON("species.json", species)
}

func TestProcess(t *testing.T) {
	rdr := strings.NewReader(`[
    {"id":1,"title":"A New Hope","name":"C-3PO","cargo_capacity":"2M","starship_class":"Destroyer",
     "films":["//////1/","//////2/","//////3/"],
     "characters":["//////4/","//////5/"],
     "residents":["//////4/"],
     "planets":["//////6/"],
     "species":["//////7/"],
     "pilots":["//////4/","//////5/"],
     "starships":["//////8/","//////9/"],
     "homeworld":"//////6/"}
  ]`)

	Objs := Process[Character](rdr, []string{"Species", "Starships", "Films", "Pilots", "Homeworld"})
	log.Printf("%+v", Objs)
	json.NewEncoder(os.Stdout).Encode(&Objs)
}

func TestStarWars(t *testing.T) {
	for _, v := range [][]string{{"A New Hope", "Raymus Antilles"}, {"Return of the Jedi", "Spock"}} {
		film, character := v[0], v[1]
		ts := time.Now()
		log.Printf("[%q %q] -> %q {%v}", film, character, StarWars(film, character), time.Since(ts))
	}
}
