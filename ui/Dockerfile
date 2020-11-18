FROM nginx:alpine

# Used for debug
#RUN apk update \
#&& apk add bash \
#&& rm -rf /var/cache/apk/*

COPY /css /usr/share/nginx/html/css
COPY /js /usr/share/nginx/html/js
COPY /*.html /usr/share/nginx/html
