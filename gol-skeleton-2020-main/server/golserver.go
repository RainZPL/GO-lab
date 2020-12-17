package main

import (
	// "errors"
	"flag"
	"fmt"
	"net"
	"time"
	"math/rand"
	// "gol-skeleton-2020-main/gol"
	// "gol-skeleton-2020-main/stubs"
	"net/rpc"
	// "uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/stubs"
)

const alive = 255
const dead = 0

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}


type DistributeOpreations struct {}

func (s *DistributeOpreations) CountAlivecells(req stubs.Request, res *stubs.Response) (err error) {
	var p Params
	countAlivecells := 0
	Copyworld := req.Copyworld
	p.ImageHeight = req.ImageHeight
	p.ImageWidth = req.ImageWidth

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if Copyworld[y][x] == 255 {
				countAlivecells++
			}
		}
	}
	res.AliveCells = countAlivecells
	return
}

func (s *DistributeOpreations) Input(req stubs.Request, res *stubs.Response) (err error) {
	var p Params
	fmt.Println("Got Message: ",req.Turns)
	p.Turns = req.Turns
	p.Threads = req.Threads
	p.ImageHeight = req.ImageHeight
	p.ImageWidth = req.ImageWidth

	world := make([][]byte, p.ImageHeight)
	for w := range world {
		world[w] = make([]byte, p.ImageWidth)
	}


	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			world[y][x] = req.Copyworld[y][x]
		}
	}

	if p.Turns == 0{
		res.Copyworld = world
		return
	}


		workerHeight := p.ImageHeight / p.Threads

		out := make([]chan [][]uint8, p.Threads)
		for a := range out {
			out[a] = make(chan [][]uint8)
		}

		newPixelData := make([][]byte, 0)
		for w := range newPixelData {
			newPixelData[w] = make([]byte, 0)
		}

		for a := 0; a < p.Threads; a++ {

			if p.ImageHeight%p.Threads == 0 {
				go worker(a*workerHeight, (a+1)*workerHeight, 0, p.ImageWidth, world, out[a], p)

			} else {

				if a == p.Threads-1 {
					go worker(a*workerHeight, ((a+1)*workerHeight)+(p.ImageHeight%p.Threads), 0, p.ImageWidth, world, out[a], p)

				} else {
					go worker(a*workerHeight, (a+1)*workerHeight, 0, p.ImageWidth, world, out[a],p)
				}
			
			
			}
			part := <-out[a]
			newPixelData = append(newPixelData, part...)
		}

		
		
		// res.Copyworld = make([][]byte, p.ImageHeight)
		// for w := range world {
		// 	res.Copyworld[w] = make([]byte, p.ImageWidth)
		// }
		res.Copyworld = newPixelData
		fmt.Println(res.Copyworld)
		


	return
}	

func worker(startY, endY, startX, endX int, world [][]byte, out chan<- [][]uint8, p Params) {
	partHeight := endY - startY
	partWidth := endX - startX

	
	imagePart := make([][]byte, partHeight)
	for i := range imagePart {
		imagePart[i] = make([]byte, partWidth)
	}

	for h := 0; h < partHeight; h++ {
		for w := 0; w < partWidth; w++ {
			neighbour := 0
			for m := -1; m <= 1; m++ {
				for n := -1; n <= 1; n++ {
					if m != 0 || n != 0 {
						if world[(startY+m+p.ImageHeight)%p.ImageHeight][(w+n+p.ImageWidth)%p.ImageWidth] == alive {
							neighbour++
						}
					}
				}
			}
			if world[startY][w] == alive {
				if (neighbour == 2) || (neighbour == 3) {
					imagePart[h][w] = alive
					// countAlivecells++
				} else {
					imagePart[h][w] = dead
					// countAlivecells--
					// c.events <- CellFlipped{turn, util.Cell{X: w, Y: startY}}
				}
			} else {
				if neighbour == 3 {
					imagePart[h][w] = alive
					// countAlivecells++
					// c.events <- CellFlipped{turn, util.Cell{X: w, Y: startY}}

				} else {
					imagePart[h][w] = dead
					// countAlivecells--
				}
			}
		}
		startY++
	}

	out <- imagePart

}

func main(){
	pAddr := flag.String("port","8030","Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&DistributeOpreations{})
	listener, err := net.Listen("tcp", ":"+*pAddr)
	if err != nil {
         panic(err)
    }
	defer listener.Close()

	for {
		// conn, err := listener.Accept()
		rpc.Accept(listener)
	}
}