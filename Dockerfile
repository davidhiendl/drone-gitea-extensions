# build
FROM golang:1.19-alpine as go
ADD . /build
RUN cd /build && go build -o drone-gitea-extensions .

# fetch certs
FROM alpine:3.6 as alpine-certs
RUN apk add -U --no-cache ca-certificates

# assemble
FROM alpine:3.6
EXPOSE 3000
ENV GODEBUG netdns=go

COPY --from=alpine-certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go /build/drone-gitea-extensions /bin/

ENTRYPOINT ["/bin/drone-gitea-extensions"]
