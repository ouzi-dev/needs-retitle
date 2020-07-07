module github.com/ouzi-dev/needs-retitle

go 1.14

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	k8s.io/client-go => k8s.io/client-go v0.17.3
)

require (
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/shurcooL/githubv4 v0.0.0-20200627185320-e003124d66e4
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1 // indirect
	golang.org/x/exp v0.0.0-20191030013958-a1ab85dbe136 // indirect
	golang.org/x/mod v0.3.0 // indirect
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/tools v0.0.0-20200702044944-0cc1aa72b347 // indirect
	google.golang.org/genproto v0.0.0-20191108220845-16a3f7862a1a // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	gotest.tools v2.2.0+incompatible
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
	k8s.io/test-infra v0.0.0-20200702033203-1e71b3526aef
	sigs.k8s.io/yaml v1.2.0
)
