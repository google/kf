FROM debian:buster-slim

# Install golang from official image
COPY --from=golang:1.23 /usr/local/go/ /usr/local/go/
ENV PATH="$PATH:/usr/local/go/bin:/root/go/bin"
RUN echo "export PATH=$PATH:/usr/local/go/bin:/root/go/bin" >> ~/.bashrc

# Install dependencies
RUN apt-get update && apt-get install -y python3 python3-pip jq curl

# Install gcloud cli
RUN apt-get install -y apt-transport-https ca-certificates gnupg
RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
RUN curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key --keyring /usr/share/keyrings/cloud.google.gpg add -
RUN apt-get update && apt-get install -y google-cloud-cli

# Install ko
RUN go install github.com/google/ko@latest

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
  install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
