package main

import "github.com/anrid/japan-stats/pkg/stats"

const japanStatsDatabase = "/tmp/japan-stats.json"
const japanStatsHomeURL = "https://www.stat.go.jp/english/data/nenkan/index.html"

func main() {
	db, found := stats.LoadIfExists(japanStatsDatabase)
	if !found {
		db = stats.NewDatabase(japanStatsHomeURL)
		db.FindYearbooks()
		db.DownloadYearbooks(stats.DownloadYearbooksArgs{
			Stats: []string{
				"Disasters and Accidents",
				"Population and Households",
			},
			Files: []string{
				"Natural Disasters by Prefecture",
				"Population by Prefecture",
			},
		})
		db.Save(japanStatsDatabase)
	}

	db.Info()
}
