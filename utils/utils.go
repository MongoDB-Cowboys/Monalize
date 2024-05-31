package utils

import (
	"encoding/csv"
	"os"
	"regexp"
	"strings"
)

func ToCsvExport(data [][]string) error { // func for export data to csv file
	file, err := os.Create("result.csv")
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	for _, value := range data {

		if err := writer.Write(value); err != nil {
			return err
		}
	}

	return nil
}

// function that edits and returns readable indexes
func JsonToStr(args string) string {
	args = strings.Replace(args, `{"$numberInt":`, "", -1)
	args = strings.Replace(args, `{"$numberDouble":`, "", -1)
	re := regexp.MustCompile(`"`)
	args = re.ReplaceAllString(args, "")
	args = strings.TrimSuffix(args, "}")

	return args
}

func FilterPrintableASCII(data []byte) string {
	var filteredData []byte
	inLine := false

	for _, b := range data {
		if b == byte('\n') {
			inLine = false
		}
		// Check if the byte is within the range of printable ASCII characters
		if b >= 32 && b < 127 {
			filteredData = append(filteredData, b)
			inLine = true
		} else if !inLine && b == byte('\n') {
			filteredData = append(filteredData, b)
		}
	}

	return string(filteredData)
}
