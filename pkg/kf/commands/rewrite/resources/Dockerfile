# Dockerfile that mimics Cloud Foundry's build process. Use the following
# command to execute it:
#
#   DOCKER_BUILDKIT=0 docker build --progress=plain --network=host -f {{.DOCKERFILE_NAME}} .
#
# This file was auto-generated. DO NOT EDIT.


###############################################################################
# Utilities
###############################################################################
FROM {{.UTILITIES_IMAGE}} AS buildutils
WORKDIR /workspace
RUN ["/ko-app/installer"]

###############################################################################
# Launcher
###############################################################################
FROM golang:latest AS launcher
WORKDIR /src
COPY launchershim/ .
RUN go mod init launcher.kf.dev
RUN go mod tidy
RUN go build -o /launcher main.go 

###############################################################################
# Build
###############################################################################
FROM {{.BUILDER_IMAGE}} AS build

# Copy utiltiies
WORKDIR /workspace
COPY --from=buildutils /workspace /workspace

# Copy source code
# TODO: generate a .dockerignore
COPY {{.LOCAL_SOURCE}} /staging/app
COPY {{.LOCAL_SOURCE}} /tmp/app

RUN CF_STACK={{.CF_STACK}} /workspace/builder \
  -buildArtifactsCacheDir=/tmp/cache \
  -buildDir=/tmp/app \
  -buildpacksDir=/tmp/buildpacks \
  -outputBuildArtifactsCache=/tmp/output-cache \
  -outputDroplet=/tmp/droplet \
  -outputMetadata=/tmp/result.json \
  "-buildpackOrder={{.BUILDPACKS}}" \
  "-skipDetect={{.SKIP_DETECT}}"

###############################################################################
# Application image
###############################################################################
FROM {{.RUN_IMAGE}}

# Set up the lifecycle that would be provided by Diego in CF.
WORKDIR /lifecycle
COPY --from=launcher /launcher /lifecycle/launcher
RUN chmod a+rx /lifecycle/launcher

# Set up the application and code.
WORKDIR /home/vcap
USER vcap:vcap
COPY --from=build /tmp/droplet droplet.tar.gz
RUN tar -xzf droplet.tar.gz && rm droplet.tar.gz
ENTRYPOINT ["/lifecycle/launcher"]
