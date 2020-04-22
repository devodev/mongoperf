FROM ubuntu:14.04

ENV HOME /root

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        g++ \
        gcc \
        libc6-dev \
        make \
        curl \
        git \
        mercurial \
    && rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.13.10
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 8a4cbc9f2b95d114c38f6cbe94a45372d48c604b707db2057c787398dfbf8e7f

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
    && echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
    && tar -C /usr/local -xzf golang.tar.gz \
    && rm golang.tar.gz

ENV GOPATH /srv
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /srv/src/github.com/devodev/mongoperf
