package k8s

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/nginxinc/kubernetes-ingress/internal/configs"
	conf_v1 "github.com/nginxinc/kubernetes-ingress/pkg/apis/configuration/v1"
	conf_v1alpha1 "github.com/nginxinc/kubernetes-ingress/pkg/apis/configuration/v1alpha1"
	"github.com/nginxinc/kubernetes-ingress/pkg/apis/configuration/validation"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ingressKind            = "Ingress"
	virtualServerKind      = "VirtualServer"
	virtualServerRouteKind = "VirtualServerRoute"
	transportServerKind    = "TransportServer"
)

// Operation defines an operation to perform for a resource.
type Operation int

const (
	// Delete the config of the resource
	Delete Operation = iota
	// AddOrUpdate the config of the resource
	AddOrUpdate
)

// Resource represents a configuration resource.
// A Resource can be a top level configuration object:
// - Regular or Master Ingress
// - VirtualServer
// - TransportServer
type Resource interface {
	GetObjectMeta() *metav1.ObjectMeta
	GetKeyWithKind() string
	Wins(resource Resource) bool
	AddWarning(warning string)
	IsEqual(resource Resource) bool
}

func chooseObjectMetaWinner(meta1 *metav1.ObjectMeta, meta2 *metav1.ObjectMeta) bool {
	if meta1.CreationTimestamp.Equal(&meta2.CreationTimestamp) {
		return meta1.UID > meta2.UID
	}

	return meta1.CreationTimestamp.Before(&meta2.CreationTimestamp)
}

// ResourceChange represents a change of the resource that needs to be reflected in the NGINX config.
type ResourceChange struct {
	// Op is an operation that needs be performed on the resource.
	Op Operation
	// Resource is the target resource.
	Resource Resource
	// Error is the error associated with the resource.
	Error string
}

// ConfigurationProblem is a problem associated with a configuration object.
type ConfigurationProblem struct {
	// Object is a configuration object.
	Object runtime.Object
	// IsError tells if the problem is an error. If it is an error, then it is expected that the status of the object
	// will be updated to the state 'invalid'. Otherwise, the state will be 'warning'.
	IsError bool
	// Reason tells the reason. It matches the reason in the events/status of our configuration objects.
	Reason string
	// Messages gives the details about the problem. It matches the message in the events/status of our configuration objects.
	Message string
}

func compareConfigurationProblems(problem1 *ConfigurationProblem, problem2 *ConfigurationProblem) bool {
	return problem1.IsError == problem2.IsError &&
		problem1.Reason == problem2.Reason &&
		problem1.Message == problem2.Message
}

// IngressConfiguration holds an Ingress resource with its minions. It implements the Resource interface.
type IngressConfiguration struct {
	// Ingress holds a regular Ingress or a master Ingress.
	Ingress *networking.Ingress
	// IsMaster is true when the Ingress is a master.
	IsMaster bool
	// Minions contains minions if the Ingress is a master.
	Minions []*MinionConfiguration
	// ValidHosts marks the hosts of the Ingress as valid (true) or invalid (false).
	// Regular Ingress resources can have multiple hosts. It is possible that some of the hosts are taken by other
	// resources. In that case, those hosts will be marked as invalid.
	ValidHosts map[string]bool
	// Warnings includes all the warnings for the resource.
	Warnings []string
	// ChildWarnings includes the warnings of the minions. The key is the namespace/name.
	ChildWarnings map[string][]string
}

// NewRegularIngressConfiguration creates an IngressConfiguration from an Ingress resource.
func NewRegularIngressConfiguration(ing *networking.Ingress) *IngressConfiguration {
	return &IngressConfiguration{
		Ingress:       ing,
		IsMaster:      false,
		ValidHosts:    make(map[string]bool),
		ChildWarnings: make(map[string][]string),
	}
}

// NewMasterIngressConfiguration creates an IngressConfiguration from a master Ingress resource.
func NewMasterIngressConfiguration(ing *networking.Ingress, minions []*MinionConfiguration, childWarnings map[string][]string) *IngressConfiguration {
	return &IngressConfiguration{
		Ingress:       ing,
		IsMaster:      true,
		Minions:       minions,
		ValidHosts:    make(map[string]bool),
		ChildWarnings: childWarnings,
	}
}

// GetObjectMeta returns the resource ObjectMeta.
func (ic *IngressConfiguration) GetObjectMeta() *metav1.ObjectMeta {
	return &ic.Ingress.ObjectMeta
}

// GetKeyWithKind returns the key of the resource with its kind. For example, Ingress/my-namespace/my-name.
func (ic *IngressConfiguration) GetKeyWithKind() string {
	key := getResourceKey(&ic.Ingress.ObjectMeta)
	return fmt.Sprintf("%s/%s", ingressKind, key)
}

// Wins tells if this resource wins over the specified resource.
func (ic *IngressConfiguration) Wins(resource Resource) bool {
	return chooseObjectMetaWinner(ic.GetObjectMeta(), resource.GetObjectMeta())
}

// AddWarning adds a warning.
func (ic *IngressConfiguration) AddWarning(warning string) {
	ic.Warnings = append(ic.Warnings, warning)
}

// IsEqual tests if the IngressConfiguration is equal to the resource.
func (ic *IngressConfiguration) IsEqual(resource Resource) bool {
	ingConfig, ok := resource.(*IngressConfiguration)
	if !ok {
		return false
	}

	if !compareObjectMetasWithAnnotations(&ic.Ingress.ObjectMeta, &ingConfig.Ingress.ObjectMeta) {
		return false
	}

	if !reflect.DeepEqual(ic.ValidHosts, ingConfig.ValidHosts) {
		return false
	}

	if ic.IsMaster != ingConfig.IsMaster {
		return false
	}

	if len(ic.Minions) != len(ingConfig.Minions) {
		return false
	}

	for i := range ic.Minions {
		if !compareObjectMetasWithAnnotations(&ic.Minions[i].Ingress.ObjectMeta, &ingConfig.Minions[i].Ingress.ObjectMeta) {
			return false
		}
	}

	return true
}

// MinionConfiguration holds a Minion resource.
type MinionConfiguration struct {
	// Ingress is the Ingress behind a minion.
	Ingress *networking.Ingress
	// ValidPaths marks the paths of the Ingress as valid (true) or invalid (false).
	// Minion Ingress resources can have multiple paths. It is possible that some of the paths are taken by other
	// Minions. In that case, those paths will be marked as invalid.
	ValidPaths map[string]bool
}

// NewMinionConfiguration creates a new MinionConfiguration.
func NewMinionConfiguration(ing *networking.Ingress) *MinionConfiguration {
	return &MinionConfiguration{
		Ingress:    ing,
		ValidPaths: make(map[string]bool),
	}
}

