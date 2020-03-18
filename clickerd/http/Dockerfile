FROM golang:1.12.13-buster as build
ARG ARCH="linux.amd64"

# todo;; real package management
RUN go get "github.com/xujiajun/gorouter" \
           "github.com/go-yaml/yaml"      \
           "github.com/jw3/ppc/cli"

COPY .  /tmp/build
WORKDIR /tmp/build/clickerd

RUN make http \
 && mkdir /tmp/dist \
 && cp dist/* /tmp/dist

# ------------------------

FROM centos:7
ARG ARCH="linux.amd64"
RUN yum -y install openssh openssh-server openssh-clients openssl-libs
COPY --from=build  /tmp/dist/clicker-http.${ARCH} /bin/clickerd-http
EXPOSE 9000
ENTRYPOINT [ "clickerd-http" ]
