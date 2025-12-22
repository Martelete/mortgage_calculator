FROM golang:1.24.3

WORKDIR /app

# Copy module files first (important for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy application files
COPY main.go .
COPY index.html .
COPY fonts ./fonts

# Build app
RUN go build -o mortgage_app main.go

# Expose web port
EXPOSE 8080

# Run app
CMD ["./mortgage_app"]
