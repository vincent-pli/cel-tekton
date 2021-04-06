module github.com/vincentpli/cel-tekton

go 1.15

require (
	github.com/tektoncd/pipeline v0.22.0
	go.uber.org/zap v1.16.0
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v0.19.7
	k8s.io/code-generator v0.19.7
	k8s.io/kube-openapi v0.0.0-20210113233702-8566a335510f
	knative.dev/hack v0.0.0-20210325223819-b6ab329907d3
	knative.dev/hack/schema v0.0.0-20210325223819-b6ab329907d3
	knative.dev/pkg v0.0.0-20210331065221-952fdd90dbb0
)

replace knative.dev/pkg => knative.dev/pkg v0.0.0-20210331065221-952fdd90dbb0
