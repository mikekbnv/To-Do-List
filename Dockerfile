FROM golang:alpine AS builder

ENV GO111MODULE=on

WORKDIR ./To-Do-List

COPY . .

RUN go build

EXPOSE 8000

CMD ["./To-Do-List"]
