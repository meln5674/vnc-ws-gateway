ARG PROXY_CACHE=
ARG BASE_IMAGE=docker.io/library/ubuntu:noble
FROM ${PROXY_CACHE}${BASE_IMAGE}

RUN apt-get update && apt-get install -y --no-install-recommends xfce4 xfce4-goodies tigervnc-standalone-server dbus-x11 tigervnc-tools

COPY bin/vnc-ws-gateway.static /usr/bin/vnc-ws-gateway

# Set this env to the password to use for VNC connections
# ENV VNC_WS_GATEWAY_PASSWORD=''

EXPOSE 8080
ENTRYPOINT ["/usr/bin/vnc-ws-gateway"]
CMD ["--listen", "0.0.0.0:8080"]


