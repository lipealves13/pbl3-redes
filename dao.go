// dao.go

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

// Definindo a dificuldade da Prova de Trabalho
const dificuldade = 3 // Número de zeros no início do hash

// Estrutura do Bloco
type Bloco struct {
    Index        int    `json:"index"`
    Timestamp    string `json:"timestamp"`
    Evento       string `json:"evento"`        // Tipo de Transação: "criar_evento", "apostar", "concluir_evento", "ajustar_saldo", etc.
    Resultado    string `json:"resultado"`     // Dados da Transação em JSON
    HashAnterior string `json:"hash_anterior"`
    HashAtual    string `json:"hash_atual"`
    Nonce        int    `json:"nonce"`
    Dificuldade  int    `json:"dificuldade"`
}

// Estrutura do Evento
type Evento struct {
    ID        int                 `json:"id"`
    Nome      string              `json:"nome"`
    Opcoes    []string            `json:"opcoes"`
    Votos     map[string][]Aposta `json:"votos"`    // Opção -> Lista de Apostas
    Resultado string              `json:"resultado"` // "opcao_vencedora"
}

// Estrutura da Aposta
type Aposta struct {
    Usuario  string  `json:"usuario"`
    Valor    float64 `json:"valor"`
    EventoID int     `json:"evento_id"` // Associado a qual evento
    Opcao    string  `json:"opcao"`     // Em qual opção está apostando
}

// Estrutura da Votação
type Voto struct {
    Usuario  string `json:"usuario"`
    EventoID int    `json:"evento_id"`
    Opcao    string `json:"opcao"`
}

// Estrutura do Blockchain
type Blockchain struct {
    Blocos []Bloco
    mu     sync.Mutex
    peers  []string // Lista de URLs dos peers
}

