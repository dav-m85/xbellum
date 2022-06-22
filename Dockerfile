FROM golang:1.16-alpine AS server

WORKDIR /src/
COPY go.* ./
RUN go mod download

WORKDIR /src/
COPY . ./
RUN go build -ldflags="-s -w" -o /bin/xbellum main.go

FROM alpine

RUN mkdir /bin/data
VOLUME /bin/data

WORKDIR /bin
COPY --from=server /bin/xbellum /bin/xbellum

EXPOSE 8082

CMD ["/bin/xbellum", "server"]
