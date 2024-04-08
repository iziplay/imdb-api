package database

import "github.com/lib/pq"

type Title struct {
	TConst         string         `gorm:"primaryKey" tsv:"tconst" json:"imdb"`
	TitleType      string         `tsv:"titleType" json:"type"`
	PrimaryTitle   string         `tsv:"primaryTitle" json:"title"`
	OriginalTitle  string         `tsv:"originalTitle" json:"originalTitle"`
	IsAdult        bool           `tsv:"isAdult" json:"isAdult"`
	StartYear      int            `tsv:"startYear" json:"startYear"`
	EndYear        int            `tsv:"endYear" json:"endYear"`
	RuntimeMinutes int            `tsv:"runtimeMinutes" json:"runtimeMinutes"`
	Genres         pq.StringArray `gorm:"type:text[]" tsv:"genres" json:"genres"`

	Episode *TitleEpisode `gorm:"foreignKey:TConst" json:"episode"`
	Akas    []TitleAka    `gorm:"foreignKey:TConst" json:"akas"`
}

type TitleAka struct {
	TConst   string `gorm:"primaryKey;index" tsv:"titleId" json:"-"`
	Region   string `gorm:"primaryKey" tsv:"region" json:"region,omitempty"`
	Language string `gorm:"primaryKey;index" tsv:"language" json:"language,omitempty"`
	Title    string `gorm:"primaryKey" tsv:"title" json:"title"`
}

type TitleEpisode struct {
	TConst  string `gorm:"primaryKey" tsv:"tconst" json:"-"`
	Parent  string `tsv:"parentTconst" json:"parent"`
	Season  int    `tsv:"seasonNumber" json:"season"`
	Episode int    `tsv:"episodeNumber" json:"episode"`
}
