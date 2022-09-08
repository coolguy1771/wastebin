ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM
ARG TARGETVARIANT=""

# BUILD FRONTEND
FROM node:18 as frontend

WORKDIR /build

COPY web .

RUN yarn && yarn build
RUN ls -la  && cd build && ls -la && cat 200.html

# BUILD BACKEND
FROM golang:1.19.0-alpine as backend

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOARM=${TARGETVARIANT}



RUN apk add --no-cache ca-certificates tini-static gcc musl-dev \
    && update-ca-certificates


WORKDIR /build

COPY . .

RUN go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o wastebin /build/cmd/wastebin/.

# RUN
FROM gcr.io/distroless/static:nonroot


USER nonroot:nonroot

COPY --from=frontend /build/build /web
COPY --from=backend --chown=nonroot:nonroot /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend --chown=nonroot:nonroot /build/wastebin /wastebin
COPY --from=backend --chown=nonroot:nonroot /sbin/tini-static /tini

ENTRYPOINT [ "/tini", "--", "/wastebin" ]
LABEL \
    org.opencontainers.image.title="wastebin" \
    org.opencontainers.image.source="https://github.com/coolguy1771/wastebin"
