# Relatório do Projeto: Sistema de Apostas Descentralizado com Blockchain Local

## Introdução

Este projeto tem como objetivo desenvolver um sistema de apostas descentralizado, baseado em blockchain, para eventos simples, como "Cara ou Coroa". A motivação principal foi explorar conceitos de blockchain aplicados a um ambiente acadêmico, onde cada participante pode executar um nó da rede em um ambiente Docker. A solução proposta permite registrar eventos, votar em resultados e validar a descentralização em uma rede local.

---

## Fundamentação Teórica

### Blockchain
Blockchain é uma tecnologia distribuída que permite o registro de transações em uma cadeia de blocos. Cada bloco contém:
- **Dados da transação**: Informações específicas do evento.
- **Hash**: Um identificador único gerado criptograficamente.
- **Hash do bloco anterior**: Relaciona o bloco atual ao anterior, garantindo integridade.

A tecnologia é amplamente utilizada em sistemas descentralizados, pois:
1. **Imutabilidade**: Os registros não podem ser alterados sem consenso.
2. **Transparência**: Todas as transações são visíveis para os participantes.
3. **Resiliência**: Funciona mesmo que partes da rede falhem.

### Sistema de Apostas
Um sistema de apostas permite que usuários façam previsões sobre o resultado de um evento, apostando valores em favor de um resultado. No contexto deste projeto:
1. **Eventos Simples**: Exemplo: "Cara" ou "Coroa".
2. **Apostas Registradas em Blockchain**: Garantem imutabilidade e transparência.
3. **Descentralização**: Cada nó da rede valida as transações, reduzindo a necessidade de uma entidade central.

### Docker e Redes Locais
O Docker é uma ferramenta que permite criar, implantar e gerenciar aplicativos em contêineres. Ele é ideal para:
1. **Isolamento**: Cada nó opera independentemente.
2. **Escalabilidade**: Facilita o teste com múltiplos nós locais.
3. **Facilidade de Deploy**: A configuração com `docker-compose` simplifica a execução de múltiplos contêineres.

---

## Metodologia

### Estrutura do Sistema
1. **Blockchain Local**: Implementada em Go, com suporte para criação de blocos, registro de eventos e validação de transações.
2. **Casa de Apostas**:
   - Registro de eventos (ex.: "Cara ou Coroa").
   - Votação em resultados.
   - Validação descentralizada.
3. **Docker Compose**: Configurado para simular uma rede local com três nós.

### Etapas de Desenvolvimento
1. **Configuração da Blockchain**:
   - Implementação básica em Go.
   - Suporte para criação e validação de blocos.
2. **Lógica de Apostas**:
   - Registro de eventos.
   - Mecanismo de votação e cálculo de resultados.
3. **Ambiente Docker**:
   - Configuração de contêineres com o Docker.
   - Integração dos nós com `docker-compose`.

### Ferramentas Utilizadas
- **Linguagem**: Go.
- **Ambiente de Execução**: Docker.
- **Gerenciamento**: Docker Compose.
- **Teste**: Simulação local com múltiplos nós.

---

## Resultados

### Implementação
- A blockchain foi desenvolvida com suporte para:
  - Registro de eventos ("Cara ou Coroa").
  - Validação de votos em resultados.
  - Registro imutável de transações.
- A solução foi containerizada com Docker, permitindo a execução de múltiplos nós.

### Testes
Os testes foram realizados em ambiente local com três nós conectados por `docker-compose`. Os seguintes cenários foram avaliados:
1. Registro de eventos e apostas.
2. Sincronização entre nós.
3. Validação descentralizada de transações.

Os resultados mostraram que o sistema funciona conforme esperado:
- Eventos e votos são registrados corretamente.
- Os nós sincronizam os blocos de forma consistente.

---

## Conclusão

O projeto demonstrou a viabilidade de utilizar blockchain para implementar um sistema de apostas descentralizado em um ambiente acadêmico. A solução proposta:
1. Explorou conceitos fundamentais de blockchain, como imutabilidade e descentralização.
2. Demonstrou a integração de tecnologias modernas, como Docker, para facilitar o desenvolvimento e os testes.
3. Ofereceu um ambiente acessível para simular aplicações descentralizadas.

Futuras melhorias incluem:
- Implementação de recompensas automáticas para usuários vencedores.
- Criação de uma interface gráfica para facilitar o uso.
- Suporte a eventos mais complexos.

---

## Referências

1. Nakamoto, S. (2008). Bitcoin: A Peer-to-Peer Electronic Cash System.
2. Docker Documentation: [https://docs.docker.com/](https://docs.docker.com/)
3. Go Programming Language: [https://golang.org/](https://golang.org/)

---
