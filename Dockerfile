FROM ubuntu

ARG os=linux
ARG aarch=arm64
ARG version=0.11.0

RUN apt update && apt install -y ca-certificates

ADD "https://github.com/MohammadBnei/go-openai-cli/releases/download/$version/go-openai-cli-$version-$os-$aarch.tar.gz" go-openai-cli-$version-$os-$aarch.tar.gz 

RUN tar -xvf go-openai-cli-$version-$os-$aarch.tar.gz

RUN chmod +x go-openai-cli

VOLUME /.config

ENV CONFIG=/.config/config.yaml

ENTRYPOINT [ "./go-openai-cli" ]

CMD ["prompt"]

