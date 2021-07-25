package stats

import "encoding/base64"

// File represents a file containing statistical data.
// This is typically a table of data in Excel format.
type File struct {
	URL           string
	Title         string
	ContentBase64 string
}

func (f *File) DownloadContent() {
	data := download(f.URL)
	f.ContentBase64 = base64.StdEncoding.EncodeToString(data)
}
