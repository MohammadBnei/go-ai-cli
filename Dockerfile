FROM golang:1.22

ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /app

COPY ["go.mod", "go.sum", "/app/"]

RUN go mod download

ADD . .

RUN go build  -o go-ai-cli

ENTRYPOINT [ "./go-ai-cli" ]

CMD ["prompt"]



