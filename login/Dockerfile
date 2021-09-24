#
# Builder
#
FROM example-services/builder:latest as builder

# 
# Runtime
#
FROM example-services/base:latest

ENV HTTP_PORT="8080" \
    REDIS_DSN="redis:6379" \
    GIN_MODE=release

EXPOSE 8080
