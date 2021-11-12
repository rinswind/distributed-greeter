#
# Builder
#
FROM example-services/builder:latest AS builder

# 
# Runtime
#
FROM example-services/base:latest

COPY config.yaml /app

ENV HTTP_PORT="8080" \
    GIN_MODE="release"

EXPOSE 8080
