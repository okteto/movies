FROM golang:1.24-bookworm AS dev

WORKDIR /usr/src/app

ADD go.mod go.sum ./
RUN go mod download all
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /usr/local/bin/api cmd/api/main.go

FROM scratch

COPY --from=dev /usr/local/bin/api /usr/local/bin/api

ENTRYPOINT ["/usr/local/bin/api"]
