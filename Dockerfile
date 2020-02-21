FROM golang:1.13

MAINTAINER "lazycat7706@gmail.com"

ENV PREFIX=/go/bin/lazycron
ENV SRC_PATH=/go/src/lazycron
ENV PATH=$PATH:$PREFIX

RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct

COPY . $SRC_PATH

WORKDIR $SRC_PATH

RUN /bin/bash ./build.sh
RUN cp -rf ./out $PREFIX

WORKDIR $PREFIX

CMD ["master", "--conf", "master.json"]