// Cria o Genesis Block
func NovoBlockchain(peers []string) *Blockchain {
    genesisTimestamp := time.Now().Format(time.RFC3339)
    genesisBloco := Bloco{
        Index:        0,
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

// Calcula o hash do bloco
func calculaHash(bloco Bloco) string {
    dados := fmt.Sprintf("%d%s%s%s%d%d", bloco.Index, bloco.Timestamp, bloco.Evento, bloco.Resultado, bloco.Nonce, bloco.Dificuldade)
    hash := sha256.Sum256([]byte(dados))
    return hex.EncodeToString(hash[:])
}

// Prova de Trabalho
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

// Adiciona um novo bloco à blockchain de forma thread-safe com PoW
func (bc *Blockchain) AdicionarBloco(evento string, resultado interface{}) Bloco {
    bc.mu.Lock()
    defer bc.mu.Unlock()

    ultimoBloco := bc.Blocos[len(bc.Blocos)-1]
    resultadoBytes, err := json.Marshal(resultado)
    if err != nil {
        log.Printf("Erro ao marshalizar resultado: %v", err)
        return Bloco{}
    }

    novoBloco := Bloco{
        Index:        len(bc.Blocos),
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

    // Notifica os peers sobre o novo bloco
    go bc.NotificarPeers(novoBloco)

    return novoBloco
}

// Valida a integridade da blockchain
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

// Valida uma nova blockchain recebida
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

// Valida um único bloco
func (bc *Blockchain) ValidarBloco(bloco Bloco) bool {
    prefixo := strings.Repeat("0", dificuldade)
    recalculadoHash := calculaHash(bloco)
    return bloco.HashAtual == recalculadoHash && strings.HasPrefix(bloco.HashAtual, prefixo)
}

// Função para exibir a blockchain via HTTP
func (bc *Blockchain) ExibirBlockchainHTTP(w http.ResponseWriter, r *http.Request) {
    bc.mu.Lock()
    defer bc.mu.Unlock()

    w.Header().Set("Content-Type", "application/json")
    // Adiciona cabeçalhos CORS
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    if r.Method == "OPTIONS" {
        return
    }

    json.NewEncoder(w).Encode(bc.Blocos)
}

// Função para imprimir a blockchain no console
func (bc *Blockchain) ImprimirBlockchain() {
    bc.mu.Lock()
    defer bc.mu.Unlock()

    for _, bloco := range bc.Blocos {
        fmt.Printf("Index: %d, Evento: %s, Resultado: %s, Hash: %s, Nonce: %d\n", bloco.Index, bloco.Evento, bloco.Resultado, bloco.HashAtual, bloco.Nonce)
    }
}

// Endpoint para adicionar um novo bloco via API
func (bc *Blockchain) HandleAdicionarBloco(w http.ResponseWriter, r *http.Request) {
    // Adiciona cabeçalhos CORS
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

// Endpoint para validar a blockchain
func (bc *Blockchain) HandleValidarBlockchain(w http.ResponseWriter, r *http.Request) {
    // Adiciona cabeçalhos CORS
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

// Endpoint para receber a blockchain de um peer
func (bc *Blockchain) ReceberBlockchain(w http.ResponseWriter, r *http.Request) {
    // Adiciona cabeçalhos CORS
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

// Endpoint para receber um novo bloco de um peer
func (bc *Blockchain) ReceberBloco(w http.ResponseWriter, r *http.Request) {
    // Adiciona cabeçalhos CORS
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
    if novoBloco.Index == ultimoBloco.Index+1 && novoBloco.HashAnterior == ultimoBloco.HashAtual && bc.ValidarBloco(novoBloco) {
        bc.Blocos = append(bc.Blocos, novoBloco)
        fmt.Fprintln(w, "Bloco adicionado com sucesso")
    } else {
        fmt.Fprintln(w, "Bloco recebido é inválido")
    }
}

// Notifica os peers sobre um novo bloco
func (bc *Blockchain) NotificarPeers(bloco Bloco) {
    for _, peer := range bc.peers {
        go func(peer string) {
            url := fmt.Sprintf("%s/receber-bloco", peer)
            jsonData, err := json.Marshal(bloco)
            if err != nil {
                log.Printf("Erro ao marshalizar bloco para %s: %v", peer, err)
                return
            }

            resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
            if err != nil {
                log.Printf("Erro ao enviar bloco para %s: %v", peer, err)
                return
            }
            defer resp.Body.Close()

            if resp.StatusCode != http.StatusOK {
                log.Printf("Erro ao enviar bloco para %s: Status %s", peer, resp.Status)
            } else {
                log.Printf("Bloco enviado com sucesso para %s", peer)
            }
        }(peer)
    }
}

// Sincroniza a blockchain com os peers
func (bc *Blockchain) SincronizarComPeers() {
    for _, peer := range bc.peers {
        go func(peer string) {
            resp, err := http.Get(fmt.Sprintf("%s/blockchain", peer))
            if err != nil {
                log.Printf("Erro ao obter blockchain de %s: %v", peer, err)
                return
            }

            var blockchainPeer []Bloco
            if err := json.NewDecoder(resp.Body).Decode(&blockchainPeer); err != nil {
                log.Printf("Erro ao decodificar blockchain de %s: %v", peer, err)
                resp.Body.Close()
                return
            }
            resp.Body.Close()

            bc.mu.Lock()
            defer bc.mu.Unlock()

            if len(blockchainPeer) > len(bc.Blocos) && bc.ValidarNovaBlockchain(blockchainPeer) {
                bc.Blocos = blockchainPeer
                log.Printf("Blockchain atualizada a partir de %s", peer)
            }
        }(peer)
    }
}

// Função para obter o próximo ID de evento
func (bc *Blockchain) ProximoIDEvento() int {
    bc.mu.Lock()
    defer bc.mu.Unlock()

    maxID := 0
    for _, bloco := range bc.Blocos {
        if bloco.Evento == "criar_evento" {
            var evento Evento
            err := json.Unmarshal([]byte(bloco.Resultado), &evento)
            if err != nil {
                log.Printf("Erro ao unmarshalar criar_evento: %v", err)
                continue
            }
            if evento.ID > maxID {
                maxID = evento.ID
            }
        }
    }
    return maxID + 1
}

// Endpoint para obter o saldo do usuário
func (bc *Blockchain) HandleSaldo(w http.ResponseWriter, r *http.Request) {
    // Adiciona cabeçalhos CORS
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

// Função para calcular o saldo do usuário
func (bc *Blockchain) CalcularSaldo(usuario string) float64 {
    bc.mu.Lock()
    defer bc.mu.Unlock()

    saldo := 0.0

    for _, bloco := range bc.Blocos {
        if bloco.Evento == "ajustar_saldo" {
            var ajuste map[string]interface{}
            err := json.Unmarshal([]byte(bloco.Resultado), &ajuste)
            if err != nil {
                log.Printf("Erro ao unmarshalar ajustar_saldo: %v", err)
                continue
            }

            if ajuste["usuario"] == usuario {
                // JSON decodifica números como float64 por padrão
                valor, ok := ajuste["valor"].(float64)
                if ok {
                    saldo += valor
                }
            }
        }
    }

    return saldo
}

// Endpoint para criar um novo evento
func (bc *Blockchain) HandleCriarEvento(w http.ResponseWriter, r *http.Request) {
    // Adiciona cabeçalhos CORS
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

    // Adiciona o evento à blockchain
    bc.AdicionarBloco("criar_evento", evento)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(evento)
}

// Endpoint para listar todos os eventos
func (bc *Blockchain) HandleListarEventos(w http.ResponseWriter, r *http.Request) {
    // Adiciona cabeçalhos CORS
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
                log.Printf("Erro ao unmarshalar criar_evento: %v", err)
                continue
            }
            eventosMap[evento.ID] = &evento
        } else if bloco.Evento == "apostar" {
            var aposta Aposta
            err := json.Unmarshal([]byte(bloco.Resultado), &aposta)
            if err != nil {
                log.Printf("Erro ao unmarshalar apostar: %v", err)
                continue
            }

            if evento, exists := eventosMap[aposta.EventoID]; exists {
                evento.Votos[aposta.Opcao] = append(evento.Votos[aposta.Opcao], aposta)
            }
        }
    }

    // Converter o mapa para uma slice
    eventos := []Evento{}
    for _, evento := range eventosMap {
        eventos = append(eventos, *evento)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(eventos)
}

// Endpoint para votar em um evento
func (bc *Blockchain) HandleVotar(w http.ResponseWriter, r *http.Request) {
    // Adiciona cabeçalhos CORS
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

    // Verifica se a opção é válida para o evento
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

    // Adiciona o voto à blockchain
    bc.AdicionarBloco("votar", voto)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(voto)
}

// Função para verificar se a opção é válida para o evento
func (bc *Blockchain) VerificarOpcaoEvento(eventoID int, opcao string) bool {
    for _, bloco := range bc.Blocos {
        if bloco.Evento == "criar_evento" {
            var evento Evento
            err := json.Unmarshal([]byte(bloco.Resultado), &evento)
            if err != nil {
                log.Printf("Erro ao unmarshalar criar_evento: %v", err)
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

// Endpoint para apostar em um evento
func (bc *Blockchain) HandleApostar(w http.ResponseWriter, r *http.Request) {
    // Adiciona cabeçalhos CORS
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

    // Verifica se o usuário tem saldo suficiente
    saldo := bc.CalcularSaldo(req.Usuario)
    if saldo < req.Valor {
        http.Error(w, "Saldo insuficiente", http.StatusBadRequest)
        return
    }

    // Verifica se a opção é válida para o evento
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

    // Adiciona a aposta à opção específica do evento
    bc.mu.Lock()
    defer bc.mu.Unlock()

    // Encontrar o evento
    var evento *Evento
    for _, bloco := range bc.Blocos {
        if bloco.Evento == "criar_evento" {
            var ev Evento
            err := json.Unmarshal([]byte(bloco.Resultado), &ev)
            if err != nil {
                continue
            }
            if ev.ID == req.EventoID {
                evento = &ev
                break
            }
        }
    }

    if evento == nil {
        http.Error(w, "Evento não encontrado", http.StatusBadRequest)
        return
    }

    // Adiciona a aposta à opção específica
    if evento.Votos == nil {
        evento.Votos = make(map[string][]Aposta)
    }
    evento.Votos[req.Opcao] = append(evento.Votos[req.Opcao], aposta)

    // Cria um bloco de aposta
    tipoEvento := "apostar"
    resultadoBytes, _ := json.Marshal(aposta)
    blocoAposta := bc.AdicionarBloco(tipoEvento, string(resultadoBytes))

    // Ajusta o saldo do usuário (deduz o valor apostado)
    ajuste := map[string]interface{}{
        "usuario": req.Usuario,
        "valor":   -req.Valor,
    }
    bc.AdicionarBloco("ajustar_saldo", ajuste)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(blocoAposta)
}

// Endpoint para sacar saldo
func (bc *Blockchain) HandleSacar(w http.ResponseWriter, r *http.Request) {
    // Adiciona cabeçalhos CORS
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

    // Verifica se o usuário tem saldo suficiente
    saldo := bc.CalcularSaldo(req.Usuario)
    if saldo < req.Valor {
        http.Error(w, "Saldo insuficiente para saque", http.StatusBadRequest)
        return
    }

    // Cria um bloco de ajuste de saldo negativo
    ajuste := map[string]interface{}{
        "usuario": req.Usuario,
        "valor":   -req.Valor,
    }
    blocoAjuste := bc.AdicionarBloco("ajustar_saldo", ajuste)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(blocoAjuste)
}

// Endpoint para depositar saldo
func (bc *Blockchain) HandleDepositar(w http.ResponseWriter, r *http.Request) {
    // Adiciona cabeçalhos CORS
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

// Endpoint para concluir um evento
func (bc *Blockchain) HandleConcluirEvento(w http.ResponseWriter, r *http.Request) {
    // Adiciona cabeçalhos CORS
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
        EventoID        int    `json:"evento_id"`
        OpcaoVencedora string `json:"opcao_vencedora"`
    }

    var req ConcluirEventoRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Dados inválidos", http.StatusBadRequest)
        return
    }

    if req.EventoID == 0 || req.OpcaoVencedora == "" {
        http.Error(w, "Todos os campos são obrigatórios", http.StatusBadRequest)
        return
    }

    bc.mu.Lock()
    defer bc.mu.Unlock()

    // Encontrar o evento
    var evento *Evento
    for _, bloco := range bc.Blocos {
        if bloco.Evento == "criar_evento" {
            var ev Evento
            err := json.Unmarshal([]byte(bloco.Resultado), &ev)
            if err != nil {
                continue
            }
            if ev.ID == req.EventoID {
                evento = &ev
                break
            }
        }
    }

    if evento == nil {
        http.Error(w, "Evento não encontrado", http.StatusBadRequest)
        return
    }

    // Verificar se a opção vencedora é válida
    opcaoValida := false
    for _, op := range evento.Opcoes {
        if op == req.OpcaoVencedora {
            opcaoValida = true
            break
        }
    }

    if !opcaoValida {
        http.Error(w, "Opção vencedora inválida", http.StatusBadRequest)
        return
    }

    // Calcular o total apostado na opção vencedora e nas perdedoras
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

    if totalVencedor == 0 {
        http.Error(w, "Nenhuma aposta na opção vencedora", http.StatusBadRequest)
        return
    }

    // Calcular a distribuição do prêmio
    premioPorAposta := totalPerdedor / totalVencedor

    // Distribuir os prêmios aos vencedores
    for opcao, apostas := range evento.Votos {
        if opcao != req.OpcaoVencedora {
            continue
        }
        for _, aposta := range apostas {
            premio := aposta.Valor * premioPorAposta
            ajuste := map[string]interface{}{
                "usuario": aposta.Usuario,
                "valor":   premio,
            }
            bc.AdicionarBloco("ajustar_saldo", ajuste)
        }
    }

    // Registrar o resultado do evento
    resultado := map[string]interface{}{
        "evento_id":        req.EventoID,
        "opcao_vencedora": req.OpcaoVencedora,
    }
    bc.AdicionarBloco("concluir_evento", resultado)

    w.Write([]byte("Evento concluído e prêmios distribuídos com sucesso."))
}

// Inicializa os endpoints HTTP
func (bc *Blockchain) InicializarEndpoints() {
    http.HandleFunc("/blockchain", bc.ExibirBlockchainHTTP)
    http.HandleFunc("/adicionar", bc.HandleAdicionarBloco)
    http.HandleFunc("/validar", bc.HandleValidarBlockchain)
    http.HandleFunc("/receber-blockchain", bc.ReceberBlockchain)
    http.HandleFunc("/receber-bloco", bc.ReceberBloco)

    // Novos Endpoints
    http.HandleFunc("/saldo", bc.HandleSaldo)
    http.HandleFunc("/criar-evento", bc.HandleCriarEvento)
    http.HandleFunc("/eventos", bc.HandleListarEventos)
    http.HandleFunc("/votar", bc.HandleVotar)

    // Endpoints para apostas
    http.HandleFunc("/apostar", bc.HandleApostar)
    http.HandleFunc("/concluir-evento", bc.HandleConcluirEvento)

    // Endpoints para gerenciamento de saldo
    http.HandleFunc("/depositar", bc.HandleDepositar)
    http.HandleFunc("/sacar", bc.HandleSacar)
}
