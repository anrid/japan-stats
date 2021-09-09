package main

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/anrid/japan-stats/pkg/stats"
	"github.com/davecgh/go-spew/spew"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const japanStatsDatabase = "/tmp/japan-stats.json"

func main() {
	db, found := stats.LoadIfExists(japanStatsDatabase)
	if !found {
		log.Panic("No database found, run the create command in `cmd/create` first.")
	}

	db.Info()

	mustTrim := func(v string) string {
		return strings.Trim(v, " \n\t\r")
	}

	mustFloat := func(v string) float64 {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Panicf("Could not parse '%s' into float64", v)
		}
		return f
	}

	pop := stats.Population{
		Year: "2021",
	}

	if popFile, found := db.GetTableFile("2021", "Population and Households", "Population by Prefecture"); found {
		var start bool

		stats.ExtractDataFromFile(popFile, func(row []string) {
			if len(row) < 34 {
				return
			}

			nameJP := mustTrim(row[1])
			name := mustTrim(row[2])
			value := mustTrim(row[33])

			if name == "Japan" {
				start = true
				pop.Total = mustFloat(value)
				return
			}
			if start && name == "" {
				start = false
			}

			if start {
				// fmt.Printf("NameJP: '%s' Name: '%s' Value: '%s'", nameJP, name, value)

				v := mustFloat(value)

				pop.Prefectures = append(pop.Prefectures, &stats.Prefecture{
					Stat:       "Population",
					Name:       name,
					NameJP:     nameJP,
					Value:      v,
					PctOfTotal: v / pop.Total,
				})
			}
		})
	}

	var disasterStats []stats.Disasters

	publishedYear := []string{
		"2021",
		"2020",
		"2019",
		"2018",
		"2017",
		"2016",
		"2015",
		"2014",
		"2013",
		"2012",
		"2011",
	}
	for _, year := range publishedYear {
		if disaFile, found := db.GetTableFile(year, "Disasters and Accidents", "Natural Disasters by Prefecture"); found {
			disa := stats.Disasters{}

			yearRegex := regexp.MustCompile(`\((\d{4})\)`)
			var start bool

			stats.ExtractDataFromFile(disaFile, func(row []string) {
				if len(row) < 4 {
					return
				}

				nameJP := mustTrim(row[0])
				name := mustTrim(row[1])
				value := mustTrim(row[3])

				// Try looking for the year of statistical data within the table.
				matches := yearRegex.FindStringSubmatch(nameJP)
				if len(matches) > 1 {
					fmt.Printf("Found year of statistical data: %s\n", matches[1])
					disa.Year = matches[1]
				}

				if name == "Japan" {
					start = true
					disa.PersonsAffectedTotal = mustFloat(value)
					return
				}
				if start && name == "" {
					start = false
				}

				if start {
					if value == "-" {
						value = "0"
					}

					v := mustFloat(value)

					disa.Prefectures = append(disa.Prefectures, &stats.Prefecture{
						Stat:       "Persons Affected",
						Name:       name,
						NameJP:     nameJP,
						Value:      v,
						PctOfTotal: v / disa.PersonsAffectedTotal,
					})
				}
			})

			disasterStats = append(disasterStats, disa)
		}
	}

	spew.Dump(disasterStats[0].PersonsAffectedTotal)
	// stats.Dump(disasterStats[0])

	// Calculate disaster stats.
	{
		var allAffectedTotal float64
		var allPrefsTotal float64
		prefs := make(map[string]float64)

		for _, d := range disasterStats {
			allAffectedTotal += d.PersonsAffectedTotal

			var prefsTotal float64
			for _, p := range d.Prefectures {
				if p.Name == "Okayama" || p.Name == "Ibaraki" || p.Name == "Saga" {
					fmt.Printf("year: %s pref: %-15s  --  %-5.f + %-5.f = %-5.f\n", d.Year, p.Name, prefs[p.Name], p.Value, prefs[p.Name]+p.Value)
				}

				prefs[p.Name] += p.Value
				prefsTotal += p.Value
				allPrefsTotal += p.Value
			}

			if prefsTotal != d.PersonsAffectedTotal {
				log.Panicf("Error in prefecture stats: %f / %f\n", prefsTotal, d.PersonsAffectedTotal)
			}
		}
		if allAffectedTotal != allPrefsTotal {
			log.Panicf("Error in total stats: %f / %f\n", allPrefsTotal, allAffectedTotal)
		}

		type entry struct {
			Name       string
			Total      float64
			PctOfTotal float64
		}

		var sorted []entry

		for p, total := range prefs {
			sorted = append(sorted, entry{
				Name:       p,
				Total:      total,
				PctOfTotal: total / allPrefsTotal,
			})
		}

		sort.SliceStable(sorted, func(i, j int) bool {
			return sorted[i].PctOfTotal > sorted[j].PctOfTotal
		})

		// New locale number printer.
		p := message.NewPrinter(language.English)

		p.Printf("\n\nNumber of People Affected by Natural Disasters in Japan (2011-2020): %.f\n\n", allPrefsTotal)
		p.Println("By Prefecture:")
		for i, e := range sorted {
			p.Printf("%03d. Pref: %-15s  --  %-5.02f%%  (%-7.f / %-7.f)\n", i+1, e.Name, e.PctOfTotal*100, e.Total, allPrefsTotal)
		}
	}
}
