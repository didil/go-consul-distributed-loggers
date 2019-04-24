FROM golang:1.12.1

WORKDIR /app

ADD Makefile Makefile

RUN make deps

ADD . .

RUN make build

CMD ["./main"]