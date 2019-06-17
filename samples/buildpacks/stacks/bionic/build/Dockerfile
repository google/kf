ARG base_image
ARG go_version=latest

FROM golang:${go_version} as helpers

RUN go get github.com/GoogleCloudPlatform/docker-credential-gcr
RUN go get github.com/awslabs/amazon-ecr-credential-helper/ecr-login/cli/docker-credential-ecr-login
RUN go get github.com/Azure/acr-docker-credential-helper/src/docker-credential-acr

FROM ${base_image}

RUN apt-get update && \
  apt-get install -y git wget && \
  rm -rf /var/lib/apt/lists/*

COPY --from=helpers /go/bin /usr/local/bin
ARG stack_id="io.buildpacks.stacks.bionic"

USER cnb
LABEL io.buildpacks.stack.id=${stack_id}
ENV CNB_STACK_ID=${stack_id}
