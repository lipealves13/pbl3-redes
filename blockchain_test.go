package main

import (
	"sync"
	"testing"
)

// Testa a adição de blocos sequencialmente
func TestAdicionarBlocosSequencial(t *testing.T) {
	bc := NovoBlockchain()

	bc.AdicionarBloco("Evento1", "Resultado1")
	bc.AdicionarBloco("Evento2", "Resultado2")

	if len(bc.Blocos) != 3 { // Genesis + 2 blocos
		t.Errorf("Esperado 3 blocos, obtido %d", len(bc.Blocos))
	}

	if !bc.ValidarBlockchain() {
		t.Error("Blockchain deveria ser válida")
	}
}

// Testa a adição de blocos concorrente
func TestAdicionarBlocosConcorrente(t *testing.T) {
	bc := NovoBlockchain()
	var wg sync.WaitGroup
	numBlocos := 100
	wg.Add(numBlocos)

	for i := 0; i < numBlocos; i++ {
		go func(i int) {
			defer wg.Done()
			evento := "EventoConcorrente"
			resultado := "Resultado" + string(i)
			bc.AdicionarBloco(evento, resultado)
		}(i)
	}

	wg.Wait()

	expectedLength := 1 + numBlocos // Genesis + blocos adicionados
	if len(bc.Blocos) != expectedLength {
		t.Errorf("Esperado %d blocos, obtido %d", expectedLength, len(bc.Blocos))
	}

	if !bc.ValidarBlockchain() {
		t.Error("Blockchain deveria ser válida após adições concorrentes")
	}
}

// Testa a integridade após adições concorrentes e apostas
func TestIntegridadeComApostasConcorrentes(t *testing.T) {
	bc := NovoBlockchain()
	var wg sync.WaitGroup
	numOperations := 100

	wg.Add(2 * numOperations)

	// Adiciona apostas e blocos concorrente
	for i := 0; i < numOperations; i++ {
		go func(i int) {
			defer wg.Done()
			AdicionarAposta("usuario"+string(i), float64(i), "Escolha"+string(i), i)
		}(i)

		go func(i int) {
			defer wg.Done()
			evento := "Evento" + string(i)
			resultado := "Resultado" + string(i)
			bc.AdicionarBloco(evento, resultado)
		}(i)
	}

	wg.Wait()

	// Verifica o número de apostas
	apostasMutex.Lock()
	if len(apostas) != numOperations {
		t.Errorf("Esperado %d apostas, obtido %d", numOperations, len(apostas))
	}
	apostasMutex.Unlock()

	// Verifica o número de blocos
	expectedBlocos := 1 + numOperations // Genesis + blocos adicionados
	if len(bc.Blocos) != expectedBlocos {
		t.Errorf("Esperado %d blocos, obtido %d", expectedBlocos, len(bc.Blocos))
	}

	// Valida a blockchain
	if !bc.ValidarBlockchain() {
		t.Error("Blockchain deveria ser válida após adições concorrentes de apostas e blocos")
	}
}
