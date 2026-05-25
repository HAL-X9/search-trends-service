package usecases

type StopListStorage interface {
	IsBanned(word string) bool
	Add(word string) error
	Remove(word string) error
}
