FROM golang

RUN apt update && apt install -y python3
RUN curl https://bootstrap.pypa.io/get-pip.py | python3

RUN curl -LOJ https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-242.0.0-linux-x86_64.tar.gz \
 && tar xfz google-cloud-sdk-242.0.0-linux-x86_64.tar.gz -C / \
 && /google-cloud-sdk/install.sh \
 && cd /usr/bin && ln -s /google-cloud-sdk/bin/gcloud gcloud

ARG k8s_version=1.14.1
RUN curl -LOJ https://storage.googleapis.com/kubernetes-release/release/v${k8s_version}/bin/linux/amd64/kubectl \
 && mv kubectl /usr/local/bin \
 && chmod +x /usr/local/bin/kubectl

RUN pip install yamllint yq
RUN apt install -y curl jq
