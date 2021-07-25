package stats

import (
	"net/url"
	"path"
	"regexp"
	"strings"
)

// Chapter represents a chapter in a statistical yearbook.
type Chapter struct {
	Name    string
	RootURL string
	Tables  []*File
}

func (s *Chapter) FindTables(patterns ...string) {
	chaptersHTML := get(s.RootURL)

	stripTags := regexp.MustCompile(`<.*?>`)

	re := regexp.MustCompile(`<a href=\"(.*?)\">(.*?)</a>`)
	res := re.FindAllStringSubmatch(chaptersHTML, -1)

	base, _ := url.Parse(s.RootURL)
	basePath := path.Dir(base.Path)

	for _, m := range res {
		relativePath := m[1]
		title := m[2]

		var found bool
		for _, p := range patterns {
			if strings.Contains(title, p) {
				found = true
				break
			}
		}
		if !found {
			continue
		}

		base.Path = path.Join(basePath, relativePath)
		url := base.String()
		cleanTitle := strings.Trim(stripTags.ReplaceAllString(title, " "), " ")

		s.Tables = append(s.Tables, &File{
			URL:   url,
			Title: cleanTitle,
		})
	}
}

func (b *Chapter) FindFile(title string) *File {
	for _, s := range b.Tables {
		if s.Title == title {
			return s
		}
	}
	return nil
}
