package util

type Int32Set map[int32]struct{}

type BackPointerTable struct {
	Set Int32Set
}

func NewBackPointerTable() *BackPointerTable {
	return &BackPointerTable{
		Set: make(Int32Set),
	}
}
