package Hackajob

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

func init() {
	log.SetFlags(0)
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

	films, _ := rsQuery("films", film)
	chars, _ := rsQuery("people", character)
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
		for _, url := range films.Results[0].Characters {
			o, _ := rsGet(url)
			Chars = append(Chars, o.Name)
		}
	}
	if chars.Count > 0 {
		for _, url := range chars.Results[0].Films {
			o, _ := rsGet(url)
			Films = append(Films, o.Title)
		}
	}

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
