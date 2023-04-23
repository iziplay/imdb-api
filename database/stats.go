package database

type Stats struct {
	SynchronizationDate string      `json:"lastSync"`
	Count               uint        `json:"count"`
	Types               []StatType  `json:"types"`
	Genres              []StatGenre `json:"genres"`
	Adult               uint        `json:"adult"`
}

type StatType struct {
	TitleType string `json:"type"`
	Count     uint   `json:"count"`
}

type StatGenre struct {
	Genre string `json:"genre"`
	Count uint   `json:"count"`
}
