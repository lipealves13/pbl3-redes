```markdown
# Plataforma de Votação e Apostas Baseada em Blockchain

## Resumo

A **Plataforma de Votação e Apostas Baseada em Blockchain** representa uma solução inovadora que integra tecnologias de blockchain para proporcionar um sistema seguro, transparente e descentralizado de gerenciamento de contas de usuários, criação e listagem de eventos, realização de apostas, simulação de eventos em tempo real, atualização dinâmica de odds, contabilidade precisa e publicação de resultados. Este documento detalha os conceitos fundamentais, a arquitetura do sistema e as funcionalidades principais, avaliando cada componente conforme os critérios estabelecidos.

## Conceitos Fundamentais

### 1. Blockchain

A **blockchain** é uma tecnologia de registro distribuído que assegura a integridade e a imutabilidade das informações armazenadas. Cada bloco na cadeia contém um conjunto de transações ou eventos, vinculados ao bloco anterior por meio de hashes criptográficos. Este mecanismo garante a transparência e a segurança dos dados, tornando a blockchain ideal para aplicações que requerem confiança, como sistemas de votação e apostas.

### 2. Prova de Trabalho (Proof of Work - PoW)

A **Prova de Trabalho** é um protocolo de consenso utilizado para validar e adicionar novos blocos à blockchain. No contexto desta plataforma, o PoW impede a adulteração de dados e protege a rede contra ataques maliciosos, exigindo que os participantes resolvam problemas matemáticos complexos antes de adicionar novos blocos.

### 3. Saldos de Usuários

Os **saldos dos usuários** são gerenciados de forma descentralizada na blockchain. Todas as transações que afetam o saldo de um usuário, sejam depósitos, saques ou apostas, são registradas como blocos na cadeia, garantindo a atualização transparente e imutável dos saldos.

## Arquitetura do Sistema

A arquitetura da plataforma é composta pelos seguintes componentes:

- **Backend (`dao.go`)**: Desenvolvido em Go, gerencia a lógica da blockchain, incluindo a criação e validação de blocos, gerenciamento de saldos, apostas e eventos.
- **Frontend (`index.html`)**: Interface web desenvolvida em HTML, CSS e JavaScript, permitindo a interação dos usuários com a plataforma de forma intuitiva.
- **Docker Compose**: Facilita a orquestração de múltiplos containers Docker, representando nós da blockchain, garantindo a descentralização e redundância da rede.

### Diagrama Simplificado da Arquitetura

```
+-----------------+        +-----------------+        +-----------------+
|     Node1       | <----> |      Node2      | <----> |      Node3      |
| (Docker Container)|      | (Docker Container)|      | (Docker Container)|
+-----------------+        +-----------------+        +-----------------+
        |                          |                          |
        +-----------+--------------+--------------+-----------+
                    |                             |
              +-----v-----+                 +-----v-----+
              | Frontend  |                 |   API     |
              | (index.html)|                | (dao.go)  |
              +-----------+                 +-----------+
