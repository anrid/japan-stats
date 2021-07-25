package stats

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	xlsx "github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/anrid/xls"
)

func ExtractDataFromFile(f *File, handler func(r []string)) {
	if strings.HasSuffix(f.URL, ".xlsx") {
		ExtractDataFromXLSX(f, handler)
	} else {
		ExtractDataFromXLS(f, handler)
	}
}

func ExtractDataFromXLS(f *File, handler func(r []string)) {
	fmt.Printf("Loading XLS data: %s\n", f.URL)

	rawData, _ := base64.StdEncoding.DecodeString(f.ContentBase64)
	reader := bytes.NewReader(rawData)
	wb, err := xls.OpenReader(reader, "utf-8")
	if err != nil {
		log.Panicf("Could not read XLS file '%s' (%s): %s", f.Title, f.URL, err.Error())
	}

	if sheet := wb.GetSheet(0); sheet != nil {
		fmt.Printf("Sheet name : %s\n", sheet.Name)
		fmt.Printf("Sheet rows : %d\n", sheet.MaxRow)

		for i := 0; i < int(sheet.MaxRow); i++ {
			row := sheet.Row(i)
			if row != nil {
				var cols []string
				for j := 0; j <= row.LastCol(); j++ {
					cols = append(cols, row.Col(j))
				}
				handler(cols)
			}
		}
	}
}

func ExtractDataFromXLSX(f *File, handler func(r []string)) {
	fmt.Printf("Loading XLSX data: %s\n", f.URL)

	rawData, _ := base64.StdEncoding.DecodeString(f.ContentBase64)
	reader := bytes.NewReader(rawData)
	wb, err := xlsx.OpenReader(reader)
	if err != nil {
		log.Panicf("Could not read XLSX file '%s' (%s): %s", f.Title, f.URL, err.Error())
	}

	defaultSheet := wb.GetSheetList()[0]

	rows, err := wb.GetRows(defaultSheet)
	if err != nil {
		log.Panicf("Could not get rows for default sheet '%s' : %s", defaultSheet, err.Error())
	}

	fmt.Printf("Sheet name : %s\n", defaultSheet)
	fmt.Printf("Sheet rows : %d\n", len(rows))

	for _, r := range rows {
		handler(r)
	}
}
