package stats

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func get(url string) string {
	return string(download(url))
}

func download(url string) []byte {
	fmt.Printf("Download: '%s'\n", url)

	// Sleep to not hammer stats site!
	time.Sleep(250 * time.Millisecond)

	resp, err := http.Get(url)
	if err != nil {
		log.Panic(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}

	return data
}
