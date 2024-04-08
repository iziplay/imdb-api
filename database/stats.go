package database

type Stats struct {
	SynchronizationDate string      `json:"lastSync"`
	Count               uint        `json:"count"`
	AkasCount           uint        `json:"akasCount"`
	Types               []StatType  `json:"types"`
	Genres              []StatGenre `json:"genres"`
	Akas                []StatAka   `json:"akas"`
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

type StatAka struct {
	Language string `json:"language"`
	Count    uint   `json:"count"`
}
