FROM golang

#ARG repo
#ARG version
WORKDIR /
ADD . /go/src/$repo
WORKDIR /go/src/$repo
#RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X $repo/main.Version=$version" -ldflags "-s" -a -o ./main
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -o ./main

FROM alpine:latest
ARG repo
RUN apk --no-cache add ca-certificates && \
    mkdir -p /root/.ssh && \
    echo "github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==" > /root/.ssh/known_hosts

ENV SSH_KNOWN_HOSTS=/root/.ssh/known_hosts

WORKDIR /root/
COPY --from=0 /go/src/$repo/main .
COPY --chmod=755 entrypoint.sh /root/entrypoint.sh
CMD ["/root/entrypoint.sh"]


