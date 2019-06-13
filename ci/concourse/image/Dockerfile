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

FROM golang:1.12-alpine

RUN apk add python3 curl jq
RUN curl https://bootstrap.pypa.io/get-pip.py | python3

RUN curl -LOJ https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-242.0.0-linux-x86_64.tar.gz \
 && tar xfz google-cloud-sdk-242.0.0-linux-x86_64.tar.gz -C / \
 && /google-cloud-sdk/install.sh
ENV PATH="/google-cloud-sdk/bin:${PATH}"

ARG k8s_version=1.14.1
RUN curl -LOJ https://storage.googleapis.com/kubernetes-release/release/v${k8s_version}/bin/linux/amd64/kubectl \
 && mv kubectl /usr/local/bin \
 && chmod +x /usr/local/bin/kubectl

RUN pip install yamllint yq

RUN gcloud components install docker-credential-gcr
