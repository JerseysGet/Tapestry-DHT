package util

type Int32Set map[int32]struct{}

type BackPointerTable struct {
	set Int32Set
}

func NewBackPointerTable() *BackPointerTable {
	return &BackPointerTable{
		set: make(Int32Set),
	}
}
