# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Build all Config Sync go binaries
FROM golang:1.19 as bins

WORKDIR /workspace

COPY . .

# Version string to embed in built binary.
ARG VERSION
ARG HELM_INFLATOR_FUNCTION_VERSION=v0.3.0

ARG HELM_VERSION=v3.11.3
ARG KUSTOMIZE_VERSION=v5.0.1

# Install Helm with license
RUN URL="https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz" && \
  URL_PREFIX="$(dirname "${URL}")" && FILENAME="$(basename "${URL}")" && \
  wget "${URL}" -O "/tmp/${FILENAME}" && \
  wget "${URL}.sha256" -O /tmp/helm_checksum.txt && \
  echo "$(cat /tmp/helm_checksum.txt)  /tmp/${FILENAME}" | sha256sum --check && \
  tar -zxvf "/tmp/${FILENAME}" -C /tmp && \
  mv /tmp/linux-amd64/helm /usr/local/bin/helm && \
  mkdir -p ./vendor/helm.sh/helm/v3 && \
  mv /tmp/linux-amd64/LICENSE ./vendor/helm.sh/helm/v3/LICENSE && \
  rm -rf /tmp/linux-amd64 "/tmp/${FILENAME}" /tmp/helm_checksum.txt

# Install Kustomize with license
RUN URL="https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_linux_amd64.tar.gz" && \
  URL_PREFIX="$(dirname "${URL}")" && FILENAME="$(basename "${URL}")" && \
  wget "${URL}" -O "/tmp/${FILENAME}" && \
  wget "${URL_PREFIX}/checksums.txt" -O /tmp/kustomize_checksums.txt && \
  echo "$(grep "${FILENAME}" /tmp/kustomize_checksums.txt | cut -d ' ' -f 1)  /tmp/${FILENAME}" | sha256sum --check && \
  tar -zxvf "/tmp/${FILENAME}" -C /tmp && \
  mv /tmp/kustomize /usr/local/bin/kustomize && \
  rm "/tmp/${FILENAME}" /tmp/kustomize_checksums.txt && \
  mkdir -p ./vendor/sigs.k8s.io/kustomize && \
  wget "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/kustomize/${KUSTOMIZE_VERSION}/LICENSE" -O ./vendor/sigs.k8s.io/kustomize/LICENSE

# Install the render-helm-chart function.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
  go install github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/render-helm-chart@${HELM_INFLATOR_FUNCTION_VERSION}

# Build all our stuff.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
  go install \
    -mod=vendor \
    -ldflags "-X kpt.dev/configsync/pkg/version.VERSION=${VERSION}" \
    ./cmd/nomos \
    ./cmd/reconciler \
    ./cmd/reconciler-manager \
    ./cmd/hydration-controller \
    ./cmd/admission-webhook \
    ./cmd/oci-sync \
    ./cmd/helm-sync

# Concatenate vendored licenses into LICENSES.txt
# Built in the container to include binary licenses (helm & kustomize)
RUN scripts/prepare-licenses.sh

# Debian non-root base image
# Uses the same nonroot UID as distroless
FROM gcr.io/gke-release/debian-base:bullseye-v1.4.3-gke.0 as debian-nonroot
WORKDIR /
ARG USERNAME=nonroot
ARG USER_UID=65532
ARG USER_GID=$USER_UID
RUN groupadd --gid $USER_GID $USERNAME && \
  useradd --uid $USER_UID --gid $USER_GID -m $USERNAME
USER nonroot:nonroot

# Hydration controller image
FROM gcr.io/distroless/static:nonroot as hydration-controller
WORKDIR /
COPY --from=bins /go/bin/hydration-controller .
COPY --from=bins /go/bin/render-helm-chart /usr/local/bin/render-helm-chart
COPY --from=bins /usr/local/bin/helm /usr/local/bin/helm
COPY --from=bins /usr/local/bin/kustomize /usr/local/bin/kustomize
COPY --from=bins /workspace/LICENSE LICENSE
COPY --from=bins /workspace/LICENSES.txt LICENSES.txt
USER nonroot:nonroot
ENTRYPOINT ["/hydration-controller"]

