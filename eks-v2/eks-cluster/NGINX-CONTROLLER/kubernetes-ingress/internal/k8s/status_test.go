package k8s

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	conf_v1 "github.com/nginxinc/kubernetes-ingress/pkg/apis/configuration/v1"
	conf_v1alpha1 "github.com/nginxinc/kubernetes-ingress/pkg/apis/configuration/v1alpha1"
	fake_v1alpha1 "github.com/nginxinc/kubernetes-ingress/pkg/client/clientset/versioned/fake"
	v1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func TestUpdateTransportServerStatus(t *testing.T) {
	ts := &conf_v1alpha1.TransportServer{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "ts-1",
			Namespace: "default",
		},
		Status: conf_v1alpha1.TransportServerStatus{
			State:   "before status",
			Reason:  "before reason",
			Message: "before message",
		},
	}

	fakeClient := fake_v1alpha1.NewSimpleClientset(
		&conf_v1alpha1.TransportServerList{
			Items: []conf_v1alpha1.TransportServer{
				*ts,
			},
		})

	tsLister := cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)

	err := tsLister.Add(ts)
	if err != nil {
		t.Errorf("Error adding TransportServer to the transportserver lister: %v", err)
	}
	su := statusUpdater{
		transportServerLister: tsLister,
		confClient:            fakeClient,
		keyFunc:               cache.DeletionHandlingMetaNamespaceKeyFunc,
	}

	err = su.UpdateTransportServerStatus(ts, "after status", "after reason", "after message")
	if err != nil {
		t.Errorf("error updating transportserver status: %v", err)
	}
	updatedTs, _ := fakeClient.K8sV1alpha1().TransportServers(ts.Namespace).Get(context.TODO(), ts.Name, meta_v1.GetOptions{})

	expectedStatus := conf_v1alpha1.TransportServerStatus{
		State:   "after status",
		Reason:  "after reason",
		Message: "after message",
	}

	if diff := cmp.Diff(expectedStatus, updatedTs.Status); diff != "" {
		t.Errorf("Unexpected status (-want +got):\n%s", diff)
	}
}

func TestUpdateTransportServerStatusIgnoreNoChange(t *testing.T) {
	ts := &conf_v1alpha1.TransportServer{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "ts-1",
			Namespace: "default",
		},
		Status: conf_v1alpha1.TransportServerStatus{
			State:   "same status",
			Reason:  "same reason",
			Message: "same message",
		},
	}

	fakeClient := fake_v1alpha1.NewSimpleClientset(
		&conf_v1alpha1.TransportServerList{
			Items: []conf_v1alpha1.TransportServer{
				*ts,
			},
		})

	tsLister, _ := cache.NewInformer(
		cache.NewListWatchFromClient(
			fakeClient.K8sV1alpha1().RESTClient(),
			"transportservers",
			"nginx-ingress",
			fields.Everything(),
		),
		&conf_v1alpha1.TransportServer{},
		2,
		nil,
	)

	err := tsLister.Add(ts)
	if err != nil {
		t.Errorf("Error adding TransportServer to the transportserver lister: %v", err)
	}
	su := statusUpdater{
		transportServerLister: tsLister,
		confClient:            fakeClient,
		keyFunc:               cache.DeletionHandlingMetaNamespaceKeyFunc,
	}

	err = su.UpdateTransportServerStatus(ts, "same status", "same reason", "same message")
	if err != nil {
		t.Errorf("error updating transportserver status: %v", err)
	}
	updatedTs, _ := fakeClient.K8sV1alpha1().TransportServers(ts.Namespace).Get(context.TODO(), ts.Name, meta_v1.GetOptions{})

	if updatedTs.Status.State != "same status" {
		t.Errorf("expected: %v actual: %v", "same status", updatedTs.Status.State)
	}
	if updatedTs.Status.Message != "same message" {
		t.Errorf("expected: %v actual: %v", "same message", updatedTs.Status.Message)
	}
	if updatedTs.Status.Reason != "same reason" {
		t.Errorf("expected: %v actual: %v", "same reason", updatedTs.Status.Reason)
	}
}

