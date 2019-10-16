# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the License);
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an AS IS BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.12

RUN apt update && apt install -y \
  python3 \
  python3-distutils \
  curl \
  jq

RUN curl https://bootstrap.pypa.io/get-pip.py | python3
RUN pip install yamllint yq

RUN go get -u github.com/google/ko/cmd/ko

ARG GOOGLE_CLOUD_SDK_VERSION=266.0.0
ARG GOOGLE_CLOUD_SDK_CHECKSUM=e2b2cd5e49e1dc73ffe1d57ba2bcc1b76620ae9549d2aa27ece05d819a8f4bbc
ARG GOOGLE_CLOUD_SDK_FILE=google-cloud-sdk-${GOOGLE_CLOUD_SDK_VERSION}-linux-x86_64.tar.gz
RUN curl -LOJ https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/${GOOGLE_CLOUD_SDK_FILE} \
  && [ "${GOOGLE_CLOUD_SDK_CHECKSUM}" = "$(cat ${GOOGLE_CLOUD_SDK_FILE} | sha256sum | cut -d' ' -f1)" ] \
  && tar xfz ${GOOGLE_CLOUD_SDK_FILE} -C / \
  && /google-cloud-sdk/install.sh
ENV PATH="/google-cloud-sdk/bin:${PATH}"

ARG k8s_version=1.14.1
RUN curl -LOJ https://storage.googleapis.com/kubernetes-release/release/v${k8s_version}/bin/linux/amd64/kubectl \
  && mv kubectl /usr/local/bin \
  && chmod +x /usr/local/bin/kubectl

RUN gcloud components install docker-credential-gcr
