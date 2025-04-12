package util

type RoutingTable struct {
	Table [][]int32
}

func NewRoutingTable() *RoutingTable {
	rt := make([][]int32, DIGITS)
	for i := 0; i < DIGITS; i++ {
		rt[i] = make([]int32, RADIX)
		for j := range(rt[i]) {
			rt[i][j] = -1
		}
	}
	return &RoutingTable{
		table: rt,
	}
}
