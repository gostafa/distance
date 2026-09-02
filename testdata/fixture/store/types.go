package store

type Store interface {
	Put(id int, v any) error
}
