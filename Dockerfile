# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

# Baixar as dependências
RUN go mod tidy

# Compilar o aplicativo
RUN go build -o festa-app .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/festa-app .
COPY ./templates /app/templates
COPY ./static /app/static

EXPOSE 3000
CMD ["./festa-app"]