// VirtualServerConfiguration holds a VirtualServer along with its VirtualServerRoutes.
type VirtualServerConfiguration struct {
	VirtualServer       *conf_v1.VirtualServer
	VirtualServerRoutes []*conf_v1.VirtualServerRoute
	Warnings            []string
}

// NewVirtualServerConfiguration creates a VirtualServerConfiguration.
func NewVirtualServerConfiguration(vs *conf_v1.VirtualServer, vsrs []*conf_v1.VirtualServerRoute, warnings []string) *VirtualServerConfiguration {
	return &VirtualServerConfiguration{
		VirtualServer:       vs,
		VirtualServerRoutes: vsrs,
		Warnings:            warnings,
	}
}

// GetObjectMeta returns the resource ObjectMeta.
func (vsc *VirtualServerConfiguration) GetObjectMeta() *metav1.ObjectMeta {
	return &vsc.VirtualServer.ObjectMeta
}

// GetKeyWithKind returns the key of the resource with its kind. For example, VirtualServer/my-namespace/my-name.
func (vsc *VirtualServerConfiguration) GetKeyWithKind() string {
	key := getResourceKey(&vsc.VirtualServer.ObjectMeta)
	return fmt.Sprintf("%s/%s", virtualServerKind, key)
}

// Wins tells if this resource wins over the specified resource.
// It is used to determine which resource should win over a host.
func (vsc *VirtualServerConfiguration) Wins(resource Resource) bool {
	return chooseObjectMetaWinner(vsc.GetObjectMeta(), resource.GetObjectMeta())
}

// AddWarning adds a warning.
func (vsc *VirtualServerConfiguration) AddWarning(warning string) {
	vsc.Warnings = append(vsc.Warnings, warning)
}

// IsEqual tests if the VirtualServerConfiguration is equal to the resource.
func (vsc *VirtualServerConfiguration) IsEqual(resource Resource) bool {
	vsConfig, ok := resource.(*VirtualServerConfiguration)
	if !ok {
		return false
	}

	if !compareObjectMetas(&vsc.VirtualServer.ObjectMeta, &vsConfig.VirtualServer.ObjectMeta) {
		return false
	}

	if len(vsc.VirtualServerRoutes) != len(vsConfig.VirtualServerRoutes) {
		return false
	}

	for i := range vsc.VirtualServerRoutes {
		if !compareObjectMetas(&vsc.VirtualServerRoutes[i].ObjectMeta, &vsConfig.VirtualServerRoutes[i].ObjectMeta) {
			return false
		}
	}

	return true
}

// TransportServerConfiguration holds a TransportServer resource.
type TransportServerConfiguration struct {
	ListenerPort    int
	TransportServer *conf_v1alpha1.TransportServer
	Warnings        []string
}

// NewTransportServerConfiguration creates a new TransportServerConfiguration.
func NewTransportServerConfiguration(ts *conf_v1alpha1.TransportServer) *TransportServerConfiguration {
	return &TransportServerConfiguration{
		TransportServer: ts,
	}
}

// GetObjectMeta returns the resource ObjectMeta.
func (tsc *TransportServerConfiguration) GetObjectMeta() *metav1.ObjectMeta {
	return &tsc.TransportServer.ObjectMeta
}

// GetKeyWithKind returns the key of the resource with its kind. For example, TransportServer/my-namespace/my-name.
func (tsc *TransportServerConfiguration) GetKeyWithKind() string {
	key := getResourceKey(&tsc.TransportServer.ObjectMeta)
	return fmt.Sprintf("%s/%s", transportServerKind, key)
}

// Wins tells if this resource wins over the specified resource.
// It is used to determine which resource should win over a host or port.
func (tsc *TransportServerConfiguration) Wins(resource Resource) bool {
	return chooseObjectMetaWinner(tsc.GetObjectMeta(), resource.GetObjectMeta())
}

// AddWarning adds a warning.
func (tsc *TransportServerConfiguration) AddWarning(warning string) {
	tsc.Warnings = append(tsc.Warnings, warning)
}

// IsEqual tests if the TransportServerConfiguration is equal to the resource.
func (tsc *TransportServerConfiguration) IsEqual(resource Resource) bool {
	tsConfig, ok := resource.(*TransportServerConfiguration)
	if !ok {
		return false
	}

	return compareObjectMetas(tsc.GetObjectMeta(), resource.GetObjectMeta()) && tsc.ListenerPort == tsConfig.ListenerPort
}

func compareObjectMetas(meta1 *metav1.ObjectMeta, meta2 *metav1.ObjectMeta) bool {
	return meta1.Namespace == meta2.Namespace &&
		meta1.Name == meta2.Name &&
		meta1.Generation == meta2.Generation
}

func compareObjectMetasWithAnnotations(meta1 *metav1.ObjectMeta, meta2 *metav1.ObjectMeta) bool {
	return compareObjectMetas(meta1, meta2) && reflect.DeepEqual(meta1.Annotations, meta2.Annotations)
}

// TransportServerMetrics holds metrics about TransportServer resources
type TransportServerMetrics struct {
	TotalTLSPassthrough int
	TotalTCP            int
	TotalUDP            int
}

// Configuration represents the configuration of the Ingress Controller - a collection of configuration objects
// (Ingresses, VirtualServers, VirtualServerRoutes) ready to be transformed into NGINX config.
// It holds the latest valid state of those objects.
// The IC needs to ensure that at any point in time the NGINX config on the filesystem reflects the state
// of the objects in the Configuration.
type Configuration struct {
	hosts     map[string]Resource
	listeners map[string]*TransportServerConfiguration

	// only valid resources with the matching IngressClass are stored
	ingresses           map[string]*networking.Ingress
	virtualServers      map[string]*conf_v1.VirtualServer
	virtualServerRoutes map[string]*conf_v1.VirtualServerRoute
	transportServers    map[string]*conf_v1alpha1.TransportServer

	globalConfiguration *conf_v1alpha1.GlobalConfiguration

	hostProblems     map[string]ConfigurationProblem
	listenerProblems map[string]ConfigurationProblem

	hasCorrectIngressClass       func(interface{}) bool
	virtualServerValidator       *validation.VirtualServerValidator
	globalConfigurationValidator *validation.GlobalConfigurationValidator
	transportServerValidator     *validation.TransportServerValidator

	secretReferenceChecker     *secretReferenceChecker
	serviceReferenceChecker    *serviceReferenceChecker
	endpointReferenceChecker   *serviceReferenceChecker
	policyReferenceChecker     *policyReferenceChecker
	appPolicyReferenceChecker  *appProtectResourceReferenceChecker
	appLogConfReferenceChecker *appProtectResourceReferenceChecker
	appDosProtectedChecker     *dosResourceReferenceChecker

	isPlus                  bool
	appProtectEnabled       bool
	appProtectDosEnabled    bool
	internalRoutesEnabled   bool
	isTLSPassthroughEnabled bool
	snippetsEnabled         bool

	lock sync.RWMutex
}

