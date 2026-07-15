// Package main implementa uma versão interativa, robusta e concorrente do
// "Jogo da Vida de Conway" para execução direta no terminal.
//
// A arquitetura deste projeto segue estritamente os princípios do SOLID
// e de Orientação a Objetos adaptados de forma idiomática para Go,
// utilizando interfaces para desacoplamento e divisão clara de responsabilidades.
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

// ============================================================================
// 1. ABSTRAÇÕES (Interfaces - Princípio da Inversão de Dependência & Segregação de Interfaces)
// ============================================================================

// Grid define a interface para manipulação de uma malha bidimensional de células.
// Permite que o motor de simulação ou renderizador consulte e altere estados sem conhecer os detalhes da matriz.
type Grid interface {
	Width() int
	Height() int
	Get(x, y int) bool
	Set(x, y int, state bool)
	Neighbors(x, y int) int
	Clone() Grid
}

// Renderer define o contrato para a exibição ou visualização do estado atual do Grid.
type Renderer interface {
	Render(g Grid) error
	Setup() error
	Cleanup() error
}

// Initializer define o contrato para popular o estado inicial do Grid.
type Initializer interface {
	Initialize(g Grid) error
}

// ============================================================================
// 2. MODELO DE DADOS (Single Responsibility: Armazenamento e Topologia Toroidal)
// ============================================================================

// Universe implementa a interface Grid. Representa o espaço físico toroidal do jogo.
type Universe struct {
	width  int
	height int
	grid   [][]bool
}

// NewUniverse cria uma nova instância de Universe limpa com as dimensões fornecidas.
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

// Width retorna a largura do universo.
func (u *Universe) Width() int {
	return u.width
}

// Height retorna a altura do universo.
func (u *Universe) Height() int {
	return u.height
}

// Get retorna o estado de vida de uma célula nas coordenadas (x, y).
func (u *Universe) Get(x, y int) bool {
	return u.grid[y][x]
}

// Set define o estado de vida de uma célula nas coordenadas (x, y).
func (u *Universe) Set(x, y int, state bool) {
	u.grid[y][x] = state
}

// Neighbors calcula a quantidade de vizinhos vivos de uma célula usando topologia toroidal (embrulho nas bordas).
func (u *Universe) Neighbors(x, y int) int {
	count := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			// Comportamento de embrulho toroidal
			nx := (x + dx + u.width) % u.width
			ny := (y + dy + u.height) % u.height
			if u.grid[ny][nx] {
				count++
			}
		}
	}
	return count
}

// Clone cria uma cópia profunda (deep copy) do estado atual do Universe.
func (u *Universe) Clone() Grid {
	clone := NewUniverse(u.width, u.height)
	for i := 0; i < u.height; i++ {
		copy(clone.grid[i], u.grid[i])
	}
	return clone
}

// ============================================================================
// 3. MOTOR DE EVOLUÇÃO (Single Responsibility: Lógica Concorrente de Transição de Estados)
// ============================================================================

// EvolutionEngine é responsável por gerenciar a evolução geracional de qualquer objeto que satisfaça a interface Grid.
type EvolutionEngine struct {
	numCPUs int
}

// NewEvolutionEngine cria uma nova instância do motor de evolução com paralelização automática.
func NewEvolutionEngine() *EvolutionEngine {
	numCPUs := runtime.NumCPU()
	if numCPUs <= 0 {
		numCPUs = 1
	}
	return &EvolutionEngine{
		numCPUs: numCPUs,
	}
}

