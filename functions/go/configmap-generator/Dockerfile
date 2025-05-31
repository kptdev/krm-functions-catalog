FROM --platform=$BUILDPLATFORM golang:1.17-alpine3.15 AS build
ENV CGO_ENABLED=0
WORKDIR /go/src/

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /usr/local/bin/function ./

#############################################

FROM alpine:3.15
COPY --from=build /usr/local/bin/function /usr/local/bin/function
ENTRYPOINT ["function"]
