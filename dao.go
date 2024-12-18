package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const dificuldade = 3

type Bloco struct {
	Indice       int    `json:"index"`
	Timestamp    string `json:"timestamp"`
	Evento       string `json:"evento"`
	Resultado    string `json:"resultado"`
	HashAnterior string `json:"hash_anterior"`
	HashAtual    string `json:"hash_atual"`
	Nonce        int    `json:"nonce"`
	Dificuldade  int    `json:"dificuldade"`
}

type Evento struct {
	ID        int                 `json:"id"`
	Nome      string              `json:"nome"`
	Opcoes    []string            `json:"opcoes"`
	Votos     map[string][]Aposta `json:"votos"`
	Resultado string              `json:"resultado"`
}

type Aposta struct {
	Usuario  string  `json:"usuario"`
	Valor    float64 `json:"valor"`
	EventoID int     `json:"evento_id"`
	Opcao    string  `json:"opcao"`
}

type Voto struct {
	Usuario  string `json:"usuario"`
	EventoID int    `json:"evento_id"`
	Opcao    string `json:"opcao"`
}

type Blockchain struct {
	Blocos []Bloco
	mu     sync.Mutex
	peers  []string
}

func NovoBlockchain(peers []string) *Blockchain {
	genesisTimestamp := time.Now().Format(time.RFC3339)
	genesisBloco := Bloco{
		Indice:       0,
		Timestamp:    genesisTimestamp,
		Evento:       "genesis",
		Resultado:    "",
		HashAnterior: "",
		Nonce:        0,
		Dificuldade:  dificuldade,
	}
	genesisBloco.HashAtual = calculaHash(genesisBloco)
	return &Blockchain{
		Blocos: []Bloco{genesisBloco},
		peers:  peers,
	}
}

func ServirIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		return
	}
	http.ServeFile(w, r, "./index.html")
}

func calculaHash(bloco Bloco) string {
	dados := fmt.Sprintf("%d%s%s%s%d%d", bloco.Indice, bloco.Timestamp, bloco.Evento, bloco.Resultado, bloco.Nonce, bloco.Dificuldade)
	hash := sha256.Sum256([]byte(dados))
	return hex.EncodeToString(hash[:])
}

func provaDeTrabalho(bloco Bloco, dificuldade int) (int, string) {
	var nonce int
	var hash string
	prefixo := strings.Repeat("0", dificuldade)
	for {
		bloco.Nonce = nonce
		hash = calculaHash(bloco)
		if strings.HasPrefix(hash, prefixo) {
			break
		}
		nonce++
	}
	return nonce, hash
}

func (bc *Blockchain) AdicionarBloco(evento string, resultado interface{}) Bloco {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	ultimoBloco := bc.Blocos[len(bc.Blocos)-1]
	resultadoBytes, err := json.Marshal(resultado)
	if err != nil {
		return Bloco{}
	}
	novoBloco := Bloco{
		Indice:       len(bc.Blocos),
		Timestamp:    time.Now().Format(time.RFC3339),
		Evento:       evento,
		Resultado:    string(resultadoBytes),
		HashAnterior: ultimoBloco.HashAtual,
		Dificuldade:  dificuldade,
	}
	nonce, hash := provaDeTrabalho(novoBloco, dificuldade)
	novoBloco.Nonce = nonce
	novoBloco.HashAtual = hash
	bc.Blocos = append(bc.Blocos, novoBloco)
	go bc.NotificarPeers(novoBloco)
	return novoBloco
}

func (bc *Blockchain) ValidarBlockchain() bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	prefixo := strings.Repeat("0", dificuldade)
	for i := 1; i < len(bc.Blocos); i++ {
		blocoAtual := bc.Blocos[i]
		blocoAnterior := bc.Blocos[i-1]
		if blocoAtual.HashAnterior != blocoAnterior.HashAtual {
			return false
		}
		recalculadoHash := calculaHash(blocoAtual)
		if blocoAtual.HashAtual != recalculadoHash {
			return false
		}
		if !strings.HasPrefix(blocoAtual.HashAtual, prefixo) {
			return false
		}
	}
	return true
}

