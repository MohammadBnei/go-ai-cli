FROM golang:1.20-bullseye

RUN go install github.com/MohammadBnei/go-openai-cli@latest

VOLUME /config.yaml

ENV CONFIG=/config.yaml

CMD ["go-openai-cli", "prompt"]

