FROM golang:1.9.2

WORKDIR /go/src/BitGainz-Bittrex
COPY . .

RUN go-wrapper download   # "go get -d -v ./..."
RUN go-wrapper install    # "go install -v ./..."

EXPOSE 8080

ENTRYPOINT ["/go/bin/BitGainz-Bittrex"]
