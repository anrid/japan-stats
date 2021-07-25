package stats

import (
	"log"
	"net/url"
	"path"
	"regexp"
)

// Yearbook
//
// The Japan Statistical Yearbook is the comprehensive and systematic
// summary of basic statistical information of Japan covering wide-ranging
// fields such as Land, Population, Economy, Society, Culture, and so on.
// The yearbook covers all fields of statistics published by government and private organisations.
//
// See: https://www.stat.go.jp/english/data/nenkan/index.html
//
type Yearbook struct {
	Year     string
	RootURL  string
	Chapters []*Chapter
}

func (b *Yearbook) FindChapters(patterns ...string) {
	rootHTML := get(b.RootURL)

	for _, p := range patterns {
		re := regexp.MustCompile(`<a href=\"(.*?)\">.*?` + p + `</a>`)
		res := re.FindStringSubmatch(rootHTML)

		if len(res) == 2 {
			relativePath := res[1]

			base, _ := url.Parse(b.RootURL)
			basePath := path.Dir(base.Path)
			base.Path = path.Join(basePath, relativePath)

			url := base.String()

			b.Chapters = append(b.Chapters, &Chapter{
				Name:    p,
				RootURL: url,
			})
		} else {
			log.Panicf("Could find any book matches for regexp: '%s'\n", re.String())
		}
	}
}
