# Build AquaChain in a stock Go builder container
FROM golang:1.9-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers

ADD . /aquachain
RUN cd /aquachain && make aquachain

# Pull AquaChain into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /aquachain/build/bin/aquad /usr/local/bin/

EXPOSE 8545 8546 21303 21303/udp 30304/udp
ENTRYPOINT ["aquachain"]
