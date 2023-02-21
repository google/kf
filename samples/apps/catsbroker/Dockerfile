FROM gcr.io/kf-releases/installer-d148684b3032e4386ff76c190d42c7d0:latest AS buildutils
WORKDIR /workspace
RUN ["/ko-app/installer"]

FROM cloudfoundry/cflinuxfs3@sha256:5219e9e30000e43e5da17906581127b38fa6417f297f522e332a801e737928f5 AS build

# copy utiltiies
WORKDIR /workspace
COPY --from=buildutils /workspace /workspace

# copy in source
# TODO: generate a .dockerignore
COPY . /staging/app
COPY . /tmp/app

RUN CF_STACK=cflinuxfs3 /workspace/builder \
  -buildArtifactsCacheDir=/tmp/cache \
  -buildDir=/tmp/app \
  -buildpacksDir=/tmp/buildpacks \
  -outputBuildArtifactsCache=/tmp/output-cache \
  -outputDroplet=/tmp/droplet \
  -outputMetadata=/tmp/result.json \
  "-buildpackOrder=https://github.com/cloudfoundry/go-buildpack" \
  "-skipDetect=false"

# Application image
FROM cloudfoundry/cflinuxfs3@sha256:5219e9e30000e43e5da17906581127b38fa6417f297f522e332a801e737928f5

# Set up the lifecycle that would be provided by
# Diego in CF.
WORKDIR /lifecycle
#COPY entrypoint.bash /lifecycle/entrypoint.bash
#COPY --from=buildutils /workspace/launcher /lifecycle/launcher
COPY containerlauncher /lifecycle/launcher
RUN chmod a+rx /lifecycle/launcher

# Set up the application and code.
WORKDIR /home/vcap
USER vcap:vcap
COPY --from=build /tmp/droplet droplet.tar.gz
RUN tar -xzf droplet.tar.gz && rm droplet.tar.gz
ENTRYPOINT ["/lifecycle/launcher"]

#  DOCKER_BUILDKIT=0 docker build --progress=plain --network=host .