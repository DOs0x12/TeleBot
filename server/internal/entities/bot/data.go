package bot

type Data struct {
	ChatID    int64
	Value     []byte
	IsCommand bool
	IsFile    bool
}
