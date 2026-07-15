# Jogo da Vida de Conway (Conway's Game of Life) - Terminal Interativo

Uma implementação robusta, interativa, de alta performance e concorrente do clássico **Jogo da Vida de Conway**, desenvolvida diretamente para rodar no terminal usando **Go (Golang)**.

O projeto foi arquitetado sob as melhores práticas de Engenharia de Software, aplicando rigorosamente os princípios de **Orientação a Objetos** e **SOLID**, sem qualquer dependência externa (utilizando apenas a biblioteca padrão do Go).

---

## 🛠️ Arquitetura e Princípios SOLID

A estrutura do projeto foi dividida em componentes modulares e desacoplados, onde cada parte possui uma única responsabilidade clara:

1. **S - Single Responsibility Principle (Princípio da Responsabilidade Única):**
   - `Universe` (Model): Responsável única e exclusivamente pelo armazenamento do estado das células da malha e pela resolução de sua topologia toroidal.
   - `EvolutionEngine` (Engine): Responsável apenas por processar a física da transição de gerações de acordo com as regras clássicas de Conway.
   - `ConsoleRenderer` (View): Responsável unicamente pela preparação, formatação e desenho otimizado do grid no console, gerenciando os códigos ANSI para evitar flicker e tratar o cursor.
   - `Initializers` (Seeders): Responsáveis apenas por definir a disposição inicial de células (ex: `RandomInitializer` para aleatoriedade ou `PatternInitializer` para padrões específicos estruturados).

2. **O - Open/Closed Principle (Princípio Aberto/Fechado) & D - Dependency Inversion Principle (Princípio da Inversão de Dependência):**
   - O motor de simulação e o renderizador comunicam-se através de abstrações (as interfaces `Grid`, `Renderer` e `Initializer`).
   - Se você desejar adicionar um novo padrão inicial clássico, um renderizador gráfico (ex: Web ou Pixel 2D), ou um novo conjunto de regras de física de autômatos celulares, basta implementar a interface correspondente sem alterar ou quebrar os motores internos existentes.

3. **I - Interface Segregation Principle (Princípio da Segregação de Interfaces):**
   - As interfaces definidas (`Grid`, `Renderer`, `Initializer`) são enxutas e focadas estritamente em suas funções específicas, garantindo que nenhum tipo dependa de métodos que não utiliza.

---

## 🚀 Como Executar o Projeto

### Pré-requisitos
* Ter o **Go** instalado em sua máquina (versão 1.18 ou superior recomendada).

### Passo 1: Clonar o repositório ou baixar os arquivos
Certifique-se de que os arquivos `main.go`, `main_test.go` e `go.mod` estão na mesma pasta.

### Passo 2: Executar diretamente
Você pode rodar a aplicação imediatamente usando o comando `go run`:

```bash
# Execução padrão (tabuleiro 80x24, preenchimento aleatório, atualização a cada 100ms)
go run main.go
```

### Passo 3: Configurar via Flags (Opções de Linha de Comando)
O programa suporta flags de CLI extremamente elegantes e idiomáticas para você controlar a simulação:

* `-width`: Largura do tabuleiro (padrão: `80`)
* `-height`: Altura do tabuleiro (padrão: `24`)
* `-pattern`: Padrão inicial do jogo (`random`, `glider` ou `pulsar`. Padrão: `random`)
* `-delay`: Intervalo de tempo entre as gerações (padrão: `100ms`, aceita formatos como `150ms`, `500ms`, `1s`, etc.)

#### Exemplos de Execução com Flags:

```bash
# Iniciar com o padrão clássico "Pulsar" (necessita de grid de no mínimo 15x15)
go run main.go -pattern pulsar -width 80 -height 30 -delay 150ms

# Iniciar com o padrão "Glider" viajando infinitamente em alta velocidade
go run main.go -pattern glider -width 40 -height 20 -delay 50ms

# Iniciar uma simulação gigante aleatória e lenta
go run main.go -pattern random -width 120 -height 40 -delay 300ms
```

---

## 🧪 Como Executar os Testes Unitários

A qualidade e a integridade da lógica de negócios (incluindo o comportamento toroidal de bordas embrulhadas e a evolução concorrente) são cobertas por um conjunto abrangente de testes automatizados.

Para executar os testes, utilize o comando nativo do Go:

```bash
go test -v .
```

---

## ⚡ Concorrência e Performance

A evolução de gerações (`Evolve`) divide de forma inteligente e dinâmica o tabuleiro em fatias de linhas verticais correspondentes à quantidade de CPUs físicas disponíveis na máquina do usuário (`runtime.NumCPU()`).

Cada fatia é computada em paralelo por uma **goroutine** dedicada e sincronizada através de um `sync.WaitGroup`. O cálculo de transição é feito de forma isolada em um grid temporário para garantir que não existam condições de corrida (*race conditions*).

---

## ⏹️ Encerramento Limpo (Graceful Shutdown)

Durante a execução da simulação, a visibilidade do cursor do console é temporariamente desabilitada para proporcionar uma experiência visual limpa e fluida (sem tremulações ou flicker).

Ao pressionar `Ctrl+C` (`SIGINT`), o sistema intercepta o sinal, chama de forma limpa as rotinas de `Cleanup()` do renderizador, restaura o cursor do terminal para o estado original e encerra com segurança o processo.