func (bc *Blockchain) ValidarNovaBlockchain(novaBlockchain []Bloco) bool {
	prefixo := strings.Repeat("0", dificuldade)
	for i := 1; i < len(novaBlockchain); i++ {
		blocoAtual := novaBlockchain[i]
		blocoAnterior := novaBlockchain[i-1]
		if blocoAtual.HashAnterior != blocoAnterior.HashAtual {
			return false
		}
		recalculadoHash := calculaHash(blocoAtual)
		if blocoAtual.HashAtual != recalculadoHash {
			return false
		}
		if !strings.HasPrefix(blocoAtual.HashAtual, prefixo) {
			return false
		}
	}
	return true
}

func (bc *Blockchain) ValidarBloco(bloco Bloco) bool {
	prefixo := strings.Repeat("0", dificuldade)
	recalculadoHash := calculaHash(bloco)
	return bloco.HashAtual == recalculadoHash && strings.HasPrefix(bloco.HashAtual, prefixo)
}

func (bc *Blockchain) ExibirBlockchainHTTP(w http.ResponseWriter, r *http.Request) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		return
	}
	json.NewEncoder(w).Encode(bc.Blocos)
}

func (bc *Blockchain) ImprimirBlockchain() {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	for _, bloco := range bc.Blocos {
		fmt.Printf("Index: %d, Evento: %s, Resultado: %s, Hash: %s, Nonce: %d\n", bloco.Indice, bloco.Evento, bloco.Resultado, bloco.HashAtual, bloco.Nonce)
	}
}

func (bc *Blockchain) HandleAdicionarBloco(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		return
	}
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

func (bc *Blockchain) HandleValidarBlockchain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		return
	}
	valido := bc.ValidarBlockchain()
	status := "inválida"
	if valido {
		status = "válida"
	}
	w.Write([]byte(fmt.Sprintf("Blockchain é %s.", status)))
}

func (bc *Blockchain) ReceberBlockchain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		return
	}
	var novaBlockchain []Bloco
	if err := json.NewDecoder(r.Body).Decode(&novaBlockchain); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if len(novaBlockchain) > len(bc.Blocos) && bc.ValidarNovaBlockchain(novaBlockchain) {
		bc.Blocos = novaBlockchain
		fmt.Fprintln(w, "Blockchain atualizada com sucesso")
	} else {
		fmt.Fprintln(w, "Blockchain recebida é inválida ou não é mais longa")
	}
}

func (bc *Blockchain) ReceberBloco(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		return
	}
	var novoBloco Bloco
	if err := json.NewDecoder(r.Body).Decode(&novoBloco); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}
	bc.mu.Lock()
	defer bc.mu.Unlock()
	ultimoBloco := bc.Blocos[len(bc.Blocos)-1]
	if novoBloco.Indice == ultimoBloco.Indice+1 && novoBloco.HashAnterior == ultimoBloco.HashAtual && bc.ValidarBloco(novoBloco) {
		bc.Blocos = append(bc.Blocos, novoBloco)
		fmt.Fprintln(w, "Bloco adicionado com sucesso")
	} else {
		fmt.Fprintln(w, "Bloco recebido é inválido")
	}
}

func (bc *Blockchain) NotificarPeers(bloco Bloco) {
	for _, peer := range bc.peers {
		go func(peer string) {
			url := fmt.Sprintf("%s/receber-bloco", peer)
			jsonData, err := json.Marshal(bloco)
			if err != nil {
				return
			}
			http.Post(url, "application/json", strings.NewReader(string(jsonData)))
		}(peer)
	}
}

func (bc *Blockchain) SincronizarComPeers() {
	for _, peer := range bc.peers {
		go func(peer string) {
			resp, err := http.Get(fmt.Sprintf("%s/blockchain", peer))
			if err != nil {
				return
			}
			var blockchainPeer []Bloco
			if err := json.NewDecoder(resp.Body).Decode(&blockchainPeer); err != nil {
				resp.Body.Close()
				return
			}
			resp.Body.Close()
			bc.mu.Lock()
			defer bc.mu.Unlock()
			if len(blockchainPeer) > len(bc.Blocos) && bc.ValidarNovaBlockchain(blockchainPeer) {
				bc.Blocos = blockchainPeer
			}
		}(peer)
	}
}

