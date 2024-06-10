FROM golang:1.22-alpine AS build

WORKDIR /app

COPY . /app

RUN go build -o ./cli ./main.go

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/cli /bin/cli

RUN ln -s /bin/cli /app/cli