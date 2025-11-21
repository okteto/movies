FROM golang:1.24-bookworm as dev

WORKDIR /usr/src/app

ADD go.mod go.sum ./
RUN go mod download all


ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /usr/local/bin/worker cmd/worker/main.go

FROM scratch

COPY --from=dev /usr/local/bin/worker /usr/local/bin/worker

ENTRYPOINT ["/usr/local/bin/worker"]
