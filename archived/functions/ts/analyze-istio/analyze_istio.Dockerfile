ARG BUILDER_IMAGE
ARG BASE_IMAGE


FROM --platform=$BUILDPLATFORM $BUILDER_IMAGE as builder

RUN apk add bash curl git && apk update

ARG TARGETOS TARGETARCH
ARG ISTIOCTL_VERSION="1.6.5"
RUN curl -fsSL -o /istio-${ISTIOCTL_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz https://github.com/istio/istio/releases/download/${ISTIOCTL_VERSION}/istio-${ISTIOCTL_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz && \
    tar -zxvf /istio-${ISTIOCTL_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz && \
    mv /istio-${ISTIOCTL_VERSION}/bin/istioctl /usr/local/bin/istioctl

RUN mkdir -p /home/node/app && \
    chown -R node:node /home/node/app

USER node

WORKDIR /home/node/app

# Install dependencies and cache them.
COPY --chown=node:node package*.json ./
RUN npm ci --ignore-scripts

# Build the source.
COPY --chown=node:node tsconfig.json .
COPY --chown=node:node src src
RUN npm run build && \
    npm prune --production && \
    rm -r src tsconfig.json

#############################################

FROM $BASE_IMAGE

# Run as non-root user as a best-practices:
# https://github.com/nodejs/docker-node/blob/master/docs/BestPractices.md
USER node

WORKDIR /home/node/app

COPY --from=builder /home/node/app /home/node/app

COPY --from=builder /usr/local/bin /usr/local/bin

ENV PATH /usr/local/bin:$PATH

ENTRYPOINT ["node", "/home/node/app/dist/istioctl_analyze_run.js"]
