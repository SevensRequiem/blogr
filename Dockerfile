# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o /app/main .

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/main /app/main
COPY ./static /app/static
COPY ./views /app/views
COPY ./assets /app/assets

# .env
COPY ./.env /app/.env

# expose port
EXPOSE 1212
CMD ["/app/main"]
