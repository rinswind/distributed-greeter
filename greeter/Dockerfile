#
# Builder
#
FROM example-services/builder as builder

# 
# Runtime
#
FROM example-services/base

ENV HTTP_PORT="8080" \
    REDIS_DSN="redis:6379" \
    GIN_MODE=release

EXPOSE 8080
