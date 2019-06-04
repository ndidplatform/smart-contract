FROM golang:1.12-alpine3.9 as builder

# This is the release of dep to pull in.
ENV DEP_VERSION v0.4.1
ENV DEP_SHA256SUM 31144e465e52ffbc0035248a10ddea61a09bf28b00784fd3fdd9882c8cbb2315

# Install tools and snappy lib
RUN apk update && apk add --no-cache --virtual .build-deps \
        g++ \
        gcc \
        make \
        git \
        openssl \
        snappy-dev

# Install LevelDB
WORKDIR /
RUN wget -q https://github.com/google/leveldb/archive/v1.20.tar.gz && \
    tar -zxvf v1.20.tar.gz && \
    cd leveldb-1.20/ && \
    make
RUN cp -r /leveldb-1.20/out-static/lib* /leveldb-1.20/out-shared/lib* /usr/local/lib/ && \
    mkdir -p /usr/local/include/leveldb && \
    cp -r /leveldb-1.20/include/leveldb /usr/local/include/ && \
    ldconfig /usr/local/lib && \
    rm -f /v1.20.tar.gz

# Install dep
RUN wget -qO /usr/local/bin/dep https://github.com/golang/dep/releases/download/${DEP_VERSION}/dep-linux-amd64 && \
    echo "${DEP_SHA256SUM}  /usr/local/bin/dep" | sha256sum -c && \
    chmod +x /usr/local/bin/dep

COPY Gopkg.toml Gopkg.lock $GOPATH/src/github.com/ndidplatform/smart-contract/
WORKDIR $GOPATH/src/github.com/ndidplatform/smart-contract
RUN dep ensure -vendor-only -v

COPY abci $GOPATH/src/github.com/ndidplatform/smart-contract/abci
COPY protos $GOPATH/src/github.com/ndidplatform/smart-contract/protos
COPY .git $GOPATH/src/github.com/ndidplatform/smart-contract/.git
COPY patches $GOPATH/src/github.com/ndidplatform/smart-contract/patches

RUN git apply $GOPATH/src/github.com/ndidplatform/smart-contract/patches/tm_goleveldb_bloom_filter.patch && \
    git apply $GOPATH/src/github.com/ndidplatform/smart-contract/patches/tm_cleveldb_cache_and_bloom_filter.patch

WORKDIR $GOPATH/src/github.com/ndidplatform/smart-contract/abci
ENV CGO_ENABLED=1
ENV CGO_LDFLAGS="-lsnappy"
RUN go install \
    -ldflags "-X github.com/ndidplatform/smart-contract/abci/version.GitCommit=`git rev-parse --short=8 HEAD`" \
    -tags "cleveldb"


FROM alpine:3.9
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
RUN apk update && apk add --no-cache bash curl jq snappy-dev && rm -rf /var/cache/apk

# Copy compiled LevelDB lib
COPY --from=builder /usr/local/lib /usr/local/lib

COPY --from=builder /go/bin/abci /usr/bin/did-tendermint
RUN mkdir -p ${TMHOME} ${ABCI_DB_DIR_PATH}

# Change owner to nobodoy:nogroup and permission to 640
RUN chown -R nobody:nogroup /usr/bin/did-tendermint ${TMHOME} ${ABCI_DB_DIR_PATH}
RUN chmod -R 740 /usr/bin/did-tendermint ${TMHOME} ${ABCI_DB_DIR_PATH}

USER nobody
ENTRYPOINT ["did-tendermint"]
CMD ["node"]
STOPSIGNAL SIGTERM