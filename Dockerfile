FROM golang:1.14-alpine AS build
ENV CGO_ENABLED=0
RUN mkdir /app
ADD . /app
WORKDIR /app

RUN go build -o /bin/parser cmd/parser/main.go

FROM alpine:latest
COPY --from=build /bin/parser /bin/parser
COPY --from=build /app/scripts/init.sh /init.sh
RUN chmod +x /init.sh

ENTRYPOINT ["/init.sh"]