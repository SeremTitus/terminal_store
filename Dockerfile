# ---- build stage ----
FROM golang:1.25.5 AS build

WORKDIR /src

# Copy module files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the app from server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./server

# ---- run stage ----
FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=build /src/app /app/app
COPY --from=build /src/migrations /app/migrations

EXPOSE 8080
CMD ["/app/app"]
