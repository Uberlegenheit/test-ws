FROM golang:1.22-alpine AS build

WORKDIR /app

COPY . /app

RUN go build -o ./cli ./main.go

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/cli /bin/cli

COPY --from=build /app/dao/postgres/migrations /app/dao/postgres/migrations

COPY --from=build /app/resources /app/resources

COPY ./firebase-auth.json /app/firebase-auth.json

RUN ln -s /bin/cli /app/cli