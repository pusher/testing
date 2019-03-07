FROM golang:1.12

RUN apt-get install -y --no-install-recommends ca-certificates git
RUN go get -u github.com/golang/dep/cmd/dep github.com/onsi/ginkgo/ginkgo github.com/alecthomas/gometalinter
RUN gometalinter --install

ENV TEST_ASSET_DIR=/usr/local/bin
ENV TEST_ASSET_KUBECTL=$TEST_ASSET_DIR/kubectl
ENV TEST_ASSET_KUBE_APISERVER=$TEST_ASSET_DIR/kube-apiserver
ENV TEST_ASSET_ETCD=$TEST_ASSET_DIR/etcd
ENV TEST_ASSET_URL=https://storage.googleapis.com/k8s-c10s-test-binaries

RUN curl $TEST_ASSET_URL/etcd-Linux-x86_64 --output $TEST_ASSET_ETCD
RUN curl $TEST_ASSET_URL/kube-apiserver-Linux-x86_64 --output $TEST_ASSET_KUBE_APISERVER
RUN curl https://storage.googleapis.com/kubernetes-release/release/v1.12.0/bin/linux/amd64/kubectl --output $TEST_ASSET_KUBECTL

RUN chmod +x $TEST_ASSET_ETCD
RUN chmod +x $TEST_ASSET_KUBE_APISERVER
RUN chmod +x $TEST_ASSET_KUBECTL

CMD ["/bin/sh"]
