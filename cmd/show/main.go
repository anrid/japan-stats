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

	popul := stats.Population{
		Country: "Japan",
		Year:    "2021",
	}

	if popFile, found := db.GetTableFile("2021", "Population and Households", "Population by Prefecture"); found {
		var start bool

		stats.ExtractDataFromFile(popFile, func(row []string) {
			if len(row) < 34 {
				return
			}

			name := mustTrim(row[2])
			value := mustTrim(row[33])

			if name == "Japan" {
				start = true
				popul.Total = mustFloat(value)
				return
			}
			if start && name == "" {
				start = false
			}

			if start {
				// fmt.Printf("NameJP: '%s' Name: '%s' Value: '%s'", nameJP, name, value)

				v := mustFloat(value)

				popul.Prefectures = append(popul.Prefectures, &stats.Prefecture{
					Name:      name,
					Statistic: "Population",
					Value:     v * 1_000,
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
			disa := stats.Disasters{
				Country: "Japan",
			}

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
						Statistic: "Persons Affected",
						Name:      name,
						Value:     v,
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
		type sum struct {
			AffectedTotal float64
			YearsTotal    int
		}

		var personsAffectedTotal float64
		var allPrefsTotal float64
		prefs := make(map[string]*sum)

		for _, d := range disasterStats {
			personsAffectedTotal += d.PersonsAffectedTotal

			var prefsTotal float64
			for _, p := range d.Prefectures {
				if _, ok := prefs[p.Name]; !ok {
					prefs[p.Name] = &sum{}
				}

				if p.Name == "Okayama" || p.Name == "Ibaraki" || p.Name == "Saga" {
					fmt.Printf("[%s] %-15s  --  %-5.f + %-5.f = %-5.f\n", d.Year, p.Name, prefs[p.Name].AffectedTotal, p.Value, prefs[p.Name].AffectedTotal+p.Value)
				}
				prefs[p.Name].AffectedTotal += p.Value
				prefs[p.Name].YearsTotal++

				prefsTotal += p.Value
				allPrefsTotal += p.Value
			}

			if prefsTotal != d.PersonsAffectedTotal {
				log.Panicf("Error in prefecture stats: %f / %f\n", prefsTotal, d.PersonsAffectedTotal)
			}
		}
		if personsAffectedTotal != allPrefsTotal {
			log.Panicf("Error in total stats: %f / %f\n", allPrefsTotal, personsAffectedTotal)
		}

		sorted := make([]*stats.DisastersInPrefecture, 0)

		for p, v := range prefs {
			avg := v.AffectedTotal / float64(v.YearsTotal)

			pop := popul.Prefectures.Find(p)
			if pop == nil {
				log.Panicf("Could not find population for prefecture %s\n", p)
			}
			if pop.Value == 0 {
				log.Panicf("No population value set for prefecture %s\n", pop.Name)
			}
			if avg > pop.Value {
				log.Panicf("Number of people affected %.f is larger than population %.f for prefecture %s\n", v.AffectedTotal, pop.Value, pop.Name)
			}

			sorted = append(sorted, &stats.DisastersInPrefecture{
				Prefecture:       p,
				AffectedTotal:    v.AffectedTotal,
				AffectedAvg:      avg,
				YearsTotal:       v.YearsTotal,
				PopulationTotal:  pop.Value,
				PopulationPct:    avg / pop.Value,
				OfAllAffectedPct: v.AffectedTotal / personsAffectedTotal,
			})
		}

		// New locale number printer.
		p := message.NewPrinter(language.English)

		// Sort by per capita.
		sort.SliceStable(sorted, func(i, j int) bool {
			return sorted[i].PopulationPct > sorted[j].PopulationPct
		})

		p.Printf("\n\nNumber of People Affected by Natural Disasters in Japan (2011-2020): %.f\n\n", personsAffectedTotal)
		p.Println("By Prefecture, Per Capita (Mean):")

		for i, e := range sorted {
			p.Printf("%02d. %-15s  --  %-5.05f%%  %6.f / %12.f\n",
				i+1, e.Prefecture,
				e.PopulationPct*100, e.AffectedAvg, e.PopulationTotal,
			)
		}

		// Sort by all affected.
		sort.SliceStable(sorted, func(i, j int) bool {
			return sorted[i].OfAllAffectedPct > sorted[j].OfAllAffectedPct
		})

		p.Printf("\n\nNumber of People Affected by Natural Disasters in Japan (2011-2020): %.f\n\n", personsAffectedTotal)
		p.Println("By Prefecture, By Total Affected:")

		for i, e := range sorted {
			p.Printf("%02d. %-15s  --  %-5.02f%%   %7.f / %9.f\n",
				i+1, e.Prefecture,
				e.OfAllAffectedPct*100, e.AffectedTotal, personsAffectedTotal,
			)
		}
	}
}
