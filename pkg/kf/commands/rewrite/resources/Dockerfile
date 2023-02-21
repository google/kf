###############################################################################
# Utilities
###############################################################################
FROM {{.UtilitiesImage}} AS buildutils
WORKDIR /workspace
RUN ["/ko-app/installer"]

###############################################################################
# Build
###############################################################################
FROM {{.BuildImage}} AS build

# Copy utiltiies
WORKDIR /workspace
COPY --from=buildutils /workspace /workspace

# Copy source code
# TODO: generate a .dockerignore
COPY {{.Source}} /staging/app
COPY {{.Source}} /tmp/app

RUN CF_STACK=cflinuxfs3 /workspace/builder \
  -buildArtifactsCacheDir=/tmp/cache \
  -buildDir=/tmp/app \
  -buildpacksDir=/tmp/buildpacks \
  -outputBuildArtifactsCache=/tmp/output-cache \
  -outputDroplet=/tmp/droplet \
  -outputMetadata=/tmp/result.json \
  "-buildpackOrder={{.Buildpacks}}" \
  "-skipDetect={{.SkipDetect}}"

###############################################################################
# Application image
###############################################################################
FROM {{.RunImage}}

# Set up the lifecycle that would be provided by Diego in CF.
WORKDIR /lifecycle
COPY entrypoint.bash /lifecycle/entrypoint.bash
#COPY --from=buildutils /workspace/launcher /lifecycle/launcher
COPY containerlauncher /lifecycle/launcher
RUN chmod a+rx /lifecycle/launcher

# Set up the application and code.
WORKDIR /home/vcap
USER vcap:vcap
COPY --from=build /tmp/droplet droplet.tar.gz
RUN tar -xzf droplet.tar.gz && rm droplet.tar.gz
ENTRYPOINT ["/lifecycle/launcher"]


# DOCKER_BUILDKIT=0 docker build --progress=plain --network=host .