// NewConfiguration creates a new Configuration.
func NewConfiguration(
	hasCorrectIngressClass func(interface{}) bool,
	isPlus bool,
	appProtectEnabled bool,
	appProtectDosEnabled bool,
	internalRoutesEnabled bool,
	virtualServerValidator *validation.VirtualServerValidator,
	globalConfigurationValidator *validation.GlobalConfigurationValidator,
	transportServerValidator *validation.TransportServerValidator,
	isTLSPassthroughEnabled bool,
	snippetsEnabled bool,
) *Configuration {
	return &Configuration{
		hosts:                        make(map[string]Resource),
		listeners:                    make(map[string]*TransportServerConfiguration),
		ingresses:                    make(map[string]*networking.Ingress),
		virtualServers:               make(map[string]*conf_v1.VirtualServer),
		virtualServerRoutes:          make(map[string]*conf_v1.VirtualServerRoute),
		transportServers:             make(map[string]*conf_v1alpha1.TransportServer),
		hostProblems:                 make(map[string]ConfigurationProblem),
		hasCorrectIngressClass:       hasCorrectIngressClass,
		virtualServerValidator:       virtualServerValidator,
		globalConfigurationValidator: globalConfigurationValidator,
		transportServerValidator:     transportServerValidator,
		secretReferenceChecker:       newSecretReferenceChecker(isPlus),
		serviceReferenceChecker:      newServiceReferenceChecker(false),
		endpointReferenceChecker:     newServiceReferenceChecker(true),
		policyReferenceChecker:       newPolicyReferenceChecker(),
		appDosProtectedChecker:       newDosResourceReferenceChecker(configs.AppProtectDosProtectedAnnotation),
		isPlus:                       isPlus,
		appProtectEnabled:            appProtectEnabled,
		appProtectDosEnabled:         appProtectDosEnabled,
		internalRoutesEnabled:        internalRoutesEnabled,
		isTLSPassthroughEnabled:      isTLSPassthroughEnabled,
		snippetsEnabled:              snippetsEnabled,
	}
}

// AddOrUpdateIngress adds or updates the Ingress resource.
func (c *Configuration) AddOrUpdateIngress(ing *networking.Ingress) ([]ResourceChange, []ConfigurationProblem) {
	c.lock.Lock()
	defer c.lock.Unlock()

	key := getResourceKey(&ing.ObjectMeta)
	var validationError error

	if !c.hasCorrectIngressClass(ing) {
		delete(c.ingresses, key)
	} else {
		validationError = validateIngress(ing, c.isPlus, c.appProtectEnabled, c.appProtectDosEnabled, c.internalRoutesEnabled, c.snippetsEnabled).ToAggregate()
		if validationError != nil {
			delete(c.ingresses, key)
		} else {
			c.ingresses[key] = ing
		}
	}

	changes, problems := c.rebuildHosts()

	if validationError != nil {
		// If the invalid resource has any active hosts, rebuildHosts will create a change
		// to remove the resource.
		// Here we add the validationErr to that change.
		keyWithKind := getResourceKeyWithKind(ingressKind, &ing.ObjectMeta)
		for i := range changes {
			k := changes[i].Resource.GetKeyWithKind()

			if k == keyWithKind {
				changes[i].Error = validationError.Error()
				return changes, problems
			}
		}

		// On the other hand, the invalid resource might not have any active hosts.
		// Or the resource was invalid before and is still invalid (in some different way).
		// In those cases,  rebuildHosts will create no change for that resource.
		// To make sure the validationErr is reported to the user, we create a problem.
		p := ConfigurationProblem{
			Object:  ing,
			IsError: true,
			Reason:  "Rejected",
			Message: validationError.Error(),
		}
		problems = append(problems, p)
	}

	return changes, problems
}

// DeleteIngress deletes an Ingress resource by the key.
func (c *Configuration) DeleteIngress(key string) ([]ResourceChange, []ConfigurationProblem) {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, exists := c.ingresses[key]
	if !exists {
		return nil, nil
	}

	delete(c.ingresses, key)

	return c.rebuildHosts()
}

// AddOrUpdateVirtualServer adds or updates the VirtualServer resource.
func (c *Configuration) AddOrUpdateVirtualServer(vs *conf_v1.VirtualServer) ([]ResourceChange, []ConfigurationProblem) {
	c.lock.Lock()
	defer c.lock.Unlock()

	key := getResourceKey(&vs.ObjectMeta)
	var validationError error

	if !c.hasCorrectIngressClass(vs) {
		delete(c.virtualServers, key)
	} else {
		validationError = c.virtualServerValidator.ValidateVirtualServer(vs)
		if validationError != nil {
			delete(c.virtualServers, key)
		} else {
			c.virtualServers[key] = vs
		}
	}

	changes, problems := c.rebuildHosts()

	if validationError != nil {
		// If the invalid resource has an active host, rebuildHosts will create a change
		// to remove the resource.
		// Here we add the validationErr to that change.
		kind := getResourceKeyWithKind(virtualServerKind, &vs.ObjectMeta)
		for i := range changes {
			k := changes[i].Resource.GetKeyWithKind()

			if k == kind {
				changes[i].Error = validationError.Error()
				return changes, problems
			}
		}

		// On the other hand, the invalid resource might not have any active host.
		// Or the resource was invalid before and is still invalid (in some different way).
		// In those cases,  rebuildHosts will create no change for that resource.
		// To make sure the validationErr is reported to the user, we create a problem.
		p := ConfigurationProblem{
			Object:  vs,
			IsError: true,
			Reason:  "Rejected",
			Message: fmt.Sprintf("VirtualServer %s was rejected with error: %s", getResourceKey(&vs.ObjectMeta), validationError.Error()),
		}
		problems = append(problems, p)
	}

	return changes, problems
}

// DeleteVirtualServer deletes a VirtualServerResource by the key.
func (c *Configuration) DeleteVirtualServer(key string) ([]ResourceChange, []ConfigurationProblem) {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, exists := c.virtualServers[key]
	if !exists {
		return nil, nil
	}

	delete(c.virtualServers, key)

	return c.rebuildHosts()
}

