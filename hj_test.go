package Hackajob

import (
	"log"
	"testing"
	"time"
)

func TestStarWars(t *testing.T) {
	for _, v := range [][]string{{"A New Hope", "Raymus Antilles"}, {"Return of the Jedi", "Spock"}} {
		film, character := v[0], v[1]
		ts := time.Now()
		log.Printf("[%q %q] -> %q {%v}", film, character, StarWars(film, character), time.Since(ts))
	}
}