# OCI-sync image
FROM gcr.io/distroless/static:latest as oci-sync
# Setting HOME ensures that whatever UID this ultimately runs as can write files.
ENV HOME=/tmp
WORKDIR /
COPY --from=bins /go/bin/oci-sync .
COPY --from=bins /workspace/LICENSE LICENSE
COPY --from=bins /workspace/LICENSES.txt LICENSES.txt
USER nonroot:nonroot
ENTRYPOINT ["/oci-sync"]

# Helm-sync image
FROM gcr.io/distroless/static:latest as helm-sync
# Setting HOME ensures that whatever UID this ultimately runs as can write files.
ENV HOME=/tmp
WORKDIR /
COPY --from=bins /go/bin/helm-sync .
COPY --from=bins /usr/local/bin/helm /usr/local/bin/helm
COPY --from=bins /workspace/LICENSE LICENSE
COPY --from=bins /workspace/LICENSES.txt LICENSES.txt
USER nonroot:nonroot
ENTRYPOINT ["/helm-sync"]

# Hydration controller image with shell
FROM debian-nonroot as hydration-controller-with-shell
WORKDIR /
USER root
COPY --from=bins /go/bin/hydration-controller .
COPY --from=bins /go/bin/render-helm-chart /usr/local/bin/render-helm-chart
COPY --from=bins /usr/local/bin/helm /usr/local/bin/helm
COPY --from=bins /usr/local/bin/kustomize /usr/local/bin/kustomize
COPY --from=bins /workspace/LICENSE LICENSE
COPY --from=bins /workspace/LICENSES.txt LICENSES.txt
RUN apt-get update && apt-get install -y git
USER nonroot:nonroot
ENTRYPOINT ["/hydration-controller"]

# Reconciler image
FROM gcr.io/distroless/static:nonroot as reconciler
WORKDIR /
COPY --from=bins /go/bin/reconciler .
COPY --from=bins /workspace/LICENSE LICENSE
COPY --from=bins /workspace/LICENSES.txt LICENSES.txt
USER nonroot:nonroot
ENTRYPOINT ["/reconciler"]

# Reconciler Manager image
FROM gcr.io/distroless/static:nonroot as reconciler-manager
WORKDIR /
COPY --from=bins /go/bin/reconciler-manager reconciler-manager
COPY --from=bins /workspace/LICENSE LICENSE
COPY --from=bins /workspace/LICENSES.txt LICENSES.txt
USER nonroot:nonroot
ENTRYPOINT ["/reconciler-manager"]

# Admission Webhook image
FROM gcr.io/distroless/static:nonroot as admission-webhook
WORKDIR /
COPY --from=bins /go/bin/admission-webhook admission-webhook
COPY --from=bins /workspace/LICENSE LICENSE
COPY --from=bins /workspace/LICENSES.txt LICENSES.txt
USER nonroot:nonroot
ENTRYPOINT ["/admission-webhook"]

# Nomos image
# Not used by Config Sync backend components. Intended for use cases with the
# nomos CLI (e.g. containerized CI/CD)
FROM debian-nonroot as nomos
USER root

# https://github.com/GoogleCloudPlatform/google-cloud-go/issues/791#issuecomment-353689746
RUN apt-get update && apt-get install -y bash git

# Install nomos CLI
RUN mkdir -p /opt/nomos/bin
WORKDIR /opt/nomos/bin
COPY --from=bins /go/bin/nomos nomos
COPY --from=bins /workspace/LICENSE LICENSE
COPY --from=bins /workspace/LICENSES.txt LICENSES.txt

# Set up a HOME directory for non-root user
RUN mkdir -p /nomos && chown nonroot:nonroot /nomos
USER nonroot:nonroot
ENV HOME="/nomos"
ENV PATH="/opt/nomos/bin:${PATH}"
ENTRYPOINT ["/opt/nomos/bin/nomos"]
