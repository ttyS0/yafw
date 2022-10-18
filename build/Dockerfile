FROM ubuntu:jammy AS base

ENV TZ=Asia/Shanghai \
    DEBIAN_FRONTEND=noninteractive
ENV LANG en_US.UTF8

RUN sed -i -E 's#https?://[^/]+#http://mirror.sjtu.edu.cn#g' /etc/apt/sources.list && \
    apt update && apt install -y locales && \
    localedef -i en_US -c -f UTF-8 -A /usr/share/locale/locale.alias en_US.UTF-8 && \
    apt update && \
    apt install -y tzdata && \
    ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo ${TZ} > /etc/timezone && \
    dpkg-reconfigure --frontend noninteractive tzdata && \
    rm -rf /var/lib/apt/lists/*

RUN apt update && apt install -y build-essential gcc-multilib gdb && rm -rf /var/lib/apt/lists/*
RUN apt update && apt install -y tmux python3 python3-pip && rm -rf /var/lib/apt/lists/*
RUN apt update && apt install -y wget curl netcat socat 'iputils-*' traceroute tcpdump net-tools iproute2 iperf3 && rm -rf /var/lib/apt/lists/*

# attack testing tools
RUN apt update && apt install -y hping3 netwox netwox-doc && rm -rf /var/lib/apt/lists/*

# logging
RUN apt update && apt install -y syslog-ng && rm -rf /var/lib/apt/lists/*

FROM base AS user

RUN apt update && apt install -y lynx && rm -rf /var/lib/apt/lists/*

FROM php:8-apache AS server

RUN apt update && apt install -y build-essential gcc-multilib tmux python3 python3-pip gdb net-tools wget curl 'iputils-*' tcpdump iproute2 iperf3 traceroute && rm -rf /var/lib/apt/lists/*

FROM base AS yafw

RUN apt update && apt install -y nftables && rm -rf /var/lib/apt/lists/*

FROM frrouting/frr:v8.3.1 AS router
