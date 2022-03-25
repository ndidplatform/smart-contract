FROM golang:1.17-alpine3.15 as builder

# Install tools and snappy lib
RUN apk update && apk add --no-cache --virtual .build-deps \
        g++ \
        gcc \
        make \
        git \
        openssl \
        snappy-dev \
        leveldb-dev \
        ca-certificates

WORKDIR /ndidplatform/smart-contract
COPY go.mod go.sum /ndidplatform/smart-contract/
COPY COPYING /ndidplatform/smart-contract/
COPY abci /ndidplatform/smart-contract/abci
COPY protos /ndidplatform/smart-contract/protos
COPY .git /ndidplatform/smart-contract/.git

# COPY patches /ndidplatform/smart-contract/patches

# Cannot be done with Go modules?
# RUN git apply /ndidplatform/smart-contract/patches/tm_goleveldb_bloom_filter.patch && \
#     git apply /ndidplatform/smart-contract/patches/tm_cleveldb_cache_and_bloom_filter.patch

ENV CGO_ENABLED=1
ENV CGO_LDFLAGS="-lsnappy"
RUN go build \
    -ldflags "-X github.com/ndidplatform/smart-contract/v7/abci/version.GitCommit=`git rev-parse --short=8 HEAD`" \
    -tags "cleveldb" \
    -o ./did-tendermint \
    ./abci


FROM alpine:3.15
LABEL maintainer="NDID IT Team <it@ndid.co.th>"

# Data directory for ABCI to store state.
ENV ABCI_DB_DIR_PATH=/DID

# The /tendermint/data dir is used by tendermint to store state.
ENV TMHOME /tendermint

# Set umask to 027
RUN umask 027 && echo "umask 0027" >> /etc/profile

COPY --from=builder /var/cache/apk /var/cache/apk

# Install snappy lib used by LevelDB.
# Install bash shell for convenience.
RUN apk add --no-cache \
      bash \
      leveldb \
      snappy \
      tzdata && \
    rm -rf /var/cache/apk

COPY --from=builder /ndidplatform/smart-contract/did-tendermint /usr/bin/did-tendermint
COPY docker/docker-entrypoint.sh /usr/bin/docker-entrypoint.sh

ENTRYPOINT [ "/usr/bin/docker-entrypoint.sh", "did-tendermint" ]
CMD [ "node" ]