func TestUpdateTransportServerStatusMissingTransportServer(t *testing.T) {
	ts := &conf_v1alpha1.TransportServer{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "ts-1",
			Namespace: "default",
		},
		Status: conf_v1alpha1.TransportServerStatus{
			State:   "before status",
			Reason:  "before reason",
			Message: "before message",
		},
	}

	fakeClient := fake_v1alpha1.NewSimpleClientset(
		&conf_v1alpha1.TransportServerList{
			Items: []conf_v1alpha1.TransportServer{},
		})

	tsLister, _ := cache.NewInformer(
		cache.NewListWatchFromClient(
			fakeClient.K8sV1alpha1().RESTClient(),
			"transportservers",
			"nginx-ingress",
			fields.Everything(),
		),
		&conf_v1alpha1.TransportServer{},
		2,
		nil,
	)

	su := statusUpdater{
		transportServerLister: tsLister,
		confClient:            fakeClient,
		keyFunc:               cache.DeletionHandlingMetaNamespaceKeyFunc,
		externalEndpoints: []conf_v1.ExternalEndpoint{
			{
				IP:    "123.123.123.123",
				Ports: "1234",
			},
		},
	}

	err := su.UpdateTransportServerStatus(ts, "after status", "after reason", "after message")
	if err != nil {
		t.Errorf("unexpected error: %v, result should be empty as no matching TransportServer is present", err)
	}

	updatedTs, _ := fakeClient.K8sV1alpha1().TransportServers(ts.Namespace).Get(context.TODO(), ts.Name, meta_v1.GetOptions{})
	if updatedTs != nil {
		t.Errorf("expected TransportServer Store would be empty as provided TransportServer was not found. Unexpected updated TransportServer: %v", updatedTs)
	}
}

func TestStatusUpdateWithExternalStatusAndExternalService(t *testing.T) {
	ing := networking.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "ing-1",
			Namespace: "namespace",
		},
		Status: networking.IngressStatus{
			LoadBalancer: v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{
					{
						IP: "1.2.3.4",
					},
				},
			},
		},
	}
	fakeClient := fake.NewSimpleClientset(
		&networking.IngressList{Items: []networking.Ingress{
			ing,
		}},
	)
	ingLister := storeToIngressLister{}
	ingLister.Store, _ = cache.NewInformer(
		cache.NewListWatchFromClient(fakeClient.NetworkingV1().RESTClient(), "ingresses", "nginx-ingress", fields.Everything()),
		&networking.Ingress{}, 2, nil)

	err := ingLister.Store.Add(&ing)
	if err != nil {
		t.Errorf("Error adding Ingress to the ingress lister: %v", err)
	}

	su := statusUpdater{
		client:                fakeClient,
		namespace:             "namespace",
		externalServiceName:   "service-name",
		externalStatusAddress: "123.123.123.123",
		ingressLister:         &ingLister,
		keyFunc:               cache.DeletionHandlingMetaNamespaceKeyFunc,
	}
	err = su.ClearIngressStatus(ing)
	if err != nil {
		t.Errorf("error clearing ing status: %v", err)
	}
	ings, _ := fakeClient.NetworkingV1().Ingresses("namespace").List(context.TODO(), meta_v1.ListOptions{})
	ingf := ings.Items[0]
	if !checkStatus("", ingf) {
		t.Errorf("expected: %v actual: %v", "", ingf.Status.LoadBalancer.Ingress[0])
	}

	su.SaveStatusFromExternalStatus("1.1.1.1")
	err = su.UpdateIngressStatus(ing)
	if err != nil {
		t.Errorf("error updating ing status: %v", err)
	}
	ring, _ := fakeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, meta_v1.GetOptions{})
	if !checkStatus("1.1.1.1", *ring) {
		t.Errorf("expected: %v actual: %v", "", ring.Status.LoadBalancer.Ingress)
	}

	svc := v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Namespace: "namespace",
			Name:      "service-name",
		},
		Status: v1.ServiceStatus{
			LoadBalancer: v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{{
					IP: "2.2.2.2",
				}},
			},
		},
	}
	su.SaveStatusFromExternalService(&svc)
	err = su.UpdateIngressStatus(ing)
	if err != nil {
		t.Errorf("error updating ing status: %v", err)
	}
	ring, _ = fakeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, meta_v1.GetOptions{})
	if !checkStatus("1.1.1.1", *ring) {
		t.Errorf("expected: %v actual: %v", "1.1.1.1", ring.Status.LoadBalancer.Ingress)
	}

	su.SaveStatusFromExternalStatus("")
	err = su.UpdateIngressStatus(ing)
	if err != nil {
		t.Errorf("error updating ing status: %v", err)
	}
	ring, _ = fakeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, meta_v1.GetOptions{})
	if !checkStatus("2.2.2.2", *ring) {
		t.Errorf("expected: %v actual: %v", "2.2.2.2", ring.Status.LoadBalancer.Ingress)
	}

	su.ClearStatusFromExternalService()
	err = su.UpdateIngressStatus(ing)
	if err != nil {
		t.Errorf("error updating ing status: %v", err)
	}
	ring, _ = fakeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, meta_v1.GetOptions{})
	if !checkStatus("", *ring) {
		t.Errorf("expected: %v actual: %v", "", ring.Status.LoadBalancer.Ingress)
	}
}