func (bc *Blockchain) ProximoIDEvento() int {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	maxID := 0
	for _, bloco := range bc.Blocos {
		if bloco.Evento == "criar_evento" {
			var evento Evento
			err := json.Unmarshal([]byte(bloco.Resultado), &evento)
			if err != nil {
				continue
			}
			if evento.ID > maxID {
				maxID = evento.ID
			}
		}
	}
	return maxID + 1
}

func (bc *Blockchain) HandleSaldo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		return
	}
	usuario := r.URL.Query().Get("usuario")
	if usuario == "" {
		http.Error(w, "Parâmetro 'usuario' é obrigatório", http.StatusBadRequest)
		return
	}
	saldo := bc.CalcularSaldo(usuario)
	type SaldoResponse struct {
		Usuario string  `json:"usuario"`
		Saldo   float64 `json:"saldo"`
	}
	response := SaldoResponse{
		Usuario: usuario,
		Saldo:   saldo,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (bc *Blockchain) CalcularSaldo(usuario string) float64 {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	saldo := 0.0
	for _, bloco := range bc.Blocos {
		if bloco.Evento == "ajustar_saldo" {
			var ajuste map[string]interface{}
			err := json.Unmarshal([]byte(bloco.Resultado), &ajuste)
			if err != nil {
				continue
			}
			if ajuste["usuario"] == usuario {
				valor, ok := ajuste["valor"].(float64)
				if ok {
					saldo += valor
				}
			}
		}
	}
	return saldo
}

func (bc *Blockchain) HandleCriarEvento(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	type CriarEventoRequest struct {
		Nome   string   `json:"nome"`
		Opcoes []string `json:"opcoes"`
	}
	var req CriarEventoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}
	if req.Nome == "" || len(req.Opcoes) < 2 {
		http.Error(w, "Nome do evento e pelo menos duas opções são obrigatórios", http.StatusBadRequest)
		return
	}
	evento := Evento{
		ID:     bc.ProximoIDEvento(),
		Nome:   req.Nome,
		Opcoes: req.Opcoes,
		Votos:  make(map[string][]Aposta),
	}
	bc.AdicionarBloco("criar_evento", evento)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evento)
}

func (bc *Blockchain) HandleListarEventos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		return
	}
	bc.mu.Lock()
	defer bc.mu.Unlock()
	eventosMap := make(map[int]*Evento)
	for _, bloco := range bc.Blocos {
		if bloco.Evento == "criar_evento" {
			var evento Evento
			err := json.Unmarshal([]byte(bloco.Resultado), &evento)
			if err != nil {
				continue
			}
			eventosMap[evento.ID] = &evento
		} else if bloco.Evento == "apostar" {
			var aposta Aposta
			err := json.Unmarshal([]byte(bloco.Resultado), &aposta)
			if err != nil {
				continue
			}
			if evento, existe := eventosMap[aposta.EventoID]; existe {
				evento.Votos[aposta.Opcao] = append(evento.Votos[aposta.Opcao], aposta)
			}
		} else if bloco.Evento == "concluir_evento" {
			var resultado map[string]interface{}
			err := json.Unmarshal([]byte(bloco.Resultado), &resultado)
			if err == nil {
				eventoID, _ := resultado["evento_id"].(float64)
				opcaoVencedora, _ := resultado["opcao_vencedora"].(string)
				if ev, existe := eventosMap[int(eventoID)]; existe {
					ev.Resultado = opcaoVencedora
				}
			}
		}
	}
	eventos := []Evento{}
	for _, evento := range eventosMap {
		eventos = append(eventos, *evento)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eventos)
}

func (bc *Blockchain) HandleVotar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	type VotarRequest struct {
		Usuario  string `json:"usuario"`
		EventoID int    `json:"evento_id"`
		Opcao    string `json:"opcao"`
	}
	var req VotarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}
	if req.Usuario == "" || req.EventoID == 0 || req.Opcao == "" {
		http.Error(w, "Todos os campos são obrigatórios", http.StatusBadRequest)
		return
	}
	eventoValido := bc.VerificarOpcaoEvento(req.EventoID, req.Opcao)
	if !eventoValido {
		http.Error(w, "Evento ou opção inválidos", http.StatusBadRequest)
		return
	}
	voto := Voto{
		Usuario:  req.Usuario,
		EventoID: req.EventoID,
		Opcao:    req.Opcao,
	}
	bc.AdicionarBloco("votar", voto)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(voto)
}