// AddOrUpdateVirtualServerRoute adds or updates the VirtualServerRoute.
func (c *Configuration) AddOrUpdateVirtualServerRoute(vsr *conf_v1.VirtualServerRoute) ([]ResourceChange, []ConfigurationProblem) {
	c.lock.Lock()
	defer c.lock.Unlock()

	key := getResourceKey(&vsr.ObjectMeta)
	var validationError error

	if !c.hasCorrectIngressClass(vsr) {
		delete(c.virtualServerRoutes, key)
	} else {
		validationError = c.virtualServerValidator.ValidateVirtualServerRoute(vsr)
		if validationError != nil {
			delete(c.virtualServerRoutes, key)
		} else {
			c.virtualServerRoutes[key] = vsr
		}
	}

	changes, problems := c.rebuildHosts()

	if validationError != nil {
		p := ConfigurationProblem{
			Object:  vsr,
			IsError: true,
			Reason:  "Rejected",
			Message: fmt.Sprintf("VirtualServerRoute %s was rejected with error: %s", getResourceKey(&vsr.ObjectMeta), validationError.Error()),
		}
		problems = append(problems, p)
	}

	return changes, problems
}

// DeleteVirtualServerRoute deletes a VirtualServerRoute by the key.
func (c *Configuration) DeleteVirtualServerRoute(key string) ([]ResourceChange, []ConfigurationProblem) {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, exists := c.virtualServerRoutes[key]
	if !exists {
		return nil, nil
	}

	delete(c.virtualServerRoutes, key)

	return c.rebuildHosts()
}

// AddOrUpdateGlobalConfiguration adds or updates the GlobalConfiguration.
func (c *Configuration) AddOrUpdateGlobalConfiguration(gc *conf_v1alpha1.GlobalConfiguration) ([]ResourceChange, []ConfigurationProblem, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	validationErr := c.globalConfigurationValidator.ValidateGlobalConfiguration(gc)
	if validationErr != nil {
		c.globalConfiguration = nil
	} else {
		c.globalConfiguration = gc
	}

	changes, problems := c.rebuildListeners()

	return changes, problems, validationErr
}

// DeleteGlobalConfiguration deletes GlobalConfiguration.
func (c *Configuration) DeleteGlobalConfiguration() ([]ResourceChange, []ConfigurationProblem) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.globalConfiguration = nil
	changes, problems := c.rebuildListeners()

	return changes, problems
}

// GetGlobalConfiguration returns the current GlobalConfiguration.
func (c *Configuration) GetGlobalConfiguration() *conf_v1alpha1.GlobalConfiguration {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.globalConfiguration
}

// AddOrUpdateTransportServer adds or updates the TransportServer.
func (c *Configuration) AddOrUpdateTransportServer(ts *conf_v1alpha1.TransportServer) ([]ResourceChange, []ConfigurationProblem) {
	c.lock.Lock()
	defer c.lock.Unlock()

	key := getResourceKey(&ts.ObjectMeta)
	var validationErr error

	if !c.hasCorrectIngressClass(ts) {
		delete(c.transportServers, key)
	} else {
		validationErr = c.transportServerValidator.ValidateTransportServer(ts)
		if validationErr != nil {
			delete(c.transportServers, key)
		} else {
			c.transportServers[key] = ts
		}
	}

	changes, problems := c.rebuildListeners()

	if c.isTLSPassthroughEnabled {
		hostChanges, hostProblems := c.rebuildHosts()

		changes = append(changes, hostChanges...)
		problems = append(problems, hostProblems...)
	}

	if validationErr != nil {
		// If the invalid resource has an active host/listener, rebuildHosts/rebuildListeners will create a change
		// to remove the resource.
		// Here we add the validationErr to that change.
		kind := getResourceKeyWithKind(transportServerKind, &ts.ObjectMeta)
		for i := range changes {
			k := changes[i].Resource.GetKeyWithKind()

			if k == kind {
				changes[i].Error = validationErr.Error()
				return changes, problems
			}
		}

		// On the other hand, the invalid resource might not have any active host/listener.
		// Or the resource was invalid before and is still invalid (in some different way).
		// In those cases,  rebuildHosts/rebuildListeners will create no change for that resource.
		// To make sure the validationErr is reported to the user, we create a problem.
		p := ConfigurationProblem{
			Object:  ts,
			IsError: true,
			Reason:  "Rejected",
			Message: fmt.Sprintf("TransportServer %s was rejected with error: %s", getResourceKey(&ts.ObjectMeta), validationErr.Error()),
		}
		problems = append(problems, p)
	}

	return changes, problems
}

// DeleteTransportServer deletes a TransportServer by the key.
func (c *Configuration) DeleteTransportServer(key string) ([]ResourceChange, []ConfigurationProblem) {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, exists := c.transportServers[key]
	if !exists {
		return nil, nil
	}

	delete(c.transportServers, key)

	changes, problems := c.rebuildListeners()

	if c.isTLSPassthroughEnabled {
		hostChanges, hostProblems := c.rebuildHosts()

		changes = append(changes, hostChanges...)
		problems = append(problems, hostProblems...)
	}

	return changes, problems
}

func (c *Configuration) rebuildListeners() ([]ResourceChange, []ConfigurationProblem) {
	newListeners, newTSConfigs := c.buildListenersAndTSConfigurations()

	removedListeners, updatedListeners, addedListeners := detectChangesInListeners(c.listeners, newListeners)
	changes := createResourceChangesForListeners(removedListeners, updatedListeners, addedListeners, c.listeners, newListeners)

	c.listeners = newListeners

	changes = squashResourceChanges(changes)

	// Note that the change will not refer to the latest version, if the TransportServerConfiguration is being removed.
	// However, referring to the latest version is necessary so that the resource latest Warnings are reported and not lost.
	// So here we make sure that changes always refer to the latest version of TransportServerConfigurations.
	for i := range changes {
		key := changes[i].Resource.GetKeyWithKind()
		if r, exists := newTSConfigs[key]; exists {
			changes[i].Resource = r
		}
	}

	newProblems := make(map[string]ConfigurationProblem)

	c.addProblemsForTSConfigsWithoutActiveListener(newTSConfigs, newProblems)

	newOrUpdatedProblems := detectChangesInProblems(newProblems, c.listenerProblems)

	// safe to update problems
	c.listenerProblems = newProblems

	return changes, newOrUpdatedProblems
}

