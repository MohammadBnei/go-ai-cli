FROM golang:1.20-alpine

WORKDIR /app

RUN go install github.com/MohammadBnei/go-openai-cli@latest

ENTRYPOINT ["go-openai-cli"]