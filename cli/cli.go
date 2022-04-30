package cli

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"alien-invasion/simulation"
)

const (
	// DefaultIterations used if number of iteration is not otherwise specified
	DefaultIterations int = 10000
	// DefaultNumberOfAliens used if number of Aliens is not otherwise specified
	DefaultNumberOfAliens int = 10
	// DefaultWorldFile used if World file is not otherwise specified
	DefaultWorldFile = "./test/example.txt"
	// DefaultIntelFile used if intel file is not otherwise specified
	DefaultIntelFile = "./test/aliens.txt"
)

var (
	entropy                              int64
	iterations, alienNumber              int
	simulationName, worldFile, intelFile string
)

// init cli flags
func init() {
	flag.Int64Var(&entropy, "entropy", 0, "random number used as entropy seed")
	flag.IntVar(&iterations, "iterations", DefaultIterations, "number of iterations")
	flag.IntVar(&alienNumber, "aliens", DefaultNumberOfAliens, "number of aliens invading")
	flag.StringVar(&simulationName, "simulation", "", "name hashed and used as entropy seed")
	flag.StringVar(&worldFile, "world", DefaultWorldFile, "a file used as world map input")
	flag.StringVar(&intelFile, "intel", DefaultIntelFile, "a file used to identify aliens")
	flag.Parse()
}

// checkFlags validates input flags
func checkFlags() error {
	if alienNumber <= 0 {
		return errors.New("Aliens number must be > 0")
	}
	if iterations <= 0 {
		return errors.New("Iterations number must be > 0")
	}
	if len(worldFile) == 0 {
		return errors.New("World map file path is empty")
	}
	return nil
}

// Execute command and set flags appropriately.
func Execute() {
	// Check input flags for errors
	if err := checkFlags(); err != nil {
		fmt.Printf("Error while checking flags: %s\n", err)
		flag.Usage()
		os.Exit(1)
	}
	// Read input file
	fmt.Printf("Reading world map from file: %s\n", worldFile)
	world, in, err := simulation.ReadWorldMapFile(worldFile)
	if err != nil {
		fmt.Printf("Could not read world from map file \"%s\" with error: %s\n", worldFile, err)
		os.Exit(1)
	}
	// Build a simulator
	fmt.Printf("Generating %d random Aliens...\n", alienNumber)
	r := buildRand()
	aliens := simulation.RandAliens(alienNumber, r)
	if intelFile != "" {
		if err := simulation.IdentifyAliens(aliens, intelFile); err != nil {
			fmt.Printf("Could not read aliens intel from file \"%s\" with error: %s\n", intelFile, err)
			os.Exit(1)
		}
	}
	sim := simulation.NewSimulation(r, iterations, world, aliens)
	// Start the simulation and print any errors
	if err := sim.Start(); err != nil {
		fmt.Printf(formatImportantMessage("Error while running simulation: %s"), err)
		os.Exit(1)
	}
	// Success
	fmt.Printf(formatImportantMessage("Simulation Success"))
	fmt.Print(in.FilterDestroyed(world))
}

// buildRand build a pseudorandom numbers generator from input flags
func buildRand() *rand.Rand {
	var seed int64
	var source rand.Source
	if entropy != 0 {
		seed = entropy
		source = rand.NewSource(entropy)
		fmt.Printf("Entropy: using provided entropy\n")
	} else if len(simulationName) > 0 {
		hash := sha256.Sum256([]byte(simulationName))
		seed = int64(binary.BigEndian.Uint64(hash[:8]))
		source = rand.NewSource(seed)
		fmt.Printf("Entropy: using first 8 bytes of sha256(\"%v\")\n", simulationName)
	} else {
		seed = time.Now().UnixNano()
		source = rand.NewSource(seed)
		fmt.Printf("Entropy: using current unix time as a random source\n")
	}
	fmt.Printf("Entropy Seed: %v\n", seed)
	return rand.New(source)
}

// formatImportantMessage formats an important message ;)
func formatImportantMessage(msg string) string {
	line := strings.Repeat("=", len(msg))
	out := fmt.Sprintf("\n\n")
	out += fmt.Sprintf("%s\n", line)
	out += fmt.Sprintf("%s\n", msg)
	out += fmt.Sprintf("%s\n", line)
	return out
}
