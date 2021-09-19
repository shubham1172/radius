// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package dapr

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/radius/pkg/model/components"
	"github.com/Azure/radius/pkg/radrp/outputresource"
	"github.com/Azure/radius/pkg/resourcekinds"
	"github.com/Azure/radius/pkg/workloads"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Renderer is the WorkloadRenderer implementation for the dapr trait decorator.
type Renderer struct {
	Inner workloads.WorkloadRenderer
}

// Allocate is the WorkloadRenderer implementation for the dapr trait decorator.
func (r Renderer) AllocateBindings(ctx context.Context, workload workloads.InstantiatedWorkload, resources []workloads.WorkloadResourceProperties) (map[string]components.BindingState, error) {
	// TODO verify return a binding for dapr invoke
	bindings, err := r.Inner.AllocateBindings(ctx, workload, resources)
	if err != nil {
		return nil, err
	}

	// If the component declares an invoke binding, handle it here so others can depend on it.
	for name, binding := range workload.Workload.Bindings {
		if binding.Kind != BindingKind {
			continue
		}

		trait := Trait{}
		found, err := workload.Workload.FindTrait(Kind, &trait)
		if err != nil {
			return nil, err
		} else if !found {
			// no trait
			return nil, fmt.Errorf("the trait %s is required to use binding %s", Kind, BindingKind)
		}

		if trait.AppID == "" {
			trait.AppID = workload.Workload.Name
		}

		bindings[name] = components.BindingState{
			Component: workload.Name,
			Binding:   name,
			Kind:      binding.Kind,
			Properties: map[string]interface{}{
				"appId": trait.AppID,
			},
		}
	}

	return bindings, nil
}

// Render is the WorkloadRenderer implementation for the dapr deployment decorator.
func (r Renderer) Render(ctx context.Context, w workloads.InstantiatedWorkload) ([]outputresource.OutputResource, error) {
	// Let the inner renderer do its work
	resources, err := r.Inner.Render(ctx, w)
	if err != nil {
		// Even if the operation fails, return the output resources created so far
		// TODO: This is temporary. Once there are no resources actually deployed during render phase,
		// we no longer need to track the output resources on error
		// See: https://github.com/Azure/radius/issues/499
		return resources, err
	}

	trait := Trait{}
	found, err := w.Workload.FindTrait(Kind, &trait)
	if !found || err != nil {
		// Even if the operation fails, return the output resources created so far
		// TODO: This is temporary. Once there are no resources actually deployed during render phase,
		// we no longer need to track the output resources on error
		// See: https://github.com/Azure/radius/issues/499
		return resources, err
	}

	// dapr detected! Update the deployment
	for _, resource := range resources {
		if resource.Kind != resourcekinds.Kubernetes {
			// Not a Kubernetes resource
			continue
		}

		o, ok := resource.Resource.(runtime.Object)
		if !ok {
			// Even if the operation fails, return the output resources created so far
			// TODO: This is temporary. Once there are no resources actually deployed during render phase,
			// we no longer need to track the output resources on error
			// See: https://github.com/Azure/radius/issues/499
			return resources, errors.New("found Kubernetes resource with non-Kubernetes payload")
		}

		annotations, ok := r.getAnnotations(o)
		if !ok {
			continue
		}

		// use the workload name
		if trait.AppID == "" {
			trait.AppID = w.Workload.Name
		}

		annotations["dapr.io/enabled"] = "true"
		annotations["dapr.io/app-id"] = trait.AppID
		if trait.AppPort != 0 {
			annotations["dapr.io/app-port"] = fmt.Sprintf("%d", trait.AppPort)
		}
		if trait.Config != "" {
			annotations["dapr.io/config"] = trait.Config
		}
		if trait.Protocol != "" {
			annotations["dapr.io/protocol"] = trait.Protocol
		}

		r.setAnnotations(o, annotations)
	}

	return resources, err
}

func (r Renderer) getAnnotations(o runtime.Object) (map[string]string, bool) {
	dep, ok := o.(*appsv1.Deployment)
	if ok {
		if dep.Spec.Template.Annotations == nil {
			dep.Spec.Template.Annotations = map[string]string{}
		}

		return dep.Spec.Template.Annotations, true
	}

	un, ok := o.(*unstructured.Unstructured)
	if ok {
		if a := un.GetAnnotations(); a != nil {
			return a, true
		}

		return map[string]string{}, true
	}

	return nil, false
}

func (r Renderer) setAnnotations(o runtime.Object, annotations map[string]string) {
	un, ok := o.(*unstructured.Unstructured)
	if ok {
		un.SetAnnotations(annotations)
	}
}