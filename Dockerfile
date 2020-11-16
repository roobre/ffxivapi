FROM golang:latest as build

LABEL maintainer="Roberto Santalla <roobre@roobre.es>"

WORKDIR /app

COPY . .
RUN make


FROM alpine:latest

RUN apk add curl
RUN mkdir -p /app/http
COPY --from=build /app/http/swagger.yaml /app/http
COPY --from=build /app/ffxivapi /app

CMD /app/ffxivapi
