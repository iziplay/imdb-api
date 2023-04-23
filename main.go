package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/iziplay/imdb-api/database"
	"github.com/iziplay/imdb-api/imdb"
	"github.com/iziplay/imdb-api/omdb"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func main() {
	running := false

	log.Println("open database")
	db, err := gorm.Open(postgres.Open(fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DATABASE"),
		os.Getenv("POSTGRES_PORT"),
	)), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "imdb_",
		},
	})
	if err != nil {
		panic("failed to connect database")
	}

	log.Println("migrate")
	db.AutoMigrate(&database.Synchronization{})
	db.AutoMigrate(&database.Title{})

	e := echo.New()
	e.HideBanner = true
	e.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	v1 := e.Group("/v1", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !running {
				return c.JSON(http.StatusServiceUnavailable, map[string]string{
					"error": "syncing",
					"info":  "First synchronization in progress, please try again later.",
				})
			}

			return next(c)
		}
	})
	v1.GET("/imdb/:imdb", func(c echo.Context) error {
		imdb := c.Param("imdb")
		switch {
		case regexp.MustCompile("^[0-9]{7,8}$").MatchString(imdb):
			imdb = "tt" + imdb
		case regexp.MustCompile("^tt[0-9]{7,8}$").MatchString(imdb):
			break
		default:
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "imdb format is invalid",
				"info":  "The IMDB ID provided is not a IMDB title ID. It should be '[tt]0000000[0]'.",
			})
		}

		var title *database.Title
		if err := db.First(&title, "t_const = ?", imdb).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "imdb title cannot be found",
					"info":  "The IMDB ID provided does not exist. Synchronization is done every 24 hours.",
				})
			}
			panic(err)
		}

		return c.JSON(http.StatusOK, title)
	})
	v1.GET("/omdb", func(c echo.Context) error {
		// this tries to emulate OMDB response
		// TODO: get by title `t` param
		// TODO: search by title `s` param
		imdb := c.QueryParam("i")
		switch {
		case regexp.MustCompile("^[0-9]{7,8}$").MatchString(imdb):
			imdb = "tt" + imdb
		case regexp.MustCompile("^tt[0-9]{7,8}$").MatchString(imdb):
			break
		default:
			return c.JSON(http.StatusOK, &omdb.Error{
				Response: "False",
				Error:    "Incorrect IMDb ID.",
			})
		}

		var title *database.Title
		if err := db.First(&title, "t_const = ?", imdb).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusOK, &omdb.Error{
					Response: "False",
					Error:    "Error getting data.",
				})
			}
			panic(err)
		}

		t := ""
		year := fmt.Sprintf("%d", title.StartYear)
		switch title.TitleType {
		case "tvSeries":
			t = "series"
			year = fmt.Sprintf("%d-%d", title.StartYear, title.EndYear)
		}
		return c.JSON(http.StatusOK, &omdb.TitleResponse{
			Title:    title.PrimaryTitle,
			Year:     year,
			Runtime:  fmt.Sprintf("%d mn", title.RuntimeMinutes),
			Genre:    strings.Join(title.Genres, ", "),
			ImdbID:   title.TConst,
			Type:     t,
			Response: "true",
		})
	})
	v1.GET("/title/:title/year/:year", func(c echo.Context) error {
		var title *database.Title
		if err := db.First(&title, "original_title = ? AND start_year = ?", c.Param("title"), c.Param("year")).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "imdb title cannot be found",
					"info":  "Cannot find a title with theses parameters.",
				})
			}
			panic(err)
		}

		return nil
	})
	v1.GET("/statistics", func(c echo.Context) error {
		stats, err := imdb.GetStatistics(db)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "statistics cannot be calculated",
				"info":  "The database cannot be accessed correctly, please try again later.",
			})
		}
		return c.JSON(http.StatusOK, stats)
	})
	go func() {
		e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
	}()

	var sync *database.Synchronization
	err = db.Order("date DESC").First(&sync).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			panic(err)
		}
		sync = nil
	}
	if sync == nil {
		if imdb.FetchTitles(db) != nil {
			panic(err)
		}
	} else {
		d, err := time.Parse(time.RFC3339, sync.Date)
		if err != nil {
			panic(err)
		}

		// sync was more than 1 day ago: sync now
		if d.Add(24 * time.Hour).Before(time.Now()) {
			if imdb.FetchTitles(db) != nil {
				panic(err)
			}
		}
	}

	log.Println("sync ok")
	running = true

	// TODO: first sync should start at last sync + 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	log.Println("next sync at " + time.Now().Add(24*time.Hour).Format(time.RFC3339))
	for {
		select {
		case t := <-ticker.C:
			if imdb.FetchTitles(db) != nil {
				panic(err)
			}
			log.Println("next sync at " + t.Add(24*time.Hour).Format(time.RFC3339))
		}
	}
}
