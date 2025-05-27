# Estágio 1: Build da aplicação Go
FROM golang:1.22-alpine AS builder

# Define o diretório de trabalho dentro do contêiner de build
WORKDIR /app

# Copie os arquivos de gerenciamento de módulos (da raiz do projeto)
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copie todo o código fonte do projeto para o diretório /app
# Isso incluirá o diretório 'cmd' e seu conteúdo.
COPY . .

# Compile a aplicação Go.
# O comando de build aponta para o diretório ./cmd onde o main package está localizado.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /app/server ./cmd

# Estágio 2: Criação da imagem final leve
FROM alpine:latest

WORKDIR /root/

# Copie o binário compilado do estágio 'builder'
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]