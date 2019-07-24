FROM golang:1.12-alpine3.10 as builder

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
    -ldflags "-X github.com/ndidplatform/smart-contract/v4/abci/version.GitCommit=`git rev-parse --short=8 HEAD`" \
    -tags "cleveldb" \
    -o ./did-tendermint \
    ./abci


FROM alpine:3.10
LABEL maintainer="NDID IT Team <it@ndid.co.th>"
ENV TERM=xterm-256color
ENV ABCI_DB_DIR_PATH=/DID

# Tendermint will be looking for genesis file in /tendermint (unless you change
# `genesis_file` in config.toml). You can put your config.toml and private
# validator file into /tendermint.
#
# The /tendermint/data dir is used by tendermint to store state.
ENV TMHOME /tendermint

# Set umask to 027
RUN umask 027 && echo "umask 0027" >> /etc/profile

COPY --from=builder /var/cache/apk /var/cache/apk

# jq and curl used for extracting `pub_key` from private validator while
# deploying tendermint with Kubernetes. It is nice to have bash so the users
# could execute bash commands.
# Install snappy lib used by LevelDB
RUN apk update && apk add --no-cache bash curl snappy leveldb && rm -rf /var/cache/apk

COPY --from=builder /ndidplatform/smart-contract/did-tendermint /usr/bin/did-tendermint
RUN mkdir -p ${TMHOME} ${ABCI_DB_DIR_PATH}

# Change owner to nobodoy:nogroup and permission to 640
RUN chown -R nobody:nogroup /usr/bin/did-tendermint ${TMHOME} ${ABCI_DB_DIR_PATH}
RUN chmod -R 740 /usr/bin/did-tendermint ${TMHOME} ${ABCI_DB_DIR_PATH}

USER nobody
ENTRYPOINT ["did-tendermint"]
CMD ["node"]
STOPSIGNAL SIGTERM