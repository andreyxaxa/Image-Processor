package dto

type Task struct {
	Data      []byte
	Operation string
	Width     *int
	Height    *int
	Text      *string
}
