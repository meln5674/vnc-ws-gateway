#!/bin/bash -xe

DOCKER=${DOCKER:-docker}

SCRIPTPATH="$(
  cd -- "$(dirname "$0")" >/dev/null 2>&1
  pwd -P
)"

# Assume the user wants an interactive shell if no args are passed, otherwise, run the provided command
if [ "$#" -eq 0 ]; then
  COMMAND=(bash)
else
  COMMAND=("$@")
fi

export DOCKER_BUILDKIT=1

IMAGE_REPO=${IMAGE_REPO:-localhost/meln5674/ankyra/build-env}

IMAGE_TAG=${IMAGE_TAG:-$(md5sum build-env/Dockerfile | awk '{ print $1 }')}

IMAGE="${IMAGE_REPO}:${IMAGE_TAG}"

if [ -z "${NO_BUILD_IMAGE}" ]; then
  ${DOCKER} build -f "${SCRIPTPATH}/Dockerfile" -t "${IMAGE}" "${SCRIPTPATH}"
fi

DOCKER_RUN_ARGS=(
  --rm
  -it
  -e IS_CI
  -e GITHUB_TOKEN
  -e HTTP_PROXY
  -e HTTPS_PROXY
  -e NO_PROXY
  -e http_proxy
  -e https_proxy
  -e no_proxy
  -e ANKYRA_E2E_DEV_MODE
  -e ANKYRA_E2E_TEST_NO_CLEANUP
  -e KIND_CLUSTER_NAME
)

if [ -t 0 ]; then
  DOCKER_RUN_ARGS+=(-t)
fi

# Make it look like we're in the same directory as we ran from
DOCKER_RUN_ARGS+=(
  -v "/${SCRIPTPATH}:/${SCRIPTPATH}"
  -w "/${PWD}"
)

# Use the same home directory
DOCKER_RUN_ARGS+=(
  -e HOME
  -v "/${HOME}:/${HOME}"
)

# If we're running as (rootful) docker, we need setup to run as the same user
if [ "${DOCKER}" == docker ]; then
  # Make it look like we're the same user
  DOCKER_RUN_ARGS+=(
    -u "$(id -u):$(id -g)"
    -v /etc/passwd:/etc/passwd:ro
    -v /etc/group:/etc/group:ro
    -v /etc/subuid:/etc/subuid:ro
    -v /etc/subgid:/etc/subgid:ro
    -v /lib/modules:/lib/modules:ro
    --privileged
  )

  for group in $(id -G); do
    DOCKER_RUN_ARGS+=(--group-add "${group}")
  done

  # Provide access to docker
  DOCKER_RUN_ARGS+=(
    -v /var/run/docker.sock:/var/run/docker.sock
    --security-opt seccomp=unconfined
  )
else
  DOCKER_RUN_ARGS+=(
    --security-opt label=disable
    --device /dev/fuse
    --device /dev/net/tun
    --userns host
  )
fi

# Provide access to an existing kind cluster, as well as enable port-forwarding for live dev env
DOCKER_RUN_ARGS+=(
  -e KUBECONFIG
  --network host
)

if [ -z "${GOPATH}" ]; then
  if command -v go; then
    export GOPATH=${GOPATH:-$(go env GOPATH)}
  else
    export GOPATH=${GOPATH:-$HOME/go}
  fi
fi

# If this doesn't exist on the first docker run, it gets owned by root, which breaks permissions
mkdir -p "${GOPATH}"

DOCKER_RUN_ARGS+=(
  -e GOPATH
  -e GOPRIVATE
  -e GONOSUMDB
  -v "/${GOPATH}:/${GOPATH}"
)

DOCKER_RUN_ARGS+=(
  # -e KIND_EXPERIMENTAL_PROVIDER=podman
  -v /run/user/$(id -u):/run/user/$(id -u)
)

DOCKER_RUN_ARGS+=(${DOCKER_RUN_EXTRA_ARGS})

exec ${DOCKER} run "${DOCKER_RUN_ARGS[@]}" "${IMAGE}" "${COMMAND[@]}"