func TestStatusUpdateWithExternalStatusAndIngressLink(t *testing.T) {
	ing := networking.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "ing-1",
			Namespace: "namespace",
		},
		Status: networking.IngressStatus{
			LoadBalancer: v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{
					{
						IP: "1.2.3.4",
					},
				},
			},
		},
	}
	fakeClient := fake.NewSimpleClientset(
		&networking.IngressList{Items: []networking.Ingress{
			ing,
		}},
	)
	ingLister := storeToIngressLister{}
	ingLister.Store, _ = cache.NewInformer(
		cache.NewListWatchFromClient(fakeClient.NetworkingV1().RESTClient(), "ingresses", "nginx-ingress", fields.Everything()),
		&networking.Ingress{}, 2, nil)

	err := ingLister.Store.Add(&ing)
	if err != nil {
		t.Errorf("Error adding Ingress to the ingress lister: %v", err)
	}

	su := statusUpdater{
		client:                fakeClient,
		namespace:             "namespace",
		externalStatusAddress: "",
		ingressLister:         &ingLister,
		keyFunc:               cache.DeletionHandlingMetaNamespaceKeyFunc,
	}

	su.SaveStatusFromIngressLink("3.3.3.3")
	err = su.UpdateIngressStatus(ing)
	if err != nil {
		t.Errorf("error updating ing status: %v", err)
	}
	ring, _ := fakeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, meta_v1.GetOptions{})
	if !checkStatus("3.3.3.3", *ring) {
		t.Errorf("expected: %v actual: %v", "3.3.3.3", ring.Status.LoadBalancer.Ingress)
	}

	su.SaveStatusFromExternalStatus("1.1.1.1")
	err = su.UpdateIngressStatus(ing)
	if err != nil {
		t.Errorf("error updating ing status: %v", err)
	}
	ring, _ = fakeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, meta_v1.GetOptions{})
	if !checkStatus("1.1.1.1", *ring) {
		t.Errorf("expected: %v actual: %v", "1.1.1.1", ring.Status.LoadBalancer.Ingress)
	}

	su.ClearStatusFromIngressLink()
	err = su.UpdateIngressStatus(ing)
	if err != nil {
		t.Errorf("error updating ing status: %v", err)
	}
	ring, _ = fakeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, meta_v1.GetOptions{})
	if !checkStatus("1.1.1.1", *ring) {
		t.Errorf("expected: %v actual: %v", "1.1.1.1", ring.Status.LoadBalancer.Ingress)
	}

	su.SaveStatusFromIngressLink("4.4.4.4")
	err = su.UpdateIngressStatus(ing)
	if err != nil {
		t.Errorf("error updating ing status: %v", err)
	}
	ring, _ = fakeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, meta_v1.GetOptions{})
	if !checkStatus("1.1.1.1", *ring) {
		t.Errorf("expected: %v actual: %v", "1.1.1.1", ring.Status.LoadBalancer.Ingress)
	}

	su.SaveStatusFromExternalStatus("")
	err = su.UpdateIngressStatus(ing)
	if err != nil {
		t.Errorf("error updating ing status: %v", err)
	}
	ring, _ = fakeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, meta_v1.GetOptions{})
	if !checkStatus("4.4.4.4", *ring) {
		t.Errorf("expected: %v actual: %v", "4.4.4.4", ring.Status.LoadBalancer.Ingress)
	}

	su.ClearStatusFromIngressLink()
	err = su.UpdateIngressStatus(ing)
	if err != nil {
		t.Errorf("error updating ing status: %v", err)
	}
	ring, _ = fakeClient.NetworkingV1().Ingresses(ing.Namespace).Get(context.TODO(), ing.Name, meta_v1.GetOptions{})
	if !checkStatus("", *ring) {
		t.Errorf("expected: %v actual: %v", "", ring.Status.LoadBalancer.Ingress)
	}
}

