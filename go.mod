module github.com/ouzi-dev/needs-retitle

go 1.14

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	k8s.io/client-go => k8s.io/client-go v0.17.3
)

require (
	github.com/shurcooL/githubv4 v0.0.0-20200627185320-e003124d66e4
	github.com/sirupsen/logrus v1.6.0
	k8s.io/test-infra v0.0.0-20200702033203-1e71b3526aef
	sigs.k8s.io/yaml v1.2.0
)