func (bc *Blockchain) VerificarOpcaoEvento(eventoID int, opcao string) bool {
	for _, bloco := range bc.Blocos {
		if bloco.Evento == "criar_evento" {
			var evento Evento
			err := json.Unmarshal([]byte(bloco.Resultado), &evento)
			if err != nil {
				continue
			}
			if evento.ID == eventoID {
				for _, op := range evento.Opcoes {
					if op == opcao {
						return true
					}
				}
				return false
			}
		}
	}
	return false
}

func (bc *Blockchain) HandleApostar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	type ApostarRequest struct {
		Usuario  string  `json:"usuario"`
		EventoID int     `json:"evento_id"`
		Opcao    string  `json:"opcao"`
		Valor    float64 `json:"valor"`
	}
	var req ApostarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}
	if req.Usuario == "" || req.EventoID == 0 || req.Opcao == "" || req.Valor <= 0 {
		http.Error(w, "Todos os campos são obrigatórios e o valor deve ser positivo", http.StatusBadRequest)
		return
	}
	saldo := bc.CalcularSaldo(req.Usuario)
	if saldo < req.Valor {
		http.Error(w, "Saldo insuficiente", http.StatusBadRequest)
		return
	}
	eventoValido := bc.VerificarOpcaoEvento(req.EventoID, req.Opcao)
	if !eventoValido {
		http.Error(w, "Evento ou opção inválidos", http.StatusBadRequest)
		return
	}
	aposta := Aposta{
		Usuario:  req.Usuario,
		Valor:    req.Valor,
		EventoID: req.EventoID,
		Opcao:    req.Opcao,
	}
	blocoAposta := bc.AdicionarBloco("apostar", aposta)
	ajuste := map[string]interface{}{
		"usuario": req.Usuario,
		"valor":   -req.Valor,
	}
	bc.AdicionarBloco("ajustar_saldo", ajuste)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blocoAposta)
}

func (bc *Blockchain) HandleSacar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	type SacarRequest struct {
		Usuario string  `json:"usuario"`
		Valor   float64 `json:"valor"`
	}
	var req SacarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}
	if req.Usuario == "" || req.Valor <= 0 {
		http.Error(w, "Todos os campos são obrigatórios e o valor deve ser positivo", http.StatusBadRequest)
		return
	}
	saldo := bc.CalcularSaldo(req.Usuario)
	if saldo < req.Valor {
		http.Error(w, "Saldo insuficiente para saque", http.StatusBadRequest)
		return
	}
	ajuste := map[string]interface{}{
		"usuario": req.Usuario,
		"valor":   -req.Valor,
	}
	blocoAjuste := bc.AdicionarBloco("ajustar_saldo", ajuste)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blocoAjuste)
}

func (bc *Blockchain) HandleDepositar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	type DepositarRequest struct {
		Usuario string  `json:"usuario"`
		Valor   float64 `json:"valor"`
	}
	var req DepositarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}
	if req.Usuario == "" || req.Valor <= 0 {
		http.Error(w, "Todos os campos são obrigatórios e o valor deve ser positivo", http.StatusBadRequest)
		return
	}
	ajuste := map[string]interface{}{
		"usuario": req.Usuario,
		"valor":   req.Valor,
	}
	blocoAjuste := bc.AdicionarBloco("ajustar_saldo", ajuste)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blocoAjuste)
}

