package util

type RoutingTable struct {
	table [][]int32
}

func NewRoutingTable() *RoutingTable {
	rt := make([][]int32, DIGITS)
	for i := 0; i < DIGITS; i++ {
		rt[i] = make([]int32, RADIX)
	}
	return &RoutingTable{
		table: rt,
	}
}
