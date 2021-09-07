package Hackajob

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
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

	Characters, Starships, Vehicles, Species, Planets []string `json:",omitempty"`

	OpeningCrawl string `json:"opening_crawl"`
	EpisodeId    int    `json:"episode_id"`
	ReleaseDate  string `json:"release_date"`
}

type Character struct {
	SWO

	Name, Height, Mass, Gender, Homeworld string
	Films, Starships, Species, Vehicles   []string `json:",omitempty"`

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

type Species struct {
	SWO

	Name, Classification, Designation, Language, Homeworld string

	People, Films []string `json:",omitempty"`

	AverageHeight   string `json:"average_height"`
	SkinColors      string `json:"skin_colors"`
	HairColors      string `json:"hair_colors"`
	EyeColors       string `json:"eye_colors"`
	AverageLifespan string `json:"average_lifespan"`
}

type SWVehicle struct {
	SWO

	Name, Model, Manufacturer, Crew, Length, Passengers, Consumables string `json:",omitempty"`

	Pilots, Films []string `json:",omitempty"`

	CostInCredits        string `json:"cost_in_credits,omitempty"`
	MaxAtmospheringSpeed string `json:"max_atmosphering_speed,omitempty"`
}

type Starship struct {
	SWVehicle

	MGLT string `json:",omitempty"`

	HyperdriveRating string `json:"hyperdrive_rating"`
	StarshipClass    string `json:"starship_class"`
}

type Vehicle struct {
	SWVehicle

	CargoCapacity string `json:"cargo_capacity"`
	VehicleClass  string `json:"vehicle_class"`
}

func Process[T any](rdr io.Reader, flds []string) []T {
	const rsPos = 6

	var Objs []T
	json.NewDecoder(rdr).Decode(&Objs)

	for i := range Objs {
		rv, rt := reflect.ValueOf(&Objs[i]), reflect.TypeOf(Objs[i])
		for _, fn := range flds {
			if _, ok := rt.FieldByName(fn); ok {
				fld := rv.Elem().FieldByName(fn)
				switch fld.Kind() {
				case reflect.Slice:
					for i := 0; i < fld.Len(); i++ {
						fld.Index(i).SetString(strings.Split(fld.Index(i).String(), "/")[rsPos])
					}
				case reflect.String:
					ts := strings.Split(fld.String(), "/")
					if len(ts) >= rsPos {
						fld.SetString(ts[rsPos])
					}
				}
			}
		}
	}

	return Objs
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

func WriteJSON(fn string, Objs any) error {
	if fn == "stdout" || fn == "" {
		jenc := json.NewEncoder(os.Stdout)
		jenc.SetIndent("", "  ")
		if err := jenc.Encode(&Objs); err != nil {
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

	jenc := json.NewEncoder(wtr)
	jenc.SetIndent("", "  ")
	if err := jenc.Encode(&Objs); err != nil {
		return err
	}
	wtr.Flush()

	log.Print(" -> ", f.Name())
	return nil
}
