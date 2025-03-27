ARG PROXY_CACHE=
ARG GOLANG_VERSION=1.24
ARG GOLANG_IMAGE=docker.io/library/golang:${GOLANG_VERSION}

FROM ${PROXY_CACHE}${GOLANG_IMAGE}

RUN apt-get update && apt-get install -y xz-utils

ARG NODEJS_VERSION=v22.14.0
ARG NODEJS_URL=https://nodejs.org/dist/${NODEJS_VERSION}/node-${NODEJS_VERSION}-linux-x64.tar.xz

RUN mkdir -p /usr/local/node \
 && curl -fL "${NODEJS_URL}" | tar -xJ -C /usr/local/node --strip-components=1 

ENV PATH=${PATH}:/usr/local/node/bin
