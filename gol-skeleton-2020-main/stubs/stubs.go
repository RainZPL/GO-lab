package stubs

var Input = "DistributeOpreations.Input"
var CountAlivecells = "DistributeOpreations.CountAlivecells"



type Request struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
	Copyworld [][]byte
}

type Response struct {
	Copyworld [][]byte
	AliveCells int
}



