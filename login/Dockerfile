#
# Builder
#
FROM example-services/builder:latest as builder

# 
# Runtime
#
FROM example-services/base:latest

ENV HTTP_PORT=8080 \
    REDIS_DSN="redis:6379" \
    ACCESS_TOKEN_SECRET=secret \
    ACCESS_TOKEN_EXPIRY=5 \
    REFRESH_TOKEN_SECRET=secret \
    REFRESH_TOKEN_EXPIRY=60 \
    GIN_MODE=release

EXPOSE 8080
