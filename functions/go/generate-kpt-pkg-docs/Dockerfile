ARG BUILDER_IMAGE
ARG BASE_IMAGE

# Use the additional context as a build stage
FROM scratch AS list-setters
COPY . .

FROM --platform=$BUILDPLATFORM $BUILDER_IMAGE AS build
ENV CGO_ENABLED=0
WORKDIR /go/src/

COPY go.mod go.sum ./
# Always copy list-setters dir (even if empty or irrelevant)
COPY --from=list-setters / /go/list-setters/

RUN go mod download

COPY . .
ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /usr/local/bin/function ./

#############################################

FROM $BASE_IMAGE
COPY --from=build /usr/local/bin/function /usr/local/bin/function
ENTRYPOINT ["function"]
