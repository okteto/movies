FROM golang:buster as dev

WORKDIR /usr/src/app

ADD go.mod go.sum ./
RUN go mod download all
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /usr/local/bin/api cmd/api/main.go

FROM scratch

# RUN --mount=type=secret,id=ST_NPM_READONLY_TOKEN,env=ST_NPM_READONLY_TOKEN npm run bootstrap

RUN --mount=type=secret,id=npm_token \
    bash -c 'echo "secret: $(cat /run/secrets/npm_token)"'

COPY --from=dev /usr/local/bin/api /usr/local/bin/api
COPY --from=dev /usr/src/app/data /data

ENTRYPOINT ["/usr/local/bin/api"]
