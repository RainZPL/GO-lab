package gol

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {

	ioCommand := make(chan ioCommand)
	ioIdle := make(chan bool)
	ioFilename := make(chan string)
	iooutput := make(chan uint8)
	ioinput := make(chan uint8)

	// ioFilename <- strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight)}, "x")

	// myfolder := `.\images\`

	// files, _ := ioutil.ReadDir(myfolder)

	// for _, file := range files {
	// 	ioFilename <- file.Name()
	// 	fmt.Println(file.Name())
	// }

	distributorChannels := distributorChannels{
		events:     events,
		ioCommand:  ioCommand,
		ioIdle:     ioIdle,
		ioFilename: ioFilename,
		input:      ioinput,
		output:     iooutput,
	}

	go distributor(p, distributorChannels, keyPresses)

	ioChannels := ioChannels{
		command:  ioCommand,
		idle:     ioIdle,
		filename: ioFilename,
		output:   iooutput,
		input:    ioinput,
	}

	go startIo(p, ioChannels)

}
