package imdb

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/iziplay/imdb-api/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func FetchTitles(db *gorm.DB) error {
	upsertColumns := []string{
		"title_type",
		"primary_title",
		"original_title",
		"is_adult",
		"start_year",
		"end_year",
		"runtime_minutes",
		"genres",
	}

	log.Println("download file")
	out, err := os.Create("/tmp/titles.basics.tsv.gz")
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get("https://datasets.imdbws.com/title.basics.tsv.gz") // ~9682270 lines
	if err != nil {
		return err
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	f, err := os.Open("/tmp/titles.basics.tsv.gz")
	if err != nil {
		return err
	}

	log.Println("ungzip")
	reader, err := gzip.NewReader(f)
	defer reader.Close()
	if err != nil {
		return err
	}

	log.Println("parse")
	data := database.Title{}
	parser, _ := NewParser(reader, &data)
	parser.SetEmptyValue("\\N")

	batch := []database.Title{}
	batchCount := 0
	total := 0
	for {
		eof, err := parser.Next()
		total += 1
		if eof {
			log.Println("got eof after line " + strconv.Itoa(total))
			if len(batch) >= 1 {
				if db.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "t_const"}},
					DoUpdates: clause.AssignmentColumns(upsertColumns),
				}).CreateInBatches(batch, len(batch)).Error != nil {
					return err
				}
			}
			return db.Create(&database.Synchronization{
				Date: time.Now().Format(time.RFC3339),
			}).Error
		}
		if err != nil {
			panic(err)
		}
		batch = append(batch, data)

		if len(batch) == 100 {
			if batchCount%200 == 0 {
				log.Println("add batch from line " + strconv.Itoa(batchCount*100))
			}
			if db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "t_const"}},
				DoUpdates: clause.AssignmentColumns(upsertColumns),
			}).CreateInBatches(batch, 100).Error != nil {
				return err
			}
			batchCount++
			batch = []database.Title{}
		}
	}
}
