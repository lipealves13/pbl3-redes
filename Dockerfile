# Dockerfile

# Base para Go
FROM golang:1.23-alpine

# Diretório de trabalho
WORKDIR /app

# Copia apenas os arquivos de módulos
COPY go.mod ./

# Instala dependências
RUN go mod download

# Copia o restante do código
COPY . .

# Compila o código
RUN go build -o blockchain .

# Porta que o aplicativo irá rodar
EXPOSE 8080

# Comando de execução
CMD ["./blockchain"]
