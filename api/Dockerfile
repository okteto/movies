FROM golang:buster as dev

WORKDIR /usr/src/app

ADD go.mod go.sum ./
RUN go mod download all
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /usr/local/bin/api cmd/api/main.go

FROM scratch

COPY --from=dev /usr/local/bin/api /usr/local/bin/api
COPY --from=dev /usr/src/app/data /data

ENTRYPOINT ["/usr/local/bin/api"]
