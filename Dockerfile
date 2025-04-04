FROM golang:1.24.0
WORKDIR /app
ENV CONFIG_PATH=./config/config_container.yaml
COPY . .
RUN go mod download
RUN go build -o sso cmd/sso/main.go
EXPOSE 50051
CMD ["/app/sso"]