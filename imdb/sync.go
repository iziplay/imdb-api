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

var currentStats *database.Stats

func Synchronize(db *gorm.DB) error {
	if err := fetchTitles(db); err != nil {
		return err
	}

	if err := fetchTitlesAkas(db); err != nil {
		return err
	}

	if err := fetchTitlesEpisodes(db); err != nil {
		return err
	}

	if err := db.Create(&database.Synchronization{
		Date: time.Now().Format(time.RFC3339),
	}).Error; err != nil {
		return err
	}
	calculateStatistics(db)

	return nil
}

func fetchTitles(db *gorm.DB) error {
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

	reader, err := getReader("title.basics.tsv.gz") // ~9,682,270 lines
	if err != nil {
		return err
	}
	defer reader.Close()

	data := database.Title{}
	parser, _ := NewParser(reader, &data)
	parser.SetEmptyValue("\\N")

	batch := []database.Title{}
	batchCount := 0
	for {
		eof, err := parser.Next()
		if eof {
			if len(batch) >= 1 {
				if db.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "t_const"}},
					DoUpdates: clause.AssignmentColumns(upsertColumns),
				}).CreateInBatches(batch, len(batch)).Error != nil {
					return err
				}
			}
			return nil
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

func fetchTitlesAkas(db *gorm.DB) error {
	reader, err := getReader("title.akas.tsv.gz") // ~47,976,664 lines
	if err != nil {
		return err
	}
	defer reader.Close()

	data := database.TitleAka{}
	parser, _ := NewParser(reader, &data)
	parser.SetEmptyValue("\\N")

	batch := []database.TitleAka{}
	batchCount := 0
	for {
		eof, err := parser.Next()
		if eof {
			if len(batch) >= 1 {
				if db.Clauses(clause.OnConflict{
					DoNothing: true,
				}).CreateInBatches(batch, len(batch)).Error != nil {
					return err
				}
			}
			return nil
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
				DoNothing: true,
			}).CreateInBatches(batch, 100).Error != nil {
				return err
			}
			batchCount++
			batch = []database.TitleAka{}
		}
	}
}

func fetchTitlesEpisodes(db *gorm.DB) error {
	upsertColumns := []string{
		"parent",
		"season",
		"episode",
	}

	reader, err := getReader("title.episode.tsv.gz") // ~8,170,872 lines
	if err != nil {
		return err
	}
	defer reader.Close()

	data := database.TitleEpisode{}
	parser, _ := NewParser(reader, &data)
	parser.SetEmptyValue("\\N")

	batch := []database.TitleEpisode{}
	batchCount := 0
	for {
		eof, err := parser.Next()
		if eof {
			if len(batch) >= 1 {
				if db.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "t_const"}},
					DoUpdates: clause.AssignmentColumns(upsertColumns),
				}).CreateInBatches(batch, len(batch)).Error != nil {
					return err
				}
			}
			return nil
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
			batch = []database.TitleEpisode{}
		}
	}
}

func getReader(file string) (*gzip.Reader, error) {
	log.Println("download file " + file)
	out, err := os.Create("/tmp/" + file)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get("https://datasets.imdbws.com/" + file)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return nil, err
	}

	if err := out.Close(); err != nil {
		return nil, err
	}

	f, err := os.Open("/tmp/" + file)
	if err != nil {
		return nil, err
	}

	return gzip.NewReader(f)
}

func GetStatistics(db *gorm.DB) (*database.Stats, error) {
	if currentStats == nil {
		calculateStatistics(db)
	}

	return currentStats, nil
}

func calculateStatistics(db *gorm.DB) {
	var count int64
	db.Model(&database.Title{}).Count(&count)

	var akasCount int64
	db.Model(&database.TitleAka{}).Count(&akasCount)

	var episodesCount int64
	db.Model(&database.TitleEpisode{}).Count(&episodesCount)

	var types [](database.StatType)
	db.Model(&database.Title{}).Select("title_type, COUNT(*)").Group("title_type").Order("title_type").Find(&types)

	var genres [](database.StatGenre)
	db.Table(
		"(?) as g",
		db.Model(&database.Title{}).Select("unnest(genres) as genre"),
	).Select("g.genre, COUNT(*)").Group("genre").Order("genre").Find(&genres)

	var akas [](database.StatAka)
	db.Model(&database.TitleAka{}).Select("language, COUNT(*)").Group("language").Order("language").Find(&akas)

	var adult int64
	db.Model(&database.Title{}).Where("is_adult = ?", true).Count(&adult)

	var sync *database.Synchronization
	db.Order("date DESC").First(&sync)

	currentStats = &database.Stats{
		SynchronizationDate: sync.Date,
		Count:               uint(count),
		AkasCount:           uint(akasCount),
		EpisodesCount:       uint(episodesCount),
		Types:               types,
		Genres:              genres,
		Akas:                akas,
		Adult:               uint(adult),
	}
}
