FROM golang:1.21-alpine AS builder

RUN apk add --no-cache gcc musl-dev
WORKDIR /app

# FIX: Copy all the code FIRST so Go knows what dependencies to look for
COPY backend/ ./

# Now run tidy to download everything
RUN go mod tidy

# Build the server
RUN CGO_ENABLED=1 GOOS=linux go build -a -o server ./cmd/server/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/server .
COPY backend/db/migrations ./db/migrations

CMD ["./server"]