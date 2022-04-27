/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package k8s

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// storeToIngressLister makes a Store that lists Ingress.
// TODO: Move this to cache/listers post 1.1.
type storeToIngressLister struct {
	cache.Store
}

// GetByKeySafe calls Store.GetByKeySafe and returns a copy of the ingress so it is
// safe to modify.
func (s *storeToIngressLister) GetByKeySafe(key string) (ing *networking.Ingress, exists bool, err error) {
	item, exists, err := s.Store.GetByKey(key)
	if !exists || err != nil {
		return nil, exists, err
	}
	ing = item.(*networking.Ingress).DeepCopy()
	return
}

// List lists all Ingress' in the store.
func (s *storeToIngressLister) List() (ing networking.IngressList, err error) {
	for _, m := range s.Store.List() {
		ing.Items = append(ing.Items, *(m.(*networking.Ingress)).DeepCopy())
	}
	return ing, nil
}

// storeToConfigMapLister makes a Store that lists ConfigMaps
type storeToConfigMapLister struct {
	cache.Store
}

// List lists all Ingress' in the store.
func (s *storeToConfigMapLister) List() (cfgm v1.ConfigMapList, err error) {
	for _, m := range s.Store.List() {
		cfgm.Items = append(cfgm.Items, *(m.(*v1.ConfigMap)))
	}
	return cfgm, nil
}

// indexerToPodLister makes a Indexer that lists Pods.
type indexerToPodLister struct {
	cache.Indexer
}

// ListByNamespace lists all Pods in the indexer for a given namespace that match the provided selector.
func (ipl indexerToPodLister) ListByNamespace(ns string, selector labels.Selector) (pods []*v1.Pod, err error) {
	err = cache.ListAllByNamespace(ipl.Indexer, ns, selector, func(m interface{}) {
		pods = append(pods, m.(*v1.Pod))
	})
	return pods, err
}

// storeToEndpointLister makes a Store that lists Endponts
type storeToEndpointLister struct {
	cache.Store
}

// GetServiceEndpoints returns the endpoints of a service, matched on service name.
func (s *storeToEndpointLister) GetServiceEndpoints(svc *v1.Service) (ep v1.Endpoints, err error) {
	for _, m := range s.Store.List() {
		ep = *m.(*v1.Endpoints)
		if svc.Name == ep.Name && svc.Namespace == ep.Namespace {
			return ep, nil
		}
	}
	return ep, fmt.Errorf("could not find endpoints for service: %v", svc.Name)
}

// findPort locates the container port for the given pod and portName.  If the
// targetPort is a number, use that.  If the targetPort is a string, look that
// string up in all named ports in all containers in the target pod.  If no
// match is found, fail.
func findPort(pod *v1.Pod, svcPort v1.ServicePort) (int32, error) {
	portName := svcPort.TargetPort
	switch portName.Type {
	case intstr.String:
		name := portName.StrVal
		for _, container := range pod.Spec.Containers {
			for _, port := range container.Ports {
				if port.Name == name && port.Protocol == svcPort.Protocol {
					return port.ContainerPort, nil
				}
			}
		}
	case intstr.Int:
		return int32(portName.IntValue()), nil
	}

	return 0, fmt.Errorf("no suitable port for manifest: %s", pod.UID)
}

// isMinion determines is an ingress is a minion or not
func isMinion(ing *networking.Ingress) bool {
	return ing.Annotations["nginx.org/mergeable-ingress-type"] == "minion"
}

// isMaster determines is an ingress is a master or not
func isMaster(ing *networking.Ingress) bool {
	return ing.Annotations["nginx.org/mergeable-ingress-type"] == "master"
}

// hasChanges determines if current ingress has changes compared to old ingress
func hasChanges(old *networking.Ingress, current *networking.Ingress) bool {
	old.Status.LoadBalancer.Ingress = current.Status.LoadBalancer.Ingress
	old.ResourceVersion = current.ResourceVersion
	return !reflect.DeepEqual(old, current)
}

// ParseNamespaceName parses the string in the <namespace>/<name> format and returns the name and the namespace.
// It returns an error in case the string does not follow the <namespace>/<name> format.
func ParseNamespaceName(value string) (ns string, name string, err error) {
	res := strings.Split(value, "/")
	if len(res) != 2 {
		return "", "", fmt.Errorf("%q must follow the format <namespace>/<name>", value)
	}
	return res[0], res[1], nil
}

// GetK8sVersion returns the running version of k8s
func GetK8sVersion(client kubernetes.Interface) (v *version.Version, err error) {
	serverVersion, err := client.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}

	runningVersion, err := version.ParseGeneric(serverVersion.String())
	if err != nil {
		return nil, fmt.Errorf("unexpected error parsing running Kubernetes version: %w", err)
	}
	glog.V(3).Infof("Kubernetes version: %v", runningVersion)

	return runningVersion, nil
}
