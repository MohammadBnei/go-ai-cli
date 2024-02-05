FROM golang:1.20

ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /app

COPY ["go.mod", "go.sum", "./"]

RUN go mod download

ADD . .

RUN go build  -o go-ai-cli

ENTRYPOINT [ "./go-ai-cli" ]

CMD ["prompt"]

# FROM ubuntu:20.04

# RUN apt update && apt install -y libc6

# COPY --from=0 /app/go-ai-cli /go-ai-cli

# VOLUME /.config

# ENV CONFIG=/.config/config.yaml


