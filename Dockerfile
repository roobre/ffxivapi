FROM golang:latest as build

LABEL maintainer="Roberto Santalla <roobre@roobre.es>"

WORKDIR /app

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ffxivapi ./http/cmd
RUN chmod 755 ffxivapi


FROM alpine:latest

RUN apk add curl
RUN mkdir -p /app/http
COPY --from=build /app/http/swagger.yaml /app/http
COPY --from=build /app/ffxivapi /app

CMD /app/ffxivapi
