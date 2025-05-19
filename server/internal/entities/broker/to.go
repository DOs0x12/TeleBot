package broker

type DataTo struct {
	CommName string
	ChatID   int64
	Value    []byte
	Token    string
	IsFile   bool
}
