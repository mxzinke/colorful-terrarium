# Build-Stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# copy source code (incl subpackages)
COPY *.go ./
COPY terrain/ ./terrain/
COPY polygon/ ./polygon/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app-binary

# Finale Stage
FROM alpine:latest

ARG USER=default

# add new user
RUN adduser -D $USER

USER $USER
WORKDIR /home/$USER

# copy binary
COPY --from=builder /app-binary .

EXPOSE 8080

# copy data (geojson)
COPY data/ ./data/

CMD ["./app-binary"]