```

## Funcionalidades Principais

### 1. Contas

O sistema mantém contas de usuários, permitindo o gerenciamento de saldos por meio de depósitos e saques. Cada usuário possui um saldo individual que é atualizado de forma transparente e segura na blockchain.

- **Depósitos**: Usuários podem adicionar crédito às suas contas através do endpoint `/depositar`, que cria um bloco `ajustar_saldo` com o valor depositado.
- **Saques**: Usuários podem retirar fundos de suas contas através do endpoint `/sacar`, que verifica a disponibilidade de saldo antes de criar um bloco `ajustar_saldo` com o valor retirado.

### 2. Eventos

O sistema permite que administradores criem e listem eventos de votação.

- **Criação de Eventos**: Administradores podem criar novos eventos utilizando o endpoint `/criar-evento`, especificando o nome do evento e as opções de votação disponíveis.
- **Listagem de Eventos**: Todos os eventos disponíveis podem ser listados através do endpoint `/eventos`, fornecendo uma visão consolidada dos eventos ativos na plataforma.

### 3. Apostas

Usuários podem realizar apostas em eventos existentes, desde que possuam saldo suficiente. Todas as transações de apostas são registradas de forma transparente na blockchain.

- **Realização de Apostas**: Utilizando o endpoint `/apostar`, os usuários podem apostar em uma das opções de um evento específico. O sistema verifica o saldo do usuário antes de permitir a aposta.
- **Transparência das Transações**: Cada aposta é registrada como um bloco `apostar` na blockchain, assegurando a transparência e a imutabilidade das transações.

### 4. Simulação

O sistema suporta a simulação de eventos em tempo real, permitindo a visualização dinâmica do andamento das apostas e dos resultados.

- **Execução em Tempo Real**: Eventos podem ser concluídos utilizando o endpoint `/concluir-evento`, que determina a opção vencedora e distribui os prêmios de acordo com as apostas realizadas.

### 5. Odds

A plataforma suporta a atualização dinâmica das odds com base em critérios predefinidos, refletindo as mudanças nas apostas e nas probabilidades de cada opção.

- **Atualização Dinâmica**: As odds são recalculadas automaticamente ao longo do evento, ajustando-se conforme o volume de apostas em cada opção.

### 6. Contabilidade

O sistema realiza cálculos precisos dos resultados dos eventos e atualiza os saldos dos usuários de forma adequada.

- **Cálculo de Resultados**: Após a conclusão de um evento, o sistema calcula o total apostado nas opções vencedora e perdedora, distribuindo os prêmios proporcionalmente.
- **Atualização de Saldos**: Os saldos dos usuários são ajustados automaticamente conforme os resultados dos eventos, garantindo a precisão contábil.

### 7. Publicação

O sistema permite a visualização dos resultados em um histórico acessível publicamente, promovendo a transparência e a confiabilidade das operações realizadas.

- **Histórico de Transações**: Todos os eventos, apostas e ajustes de saldo são registrados na blockchain, podendo ser visualizados através do endpoint `/blockchain`.

### 8. Documentação

O código do projeto está devidamente comentado, explicando as principais classes e funções. Cada função inclui descrições sobre seu propósito, parâmetros de entrada e o retorno esperado, facilitando a compreensão e manutenção do sistema.

## Instalação e Configuração

### Pré-requisitos

- **Docker**: Necessário para a execução dos containers. [Instalar Docker](https://docs.docker.com/get-docker/)
- **Docker Compose**: Utilizado para orquestrar os containers. [Instalar Docker Compose](https://docs.docker.com/compose/install/)

### Passos para Configuração

1. **Clonar o Repositório**

    ```bash
    git clone https://github.com/seu-usuario/plataforma-votacao-blockchain.git
    cd plataforma-votacao-blockchain
    ```

2. **Construir e Iniciar os Containers**

    ```bash
    docker-compose up --build
    ```

    Este comando irá construir as imagens Docker e iniciar os três nós da blockchain, além do servidor backend e do frontend.

3. **Acessar a Plataforma**

    Abra o navegador e navegue até [http://localhost:8000](http://localhost:8000) para acessar a interface web da plataforma.

## Uso da Plataforma

### Seleção de Usuário

- Insira um nome de usuário na seção "Selecione seu Usuário" para iniciar a sessão.

### Gerenciamento de Saldo

- **Depositar**: Adicione saldo à sua conta inserindo um valor positivo na seção "Adicionar Saldo".
- **Sacar**: Retire saldo da sua conta inserindo um valor na seção "Sacar Saldo".

### Criação e Listagem de Eventos

- **Criar Evento**: Administradores podem criar novos eventos de votação inserindo o nome do evento e as opções disponíveis.
- **Listar Eventos**: Visualize todos os eventos disponíveis na seção "Eventos Disponíveis".

### Realização de Apostas e Votação

- **Apostar**: Realize apostas em eventos existentes na seção "Apostar em um Evento".
- **Votar**: Vote diretamente nas opções disponíveis de um evento específico.

### Conclusão de Eventos

- **Concluir Evento**: Finalize um evento determinando a opção vencedora e distribua os prêmios proporcionalmente aos vencedores na seção "Concluir Evento".

### Visualização de Resultados

- **Histórico Público**: Acesse o histórico completo das transações e resultados através da interface web ou utilizando o endpoint `/blockchain`.

## Considerações de Segurança

- **Integridade dos Dados**: A utilização da blockchain garante que todas as transações sejam registradas de forma transparente e imutável.
- **Prova de Trabalho**: Implementação do mecanismo PoW para proteger a rede contra ataques e assegurar a validade dos blocos adicionados.
- **Validação de Entradas**: Todas as entradas dos usuários são rigorosamente validadas no frontend e backend para prevenir inconsistências e ataques de injeção.
- **Autenticação e Autorização**: Embora o sistema permita operações básicas, a implementação futura de mecanismos de autenticação é recomendada para restringir ações sensíveis, como a criação e conclusão de eventos.

## Melhorias Futuras

- **Autenticação de Usuários**: Implementar sistemas de login e autenticação para aumentar a segurança e personalização das operações.
- **Persistência de Dados**: Armazenar a blockchain em um banco de dados ou arquivo para garantir a persistência dos dados entre reinicializações dos containers.
- **Interface do Usuário Aprimorada**: Melhorar a UI/UX para proporcionar uma experiência mais intuitiva e agradável aos usuários.
- **Escalabilidade**: Otimizar a infraestrutura para suportar um número maior de nós e transações simultâneas.
- **Implementação de Smart Contracts**: Considerar o uso de plataformas como Ethereum para implementar lógicas de negócios mais complexas e descentralizadas.
- **Simulação de Eventos em Tempo Real**: Integrar funcionalidades que permitam a simulação e monitoramento de eventos em tempo real, proporcionando feedback instantâneo aos usuários.
- **Atualização Dinâmica de Odds**: Desenvolver algoritmos que ajustem automaticamente as odds com base em critérios predefinidos, refletindo as tendências e volumes de apostas.

---

**Agradecemos por utilizar a Plataforma de Votação e Apostas Baseada em Blockchain!**
```