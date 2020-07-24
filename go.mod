module github.com/jauninb/tektondoc

go 1.13

// Needed because of wrong reolution for k8s.io/client-go@latest
// https://github.com/kubernetes/client-go/issues/749
replace k8s.io/client-go => k8s.io/client-go v0.17.6

require (
	github.com/tektoncd/pipeline v0.14.2
	k8s.io/api v0.17.6
	k8s.io/apimachinery v0.17.6
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
)