func (bc *Blockchain) HandleConcluirEvento(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	type ConcluirEventoRequest struct {
		EventoID       int    `json:"evento_id"`
		OpcaoVencedora string `json:"opcao_vencedora"`
	}

	var req ConcluirEventoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Erro ao decodificar request: %v", err)
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if req.EventoID == 0 || req.OpcaoVencedora == "" {
		log.Printf("Dados insuficientes: EventoID=%d OpcaoVencedora=%s", req.EventoID, req.OpcaoVencedora)
		http.Error(w, "Todos os campos são obrigatórios", http.StatusBadRequest)
		return
	}

	bc.mu.Lock()
	log.Printf("Mutex bloqueado para leitura dos eventos")
	eventosMap := make(map[int]*Evento)
	for _, bloco := range bc.Blocos {
		if bloco.Evento == "criar_evento" {
			var evento Evento
			err := json.Unmarshal([]byte(bloco.Resultado), &evento)
			if err != nil {
				log.Printf("Erro ao deserializar criar_evento: %v", err)
				continue
			}
			eventosMap[evento.ID] = &evento
		} else if bloco.Evento == "apostar" {
			var aposta Aposta
			err := json.Unmarshal([]byte(bloco.Resultado), &aposta)
			if err != nil {
				log.Printf("Erro ao deserializar apostar: %v", err)
				continue
			}
			if evento, existe := eventosMap[aposta.EventoID]; existe {
				evento.Votos[aposta.Opcao] = append(evento.Votos[aposta.Opcao], aposta)
			}
		}
	}

	evento, existe := eventosMap[req.EventoID]
	if !existe {
		bc.mu.Unlock()
		log.Printf("Evento %d não encontrado", req.EventoID)
		http.Error(w, "Evento não encontrado", http.StatusBadRequest)
		return
	}

	opcaoValida := false
	for _, op := range evento.Opcoes {
		if op == req.OpcaoVencedora {
			opcaoValida = true
			break
		}
	}

	if !opcaoValida {
		bc.mu.Unlock()
		log.Printf("Opção vencedora %s não é válida para evento %d", req.OpcaoVencedora, req.EventoID)
		http.Error(w, "Opção vencedora inválida", http.StatusBadRequest)
		return
	}

	totalVencedor := 0.0
	totalPerdedor := 0.0
	for opcao, apostas := range evento.Votos {
		for _, aposta := range apostas {
			if opcao == req.OpcaoVencedora {
				totalVencedor += aposta.Valor
			} else {
				totalPerdedor += aposta.Valor
			}
		}
	}

	log.Printf("Evento %d: totalVencedor=%.2f totalPerdedor=%.2f", req.EventoID, totalVencedor, totalPerdedor)

	bc.mu.Unlock()
	log.Printf("Mutex liberado antes de ajustar saldos")

	var premioPorAposta float64
	if totalVencedor > 0 {
		premioPorAposta = totalPerdedor / totalVencedor
		log.Printf("Evento %d: premioPorAposta=%.2f", req.EventoID, premioPorAposta)
		for opcao, apostas := range evento.Votos {
			if opcao != req.OpcaoVencedora {
				continue
			}
			for _, aposta := range apostas {
				ajuste := map[string]interface{}{
					"usuario": aposta.Usuario,
					"valor":   aposta.Valor * premioPorAposta,
				}
				log.Printf("Ajustando saldo para %s valor=%.2f", aposta.Usuario, aposta.Valor*premioPorAposta)
				bc.AdicionarBloco("ajustar_saldo", ajuste)
			}
		}
	} else {
		log.Printf("Nenhum vencedor no evento %d", req.EventoID)
	}

	resultado := map[string]interface{}{
		"evento_id":       req.EventoID,
		"opcao_vencedora": req.OpcaoVencedora,
	}
	log.Printf("Adicionando bloco concluir_evento para evento %d", req.EventoID)
	bc.AdicionarBloco("concluir_evento", resultado)

	w.Write([]byte("Evento concluído e prêmios distribuídos com sucesso."))
	log.Printf("Conclusão do evento %d finalizada", req.EventoID)
}

func (bc *Blockchain) InicializarEndpoints() {
	http.HandleFunc("/", ServirIndex)
	http.HandleFunc("/blockchain", bc.ExibirBlockchainHTTP)
	http.HandleFunc("/adicionar", bc.HandleAdicionarBloco)
	http.HandleFunc("/validar", bc.HandleValidarBlockchain)
	http.HandleFunc("/receber-blockchain", bc.ReceberBlockchain)
	http.HandleFunc("/receber-bloco", bc.ReceberBloco)
	http.HandleFunc("/saldo", bc.HandleSaldo)
	http.HandleFunc("/criar-evento", bc.HandleCriarEvento)
	http.HandleFunc("/eventos", bc.HandleListarEventos)
	http.HandleFunc("/votar", bc.HandleVotar)
	http.HandleFunc("/apostar", bc.HandleApostar)
	http.HandleFunc("/concluir-evento", bc.HandleConcluirEvento)
	http.HandleFunc("/depositar", bc.HandleDepositar)
	http.HandleFunc("/sacar", bc.HandleSacar)
}