func checkStatus(expected string, actual networking.Ingress) bool {
	if len(actual.Status.LoadBalancer.Ingress) == 0 {
		return expected == ""
	}
	return expected == actual.Status.LoadBalancer.Ingress[0].IP
}

func TestGenerateExternalEndpointsFromStatus(t *testing.T) {
	su := statusUpdater{
		status: []v1.LoadBalancerIngress{
			{
				IP: "8.8.8.8",
			},
		},
	}

	expectedEndpoints := []conf_v1.ExternalEndpoint{
		{IP: "8.8.8.8", Ports: ""},
	}

	endpoints := su.generateExternalEndpointsFromStatus(su.status)

	if !reflect.DeepEqual(endpoints, expectedEndpoints) {
		t.Errorf("generateExternalEndpointsFromStatus(%v) returned %v but expected %v", su.status, endpoints, expectedEndpoints)
	}
}

func TestHasVsStatusChanged(t *testing.T) {
	state := "Valid"
	reason := "AddedOrUpdated"
	msg := "Configuration was added or updated"

	tests := []struct {
		expected bool
		vs       conf_v1.VirtualServer
	}{
		{
			expected: false,
			vs: conf_v1.VirtualServer{
				Status: conf_v1.VirtualServerStatus{
					State:   state,
					Reason:  reason,
					Message: msg,
				},
			},
		},
		{
			expected: true,
			vs: conf_v1.VirtualServer{
				Status: conf_v1.VirtualServerStatus{
					State:   "DifferentState",
					Reason:  reason,
					Message: msg,
				},
			},
		},
		{
			expected: true,
			vs: conf_v1.VirtualServer{
				Status: conf_v1.VirtualServerStatus{
					State:   state,
					Reason:  "DifferentReason",
					Message: msg,
				},
			},
		},
		{
			expected: true,
			vs: conf_v1.VirtualServer{
				Status: conf_v1.VirtualServerStatus{
					State:   state,
					Reason:  reason,
					Message: "DifferentMessage",
				},
			},
		},
	}

	for _, test := range tests {
		changed := hasVsStatusChanged(&test.vs, state, reason, msg)

		if changed != test.expected {
			t.Errorf("hasVsStatusChanged(%v, %v, %v, %v) returned %v but expected %v.", test.vs, state, reason, msg, changed, test.expected)
		}
	}
}

