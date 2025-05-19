FROM golang:buster AS dev

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download all

COPY pkg pkg
COPY cmd cmd

RUN --mount=type=cache,sharing=private,target=/root/.cache/go-build \ 
    CGO_ENABLED=0 GOOS=linux go build -v -o /usr/local/bin/worker cmd/worker/main.go

FROM scratch

COPY --from=dev /usr/local/bin/worker /usr/local/bin/worker

ENTRYPOINT ["/usr/local/bin/worker"]
