package mongo

type Record interface {
	GetCollection() string
}
