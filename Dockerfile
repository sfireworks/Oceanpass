FROM golang:1.19-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOPROXY https://goproxy.cn,direct
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go build  -o /oceanpass_http_main src/oceanpass_http_main.go


FROM ubuntu:18.04
#scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ Asia/Shanghai

WORKDIR /root
COPY --from=builder /oceanpass_http_main /root/oceanpass_http_main
COPY --from=builder /build/conf /root/conf
COPY --from=builder /build/credentials /root/credentials
RUN cat /root/conf/*
RUN cat /root/credentials
RUN mkdir .oceanpass
RUN cp -r credentials .oceanpass
CMD ./oceanpass_http_main
