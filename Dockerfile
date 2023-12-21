FROM alpine

RUN apk update && apk upgrade && apk add bash curl

COPY test.sh .

CMD ["bash", "test.sh"]