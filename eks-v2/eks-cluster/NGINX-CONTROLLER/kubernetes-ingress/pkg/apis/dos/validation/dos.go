package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	validation2 "github.com/nginxinc/kubernetes-ingress/pkg/apis/configuration/validation"
	"github.com/nginxinc/kubernetes-ingress/pkg/apis/dos/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var appProtectDosPolicyRequiredFields = [][]string{
	{"spec"},
}

var appProtectDosLogConfRequiredFields = [][]string{
	{"spec", "content"},
	{"spec", "filter"},
}

const maxNameLength = 63

// ValidateDosProtectedResource validates a dos protected resource.
func ValidateDosProtectedResource(protected *v1beta1.DosProtectedResource) error {
	var err error

	// name
	if protected.Spec.Name == "" {
		return fmt.Errorf("error validating DosProtectedResource: %v missing value for field: %v", protected.Name, "name")
	}
	err = validateAppProtectDosName(protected.Spec.Name)
	if err != nil {
		return fmt.Errorf("error validating DosProtectedResource: %v invalid field: %v err: %w", protected.Name, "name", err)
	}

	// apDosMonitor
	if protected.Spec.ApDosMonitor != nil {
		err = validateAppProtectDosMonitor(*protected.Spec.ApDosMonitor)
		if err != nil {
			return fmt.Errorf("error validating DosProtectedResource: %v invalid field: %v err: %w", protected.Name, "apDosMonitor", err)
		}
	}

	// dosAccessLogDest
	if protected.Spec.DosAccessLogDest == "" {
		return fmt.Errorf("error validating DosProtectedResource: %v missing value for field: %v", protected.Name, "dosAccessLogDest")
	}
	err = validateAppProtectDosLogDest(protected.Spec.DosAccessLogDest)
	if err != nil {
		return fmt.Errorf("error validating DosProtectedResource: %v invalid field: %v err: %w", protected.Name, "dosAccessLogDest", err)
	}

	// apDosPolicy
	if protected.Spec.ApDosPolicy != "" {
		err = validateResourceReference(protected.Spec.ApDosPolicy)
		if err != nil {
			return fmt.Errorf("error validating DosProtectedResource: %v invalid field: %v err: %w", protected.Name, "apDosPolicy", err)
		}
	}

	// dosSecurityLog
	if protected.Spec.DosSecurityLog != nil {
		// dosLogDest
		err = validateAppProtectDosLogDest(protected.Spec.DosSecurityLog.DosLogDest)
		if err != nil {
			return fmt.Errorf("error validating DosProtectedResource: %v invalid field: %v err: %w", protected.Name, "dosSecurityLog/dosLogDest", err)
		}
		// apDosLogConf
		err = validateResourceReference(protected.Spec.DosSecurityLog.ApDosLogConf)
		if err != nil {
			return fmt.Errorf("error validating DosProtectedResource: %v invalid field: %v err: %w", protected.Name, "dosSecurityLog/apDosLogConf", err)
		}
	}

	return nil
}

// validateResourceReference validates a resource reference. A valid resource can be either namespace/name or name.
func validateResourceReference(ref string) error {
	errs := validation.IsQualifiedName(ref)
	if len(errs) != 0 {
		return fmt.Errorf("reference name is invalid: %v", ref)
	}

	return nil
}

// ValidateAppProtectDosLogConf validates LogConfiguration resource
func ValidateAppProtectDosLogConf(logConf *unstructured.Unstructured) error {
	lcName := logConf.GetName()
	err := validation2.ValidateRequiredFields(logConf, appProtectDosLogConfRequiredFields)
	if err != nil {
		return fmt.Errorf("error validating App Protect Dos Log Configuration %v: %w", lcName, err)
	}

	return nil
}

var (
	validDNSRegex       = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9-]{1,62}\.)([A-Za-z0-9-]{1,63}\.)*[A-Za-z]{2,6}:\d{1,5}$`)
	validIPRegex        = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}:\d{1,5}$`)
	validLocalhostRegex = regexp.MustCompile(`^localhost:\d{1,5}$`)
)

func validateAppProtectDosLogDest(dstAntn string) error {
	if validIPRegex.MatchString(dstAntn) || validDNSRegex.MatchString(dstAntn) || validLocalhostRegex.MatchString(dstAntn) {
		chunks := strings.Split(dstAntn, ":")
		err := validatePort(chunks[1])
		if err != nil {
			return fmt.Errorf("invalid log destination: %w", err)
		}
		return nil
	}
	if dstAntn == "stderr" {
		return nil
	}

	return fmt.Errorf("invalid log destination: %s, must follow format: <ip-address | localhost | dns name>:<port> or stderr", dstAntn)
}

func validatePort(value string) error {
	port, _ := strconv.Atoi(value)
	if port > 65535 || port < 1 {
		return fmt.Errorf("error parsing port: %v not a valid port number", port)
	}
	return nil
}

func validateAppProtectDosName(name string) error {
	if len(name) > maxNameLength {
		return fmt.Errorf("app Protect Dos Name max length is %v", maxNameLength)
	}

	return validation2.ValidateEscapedString(name, "protected-object-one")
}

var validMonitorProtocol = map[string]bool{
	"http1": true,
	"http2": true,
	"grpc":  true,
}

func validateAppProtectDosMonitor(apDosMonitor v1beta1.ApDosMonitor) error {
	_, err := url.Parse(apDosMonitor.URI)
	if err != nil {
		return fmt.Errorf("app Protect Dos Monitor must have valid URL")
	}

	if err := validation2.ValidateEscapedString(apDosMonitor.URI, "http://www.example.com"); err != nil {
		return err
	}

	if apDosMonitor.Protocol != "" {
		allErrs := field.ErrorList{}
		fieldPath := field.NewPath("dosMonitorProtocol")
		allErrs = append(allErrs, validation2.ValidateParameter(apDosMonitor.Protocol, validMonitorProtocol, fieldPath)...)
		err := allErrs.ToAggregate()
		if err != nil {
			return fmt.Errorf("app Protect Dos Monitor Protocol must be: %v", err)
		}
	}

	return nil
}

// ValidateAppProtectDosPolicy validates Policy resource.
func ValidateAppProtectDosPolicy(policy *unstructured.Unstructured) error {
	polName := policy.GetName()

	err := validation2.ValidateRequiredFields(policy, appProtectDosPolicyRequiredFields)
	if err != nil {
		return fmt.Errorf("error validating DosPolicy %v: %w", polName, err)
	}

	return nil
}
