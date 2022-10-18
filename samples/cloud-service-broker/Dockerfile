# Step 1: Build the Cloud Service Broker
FROM golang:1.18 as build

RUN git clone --single-branch --branch main https://github.com/cloudfoundry-incubator/cloud-service-broker.git /csb-src
WORKDIR /csb-src
RUN git checkout 1f4b0b8d51409ff040786949ef407a235c3120c5
RUN CGO_ENABLED=0 GOOS=linux go build -o cloud-service-broker

# Step 2: Build the brokerpak
WORKDIR /
RUN git clone --single-branch --branch main https://github.com/cloudfoundry-incubator/csb-brokerpak-gcp.git /brokerpak-src
WORKDIR /brokerpak-src
RUN git checkout 3a52bb28c7144b4eeda49aa024ac3fa839e1e30c
WORKDIR /
RUN /csb-src/cloud-service-broker pak build /brokerpak-src

# Step 3: Build the final container
FROM gcr.io/distroless/static:latest

COPY --from=build /csb-src/cloud-service-broker /bin/cloud-service-broker
COPY --from=build /gcp-services-*.brokerpak /bin

ENV PORT 8080
EXPOSE 8080/tcp

WORKDIR /bin
ENTRYPOINT ["/bin/cloud-service-broker"]
CMD ["help"]
