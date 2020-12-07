package gol

import (
	"fmt"
	"time"

	"uk.ac.bris.cs/gameoflife/util"
)

const alive = 255
const dead = 0

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	output     chan<- uint8
	input      <-chan uint8
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels, keyPresses <-chan rune) {

	// Create a 2D slice to store the world.
	world := make([][]byte, p.ImageHeight)
	for w := range world {
		world[w] = make([]byte, p.ImageWidth)
	}

	//change iocommand into input state, read the image
	c.ioCommand <- ioInput
	//write the filename with "p.ImageWidth X p.ImageHeight"
	filename := fmt.Sprintf("%vx%v", p.ImageWidth, p.ImageHeight)
	//send filename to ioFilename
	c.ioFilename <- filename

	//send every input pixel to the world 2D matrix byte by byte in rows
	//finish reading the image
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			world[y][x] = <-c.input
		}
	}

	//For all initially alive cells send a CellFlipped Event.
	turn := 0
	var alivecells []util.Cell

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == alive {
				c.events <- CellFlipped{turn, util.Cell{X: x, Y: y}}

			}
		}
	}

	//create a 2s ticker
	ticker := time.NewTicker(2 * time.Second)

	// TODO: Execute all turns of the Game of Life.

	for turn < p.Turns {

		select {

		case <-ticker.C:
			countAlivecells := 0
			if turn > 0 {
				for y := 0; y < p.ImageHeight; y++ {
					for x := 0; x < p.ImageWidth; x++ {
						if world[y][x] == alive {
							countAlivecells++
						}
					}
				}
				c.events <- AliveCellsCount{turn, countAlivecells}
			}
			c.events <- AliveCellsCount{turn, countAlivecells}

		case press := <-keyPresses:
			if press == 's' {
				writePgm(p, c, world, turn)
			} else if press == 'q' {
				writePgm(p, c, world, turn)
				c.events <- StateChange{turn, Quitting}
				return
			} else if press == 'p' {
				c.events <- StateChange{turn, Paused}
				fmt.Println(turn)
				for {
					continuepress := <-keyPresses
					if continuepress == 'p' {
						c.events <- StateChange{turn, Executing}
						fmt.Println("Continuing")
						break
					}
				}

			}

		default:

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
					go worker(a*workerHeight, (a+1)*workerHeight, 0, p.ImageWidth, world, out[a], p, c, turn)

				} else {

					if a == p.Threads-1 {
						go worker(a*workerHeight, ((a+1)*workerHeight)+(p.ImageHeight%p.Threads), 0, p.ImageWidth, world, out[a], p, c, turn)

					} else {
						go worker(a*workerHeight, (a+1)*workerHeight, 0, p.ImageWidth, world, out[a], p, c, turn)
					}
				}
				part := <-out[a]
				newPixelData = append(newPixelData, part...)
			}

			world = newPixelData

			// c.ioCommand <- ioOutput
			// filename = fmt.Sprintf("%vx%v%v", p.ImageWidth, p.ImageHeight, turn)
			// c.ioFilename <- filename
			// for y := 0; y < p.ImageHeight; y++ {
			// 	for x := 0; x < p.ImageWidth; x++ {
			// 		c.output <- world[y][x]
			// 	}
			// }

			c.events <- TurnComplete{turn}
			turn++
		}
	}

	ticker.Stop()

	// c.ioCommand <- ioOutput
	// pgmFilename := fmt.Sprintf("%vx%vx%v", p.ImageWidth, p.ImageHeight, turn)
	// c.ioFilename <- pgmFilename
	// for y := 0; y < p.ImageHeight; y++ {
	// 	for x := 0; x < p.ImageWidth; x++ {
	// 		c.output <- world[y][x]
	// 	}
	// }

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == alive {
				alivecells = append(alivecells, util.Cell{X: x, Y: y})
			}
		}
	}

	// TODO: Send correct Events when required, e.g. CellFlipped, TurnComplete and FinalTurnComplete.
	//		 See event.go for a list of all events.

	// c.events <- AliveCellsCount{turn, countAlivecells}

	// c.events <- TurnComplete{turn}
	c.events <- FinalTurnComplete{turn, alivecells}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

func worker(startY, endY, startX, endX int, world [][]byte, out chan<- [][]uint8, p Params, c distributorChannels, turn int) {
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
					c.events <- CellFlipped{turn, util.Cell{X: w, Y: startY}}
				}
			} else {
				if neighbour == 3 {
					imagePart[h][w] = alive
					// countAlivecells++
					c.events <- CellFlipped{turn, util.Cell{X: w, Y: startY}}

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

func writePgm(p Params, c distributorChannels, world [][]byte, turn int) {
	c.ioCommand <- ioOutput
	pgmFilename := fmt.Sprintf("%vx%vx%v", p.ImageWidth, p.ImageHeight, turn)
	c.ioFilename <- pgmFilename
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			c.output <- world[y][x]
		}
	}
	c.events <- ImageOutputComplete{turn, pgmFilename}
}
