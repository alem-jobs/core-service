package model

type Category struct {
	ID       int
	Name     string
	ParentID *int
	Left     int
	Right    int
	Depth    int
}
