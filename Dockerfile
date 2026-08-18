FROM golang:1.23-alpine AS build
RUN apk add --no-cache gcc musl-dev
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /statusly ./cmd/statusly

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /statusly /app/statusly
COPY config.yaml /app/config.yaml
COPY web /app/web
RUN mkdir -p /app/data
EXPOSE 8080
VOLUME ["/app/data"]
ENTRYPOINT ["/app/statusly"]
