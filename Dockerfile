FROM golang:1.27.0-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/openuss ./cmd/openuss

FROM alpine:3.21
RUN adduser -D -u 10001 openuss
COPY --from=build /out/openuss /usr/local/bin/openuss
USER openuss
EXPOSE 80
ENTRYPOINT ["/usr/local/bin/openuss"]