// Evolve calcula a próxima geração do Grid de forma paralela e retorna um novo Grid evoluído.
// Aplica rigorosamente as quatro regras clássicas de Conway sem condições de corrida.
func (e *EvolutionEngine) Evolve(g Grid) Grid {
	nextG := g.Clone()
	height := g.Height()
	width := g.Width()

	// Ajusta o número de goroutines se a altura do grid for menor que as CPUs disponíveis.
	workers := e.numCPUs
	if workers > height {
		workers = height
	}

	var wg sync.WaitGroup
	wg.Add(workers)

	chunkSize := height / workers
	for i := 0; i < workers; i++ {
		startY := i * chunkSize
		endY := startY + chunkSize
		if i == workers-1 {
			endY = height // Garante o processamento de linhas restantes
		}

		go func(start, end int) {
			defer wg.Done()
			for y := start; y < end; y++ {
				for x := 0; x < width; x++ {
					neighbors := g.Neighbors(x, y)
					alive := g.Get(x, y)

					if alive {
						// Sobrevivência: 2 ou 3 vizinhos vivos
						nextG.Set(x, y, neighbors == 2 || neighbors == 3)
					} else {
						// Renascimento: exatamente 3 vizinhos vivos
						nextG.Set(x, y, neighbors == 3)
					}
				}
			}
		}(startY, endY)
	}

	wg.Wait()
	return nextG
}

// ============================================================================
// 4. RENDERIZADORES (Single Responsibility: Visualização / Input-Output)
// ============================================================================

// ConsoleRenderer é responsável por desenhar o Grid de forma limpa e otimizada no terminal.
type ConsoleRenderer struct {
	liveChar  string
	deadChar  string
	outBuffer []byte
}

// NewConsoleRenderer cria um renderizador para terminal, utilizando caracteres unicode de alto contraste.
func NewConsoleRenderer() *ConsoleRenderer {
	return &ConsoleRenderer{
		liveChar: "█",
		deadChar: " ",
	}
}

// Setup prepara o console ocultando o cursor e limpando a tela inicial.
func (cr *ConsoleRenderer) Setup() error {
	fmt.Print("\033[?25l")     // Oculta o cursor do terminal
	fmt.Print("\033[H\033[2J") // Limpa a tela inteira inicialmente
	return nil
}

// Cleanup restaura as configurações originais do console.
func (cr *ConsoleRenderer) Cleanup() error {
	fmt.Print("\033[?25h") // Restaura a visibilidade do cursor
	fmt.Println()          // Pula linha para evitar sobreposição no prompt do sistema
	return nil
}

// Render desenha o estado atual do Grid de forma suave e rápida para evitar flicker.
func (cr *ConsoleRenderer) Render(g Grid) error {
	height := g.Height()
	width := g.Width()

	// Pré-aloca o buffer de bytes do frame para otimizar escrita de IO de uma única vez.
	// Cada caractere vivo ocupa 3 bytes em UTF-8. Espaço ocupa 1 byte.
	// Vamos alocar uma capacidade segura baseada na largura e altura.
	expectedSize := height * (width*3 + 1)
	if cap(cr.outBuffer) < expectedSize {
		cr.outBuffer = make([]byte, 0, expectedSize)
	} else {
		cr.outBuffer = cr.outBuffer[:0]
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if g.Get(x, y) {
				cr.outBuffer = append(cr.outBuffer, cr.liveChar...)
			} else {
				cr.outBuffer = append(cr.outBuffer, cr.deadChar...)
			}
		}
		cr.outBuffer = append(cr.outBuffer, '\n')
	}

	// Posiciona o cursor no canto superior esquerdo (0,0) de forma ultra rápida antes de reescrever
	fmt.Print("\033[H")
	_, err := os.Stdout.Write(cr.outBuffer)
	return err
}

// ============================================================================
// 5. INICIALIZADORES (Single Responsibility: Geração Inicial e Injeção de Padrões)
// ============================================================================

// RandomInitializer preenche o universo de forma aleatória com densidade equilibrada.
type RandomInitializer struct {
	seed int64
}

// NewRandomInitializer cria um inicializador aleatório baseado na semente fornecida.
func NewRandomInitializer(seed int64) *RandomInitializer {
	return &RandomInitializer{seed: seed}
}

// Initialize preenche aleatoriamente o Grid fornecido.
func (ri *RandomInitializer) Initialize(g Grid) error {
	r := rand.New(rand.NewSource(ri.seed))
	for y := 0; y < g.Height(); y++ {
		for x := 0; x < g.Width(); x++ {
			g.Set(x, y, r.Intn(2) == 1)
		}
	}
	return nil
}

