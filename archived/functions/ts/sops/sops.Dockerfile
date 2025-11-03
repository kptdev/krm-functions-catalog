ARG BUILDER_IMAGE
ARG BASE_IMAGE


FROM --platform=$BUILDPLATFORM $BUILDER_IMAGE as builder

RUN apk add bash wget git && apk update

ARG TARGETOS TARGETARCH
ARG SOPS_VERSION="v3.10.2"
RUN wget --retry-connrefused --waitretry=1 --read-timeout=20 --timeout=15 -t 3 \
         -O /usr/local/bin/sops \
         https://github.com/getsops/sops/releases/download/${SOPS_VERSION}/sops-${SOPS_VERSION}.${TARGETOS}.${TARGETARCH} && \
    chmod +x /usr/local/bin/sops

RUN mkdir -p /home/node/app && \
    chown -R node:node /home/node/app

USER node

WORKDIR /home/node/app

# Install dependencies and cache them.
COPY --chown=node:node package*.json ./
# Make rw package work
COPY --chown=node:node @types @types

RUN npm ci --ignore-scripts

# Copy the source.
COPY --chown=node:node tsconfig.json .
COPY --chown=node:node src src

# Build the source.
RUN npm run build && \
    npm prune --production && \
    rm -r src tsconfig.json

#############################################

FROM $BASE_IMAGE

RUN apk add git gnupg

# Run as non-root user as a best-practices:
# https://github.com/nodejs/docker-node/blob/master/docs/BestPractices.md
USER node

WORKDIR /home/node/app

COPY --from=builder /home/node/app /home/node/app

COPY --from=builder /usr/local/bin /usr/local/bin

ENV PATH /usr/local/bin:$PATH
ENV GNUPGHOME /tmp
ENV XDG_CONFIG_HOME /tmp
RUN mkdir -p -m 0777 $XDG_CONFIG_HOME/sops/age/

ENTRYPOINT ["node", "/home/node/app/dist/sops_run.js"]
