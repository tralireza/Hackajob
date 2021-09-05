package Hackajob

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func TestClone(t *testing.T) {
	writeOutput := func(fn string, rdr io.Reader) error {
		if fn == "stdout" || fn == "" {
			if _, err := io.Copy(os.Stdout, rdr); err != nil {
				return err
			}
			return nil
		}

		f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fs.FileMode(0640))
		if err != nil {
			return err
		}
		defer f.Close()

		wtr := bufio.NewWriter(f)
		if _, err := io.Copy(wtr, rdr); err != nil {
			return err
		}
		wtr.Flush()

		log.Print(" -> ", f.Name())
		return nil
	}

	bfr := bytes.Buffer{}
	Clone[Film]("films", 7, &bfr)
	writeOutput("sw_films.json", &bfr)

	bfr.Reset()
	Clone[Character]("people", 88, &bfr)
	writeOutput("sw_characters.json", &bfr)

	bfr.Reset()
	Clone[Planet]("planets", 60, &bfr)
	writeOutput("sw_planets.json", &bfr)

	bfr.Reset()
	Clone[Starship]("starships", 76, &bfr)
	writeOutput("sw_starships.json", &bfr)

	bfr.Reset()
	Clone[Vehicle]("vehicles", 76, &bfr)
	writeOutput("sw_vehicles.json", &bfr)
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
