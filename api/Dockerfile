FROM golang:buster as dev

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download all

COPY pkg pkg
COPY cmd cmd

RUN --mount=type=cache,sharing=private,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -v -o /usr/local/bin/api cmd/api/main.go

FROM scratch

COPY --from=dev /usr/local/bin/api /usr/local/bin/api

ENTRYPOINT ["/usr/local/bin/api"]
