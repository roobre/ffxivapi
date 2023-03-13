FROM golang:1.20-alpine3.17 as build

LABEL maintainer="Roberto Santalla <roobre@roobre.es>"

WORKDIR /app

COPY . .
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ffxivapi ./http/cmd && \
    chmod +rx ffxivapi

FROM alpine:3.17

RUN apk add curl
RUN mkdir -p /app/http
COPY --from=build /app/http/swagger.yaml /app/http
COPY --from=build /app/ffxivapi /app

WORKDIR /app
CMD /app/ffxivapi
