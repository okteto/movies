FROM golang:buster as dev

WORKDIR /usr/src/app

ADD go.mod go.sum ./
RUN go mod download all

ADD . .
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux go build -v -o /usr/local/bin/api cmd/api/main.go
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux go build -v -o /usr/local/bin/worker cmd/worker/main.go

FROM scratch

COPY --from=dev /usr/local/bin/api /usr/local/bin/api
COPY --from=dev /usr/local/bin/worker /usr/local/bin/worker

ENTRYPOINT ["/usr/local/bin/api"]