func (c *Configuration) buildListenersAndTSConfigurations() (newListeners map[string]*TransportServerConfiguration, newTSConfigs map[string]*TransportServerConfiguration) {
	newListeners = make(map[string]*TransportServerConfiguration)
	newTSConfigs = make(map[string]*TransportServerConfiguration)

	for key, ts := range c.transportServers {
		if ts.Spec.Listener.Protocol == conf_v1alpha1.TLSPassthroughListenerProtocol {
			continue
		}

		tsc := NewTransportServerConfiguration(ts)
		newTSConfigs[key] = tsc

		if c.globalConfiguration == nil {
			continue
		}

		found := false
		var listener conf_v1alpha1.Listener
		for _, l := range c.globalConfiguration.Spec.Listeners {
			if ts.Spec.Listener.Name == l.Name && ts.Spec.Listener.Protocol == l.Protocol {
				listener = l
				found = true
				break
			}
		}

		if !found {
			continue
		}

		tsc.ListenerPort = listener.Port

		holder, exists := newListeners[listener.Name]
		if !exists {
			newListeners[listener.Name] = tsc
			continue
		}

		warning := fmt.Sprintf("listener %s is taken by another resource", listener.Name)

		if !holder.Wins(tsc) {
			holder.AddWarning(warning)
			newListeners[listener.Name] = tsc
		} else {
			tsc.AddWarning(warning)
		}
	}

	return newListeners, newTSConfigs
}

// GetResources returns all configuration resources.
func (c *Configuration) GetResources() []Resource {
	return c.GetResourcesWithFilter(resourceFilter{
		Ingresses:        true,
		VirtualServers:   true,
		TransportServers: true,
	})
}

type resourceFilter struct {
	Ingresses        bool
	VirtualServers   bool
	TransportServers bool
}

// GetResourcesWithFilter returns resources using the filter.
func (c *Configuration) GetResourcesWithFilter(filter resourceFilter) []Resource {
	c.lock.RLock()
	defer c.lock.RUnlock()

	resources := make(map[string]Resource)

	for _, r := range c.hosts {
		switch r.(type) {
		case *IngressConfiguration:
			if filter.Ingresses {
				resources[r.GetKeyWithKind()] = r
			}
		case *VirtualServerConfiguration:
			if filter.VirtualServers {
				resources[r.GetKeyWithKind()] = r
			}
		case *TransportServerConfiguration:
			if filter.TransportServers {
				resources[r.GetKeyWithKind()] = r
			}
		}
	}

	if filter.TransportServers {
		for _, r := range c.listeners {
			resources[r.GetKeyWithKind()] = r
		}
	}

	var result []Resource
	for _, key := range getSortedResourceKeys(resources) {
		result = append(result, resources[key])
	}

	return result
}

// FindResourcesForService finds resources that reference the specified service.
func (c *Configuration) FindResourcesForService(svcNamespace string, svcName string) []Resource {
	return c.findResourcesForResourceReference(svcNamespace, svcName, c.serviceReferenceChecker)
}

// FindResourcesForEndpoints finds resources that reference the specified endpoints.
func (c *Configuration) FindResourcesForEndpoints(endpointsNamespace string, endpointsName string) []Resource {
	// Resources reference not endpoints but the corresponding service, which has the same namespace and name
	return c.findResourcesForResourceReference(endpointsNamespace, endpointsName, c.endpointReferenceChecker)
}

// FindResourcesForSecret finds resources that reference the specified secret.
func (c *Configuration) FindResourcesForSecret(secretNamespace string, secretName string) []Resource {
	return c.findResourcesForResourceReference(secretNamespace, secretName, c.secretReferenceChecker)
}

// FindResourcesForPolicy finds resources that reference the specified policy.
func (c *Configuration) FindResourcesForPolicy(policyNamespace string, policyName string) []Resource {
	return c.findResourcesForResourceReference(policyNamespace, policyName, c.policyReferenceChecker)
}

// FindResourcesForAppProtectPolicyAnnotation finds resources that reference the specified AppProtect policy via annotation.
func (c *Configuration) FindResourcesForAppProtectPolicyAnnotation(policyNamespace string, policyName string) []Resource {
	return c.findResourcesForResourceReference(policyNamespace, policyName, c.appPolicyReferenceChecker)
}

// FindResourcesForAppProtectLogConfAnnotation finds resources that reference the specified AppProtect LogConf.
func (c *Configuration) FindResourcesForAppProtectLogConfAnnotation(logConfNamespace string, logConfName string) []Resource {
	return c.findResourcesForResourceReference(logConfNamespace, logConfName, c.appLogConfReferenceChecker)
}

// FindResourcesForAppProtectDosProtected finds resources that reference the specified AppProtectDos DosLogConf.
func (c *Configuration) FindResourcesForAppProtectDosProtected(namespace string, name string) []Resource {
	return c.findResourcesForResourceReference(namespace, name, c.appDosProtectedChecker)
}

func (c *Configuration) findResourcesForResourceReference(namespace string, name string, checker resourceReferenceChecker) []Resource {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var result []Resource

	for _, h := range getSortedResourceKeys(c.hosts) {
		r := c.hosts[h]

		switch impl := r.(type) {
		case *IngressConfiguration:
			if checker.IsReferencedByIngress(namespace, name, impl.Ingress) {
				result = append(result, r)
				continue
			}

			for _, fm := range impl.Minions {
				if checker.IsReferencedByMinion(namespace, name, fm.Ingress) {
					result = append(result, r)
					break
				}
			}
		case *VirtualServerConfiguration:
			if checker.IsReferencedByVirtualServer(namespace, name, impl.VirtualServer) {
				result = append(result, r)
				continue
			}

			for _, vsr := range impl.VirtualServerRoutes {
				if checker.IsReferencedByVirtualServerRoute(namespace, name, vsr) {
					result = append(result, r)
					break
				}
			}
		case *TransportServerConfiguration:
			if checker.IsReferencedByTransportServer(namespace, name, impl.TransportServer) {
				result = append(result, r)
				continue
			}
		}
	}

	for _, l := range getSortedTransportServerConfigurationKeys(c.listeners) {
		tsConfig := c.listeners[l]

		if checker.IsReferencedByTransportServer(namespace, name, tsConfig.TransportServer) {
			result = append(result, tsConfig)
			continue
		}
	}

	return result
}

func getResourceKey(meta *metav1.ObjectMeta) string {
	return fmt.Sprintf("%s/%s", meta.Namespace, meta.Name)
}