// PatternInitializer injeta um padrão clássico específico centralizado no Grid.
type PatternInitializer struct {
	name    string
	pattern [][]bool
}

// NewPatternInitializer cria um inicializador de padrões clássicos.
func NewPatternInitializer(name string, pattern [][]bool) *PatternInitializer {
	return &PatternInitializer{
		name:    name,
		pattern: pattern,
	}
}

// Initialize valida e posiciona o padrão no centro do Grid.
func (pi *PatternInitializer) Initialize(g Grid) error {
	pHeight := len(pi.pattern)
	if pHeight == 0 {
		return nil
	}
	pWidth := len(pi.pattern[0])

	if pWidth > g.Width() || pHeight > g.Height() {
		return fmt.Errorf("dimensões do tabuleiro (%dx%d) são insuficientes para o padrão '%s' (%dx%d)",
			g.Width(), g.Height(), pi.name, pWidth, pHeight)
	}

	// Cálculo para centralização perfeita no Grid
	offsetX := (g.Width() - pWidth) / 2
	offsetY := (g.Height() - pHeight) / 2

	for y := 0; y < pHeight; y++ {
		for x := 0; x < pWidth; x++ {
			g.Set(offsetX+x, offsetY+y, pi.pattern[y][x])
		}
	}
	return nil
}

// ============================================================================
// 6. PATTERNS CLÁSSICOS DISPONÍVEIS
// ============================================================================

var (
	gliderPattern = [][]bool{
		{false, true, false},
		{false, false, true},
		{true, true, true},
	}

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

// ============================================================================
// 7. ORQUESTRADOR / PONTO DE ENTRADA
// ============================================================================

func main() {
	width := flag.Int("width", 80, "Largura do tabuleiro de simulação")
	height := flag.Int("height", 24, "Altura do tabuleiro de simulação")
	patternOpt := flag.String("pattern", "random", "Padrão de inicialização ('random', 'glider', 'pulsar')")
	delay := flag.Duration("delay", 100*time.Millisecond, "Intervalo de tempo entre cada geração")

	flag.Parse()

	if *width <= 0 || *height <= 0 {
		fmt.Fprintln(os.Stderr, "Erro: largura e altura do tabuleiro devem ser maiores que zero.")
		os.Exit(1)
	}

	// Criação do Grid
	var u Grid = NewUniverse(*width, *height)

	// Seleção de Initializer de acordo com as diretrizes do Princípio Aberto/Fechado (Ocp)
	var initializer Initializer

	switch *patternOpt {
	case "random":
		initializer = NewRandomInitializer(time.Now().UnixNano())
	case "glider":
		initializer = NewPatternInitializer("Glider", gliderPattern)
	case "pulsar":
		initializer = NewPatternInitializer("Pulsar", pulsarPattern)
	default:
		fmt.Fprintf(os.Stderr, "Erro: Padrão '%s' desconhecido. Escolha entre 'random', 'glider' ou 'pulsar'.\n", *patternOpt)
		os.Exit(1)
	}

	// Inicializa o tabuleiro
	if err := initializer.Initialize(u); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar tabuleiro: %v\n", err)
		os.Exit(1)
	}

	// Instanciação de dependências concretas injetadas via interfaces
	var renderer Renderer = NewConsoleRenderer()
	engine := NewEvolutionEngine()

	// Preparação do console
	if err := renderer.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao configurar console: %v\n", err)
		os.Exit(1)
	}

	// Captura de sinais do sistema operacional para encerramento graceful
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(*delay)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			renderer.Cleanup()
			return
		case <-ticker.C:
			if err := renderer.Render(u); err != nil {
				renderer.Cleanup()
				fmt.Fprintf(os.Stderr, "Erro durante a renderização: %v\n", err)
				os.Exit(1)
			}
			u = engine.Evolve(u)
		}
	}
}
