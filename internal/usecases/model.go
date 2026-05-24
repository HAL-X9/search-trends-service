package usecases

type SearchEvent struct {
	Query     string
	UserID    string
	IPAddress string
	Timestamp int64
}

type WordStat struct {
	Word  string
	Count int64
}
