/*
Copyright 2021 Kontain
*/
// Code generated by informer-gen. DO NOT EDIT.

package kontain

import (
	internalinterfaces "pkg/generated/informers/externalversions/internalinterfaces"
	v1beta "pkg/generated/informers/externalversions/kontain.app/v1beta"
)

// Interface provides access to each of this group's versions.
type Interface interface {
	// V1beta provides access to shared informers for resources in V1beta.
	V1beta() v1beta.Interface
}

type group struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &group{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// V1beta returns a new v1beta.Interface.
func (g *group) V1beta() v1beta.Interface {
	return v1beta.New(g.factory, g.namespace, g.tweakListOptions)
}