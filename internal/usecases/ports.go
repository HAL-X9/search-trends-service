package usecases

type StopListStorage interface {
	IsBanned(word string) bool
}
