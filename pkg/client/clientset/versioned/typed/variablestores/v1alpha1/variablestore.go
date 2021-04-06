/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/vincentpli/cel-tekton/pkg/apis/variablestores/v1alpha1"
	scheme "github.com/vincentpli/cel-tekton/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// VariableStoresGetter has a method to return a VariableStoreInterface.
// A group's client should implement this interface.
type VariableStoresGetter interface {
	VariableStores(namespace string) VariableStoreInterface
}

// VariableStoreInterface has methods to work with VariableStore resources.
type VariableStoreInterface interface {
	Create(ctx context.Context, variableStore *v1alpha1.VariableStore, opts v1.CreateOptions) (*v1alpha1.VariableStore, error)
	Update(ctx context.Context, variableStore *v1alpha1.VariableStore, opts v1.UpdateOptions) (*v1alpha1.VariableStore, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.VariableStore, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.VariableStoreList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.VariableStore, err error)
	VariableStoreExpansion
}

// variableStores implements VariableStoreInterface
type variableStores struct {
	client rest.Interface
	ns     string
}

// newVariableStores returns a VariableStores
func newVariableStores(c *SamplesV1alpha1Client, namespace string) *variableStores {
	return &variableStores{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the variableStore, and returns the corresponding variableStore object, and an error if there is any.
func (c *variableStores) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.VariableStore, err error) {
	result = &v1alpha1.VariableStore{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("variablestores").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of VariableStores that match those selectors.
func (c *variableStores) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.VariableStoreList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.VariableStoreList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("variablestores").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested variableStores.
func (c *variableStores) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("variablestores").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a variableStore and creates it.  Returns the server's representation of the variableStore, and an error, if there is any.
func (c *variableStores) Create(ctx context.Context, variableStore *v1alpha1.VariableStore, opts v1.CreateOptions) (result *v1alpha1.VariableStore, err error) {
	result = &v1alpha1.VariableStore{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("variablestores").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(variableStore).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a variableStore and updates it. Returns the server's representation of the variableStore, and an error, if there is any.
func (c *variableStores) Update(ctx context.Context, variableStore *v1alpha1.VariableStore, opts v1.UpdateOptions) (result *v1alpha1.VariableStore, err error) {
	result = &v1alpha1.VariableStore{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("variablestores").
		Name(variableStore.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(variableStore).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the variableStore and deletes it. Returns an error if one occurs.
func (c *variableStores) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("variablestores").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *variableStores) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("variablestores").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched variableStore.
func (c *variableStores) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.VariableStore, err error) {
	result = &v1alpha1.VariableStore{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("variablestores").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}