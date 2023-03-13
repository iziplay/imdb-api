FROM golang:1.20.2-bullseye as builder

# Copy dependencies source files
RUN mkdir /data
COPY go.mod /data
COPY go.sum /data

WORKDIR /data

# Download dependencies
RUN go mod download

# Add source code
COPY . .

# Build
RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian11
COPY --from=builder /go/bin/app /
CMD ["/app"]