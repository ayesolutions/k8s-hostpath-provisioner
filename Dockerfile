FROM golang:1.16 as builder
WORKDIR /usr/src
COPY . .
RUN go build

FROM debian:buster-slim
COPY --from=builder /usr/src/k8s-hostpath-provisioner /usr/local/bin/k8s-hostpath-provisioner
CMD ["k8s-hostpath-provisioner"]
