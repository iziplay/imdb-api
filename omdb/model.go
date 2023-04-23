package omdb

type TitleResponse struct {
	Title        string   `json:"Title"`
	Year         string   `json:"Year"`
	Rated        string   `json:"Rated"`
	Released     string   `json:"Released"`
	Runtime      string   `json:"Runtime"`
	Genre        string   `json:"Genre"`
	Director     string   `json:"Director"`
	Writer       string   `json:"Writer"`
	Actors       string   `json:"Actors"`
	Plot         string   `json:"Plot"`
	Language     string   `json:"Language"`
	Country      string   `json:"Country"`
	Awards       string   `json:"Awards"`
	Poster       string   `json:"Poster"`
	Ratings      []Rating `json:"Ratings"`
	Metascore    string   `json:"Metascore"`
	ImdbRating   string   `json:"imdbRating"`
	ImdbVotes    string   `json:"imdbVotes"`
	ImdbID       string   `json:"imdbID"`
	Type         string   `json:"Type"`
	TotalSeasons string   `json:"totalSeasons,omitempty"` // for series
	Dvd          string   `json:"DVD,omitempty"`          // for movies
	BoxOffice    string   `json:"BoxOffice,omitempty"`    // for movies
	Production   string   `json:"Production,omitempty"`   // for movies
	Website      string   `json:"Website,omitempty"`      // for movies
	Response     string   `json:"Response"`               // boolean "True" or "False"
}
type SearchResponse struct {
	Searc        string `json:"Search"`
	TotalResults string `json:"totalResults"` // number
}
type SearchResult struct {
	Title  string   `json:"Title"`
	Year   string   `json:"Year"`
	ImdbID string   `json:"imdbID"`
	Type   string   `json:"Type"`
	Poster []string `json:"Poster"`
}
type Error struct {
	Response string `json:"Response"`
	Error    string `json:"Error"`
}
type Rating struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
}

type Ratings struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
}
