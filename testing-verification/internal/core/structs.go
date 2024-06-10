package core

type ComicWithID struct {
	ID    int
	Comic Comic
}

type ComicCount struct {
	UpdatedComics int `json:"updated_comics"`
	TotalComics   int `json:"total_comics"`
}

type Comic struct {
	ID       int
	URL      string
	Keywords string
}

type User struct {
	ID       int
	Username string
	Password string
	Role     string
}

var ComicInfo struct {
	Img        string `json:"img"`
	Alt        string `json:"alt"`
	Transcript string `json:"transcript"`
	Title      string `json:"title"`
}

type Config struct {
	SourceURL string `yaml:"source_url"`
	DBFile    string `yaml:"db_file"`
	Parallel  int    `yaml:"parallel"`
	IndexFile string `yaml:"index_file"`
	Port      int    `yaml:"port"`
	DSN       string `yaml:"dsn"`
	TokenTime int    `yaml:"token_max_time"`
	ConcLim   int    `yaml:"concurrency_limit"`
	RateLim   int    `yaml:"rate_limit"`
}