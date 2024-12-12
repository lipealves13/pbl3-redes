package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Estrutura do Bloco
type Bloco struct {
	Index        int    `json:"index"`
	Timestamp    string `json:"timestamp"`
	Evento       string `json:"evento"`
	Resultado    string `json:"resultado"`
	HashAnterior string `json:"hash_anterior"`
	HashAtual    string `json:"hash_atual"`
}

// Estrutura da Aposta
type Aposta struct {
	Usuario  string  `json:"usuario"`
	Valor    float64 `json:"valor"`
	Escolha  string  `json:"escolha"`
	EventoID int     `json:"evento_id"`
	Aprovada bool    `json:"aprovada"`
}

var (
	apostas      []Aposta
	apostasMutex sync.Mutex
)

// Adiciona uma aposta de forma thread-safe
func AdicionarAposta(usuario string, valor float64, escolha string, eventoID int) {
	apostasMutex.Lock()
	defer apostasMutex.Unlock()

	aposta := Aposta{
		Usuario:  usuario,
		Valor:    valor,
		Escolha:  escolha,
		EventoID: eventoID,
		Aprovada: false,
	}
	apostas = append(apostas, aposta)
}

// Blockchain
type Blockchain struct {
	Blocos []Bloco
	mu     sync.Mutex
}

// Cria o Genesis Block
func NovoBlockchain() *Blockchain {
	genesisTimestamp := time.Now().Format(time.RFC3339)
	genesis := Bloco{
		Index:        0,
		Timestamp:    genesisTimestamp,
		Evento:       "Genesis Block",
		Resultado:    "",
		HashAnterior: "",
		HashAtual:    calculaHash(0, genesisTimestamp, "Genesis Block", "", ""),
	}
	return &Blockchain{Blocos: []Bloco{genesis}}
}

// Calcula o hash do bloco
func calculaHash(index int, timestamp, evento, resultado, hashAnterior string) string {
	dados := fmt.Sprintf("%d%s%s%s%s", index, timestamp, evento, resultado, hashAnterior)
	hash := sha256.Sum256([]byte(dados))
	return hex.EncodeToString(hash[:])
}

// Adiciona um novo bloco à blockchain de forma thread-safe
func (bc *Blockchain) AdicionarBloco(evento, resultado string) Bloco {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	ultimoBloco := bc.Blocos[len(bc.Blocos)-1]
	novoTimestamp := time.Now().Format(time.RFC3339)
	novoBloco := Bloco{
		Index:        len(bc.Blocos),
		Timestamp:    novoTimestamp,
		Evento:       evento,
		Resultado:    resultado,
		HashAnterior: ultimoBloco.HashAtual,
		HashAtual:    calculaHash(len(bc.Blocos), novoTimestamp, evento, resultado, ultimoBloco.HashAtual),
	}

	bc.Blocos = append(bc.Blocos, novoBloco)
	return novoBloco
}

// Valida a integridade da blockchain
func (bc *Blockchain) ValidarBlockchain() bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	for i := 1; i < len(bc.Blocos); i++ {
		blocoAtual := bc.Blocos[i]
		blocoAnterior := bc.Blocos[i-1]

		if blocoAtual.HashAnterior != blocoAnterior.HashAtual {
			return false
		}

		calculadoHash := calculaHash(blocoAtual.Index, blocoAtual.Timestamp, blocoAtual.Evento, blocoAtual.Resultado, blocoAtual.HashAnterior)
		if blocoAtual.HashAtual != calculadoHash {
			return false
		}
	}
	return true
}

// Função para exibir a blockchain via HTTP
func (bc *Blockchain) ExibirBlockchainHTTP(w http.ResponseWriter, r *http.Request) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bc.Blocos)
}

// Função para imprimir a blockchain no console
func (bc *Blockchain) ImprimirBlockchain() {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	for _, bloco := range bc.Blocos {
		fmt.Printf("Index: %d, Evento: %s, Resultado: %s, Hash: %s\n", bloco.Index, bloco.Evento, bloco.Resultado, bloco.HashAtual)
	}
}

// Endpoint para adicionar um novo bloco
func (bc *Blockchain) HandleAdicionarBloco(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Evento    string `json:"evento"`
		Resultado string `json:"resultado"`
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	novoBloco := bc.AdicionarBloco(req.Evento, req.Resultado)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(novoBloco)
}

// Endpoint para validar a blockchain
func (bc *Blockchain) HandleValidarBlockchain(w http.ResponseWriter, r *http.Request) {
	valido := bc.ValidarBlockchain()
	status := "inválida"
	if valido {
		status = "válida"
	}
	w.Write([]byte(fmt.Sprintf("Blockchain é %s.", status)))
}

// Inicializa os endpoints HTTP
func (bc *Blockchain) InicializarEndpoints() {
	http.HandleFunc("/blockchain", bc.ExibirBlockchainHTTP)
	http.HandleFunc("/adicionar", bc.HandleAdicionarBloco)
	http.HandleFunc("/validar", bc.HandleValidarBlockchain)
}

// Função principal
func main() {
	blockchain := NovoBlockchain()

	// Inicializa os endpoints HTTP
	blockchain.InicializarEndpoints()

	// Inicia o servidor HTTP
	go func() {
		log.Println("Servidor HTTP iniciado na porta 8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Erro ao iniciar o servidor: %v", err)
		}
	}()

	// Simulação de apostas e adição de blocos
	AdicionarAposta("usuario1", 100.0, "Cara", 1)
	AdicionarAposta("usuario2", 150.0, "Coroa", 1)

	// Adiciona alguns blocos de teste
	blockchain.AdicionarBloco("Cara ou Coroa", "Cara")
	blockchain.AdicionarBloco("Cara ou Coroa", "Coroa")

	// Exibe a blockchain no console
	blockchain.ImprimirBlockchain()

	// Mantém o programa rodando
	select {}
}
