FROM golang:latest AS builder

WORKDIR /build

COPY go.mod go.sum scheduler.db .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./scheduler

FROM scratch

WORKDIR /app

COPY --from=builder /build/scheduler .
COPY --from=builder /build/web ./web

ENV TODO_PORT=7540
ENV TODO_DBFILE=scheduler.db

EXPOSE ${TODO_PORT}/tcp

CMD ["/app/scheduler"]