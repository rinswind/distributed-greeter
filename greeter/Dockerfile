#
# Builder
#
FROM example-services/builder as builder

COPY ../msi 

# 
# Runtime
#
FROM example-services/base

ENV HTTP_PORT="8080" \
    GIN_MODE="release"

EXPOSE 8080
