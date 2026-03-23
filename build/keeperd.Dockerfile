FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /keeperd github.com/arvaliullin/goph-keeper/cmd/keeperd

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /keeperd .
RUN mkdir -p migrations

EXPOSE 8080

CMD ["./keeperd"]
