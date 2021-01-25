package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const usage = "Usage: generate <num_of_obj> <x> <y> [-s] \n" +
	"\t <num_of_obj> = the number of objects you want to generate\n" +
	"\t <x> = the width of the window. Integer\n" +
	"\t <y> = the height of the window. Integer\n" +
	"\t -s = Have all objects start with 0 initial velocity."

// Constants to change values
const MaxMass = 3
const MinMass = 0.5
const MaxVel = 2
const MinVel = -2

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	// Read in CL arguments
	args := os.Args[1:]
	if len(args) < 3 || len(args) > 4 {
		fmt.Println(usage)
		os.Exit(0)
	}

	// Parse the arguments
	numObj, _ := strconv.Atoi(args[0])
	X, _ := strconv.Atoi(args[1])
	Y, _ := strconv.Atoi(args[2])
	stationary := false
	if len(args) == 4 && args[3] == "-s" {
		stationary = true
	}

	// Decoder to output JSON
	dec := json.NewEncoder(os.Stdout)

	for i := 0; i < numObj; i++ {
		// Holds the JSON object_data
		objectData := make(map[string]interface{})
		objectData["Command"] = "ADD"
		objectData["Mass"] = (rand.Float32() * (MaxMass - MinMass)) + MinMass
		objectData["Position"] = [2]int{rand.Intn(X), rand.Intn(Y)}
		vel := [2]float32{0, 0}
		if !stationary {
			vel[0] = (rand.Float32() * (MaxVel - MinVel)) + MinVel
			vel[1] = (rand.Float32() * (MaxVel - MinVel)) + MinVel
		}
		objectData["Velocity"] = vel
		objectData["Id"] = i
		_ = dec.Encode(objectData)
	}
}
