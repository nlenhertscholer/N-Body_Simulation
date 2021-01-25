package main

import (
	"encoding/json"
	"flag"
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"io"
	"math/rand"
	"os"
	"proj3/phys"
	"proj3/qtree"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const usage = "Usage: ./sim [-w | -i=INTEGER] <X> <Y> <thread_count>\n" +
	"\t -w = Run this program in GUI mode.\n" +
	"\t -i = Number of updates to run. Must be greater than 0. (Note only have -w or -i, not both.\n" +
	"\t <X> = The width of the window. Positive Integer.\n" +
	"\t <Y> = The height of the window. Positive Integer.\n" +
	"\t <thread_count> = Number of maximum threads to use. Set to 0 to run in sequential mode."

// Global variables
var WindowWidth int
var WindowHeight int
var ThreadCount int

/*
 * Initialize the data array to hold the positions of the bodies
 *
 * bodies: slice holding physics bodies
 *
 * return: slice of maps holding the data for each object
 */
func initData(bodies []phys.Body) []map[string]interface{} {

	data := make([]map[string]interface{}, len(bodies))

	// Add the bodies into the tree
	for i := 0; i < len(bodies); i++ {
		origPosition := [][]float32{{bodies[i].Position.X, bodies[i].Position.Y}}
		data[bodies[i].Id] = map[string]interface{}{"Id": bodies[i].Id, "Position": origPosition}
	}

	return data

}

/*
 * Reads from a channel of bodies and adds them to a BHTree
 *
 * bodies: channel holding physics bodies to read from
 * cTree: channel to send a BHTree into
 */
func addToTree(bodies chan phys.Body, cTree chan *qtree.BHTree) {

	tree := qtree.NewBHTree(rl.NewRectangle(0, 0, float32(WindowWidth), float32(WindowHeight)))

	for body := range bodies {
		tree.Insert(body, 0)
	}

	cTree <- tree
}

/*
 * Calculate the physics applied to a body
 *
 * bodies: slice of physics body objects
 * data: slice of maps holding physics object data
 * draw: bool controlling the drawing of the tree
 */
func seqProcess(bodies []phys.Body, data []map[string]interface{}, draw bool) {

	// Channels for communication
	cBodies := make(chan phys.Body, len(bodies))
	cTree := make(chan *qtree.BHTree, 1)

	// Add the bodies into the tree
	for i := 0; i < len(bodies); i++ {
		cBodies <- bodies[i]
	}

	close(cBodies)

	// Add the bodies to the tree
	addToTree(cBodies, cTree)
	tree := <-cTree

	// Calculate the physics on each object
	for i := 0; i < len(bodies); i++ {
		tree.CalculateForces(&bodies[i])
		bodies[i].Position = bodies[i].Update()
		bodies[i].ZeroForce()

		// Add the updated position data
		if data != nil {
			data[i]["Position"] = append(data[i]["Position"].([][]float32),
				[]float32{bodies[i].Position.X, bodies[i].Position.Y})
		}
	}

	// Draw the tree
	if data == nil && draw {
		tree.DrawTree()
	}
}

/*
 * Process the data for each thread
 *
 * bodies: slice of physcis body objects
 * tree: pointer to a BHTree
 * cData: Channel to send updated bodies into
 * cBodies: Channel to send bodies for the creation of a tree
 * done: Channel signifying that all of the threads are done
 */
func parProcess(bodies []phys.Body, tree *qtree.BHTree,
	cData chan phys.Body, cBodies chan phys.Body, done chan bool) {

	// Iterate over the objects and apply the physics on them
	for i := 0; i < len(bodies); i++ {
		tree.CalculateForces(&bodies[i])
		bodies[i].Position = bodies[i].Update()
		bodies[i].ZeroForce()

		// Send to create the next tree
		cBodies <- bodies[i]

		// Send to start reading data
		if cData != nil {
			cData <- bodies[i]
		}
	}

	// This thread is done processing and if it is the last one, then close the channels
	done <- true
	if len(done) == ThreadCount {
		if cData != nil {
			close(cData)
		}
		close(cBodies)
	}
}

/*
 * Read in the data from Stdin
 *
 * b: Slice of physics bodies
 * wg: Waitgroup
 * mtx: Mutex
 * dec: Json Decoder
 * cBodies: Channel to send physics objects into to create trea
 * readDone: Channel to sync up the termination of threads
 */
func readData(b *[]phys.Body, wg *sync.WaitGroup, mtx *sync.Mutex,
	dec *json.Decoder, cBodies chan phys.Body, readDone chan bool) {

	if wg != nil {
		defer wg.Done()
	}

	for {

		// Holds JSON object specifying the command
		var inData map[string]interface{}

		var err error
		if mtx != nil {
			mtx.Lock()
			err = dec.Decode(&inData)
			mtx.Unlock()
		} else {
			err = dec.Decode(&inData)
		}

		// Process any error that might occur
		if err != nil {
			if err != io.EOF {
				panic(err)
			}
			// End of data
			if readDone != nil {
				// Signify other threads that I'm done
				readDone <- true
				if len(readDone) == ThreadCount {
					// CLose channel if this is the last thread
					close(cBodies)
				}
			}
			return
		}

		// Create a new physics body from the data
		pos := rl.NewVector2(float32(inData["Position"].([]interface{})[0].(float64)),
			float32(inData["Position"].([]interface{})[1].(float64)))
		vel := rl.NewVector2(float32(inData["Velocity"].([]interface{})[0].(float64)),
			float32(inData["Velocity"].([]interface{})[1].(float64)))
		body := phys.NewBody(float32(inData["Mass"].(float64)), int(inData["Id"].(float64)), pos, vel)

		// Add it to the slice
		if mtx != nil {
			mtx.Lock()
			*b = append(*b, body)
			mtx.Unlock()
		} else {
			*b = append(*b, body)
		}

		// Send body for tree creation
		if cBodies != nil {
			cBodies <- body
		}

		runtime.Gosched()

	}
}

/*
 * Run the sequential version of the program
 *
 * numIterations: Integer - number of time-steps to calculate
 */
func sequential(numIterations int) {

	// Read in the physics bodies
	dec := json.NewDecoder(os.Stdin)
	bodies := make([]phys.Body, 0)
	readData(&bodies, nil, nil, dec, nil, nil)

	// Slice to hold the data for each object
	bodiesData := initData(bodies)

	// Calculate the changed position for each object numIterations number of times
	for count := 0; count < numIterations; count++ {
		seqProcess(bodies, bodiesData, false)
	}

	// Output the data
	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(bodiesData)

}

/*
 * GUI version to run the parallel code
 *
 * bodies: Slice of physics objects
 * tree: Pointer to a BHTree
 * cTree: Channel to pass to addToTree() to send back a built tree
 * draw: Bool - draw the tree
 */
func guiParallel(bodies []phys.Body, tree *qtree.BHTree, cTree chan *qtree.BHTree, draw bool) {

	// Create a new channel of physics objects
	cBodies := make(chan phys.Body, len(bodies))

	// Calculate the length of data each thread will operate on
	sublength := len(bodies) / ThreadCount
	if sublength == 0 {
		ThreadCount = len(bodies)
		sublength = 1
	}

	// Channel to signal when each thread is done
	workersDone := make(chan bool, ThreadCount)

	// Start building the tree
	go addToTree(cBodies, cTree)

	// Send each thread to work on their respective subsections
	for i := 0; i < ThreadCount; i++ {

		min := sublength * i
		var max int
		if i == ThreadCount-1 {
			max = len(bodies)
		} else {
			max = min + sublength
		}

		go parProcess(bodies[min:max], tree, nil, cBodies, workersDone)
	}

	// Wait until the threads are done
	for {
		if len(workersDone) == ThreadCount {
			if draw {
				tree.DrawTree()
			}
			return
		}
		runtime.Gosched()
	}

}

/*
 * Run the parallel version of the console program
 *
 * numIterations: Integer - number of time-steps to calculate
 */
func parallel(numIterations int) {

	// Waitgroup and lock for reading in data
	var mtx sync.Mutex
	var wg sync.WaitGroup

	// Stdin reader and channels for reader threads
	dec := json.NewDecoder(os.Stdin)
	cBodies := make(chan phys.Body)             // readers send physics objects here so the tree can be built
	cTree := make(chan *qtree.BHTree)           // Tree builder sends tree into here
	readersDone := make(chan bool, ThreadCount) // Used to sync up the readers

	// Start building the tree
	go addToTree(cBodies, cTree)

	// Read in the data
	bodies := make([]phys.Body, 0)
	for i := 0; i < ThreadCount; i++ {
		wg.Add(1)
		go readData(&bodies, &wg, &mtx, dec, cBodies, readersDone)
	}
	wg.Wait()

	//Data to hold updated positions
	bodiesData := initData(bodies)

	// Calculate subsection of slice that each thread is going to work on
	sublength := len(bodies) / ThreadCount
	if sublength == 0 {
		ThreadCount = len(bodies)
		sublength = 1
	}

	for count := 0; count < numIterations; count++ {

		// Wait for tree to finish building --- Synchronous Barrier
		tree := <-cTree

		// Data to hold updated positions
		cBodies = make(chan phys.Body, len(bodies))
		cData := make(chan phys.Body, len(bodies))
		workersDone := make(chan bool, ThreadCount)

		// Start building the new tree
		go addToTree(cBodies, cTree)

		// Spawn off each thread to work on its part of the slice
		for i := 0; i < ThreadCount; i++ {
			min := sublength * i
			var max int
			if i == ThreadCount-1 {
				max = len(bodies)
			} else {
				max = min + sublength
			}
			go parProcess(bodies[min:max], tree, cData, cBodies, workersDone)

		}

		// Read in the data for each body
		for b := range cData {
			bodiesData[b.Id]["Position"] = append(bodiesData[b.Id]["Position"].([][]float32),
				[]float32{b.Position.X, b.Position.Y})
		}
	}

	// Output the data
	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(bodiesData)

}

/*
 * Run the program through the command line without a GUI
 *
 * numIterations: Integer - number of time-slices to calculate
 */
func consoleMode(numIterations int) {

	if ThreadCount == 0 {
		// Run in sequential mode
		sequential(numIterations)
	} else {
		// Run the parallel version
		runtime.GOMAXPROCS(ThreadCount)
		parallel(numIterations)
	}

}

/*
 *  Run the program through a GUI interface
 */
func guiMode() {

	// GUI mode is ran in both sequential and parallel, so there must be more than 0 threads
	if ThreadCount == 0 {
		fmt.Println("GUI Must be ran with more than 0 threads")
		fmt.Println(usage)
		return
	}

	runtime.GOMAXPROCS(ThreadCount)

	// Initialize GUI settings
	rand.Seed(time.Now().UnixNano())
	rl.InitWindow(int32(WindowWidth), int32(WindowHeight), "N-Body Simulation")
	rl.SetTargetFPS(60)

	// Read in the physics bodies
	dec := json.NewDecoder(os.Stdin)
	bodies := make([]phys.Body, 0)
	cTree := make(chan *qtree.BHTree)
	readData(&bodies, nil, nil, dec, nil, nil)

	var drawTree = false
	var parallelMode = false
	var first = true
	var setting = "Sequential (Press Space to Change)"

	// GUI loop
	for !rl.WindowShouldClose() {

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		var tree *qtree.BHTree // Declare the tree object

		if rl.IsKeyPressed(rl.KeyB) {
			// Check whether we should draw the tree or not
			drawTree = !drawTree
		}

		if rl.IsKeyPressed(rl.KeySpace) {
			// Switch between parallel and sequential mode
			parallelMode = !parallelMode

			if parallelMode {

				// This clears the previous tree that's waiting if this isn't the first switch to parallel
				switch first {
				case true:
					first = false
				case false:
					<-cTree
				}

				// Add the bodies to the tree
				cBodies := make(chan phys.Body, len(bodies))
				go addToTree(cBodies, cTree)
				for i := 0; i < len(bodies); i++ {
					cBodies <- bodies[i]
				}
				close(cBodies)

				setting = "Parallel"
			} else {
				setting = "Sequential"
			}
		}

		// Compute the physics
		if parallelMode {
			tree = <-cTree // --- Synchronous Barrier
			guiParallel(bodies, tree, cTree, drawTree)
		} else {
			seqProcess(bodies, nil, drawTree)
		}

		// Draw each object
		for i := 0; i < len(bodies); i++ {
			rl.DrawCircleLines(int32(bodies[i].Position.X), int32(bodies[i].Position.Y), bodies[i].Radius, rl.Green)
		}

		// Print MetaData to Screen
		frameTime := fmt.Sprintf("FrameTime: %.4fms", rl.GetFrameTime()*1000)
		fps := fmt.Sprintf("FPS: %v", rl.GetFPS())
		rl.DrawText(fps, int32(WindowWidth)-200, 10, 15, rl.White)
		rl.DrawText(frameTime, int32(WindowWidth)-200, 35, 15, rl.White)
		rl.DrawText(setting, 20, 10, 15, rl.White)

		rl.EndDrawing()
	}

	rl.CloseWindow()
}

func main() {

	// Flag commands for CLI
	wPtr := flag.Bool("w", false, "Run this program in GUI mode.")
	iPtr := flag.Int("i", -1, "Number of updates to run.")

	// Parse commands and error check the input
	flag.Parse()
	args := flag.Args()
	if (*wPtr == false && *iPtr == -1) || (*wPtr == true && *iPtr != -1) || len(args) != 3 {
		fmt.Println(usage)
		os.Exit(0)
	}

	WindowWidth, _ = strconv.Atoi(args[0])
	WindowHeight, _ = strconv.Atoi(args[1])
	ThreadCount, _ = strconv.Atoi(args[2])

	if WindowWidth <= 0 || WindowHeight <= 0 || ThreadCount < 0 {
		fmt.Printf("X and Y must be greater than 0.\n"+
			"thread_count must be greater than -1. Not [%v, %v, %v]\n", WindowWidth, WindowHeight, ThreadCount)
		fmt.Println(usage)
		os.Exit(0)
	}

	if *wPtr {
		guiMode()
	} else {
		consoleMode(*iPtr)
	}
}
