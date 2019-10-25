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

FROM google/cloud-sdk

RUN apt update -y
RUN apt-get install -y curl software-properties-common git
RUN curl -sL https://deb.nodesource.com/setup_12.x | bash -
RUN apt-get install -y nodejs

# Docsy deps
RUN npm install -g --no-save autoprefixer
RUN npm install -g --no-save postcss-cli
RUN curl -LO https://github.com/gohugoio/hugo/releases/download/v0.56.3/hugo_extended_0.56.3_Linux-64bit.deb \
  && dpkg -i hugo*.deb
