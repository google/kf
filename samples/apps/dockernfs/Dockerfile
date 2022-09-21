# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.11-alpine AS build

COPY . src
WORKDIR src
RUN GO111MODULE=on CGO_ENABLED=0 go build -o /bin/hello

FROM gcr.io/kf-releases/fusesidecar:v2.11.2 AS builder

FROM cloudfoundry/cflinuxfs3
COPY --from=build /bin/hello /bin/hello
COPY --from=builder --chown=root:vcap /bin/mapfs /bin/mapfs

# Allow users other than root to use fuse.
RUN echo "user_allow_other" >> /etc/fuse.conf
RUN chmod 644 /etc/fuse.conf

RUN chmod 750 /bin/mapfs
# Allow setuid so the mapfs binary is run as root.
RUN chmod u+s /bin/mapfs


# Setting the PORT and EXPOSE allows the app to be run as a normal Docker
# container outside kf, you don't need them if you're building images only for
# kf.
ENV PORT 8080
EXPOSE 8080/tcp

ENTRYPOINT ["/bin/hello"]
