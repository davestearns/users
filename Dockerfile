FROM alpine
RUN apk add --no-cache ca-certificates
COPY userservice /userservice
ENTRYPOINT [ "/userservice" ]