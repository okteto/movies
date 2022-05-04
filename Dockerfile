FROM golang:buster as dev

# setup okteto message
COPY bashrc /root/.bashrc

WORKDIR /usr/src/app

ADD go.mod go.sum ./
RUN go mod download all

RUN go install github.com/go-delve/delve/cmd/dlv && \
    curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b /usr/bin

ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /usr/local/bin/api cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /usr/local/bin/worker cmd/worker/main.go

FROM scratch

COPY --from=dev /usr/local/bin/api /usr/local/bin/api
COPY --from=dev /usr/local/bin/worker /usr/local/bin/worker

ENTRYPOINT ["/usr/local/bin/api"]
