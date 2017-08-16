FROM golang:1.8.3-stretch

WORKDIR /go/src/minikube-lb-patch

COPY . .

RUN go get
RUN go build

FROM debian:stretch-20170723

COPY --from=0 /go/src/minikube-lb-patch/minikube-lb-patch /app/

ENTRYPOINT ["/app/minikube-lb-patch"]
