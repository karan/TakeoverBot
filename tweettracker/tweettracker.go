package tweettracker

import (
	"encoding/csv"
	"log"
	"os"
)

type DataLine struct {
	UserName     string
	Domain       string
	ReplyTweetId string

	IDStr         string
	CreatedAt     string
	FavoriteCount string
	ReplyCount    string
	RetweetCount  string
	QuoteCount    string
	FullText      string
}

const (
	DataFilePath = "posted.csv"
)

var (
	// TODO: make this a map for faster lookups
	Data []*DataLine
)

// Read a local file of CSV and convert into list
func Init() {
	file, err := os.OpenFile(DataFilePath, os.O_RDWR|os.O_CREATE, 0755)
	log.Printf("tweettracker init with file %s", file.Name())
	defer file.Close()
	if err != nil {
		log.Fatalf(err.Error())
	}
	reader := csv.NewReader(file)
	lines, _ := reader.ReadAll()
	defer log.Printf("Ready to rumble...")

	log.Printf("Read %s, and found %d lines\n", file.Name(), len(lines))
	if len(lines) == 0 {
		return
	}

	lines = append(lines[:0], lines[0+1:]...)

	for _, line := range lines {
		Data = append(Data, &DataLine{
			UserName:      line[0],
			Domain:        line[1],
			ReplyTweetId:  line[2],
			IDStr:         line[3],
			CreatedAt:     line[4],
			FavoriteCount: line[5],
			ReplyCount:    line[6],
			RetweetCount:  line[7],
			QuoteCount:    line[8],
			FullText:      line[9],
		})
	}
}

// Returns true if the passed idStr already exists.
func Exists(idStr string) bool {
	for _, dataLine := range Data {
		if dataLine.IDStr == idStr {
			log.Printf("Already replied to %s with tweet ID", idStr, dataLine.ReplyTweetId)
			return true
		}
	}
	return false
}

func Add(line *DataLine) {
	Data = append(Data, line)
	file, err := os.OpenFile(DataFilePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)

	for _, dataLine := range Data {
		var row []string
		row = append(row, dataLine.UserName)
		row = append(row, dataLine.Domain)
		row = append(row, dataLine.ReplyTweetId)
		row = append(row, dataLine.IDStr)
		row = append(row, dataLine.CreatedAt)
		row = append(row, dataLine.FavoriteCount)
		row = append(row, dataLine.ReplyCount)
		row = append(row, dataLine.RetweetCount)
		row = append(row, dataLine.QuoteCount)
		row = append(row, dataLine.FullText)
		err := writer.Write(row)
		if err != nil {
			panic(err)
		}
	}
	writer.Flush()
	err = writer.Error()
	if err != nil {
		panic(err)
	}
}