// rebuildHosts rebuilds the Configuration and returns the changes to it and the new problems.
func (c *Configuration) rebuildHosts() ([]ResourceChange, []ConfigurationProblem) {
	newHosts, newResources := c.buildHostsAndResources()

	updateActiveHostsForIngresses(newHosts, newResources)

	removedHosts, updatedHosts, addedHosts := detectChangesInHosts(c.hosts, newHosts)
	changes := createResourceChangesForHosts(removedHosts, updatedHosts, addedHosts, c.hosts, newHosts)

	// safe to update hosts
	c.hosts = newHosts

	changes = squashResourceChanges(changes)

	// Note that the change will not refer to the latest version, if the resource is being removed.
	// However, referring to the latest version is necessary so that the resource latest Warnings are reported and not lost.
	// So here we make sure that changes always refer to the latest version of resources.
	for i := range changes {
		key := changes[i].Resource.GetKeyWithKind()
		if r, exists := newResources[key]; exists {
			changes[i].Resource = r
		}
	}

	newProblems := make(map[string]ConfigurationProblem)

	c.addProblemsForResourcesWithoutActiveHost(newResources, newProblems)
	c.addProblemsForOrphanMinions(newProblems)
	c.addProblemsForOrphanOrIgnoredVsrs(newProblems)

	newOrUpdatedProblems := detectChangesInProblems(newProblems, c.hostProblems)

	// safe to update problems
	c.hostProblems = newProblems

	return changes, newOrUpdatedProblems
}

func updateActiveHostsForIngresses(hosts map[string]Resource, resources map[string]Resource) {
	for _, r := range resources {
		ingConfig, ok := r.(*IngressConfiguration)
		if !ok {
			continue
		}

		for _, rule := range ingConfig.Ingress.Spec.Rules {
			res := hosts[rule.Host]
			ingConfig.ValidHosts[rule.Host] = res.GetKeyWithKind() == r.GetKeyWithKind()
		}
	}
}

func detectChangesInProblems(newProblems map[string]ConfigurationProblem, oldProblems map[string]ConfigurationProblem) []ConfigurationProblem {
	var result []ConfigurationProblem

	for _, key := range getSortedProblemKeys(newProblems) {
		newP := newProblems[key]

		oldP, exists := oldProblems[key]
		if !exists {
			result = append(result, newP)
			continue
		}

		if !compareConfigurationProblems(&newP, &oldP) {
			result = append(result, newP)
		}
	}

	return result
}

func (c *Configuration) addProblemsForTSConfigsWithoutActiveListener(tsConfigs map[string]*TransportServerConfiguration, problems map[string]ConfigurationProblem) {
	for _, tsc := range tsConfigs {
		holder, exists := c.listeners[tsc.TransportServer.Spec.Listener.Name]
		if !exists {
			p := ConfigurationProblem{
				Object:  tsc.TransportServer,
				IsError: false,
				Reason:  "Rejected",
				Message: fmt.Sprintf("Listener %s doesn't exist", tsc.TransportServer.Spec.Listener.Name),
			}
			problems[tsc.GetKeyWithKind()] = p
			continue
		}

		if !tsc.IsEqual(holder) {
			p := ConfigurationProblem{
				Object:  tsc.TransportServer,
				IsError: false,
				Reason:  "Rejected",
				Message: fmt.Sprintf("Listener %s is taken by another resource", tsc.TransportServer.Spec.Listener.Name),
			}
			problems[tsc.GetKeyWithKind()] = p
		}
	}
}

func (c *Configuration) addProblemsForResourcesWithoutActiveHost(resources map[string]Resource, problems map[string]ConfigurationProblem) {
	for _, r := range resources {
		switch impl := r.(type) {
		case *IngressConfiguration:
			atLeastOneValidHost := false
			for _, v := range impl.ValidHosts {
				if v {
					atLeastOneValidHost = true
					break
				}
			}
			if !atLeastOneValidHost {
				p := ConfigurationProblem{
					Object:  impl.Ingress,
					IsError: false,
					Reason:  "Rejected",
					Message: "All hosts are taken by other resources",
				}
				problems[r.GetKeyWithKind()] = p
			}
		case *VirtualServerConfiguration:
			res := c.hosts[impl.VirtualServer.Spec.Host]

			if res.GetKeyWithKind() != r.GetKeyWithKind() {
				p := ConfigurationProblem{
					Object:  impl.VirtualServer,
					IsError: false,
					Reason:  "Rejected",
					Message: "Host is taken by another resource",
				}
				problems[r.GetKeyWithKind()] = p
			}
		case *TransportServerConfiguration:
			res := c.hosts[impl.TransportServer.Spec.Host]

			if res.GetKeyWithKind() != r.GetKeyWithKind() {
				p := ConfigurationProblem{
					Object:  impl.TransportServer,
					IsError: false,
					Reason:  "Rejected",
					Message: "Host is taken by another resource",
				}
				problems[r.GetKeyWithKind()] = p
			}
		}
	}
}

func (c *Configuration) addProblemsForOrphanMinions(problems map[string]ConfigurationProblem) {
	for _, key := range getSortedIngressKeys(c.ingresses) {
		ing := c.ingresses[key]

		if !isMinion(ing) {
			continue
		}

		r, exists := c.hosts[ing.Spec.Rules[0].Host]
		ingressConf, ok := r.(*IngressConfiguration)

		if !exists || !ok || !ingressConf.IsMaster {
			p := ConfigurationProblem{
				Object:  ing,
				IsError: false,
				Reason:  "NoIngressMasterFound",
				Message: "Ingress master is invalid or doesn't exist",
			}
			k := getResourceKeyWithKind(ingressKind, &ing.ObjectMeta)
			problems[k] = p
		}
	}
}

func (c *Configuration) addProblemsForOrphanOrIgnoredVsrs(problems map[string]ConfigurationProblem) {
	for _, key := range getSortedVirtualServerRouteKeys(c.virtualServerRoutes) {
		vsr := c.virtualServerRoutes[key]

		r, exists := c.hosts[vsr.Spec.Host]
		vsConfig, ok := r.(*VirtualServerConfiguration)

		if !exists || !ok {
			p := ConfigurationProblem{
				Object:  vsr,
				IsError: false,
				Reason:  "NoVirtualServerFound",
				Message: "VirtualServer is invalid or doesn't exist",
			}
			k := getResourceKeyWithKind(virtualServerRouteKind, &vsr.ObjectMeta)
			problems[k] = p
			continue
		}

		found := false
		for _, v := range vsConfig.VirtualServerRoutes {
			if vsr.Namespace == v.Namespace && vsr.Name == v.Name {
				found = true
				break
			}
		}

		if !found {
			p := ConfigurationProblem{
				Object:  vsr,
				IsError: false,
				Reason:  "Ignored",
				Message: fmt.Sprintf("VirtualServer %s ignores VirtualServerRoute", getResourceKey(&vsConfig.VirtualServer.ObjectMeta)),
			}
			k := getResourceKeyWithKind(virtualServerRouteKind, &vsr.ObjectMeta)
			problems[k] = p
		}
	}
}

