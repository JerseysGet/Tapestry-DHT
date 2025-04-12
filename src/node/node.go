package main

import (
	util "Tapestry/util"
)

type Node struct {
	RT util.RoutingTable
	BP util.BackPointerTable
	ID uint64
}
