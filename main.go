// main.go

package main

import (
    "log"
    "net/http"
    "os"
    "strings"
    "time"
)

func main() {
    // Obtém a lista de peers a partir de variáveis de ambiente
    peersEnv := os.Getenv("PEERS")
    var peers []string
    if peersEnv != "" {
        peers = strings.Split(peersEnv, ",")
    }

    blockchain := NovoBlockchain(peers)

    // Inicializa os endpoints HTTP
    blockchain.InicializarEndpoints()

    // Inicia o servidor HTTP
    go func() {
        log.Println("Servidor HTTP iniciado na porta 8080")
        if err := http.ListenAndServe(":8080", nil); err != nil {
            log.Fatalf("Erro ao iniciar o servidor: %v", err)
        }
    }()

    // Sincroniza com os peers periodicamente
    go func() {
        for {
            blockchain.SincronizarComPeers()
            time.Sleep(10 * time.Second) // Intervalo de sincronização
        }
    }()

    // Exibe a blockchain no console
    blockchain.ImprimirBlockchain()

    // Mantém o programa rodando
    select {}
}
