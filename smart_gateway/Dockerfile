FROM golang:1.24 AS build
WORKDIR /app
COPY . .
RUN go mod download
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/rabbit_sql_pipeline ./cmd/rabbit_sql_pipeline

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /bin/rabbit_sql_pipeline /app/rabbit_sql_pipeline
ENV TZ=Asia/Shanghai
ENTRYPOINT ["/app/rabbit_sql_pipeline"]