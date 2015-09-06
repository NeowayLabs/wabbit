package server

type Queue struct {
	name string
	data chan []byte
}
