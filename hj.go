package Hackajob

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
)

func init() {
	log.SetFlags(0)
}

type SWO struct {
	Id int
}

type Film struct {
	SWO

	Title              string
	Director, Producer string

	Characters, Starships, Vehicles, Species, Planets []string

	OpeningCrawl string `json:"opening_crawl"`
	EpisodeId    int    `json:"episode_id"`
	ReleaseDate  string `json:"release_date"`
}

type Character struct {
	SWO

	Name, Height, Mass, Gender, Homeworld string
	Films, Starships, Species, Vehicles   []string

	HairColor string `json:"hair_color"`
	SkinColor string `json:"skin_color"`
	EyeColor  string `json:"eye_color"`
	BirthYear string `json:"birth_year"`
}

type Planet struct {
	SWO

	Name, Diameter, Gravity, Climate, Terrain, Population string

	Films, Residents []string

	RotationPeriod string `json:"rotation_period"`
	OrbitalPeriod  string `json:"orbital_period"`
	SurfaceWater   string `json:"surface_water"`
}

type SWVehicle struct {
	SWO

	Name, Model, Manufacturer, Crew, Length, Passengers, Consumables string

	Pilots, Films []string

	CostInCredits        string `json:"cost_in_credits"`
	MaxAtmospheringSpeed string `json:"max_atmosphering_speed"`
}

type Starship struct {
	SWVehicle

	MGLT string

	HyperdriveRating string `json:"hyperdrive_rating"`
	StarshipClass    string `json:"starship_class"`
}

type Vehicle struct {
	SWVehicle

	CargoCapacity string `json:"cargo_capacity"`
	VehicleClass  string `json:"vehicle_class"`
}

func Clone[T any](rs string, rsMax int, w io.Writer) error {
	const apiUrl = "https://challenges.hackajob.co/swapi/api"
	var Objs []*T

	for i := range rsMax {
		rsp, err := http.Get(fmt.Sprintf("%s/%s/%d/", apiUrl, rs, i+1))
		if err != nil {
			return err
		}

		if rsp.StatusCode != http.StatusOK {
			rsp.Body.Close()
			continue
		}

		o := new(T)
		json.NewDecoder(rsp.Body).Decode(o)
		rsp.Body.Close()

		rv := reflect.ValueOf(o)
		rv.Elem().FieldByName("Id").Set(reflect.ValueOf(i + 1))

		Objs = append(Objs, o)
	}

	log.Print(" -> ", len(Objs))

	jenc := json.NewEncoder(w)
	jenc.SetIndent("", "  ")
	return jenc.Encode(Objs)
}

func StarWars(film, character string) string {
	type Resource struct{ Title, Name string }
	type Result struct {
		Count   int
		Results []struct {
			Characters []string
			Films      []string
		}
	}

	if tp, ok := http.DefaultTransport.(*http.Transport); ok {
		f := tp.DialContext
		tp.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			ts := time.Now()
			cnn, err := f(ctx, network, addr)
			status := '🎉'
			if err != nil {
				status = '🔥'
			}
			log.Printf("%c [%v] --|%v|-> %v", status, time.Since(ts), network, addr)
			return cnn, err
		}
	}

	rsQuery := func(rs, qry string) (*Result, error) {
		const apiUrl = "https://challenges.hackajob.co/swapi/api"
		rsp, err := http.Get(fmt.Sprintf("%s/%s/?search=%s", apiUrl, rs, url.QueryEscape(qry)))
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()

		var o Result
		json.NewDecoder(rsp.Body).Decode(&o)
		return &o, nil
	}

	var wg sync.WaitGroup
	var films, chars *Result

	wg.Add(2)
	go func() {
		defer wg.Done()
		films, _ = rsQuery("films", film)
	}()
	go func() {
		defer wg.Done()
		chars, _ = rsQuery("people", character)
	}()
	wg.Wait()

	if films == nil || chars == nil {
		return ""
	}

	log.Print(" -> ", len(films.Results))
	log.Print(" -> ", len(chars.Results))

	rsGet := func(url string) (*Resource, error) {
		rsp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()

		var o Resource
		json.NewDecoder(rsp.Body).Decode(&o)
		return &o, nil
	}

	var Films, Chars []string
	if films.Count > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, url := range films.Results[0].Characters {
				o, _ := rsGet(url)
				Chars = append(Chars, o.Name)
			}
		}()
	}
	if chars.Count > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, url := range chars.Results[0].Films {
				o, _ := rsGet(url)
				Films = append(Films, o.Title)
			}
		}()
	}
	wg.Wait()

	sort.Strings(Chars)
	r := film + ": " + strings.Join(Chars, ", ") + "; "
	r += character + ": "
	if len(Films) > 0 {
		sort.Strings(Films)
		r += strings.Join(Films, ", ")
	} else {
		r += "none"
	}
	return r
}
