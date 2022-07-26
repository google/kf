# builder to build mapfs
FROM golang:1.18.2-bullseye as builder

WORKDIR /app
COPY src/go.* ./
RUN go mod download

COPY src/. ./
RUN go build -v -o mapfs_binary


# Dockerfile for fuse sidecar
FROM us.gcr.io/k8s-artifacts-prod/build-image/debian-base-amd64:bullseye-v1.3.0

RUN apt-get update && apt-get install -y fuse
# Add vcap user. Mounting command need to be executed by the vcap user to ensure permissions are right.
RUN addgroup vcap --gid 2000 || true
RUN adduser \
    --disabled-password \
    --gecos "" \
    --ingroup vcap \
    --no-create-home \
    --uid 2000 \
    vcap

COPY --from=builder --chown=root:vcap /app/mapfs_binary /bin/mapfs

# need this to allow users other than root to use fuse.
RUN echo "user_allow_other" >> /etc/fuse.conf
RUN chmod 644 /etc/fuse.conf

RUN chmod 750 /bin/mapfs
# so that whoever runs this has the privileges of the owner(root).
RUN chmod u+s /bin/mapfs

USER vcap:vcap