func TestHasVsrStatusChanged(t *testing.T) {
	referencedBy := "namespace/name"
	state := "Valid"
	reason := "AddedOrUpdated"
	msg := "Configuration was added or updated"

	tests := []struct {
		expected bool
		vsr      conf_v1.VirtualServerRoute
	}{
		{
			expected: false,
			vsr: conf_v1.VirtualServerRoute{
				Status: conf_v1.VirtualServerRouteStatus{
					State:        state,
					Reason:       reason,
					Message:      msg,
					ReferencedBy: referencedBy,
				},
			},
		},
		{
			expected: true,
			vsr: conf_v1.VirtualServerRoute{
				Status: conf_v1.VirtualServerRouteStatus{
					State:        "DifferentState",
					Reason:       reason,
					Message:      msg,
					ReferencedBy: referencedBy,
				},
			},
		},
		{
			expected: true,
			vsr: conf_v1.VirtualServerRoute{
				Status: conf_v1.VirtualServerRouteStatus{
					State:        state,
					Reason:       "DifferentReason",
					Message:      msg,
					ReferencedBy: referencedBy,
				},
			},
		},
		{
			expected: true,
			vsr: conf_v1.VirtualServerRoute{
				Status: conf_v1.VirtualServerRouteStatus{
					State:        state,
					Reason:       reason,
					Message:      "DifferentMessage",
					ReferencedBy: referencedBy,
				},
			},
		},
		{
			expected: true,
			vsr: conf_v1.VirtualServerRoute{
				Status: conf_v1.VirtualServerRouteStatus{
					State:        state,
					Reason:       reason,
					Message:      msg,
					ReferencedBy: "DifferentReferencedBy",
				},
			},
		},
	}

	for _, test := range tests {
		changed := hasVsrStatusChanged(&test.vsr, state, reason, msg, referencedBy)

		if changed != test.expected {
			t.Errorf("hasVsrStatusChanged(%v, %v, %v, %v) returned %v but expected %v.", test.vsr, state, reason, msg, changed, test.expected)
		}
	}
}

func TestGetExternalServicePorts(t *testing.T) {
	svc := v1.Service{
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port: int32(80),
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 80,
					},
				},
				{
					Port: int32(443),
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 443,
					},
				},
			},
		},
	}

	expected := "[80,443]"
	ports := getExternalServicePorts(&svc)

	if ports != expected {
		t.Errorf("getExternalServicePorts(%v) returned %v but expected %v", svc, ports, expected)
	}
}

func TestIsRequiredPort(t *testing.T) {
	tests := []struct {
		port     intstr.IntOrString
		expected bool
	}{
		{
			port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 999,
			},
			expected: false,
		},
		{
			port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 80,
			},
			expected: true,
		},
		{
			port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 443,
			},
			expected: true,
		},
		{
			port: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "name",
			},
			expected: false,
		},
		{
			port: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "http",
			},
			expected: true,
		},
		{
			port: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "https",
			},
			expected: true,
		},
	}

	for _, test := range tests {
		result := isRequiredPort(test.port)

		if result != test.expected {
			t.Errorf("isRequiredPort(%+v) returned %v but expected %v", test.port, result, test.expected)
		}
	}
}

func TestHasPolicyStatusChanged(t *testing.T) {
	state := "Valid"
	reason := "AddedOrUpdated"
	msg := "Configuration was added or updated"

	tests := []struct {
		expected bool
		pol      conf_v1.Policy
	}{
		{
			expected: false,
			pol: conf_v1.Policy{
				Status: conf_v1.PolicyStatus{
					State:   state,
					Reason:  reason,
					Message: msg,
				},
			},
		},
		{
			expected: true,
			pol: conf_v1.Policy{
				Status: conf_v1.PolicyStatus{
					State:   "DifferentState",
					Reason:  reason,
					Message: msg,
				},
			},
		},
		{
			expected: true,
			pol: conf_v1.Policy{
				Status: conf_v1.PolicyStatus{
					State:   state,
					Reason:  "DifferentReason",
					Message: msg,
				},
			},
		},
		{
			expected: true,
			pol: conf_v1.Policy{
				Status: conf_v1.PolicyStatus{
					State:   state,
					Reason:  reason,
					Message: "DifferentMessage",
				},
			},
		},
	}

	for _, test := range tests {
		changed := hasPolicyStatusChanged(&test.pol, state, reason, msg)

		if changed != test.expected {
			t.Errorf("hasPolicyStatusChanged(%v, %v, %v, %v) returned %v but expected %v.", test.pol, state, reason, msg, changed, test.expected)
		}
	}
}
