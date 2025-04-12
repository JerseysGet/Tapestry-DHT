package util

type RoutingTable struct {
	Table [][]int
}

func NewRoutingTable() *RoutingTable {
	rt := make([][]int, DIGITS)
	for i := 0; i < DIGITS; i++ {
		rt[i] = make([]int, RADIX)
		for j := range(rt[i]) {
			rt[i][j] = -1
		}
	}
	return &RoutingTable{
		Table: rt,
	}
}
