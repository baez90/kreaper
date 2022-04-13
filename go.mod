module github.com/baez90/kreaper

go 1.18

require (
k8s.io/client-go v0.23.5
	k8s.io/api v0.23.5 // indirect
	k8s.io/apimachinery v0.23.5 // indirect
	sigs.k8s.io/controller-runtime v0.11.2
)

replace (
	k8s.io/api => k8s.io/api v0.23.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.23.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.23.1
	k8s.io/client-go => k8s.io/client-go v0.23.1
	k8s.io/component-base => k8s.io/component-base v0.23.1
)