func getResourceKeyWithKind(kind string, objectMeta *metav1.ObjectMeta) string {
	return fmt.Sprintf("%s/%s/%s", kind, objectMeta.Namespace, objectMeta.Name)
}

func createResourceChangesForHosts(removedHosts []string, updatedHosts []string, addedHosts []string, oldHosts map[string]Resource, newHosts map[string]Resource) []ResourceChange {
	var changes []ResourceChange
	var deleteChanges []ResourceChange

	for _, h := range removedHosts {
		change := ResourceChange{
			Op:       Delete,
			Resource: oldHosts[h],
		}
		deleteChanges = append(deleteChanges, change)
	}

	for _, h := range updatedHosts {
		if oldHosts[h].GetKeyWithKind() != newHosts[h].GetKeyWithKind() {
			deleteChange := ResourceChange{
				Op:       Delete,
				Resource: oldHosts[h],
			}
			deleteChanges = append(deleteChanges, deleteChange)
		}

		change := ResourceChange{
			Op:       AddOrUpdate,
			Resource: newHosts[h],
		}
		changes = append(changes, change)
	}

	for _, h := range addedHosts {
		change := ResourceChange{
			Op:       AddOrUpdate,
			Resource: newHosts[h],
		}
		changes = append(changes, change)
	}

	// We need to ensure that delete changes come first.
	// This way an addOrUpdate change, which might include a resource that uses the same host as a resource
	// in a delete change, will be processed only after the config of the delete change is removed.
	// That will prevent any host collisions in the NGINX config in the state between the changes.
	return append(deleteChanges, changes...)
}

func createResourceChangesForListeners(removedListeners []string, updatedListeners []string, addedListeners []string, oldListeners map[string]*TransportServerConfiguration,
	newListeners map[string]*TransportServerConfiguration) []ResourceChange {
	var changes []ResourceChange
	var deleteChanges []ResourceChange

	for _, l := range removedListeners {
		change := ResourceChange{
			Op:       Delete,
			Resource: oldListeners[l],
		}
		deleteChanges = append(deleteChanges, change)
	}

	for _, l := range updatedListeners {
		if oldListeners[l].GetKeyWithKind() != newListeners[l].GetKeyWithKind() {
			deleteChange := ResourceChange{
				Op:       Delete,
				Resource: oldListeners[l],
			}
			deleteChanges = append(deleteChanges, deleteChange)
		}

		change := ResourceChange{
			Op:       AddOrUpdate,
			Resource: newListeners[l],
		}
		changes = append(changes, change)
	}

	for _, l := range addedListeners {
		change := ResourceChange{
			Op:       AddOrUpdate,
			Resource: newListeners[l],
		}
		changes = append(changes, change)
	}

	// We need to ensure that delete changes come first.
	// This way an addOrUpdate change, which might include a resource that uses the same listener as a resource
	// in a delete change, will be processed only after the config of the delete change is removed.
	// That will prevent any listener collisions in the NGINX config in the state between the changes.
	return append(deleteChanges, changes...)
}

func squashResourceChanges(changes []ResourceChange) []ResourceChange {
	// deletes for the same resource become a single delete
	// updates for the same resource become a single update
	// delete and update for the same resource become a single update

	var deletes []ResourceChange
	var updates []ResourceChange

	changesPerResource := make(map[string][]ResourceChange)

	for _, c := range changes {
		key := c.Resource.GetKeyWithKind()
		changesPerResource[key] = append(changesPerResource[key], c)
	}

	// we range over the changes again to preserver the original order
	for _, c := range changes {
		key := c.Resource.GetKeyWithKind()
		resChanges, exists := changesPerResource[key]

		if !exists {
			continue
		}

		// the last element will be an update (if it exists) or a delete
		squashedChanged := resChanges[len(resChanges)-1]
		if squashedChanged.Op == Delete {
			deletes = append(deletes, squashedChanged)
		} else {
			updates = append(updates, squashedChanged)
		}

		delete(changesPerResource, key)
	}

	// We need to ensure that delete changes come first.
	// This way an addOrUpdate change, which might include a resource that uses the same host/listener as a resource
	// in a delete change, will be processed only after the config of the delete change is removed.
	// That will prevent any host/listener collisions in the NGINX config in the state between the changes.
	return append(deletes, updates...)
}

func (c *Configuration) buildHostsAndResources() (newHosts map[string]Resource, newResources map[string]Resource) {
	newHosts = make(map[string]Resource)
	newResources = make(map[string]Resource)

	// Step 1 - Build hosts from Ingress resources

	for _, key := range getSortedIngressKeys(c.ingresses) {
		ing := c.ingresses[key]

		if isMinion(ing) {
			continue
		}

		var resource *IngressConfiguration

		if isMaster(ing) {
			minions, childWarnings := c.buildMinionConfigs(ing.Spec.Rules[0].Host)
			resource = NewMasterIngressConfiguration(ing, minions, childWarnings)
		} else {
			resource = NewRegularIngressConfiguration(ing)
		}

		newResources[resource.GetKeyWithKind()] = resource

		for _, rule := range ing.Spec.Rules {
			holder, exists := newHosts[rule.Host]
			if !exists {
				newHosts[rule.Host] = resource
				continue
			}

			warning := fmt.Sprintf("host %s is taken by another resource", rule.Host)

			if !holder.Wins(resource) {
				holder.AddWarning(warning)
				newHosts[rule.Host] = resource
			} else {
				resource.AddWarning(warning)
			}
		}
	}

	// Step 2 - Build hosts from VirtualServer resources

	for _, key := range getSortedVirtualServerKeys(c.virtualServers) {
		vs := c.virtualServers[key]

		vsrs, warnings := c.buildVirtualServerRoutes(vs)
		resource := NewVirtualServerConfiguration(vs, vsrs, warnings)

		newResources[resource.GetKeyWithKind()] = resource

		holder, exists := newHosts[vs.Spec.Host]
		if !exists {
			newHosts[vs.Spec.Host] = resource
			continue
		}

		warning := fmt.Sprintf("host %s is taken by another resource", vs.Spec.Host)

		if !holder.Wins(resource) {
			newHosts[vs.Spec.Host] = resource
			holder.AddWarning(warning)
		} else {
			resource.AddWarning(warning)
		}
	}

	// Step - 3 - Build hosts from TransportServer resources if TLS Passthrough is enabled

	if c.isTLSPassthroughEnabled {
		for _, key := range getSortedTransportServerKeys(c.transportServers) {
			ts := c.transportServers[key]

			if ts.Spec.Listener.Name != conf_v1alpha1.TLSPassthroughListenerName && ts.Spec.Listener.Protocol != conf_v1alpha1.TLSPassthroughListenerProtocol {
				continue
			}

			resource := NewTransportServerConfiguration(ts)
			newResources[resource.GetKeyWithKind()] = resource

			holder, exists := newHosts[ts.Spec.Host]
			if !exists {
				newHosts[ts.Spec.Host] = resource
				continue
			}

			warning := fmt.Sprintf("host %s is taken by another resource", ts.Spec.Host)

			if !holder.Wins(resource) {
				newHosts[ts.Spec.Host] = resource
				holder.AddWarning(warning)
			} else {
				resource.AddWarning(warning)
			}
		}
	}

	return newHosts, newResources
}

