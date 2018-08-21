FROM scratch
COPY userservice /userservice
ENTRYPOINT [ "/userservice" ]