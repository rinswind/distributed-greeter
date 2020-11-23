#
# Builder
#
FROM golang:alpine as builder

# Git is required for fetching the dependencies.
RUN apk update && \
    apk add --no-cache git ca-certificates tzdata && \
    update-ca-certificates

# Create appuser
ENV USER=appuser
ENV UID=10001

RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "${UID}" \    
    "${USER}"

# Copy code into builder
RUN mkdir /app
COPY . /app

WORKDIR /app

# Fetch dependencies
RUN go mod download
RUN go mod verify

# Build statically linked app
RUN CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o main ./cmd/main.go

# 
# Runtime
#
FROM scratch

# Import from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /app/main .

# Use an unprivileged user.
USER appuser:appuser

ENTRYPOINT ["/main"]
CMD ["-port", "8080"]

ENV JWT_SECRET="secret"

EXPOSE 8080