func (c *Configuration) buildMinionConfigs(masterHost string) ([]*MinionConfiguration, map[string][]string) {
	var minionConfigs []*MinionConfiguration
	childWarnings := make(map[string][]string)
	paths := make(map[string]*MinionConfiguration)

	for _, minionKey := range getSortedIngressKeys(c.ingresses) {
		ingress := c.ingresses[minionKey]

		if !isMinion(ingress) {
			continue
		}

		if masterHost != ingress.Spec.Rules[0].Host {
			continue
		}

		minionConfig := NewMinionConfiguration(ingress)

		for _, p := range ingress.Spec.Rules[0].HTTP.Paths {
			holder, exists := paths[p.Path]
			if !exists {
				paths[p.Path] = minionConfig
				minionConfig.ValidPaths[p.Path] = true
				continue
			}

			warning := fmt.Sprintf("path %s is taken by another resource", p.Path)

			if !chooseObjectMetaWinner(&holder.Ingress.ObjectMeta, &ingress.ObjectMeta) {
				paths[p.Path] = minionConfig
				minionConfig.ValidPaths[p.Path] = true

				holder.ValidPaths[p.Path] = false
				key := getResourceKey(&holder.Ingress.ObjectMeta)
				childWarnings[key] = append(childWarnings[key], warning)
			} else {
				key := getResourceKey(&minionConfig.Ingress.ObjectMeta)
				childWarnings[key] = append(childWarnings[key], warning)
			}
		}

		minionConfigs = append(minionConfigs, minionConfig)
	}

	return minionConfigs, childWarnings
}

func (c *Configuration) buildVirtualServerRoutes(vs *conf_v1.VirtualServer) ([]*conf_v1.VirtualServerRoute, []string) {
	var vsrs []*conf_v1.VirtualServerRoute
	var warnings []string

	for _, r := range vs.Spec.Routes {
		if r.Route == "" {
			continue
		}

		vsrKey := r.Route

		// if route is defined without a namespace, use the namespace of VirtualServer.
		if !strings.Contains(r.Route, "/") {
			vsrKey = fmt.Sprintf("%s/%s", vs.Namespace, r.Route)
		}

		vsr, exists := c.virtualServerRoutes[vsrKey]
		if !exists {
			warning := fmt.Sprintf("VirtualServerRoute %s doesn't exist or invalid", vsrKey)
			warnings = append(warnings, warning)
			continue
		}

		err := c.virtualServerValidator.ValidateVirtualServerRouteForVirtualServer(vsr, vs.Spec.Host, r.Path)
		if err != nil {
			warning := fmt.Sprintf("VirtualServerRoute %s is invalid: %v", vsrKey, err)
			warnings = append(warnings, warning)
			continue
		}

		vsrs = append(vsrs, vsr)
	}

	return vsrs, warnings
}

// GetTransportServerMetrics returns metrics about TransportServers
func (c *Configuration) GetTransportServerMetrics() *TransportServerMetrics {
	var metrics TransportServerMetrics

	if c.isTLSPassthroughEnabled {
		for _, resource := range c.hosts {
			_, ok := resource.(*TransportServerConfiguration)
			if ok {
				metrics.TotalTLSPassthrough++
			}
		}
	}

	for _, tsConfig := range c.listeners {
		if tsConfig.TransportServer.Spec.Listener.Protocol == "TCP" {
			metrics.TotalTCP++
		} else {
			metrics.TotalUDP++
		}
	}

	return &metrics
}

func getSortedIngressKeys(m map[string]*networking.Ingress) []string {
	var keys []string

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func getSortedVirtualServerKeys(m map[string]*conf_v1.VirtualServer) []string {
	var keys []string

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func getSortedVirtualServerRouteKeys(m map[string]*conf_v1.VirtualServerRoute) []string {
	var keys []string

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func getSortedProblemKeys(m map[string]ConfigurationProblem) []string {
	var keys []string

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func getSortedResourceKeys(m map[string]Resource) []string {
	var keys []string

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func getSortedTransportServerKeys(m map[string]*conf_v1alpha1.TransportServer) []string {
	var keys []string

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func getSortedTransportServerConfigurationKeys(m map[string]*TransportServerConfiguration) []string {
	var keys []string

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func detectChangesInHosts(oldHosts map[string]Resource, newHosts map[string]Resource) (removedHosts []string, updatedHosts []string, addedHosts []string) {
	for _, h := range getSortedResourceKeys(oldHosts) {
		_, exists := newHosts[h]
		if !exists {
			removedHosts = append(removedHosts, h)
		}
	}

	for _, h := range getSortedResourceKeys(newHosts) {
		_, exists := oldHosts[h]
		if !exists {
			addedHosts = append(addedHosts, h)
		}
	}

	for _, h := range getSortedResourceKeys(newHosts) {
		oldR, exists := oldHosts[h]
		if !exists {
			continue
		}

		if !oldR.IsEqual(newHosts[h]) {
			updatedHosts = append(updatedHosts, h)
		}
	}

	return removedHosts, updatedHosts, addedHosts
}

func detectChangesInListeners(oldListeners map[string]*TransportServerConfiguration, newListeners map[string]*TransportServerConfiguration) (removedListeners []string,
	updatedListeners []string, addedListeners []string) {
	for _, l := range getSortedTransportServerConfigurationKeys(oldListeners) {
		_, exists := newListeners[l]
		if !exists {
			removedListeners = append(removedListeners, l)
		}
	}

	for _, l := range getSortedTransportServerConfigurationKeys(newListeners) {
		_, exists := oldListeners[l]
		if !exists {
			addedListeners = append(addedListeners, l)
		}
	}

	for _, l := range getSortedTransportServerConfigurationKeys(newListeners) {
		oldR, exists := oldListeners[l]
		if !exists {
			continue
		}

		if !oldR.IsEqual(newListeners[l]) {
			updatedListeners = append(updatedListeners, l)
		}
	}

	return removedListeners, updatedListeners, addedListeners
}
