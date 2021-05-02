FROM golang:1.16 as builder
WORKDIR /usr/src
COPY . .
RUN go build

FROM debian:buster-slim
#RUN apt-get update && apt-get install -y extra-runtime-dependencies && rm -rf /var/lib/apt/lists/*
COPY --from=builder /usr/src/template-golang /usr/local/bin/template-golang
CMD ["template-golang"]
