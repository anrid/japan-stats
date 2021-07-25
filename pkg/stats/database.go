package stats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type Database struct {
	IndexURL   string
	Books      []*Yearbook
	Downloaded time.Time
}

func NewDatabase(indexURL string) *Database {
	return &Database{IndexURL: indexURL}
}

func LoadIfExists(dbFile string) (db *Database, found bool) {
	_, err := os.Stat(dbFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}
		log.Panic(err)
	}

	data, err := ioutil.ReadFile(dbFile)
	if err != nil {
		log.Panic(err)
	}

	db = new(Database)
	err = json.Unmarshal(data, db)
	if err != nil {
		log.Panic(err)
	}

	return db, true
}

func (db *Database) Info() {
	uniqueFiles := make(map[string]bool)
	contentSize := 0
	firstYear := ""
	lastYear := ""

	for _, b := range db.Books {
		if firstYear == "" || firstYear > b.Year {
			firstYear = b.Year
		}
		if lastYear == "" || lastYear < b.Year {
			lastYear = b.Year
		}
		for _, s := range b.Chapters {
			for _, f := range s.Tables {
				_, found := uniqueFiles[f.URL]
				if found {
					log.Panicf("Found duplicate stats file URL: '%s'\n", f.URL)
				}
				uniqueFiles[f.URL] = true
				contentSize += len(f.ContentBase64)
			}
		}
	}

	fmt.Printf(`
	Years        : %s - %s
	Unique Files : %d
	Content Size : %d
	`, firstYear, lastYear, len(uniqueFiles), contentSize)
	fmt.Println("")
}

func (db *Database) Save(dbFile string) {
	js, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		log.Panic(err)
	}
	err = ioutil.WriteFile(dbFile, js, 0777)
	if err != nil {
		log.Panic(err)
	}
}

func Dump(o interface{}) {
	js, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		log.Panic(err)
	}
	println(string(js))
}

type DownloadYearbooksArgs struct {
	Stats []string
	Files []string
}

func (db *Database) DownloadYearbooks(a DownloadYearbooksArgs) {
	// Find matching stats categories.
	for _, b := range db.Books {
		b.FindChapters(a.Stats...)
	}

	// Find matching files for each stats category.
	for _, b := range db.Books {
		for _, s := range b.Chapters {
			s.FindTables(a.Files...)

			// Download file contents.
			for _, f := range s.Tables {
				f.DownloadContent()
			}
		}
	}

	db.Downloaded = time.Now()
}

func (db *Database) FindYearbooks() {
	html := get(db.IndexURL)

	re := regexp.MustCompile(`<a href=\"(.*?)\">.*? Yearbook (\d{4}).*?</a>`)
	res := re.FindAllStringSubmatch(html, -1)

	base, _ := url.Parse(db.IndexURL)
	basePath := path.Dir(base.Path)

	for _, m := range res {
		if len(m) == 3 {
			relativePath := m[1]
			year := m[2]

			base.Path = path.Join(basePath, relativePath)
			url := base.String()

			db.Books = append(db.Books, &Yearbook{RootURL: url, Year: year})
		}
	}
}

func (db *Database) GetTableFile(year, chapter, table string) (f *File, found bool) {
	// For example:
	// year    : "2020"
	// chapter : "Population and Households"
	// table   : "Population by Prefecture"
	for _, b := range db.Books {
		if b.Year != year {
			continue
		}

		for _, c := range b.Chapters {
			if !strings.Contains(c.Name, chapter) {
				continue
			}

			for _, t := range c.Tables {
				if !strings.Contains(t.Title, table) {
					continue
				}

				return t, true
			}
		}
	}
	return nil, false
}
