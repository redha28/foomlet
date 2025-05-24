FROM golang:1.23.5-alpine

WORKDIR /app

# Install dependencies
RUN apk add --no-cache curl unzip

# Install migrate CLI
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz \
    | tar xvz -C /usr/local/bin

# Copy semua file project
COPY . .

# Build project
RUN go build -o main ./cmd/main.go

# Copy dan buat script bisa dieksekusi
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

EXPOSE 8080

CMD ["./entrypoint.sh"]
