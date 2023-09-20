module github.com/pusher/testing

go 1.12

require (
	github.com/gophercloud/gophercloud v0.4.0 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/sirupsen/logrus v1.4.2
	k8s.io/apiserver v0.15.10
	k8s.io/klog v1.0.0 // indirect
	k8s.io/test-infra v0.0.0-20190918113529-2f64091d118a
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
