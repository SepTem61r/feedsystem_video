FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /server ./backend/cmd/main.go
RUN CGO_ENABLED=0 go build -o /worker ./backend/cmd/worker/main.go

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /server /app/server
COPY --from=builder /worker /app/worker

ENTRYPOINT ["/app/server"]
