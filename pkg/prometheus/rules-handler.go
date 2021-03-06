package prometheus

import (
	"fmt"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/mitchellh/mapstructure"
)

// RuleHandler is a Grizzly Provider for Grafana datasources
type RuleHandler struct{}

// NewRuleHandler returns configuration defining a new Grafana Provider
func NewRuleHandler() *RuleHandler {
	return &RuleHandler{}
}

// GetName returns the name for this provider
func (h *RuleHandler) GetName() string {
	return "prometheus"
}

// GetFullName returns the name for this provider
func (h *RuleHandler) GetFullName() string {
	return "prometheus.rulegroup"
}

const prometheusAlertsPath = "prometheusAlerts"
const prometheusRulesPath = "prometheusRules"

// GetJSONPaths returns paths within Jsonnet output that this provider will consume
func (h *RuleHandler) GetJSONPaths() []string {
	return []string{
		prometheusAlertsPath,
		prometheusRulesPath,
	}
}

// GetExtension returns the file name extension for a rule grouping
func (h *RuleHandler) GetExtension() string {
	return "yaml"
}

func (h *RuleHandler) newRuleGroupingResource(path string, group RuleGroup) grizzly.Resource {
	resource := grizzly.Resource{
		UID:      group.UID(),
		Filename: group.UID(),
		Handler:  h,
		Detail:   group,
		JSONPath: path,
	}
	return resource
}

// Parse parses an interface{} object into a struct for this resource type
func (h *RuleHandler) Parse(path string, i interface{}) (grizzly.ResourceList, error) {
	resources := grizzly.ResourceList{}
	msi := i.(map[string]interface{})
	groupings := map[string]RuleGrouping{}
	err := mapstructure.Decode(msi, &groupings)
	if err != nil {
		return nil, err
	}
	for k, grouping := range groupings {
		for _, group := range grouping.Groups {
			group.Namespace = k
			resource := h.newRuleGroupingResource(path, group)
			key := resource.Key()
			resources[key] = resource
		}
	}
	return resources, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *RuleHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *RuleHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	return &resource
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *RuleHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	group, err := getRemoteRuleGroup(UID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving datasource %s: %v", UID, err)
	}
	resource := h.newRuleGroupingResource(prometheusAlertsPath, *group)
	return &resource, nil
}

// GetRepresentation renders a resource as JSON or YAML as appropriate
func (h *RuleHandler) GetRepresentation(uid string, resource grizzly.Resource) (string, error) {
	g := resource.Detail.(RuleGroup)
	return g.toYAML()
}

// GetRemoteRepresentation retrieves a datasource as JSON
func (h *RuleHandler) GetRemoteRepresentation(uid string) (string, error) {
	group, err := getRemoteRuleGroup(uid)
	if err != nil {
		return "", err
	}
	return group.toYAML()
}

// GetRemote retrieves a datasource as a Resource
func (h *RuleHandler) GetRemote(uid string) (*grizzly.Resource, error) {
	group, err := getRemoteRuleGroup(uid)
	if err != nil {
		return nil, err
	}
	resource := h.newRuleGroupingResource("", *group)
	return &resource, nil
}

// Add pushes a datasource to Grafana via the API
func (h *RuleHandler) Add(resource grizzly.Resource) error {
	g := resource.Detail.(RuleGroup)
	return writeRuleGroup(g)
}

// Update pushes a datasource to Grafana via the API
func (h *RuleHandler) Update(existing, resource grizzly.Resource) error {
	g := resource.Detail.(RuleGroup)
	return writeRuleGroup(g)
}

// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
func (h *RuleHandler) Preview(resource grizzly.Resource, notifier grizzly.Notifier, opts *grizzly.PreviewOpts) error {
	return grizzly.ErrNotImplemented
}
