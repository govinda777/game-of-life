package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// Universe represents the 2D grid containing GoL cells.
type Universe struct {
	width  int
	height int
	grid   [][]bool
}

// NewUniverse creates a new blank Universe.
func NewUniverse(width, height int) *Universe {
	grid := make([][]bool, height)
	for i := range grid {
		grid[i] = make([]bool, width)
	}
	return &Universe{
		width:  width,
		height: height,
		grid:   grid,
	}
}

// Clone creates a deep copy of the Universe.
func (u *Universe) Clone() *Universe {
	clone := NewUniverse(u.width, u.height)
	for i := 0; i < u.height; i++ {
		copy(clone.grid[i], u.grid[i])
	}
	return clone
}

// Set sets the state of a cell.
func (u *Universe) Set(x, y int, state bool) {
	u.grid[y][x] = state
}

// Get gets the state of a cell.
func (u *Universe) Get(x, y int) bool {
	return u.grid[y][x]
}

// Neighbors returns the number of live neighbors of a cell at (x, y) with toroidal wrap-around.
func (u *Universe) Neighbors(x, y int) int {
	count := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			// Toroidal wrap-around logic
			nx := (x + dx + u.width) % u.width
			ny := (y + dy + u.height) % u.height
			if u.grid[ny][nx] {
				count++
			}
		}
	}
	return count
}

// Next calculates the next generation of the Universe using concurrency.
func (u *Universe) Next() *Universe {
	nextU := NewUniverse(u.width, u.height)
	numCPUs := runtime.NumCPU()
	if numCPUs <= 0 {
		numCPUs = 1
	}
	if numCPUs > u.height {
		numCPUs = u.height
	}

	var wg sync.WaitGroup
	wg.Add(numCPUs)

	// Determine chunks of rows to process per goroutine.
	chunkSize := u.height / numCPUs
	for i := 0; i < numCPUs; i++ {
		startY := i * chunkSize
		endY := startY + chunkSize
		if i == numCPUs-1 {
			endY = u.height // Ensure last goroutine handles remainder
		}

		go func(start, end int) {
			defer wg.Done()
			for y := start; y < end; y++ {
				for x := 0; x < u.width; x++ {
					neighbors := u.Neighbors(x, y)
					alive := u.grid[y][x]
					if alive {
						// Survive if 2 or 3 neighbors, otherwise die
						nextU.grid[y][x] = neighbors == 2 || neighbors == 3
					} else {
						// Repopulate if exactly 3 neighbors
						nextU.grid[y][x] = neighbors == 3
					}
				}
			}
		}(startY, endY)
	}

	wg.Wait()
	return nextU
}

// Randomize populates the grid randomly with a given density.
func (u *Universe) Randomize() {
	// Seed math/rand with current time for varied random outputs
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for y := 0; y < u.height; y++ {
		for x := 0; x < u.width; x++ {
			u.grid[y][x] = r.Intn(2) == 1
		}
	}
}

// InsertPattern centers and inserts a pattern into the universe.
// It returns an error if the pattern does not fit.
func (u *Universe) InsertPattern(pattern [][]bool) error {
	pHeight := len(pattern)
	if pHeight == 0 {
		return nil
	}
	pWidth := len(pattern[0])

	if pWidth > u.width || pHeight > u.height {
		return fmt.Errorf("grid dimensions (%dx%d) are too small for the chosen pattern (%dx%d)", u.width, u.height, pWidth, pHeight)
	}

	// Calculate center offset
	offsetX := (u.width - pWidth) / 2
	offsetY := (u.height - pHeight) / 2

	for y := 0; y < pHeight; y++ {
		for x := 0; x < pWidth; x++ {
			u.grid[offsetY+y][offsetX+x] = pattern[y][x]
		}
	}
	return nil
}

// Draw prints the current Universe state to stdout.
func (u *Universe) Draw() {
	// Pre-allocate buffer to render frame efficiently in one go
	// Each row has width * 3 bytes (for "█" or " ") + 1 byte for newline
	buf := make([]byte, 0, u.height*(u.width*3+1))
	for y := 0; y < u.height; y++ {
		for x := 0; x < u.width; x++ {
			if u.grid[y][x] {
				buf = append(buf, "█"...)
			} else {
				buf = append(buf, ' ')
			}
		}
		buf = append(buf, '\n')
	}
	os.Stdout.Write(buf)
}

// Classic Patterns
var (
	gliderPattern = [][]bool{
		{false, true, false},
		{false, false, true},
		{true, true, true},
	}

	// Pulsar is a period-3 oscillator.
	// Centered in a 15x15 or 17x17 grid (symmetry from the Wikipedia/Matplotlib layout)
	// We use the 15x15 pattern layout:
	pulsarPattern = [][]bool{
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, true, true, true, false, false, false, true, true, true, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, true, false, false, false, false, true, false, true, false, false, false, false, true, false},
		{false, true, false, false, false, false, true, false, true, false, false, false, false, true, false},
		{false, true, false, false, false, false, true, false, true, false, false, false, false, true, false},
		{false, false, false, true, true, true, false, false, false, true, true, true, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, true, true, true, false, false, false, true, true, true, false, false, false},
		{false, true, false, false, false, false, true, false, true, false, false, false, false, true, false},
		{false, true, false, false, false, false, true, false, true, false, false, false, false, true, false},
		{false, true, false, false, false, false, true, false, true, false, false, false, false, true, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, true, true, true, false, false, false, true, true, true, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
	}
)

func main() {
	width := flag.Int("width", 80, "Grid width")
	height := flag.Int("height", 24, "Grid height")
	pattern := flag.String("pattern", "random", "Initial pattern ('random', 'glider', 'pulsar')")
	delay := flag.Duration("delay", 100*time.Millisecond, "Time delay between generations")

	flag.Parse()

	if *width <= 0 || *height <= 0 {
		fmt.Fprintln(os.Stderr, "Error: width and height must be greater than 0.")
		os.Exit(1)
	}

	u := NewUniverse(*width, *height)

	switch *pattern {
	case "random":
		u.Randomize()
	case "glider":
		if err := u.InsertPattern(gliderPattern); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "pulsar":
		if err := u.InsertPattern(pulsarPattern); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown pattern %q. Choose 'random', 'glider' or 'pulsar'.\n", *pattern)
		os.Exit(1)
	}

	// Channel to catch termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Hide cursor on start
	fmt.Print("\033[?25l")

	// Helper to restore cursor on graceful shutdown
	cleanup := func() {
		fmt.Print("\033[?25h") // Restore cursor
		fmt.Println()          // Newline to avoid overlapping terminal prompts
	}

	// Run simulation in a separate goroutine or handle it in the main thread with select
	ticker := time.NewTicker(*delay)
	defer ticker.Stop()

	// Clear terminal initially
	fmt.Print("\033[H\033[2J")

	for {
		select {
		case <-sigChan:
			cleanup()
			return
		case <-ticker.C:
			// Move cursor back to top-left (0,0) to redraw without flicker
			fmt.Print("\033[H")
			u.Draw()
			u = u.Next()
		}
	}
}
