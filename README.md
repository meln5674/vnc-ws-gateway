# VNC Websocket Gateway

This tool is a lightweight virtual desktop solution intended for use in containers or MicroVM's. 
It runs a web server with a single page, each visit to which will spawn a new TigerVNC server
and tunnel a connection to that server over a websocket, which is displayed using a
[NoVNC](https://github.com/novnc/noVNC) display. The server (and its corresponding session)
will be destroyed when the page is closed or the connection is lost.

That's it.

WARNING: This tool does not implement any authentication beyond the static VNC password, nor any
authorization. It will run a session as the user the server is started as, which may be root.
You should proxy this service behind an authenticating proxy like
[OAuth2 Proxy](https://github.com/oauth2-proxy/oauth2-proxy).

## Running

```bash
export VNC_wS_GATEWAY_PASSWORD=super-secure-password-goes-here
vnc-ws-gateway [--listen addr:port] [--vnc-args -extra,args,-to,the,-vncserver,...]
```

## Building from Source

Building this tool requires Golang and NPM. A pre-made environment for this task can be created by Running

```bash
./build-env/run.sh [command...]
```

To build run `make bin/vnc-ws-gateway` . To compile statically (e.g. for see below), instead run `make bin/vnc-ws-gateway.static`

## Example Dockerfile

An example [Dockerfile](./Dockerfile) is provided which demonstrates the tool in an Ubuntu LTS environment with XFCE.
