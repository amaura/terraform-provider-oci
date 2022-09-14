package commonexport

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/oracle/terraform-provider-oci/internal/utils"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	tf_client "github.com/oracle/terraform-provider-oci/internal/client"
)

type TerraformResourceGraph map[string][]TerraformResourceAssociation
type TerraformResourceAssociation struct {
	*TerraformResourceHints
	DatasourceQueryParams map[string]string // Mapping of data source inputs and the source attribute from a parent resource
}
type TerraformResourceHints struct {
	// Information about this resource
	ResourceClass        string // The name of the resource class (e.g. oci_core_vcn)
	ResourceAbbreviation string // An abbreviated version of the resource class used for generating shorter resource names (e.g. vcn)

	// Hints to help with discovering this resource using data sources
	DatasourceClass              string                  // The name of the data source class to use for discovering resources (e.g. oci_core_vcns)
	DatasourceItemsAttr          string                  // The attribute with the data source that contains the discovered resources returned by the data source (e.g. virtual_networks)
	IsDatasourceCollection       bool                    // True if list datasource is modeled as a collection with `items` field under DatasourceItemsAttr
	RequireResourceRefresh       bool                    // Whether to use the resource to fill in missing information from datasource (e.g. when datasources only return summary information)
	DiscoverableLifecycleStates  []string                // List of lifecycle states that should be discovered. If empty, then all lifecycle states are discoverable.
	ProcessDiscoveredResourcesFn ProcessOCIResourcesFunc // Custom function for processing resources discovered by the data source
	AlwaysExportable             bool                    // Some resources always need to be exportable, regardless of whether they are being targeted for export
	IsDataSource                 bool
	GetIdFn                      func(*OCIResource) (string, error) // If the resource has no OCID generated by services, then implement this to generate one from the OCIResource. Typically used for composite IDs.

	// Override function for discovering resources. To be used when there is no datasource implementation to help with discovery.
	FindResourcesOverrideFn func(*ResourceDiscoveryContext, *TerraformResourceAssociation, *OCIResource, *TerraformResourceGraph) ([]*OCIResource, error)

	// Hints to help with generating HCL representation from this resource
	GetHCLStringOverrideFn func(*strings.Builder, *OCIResource, map[string]string) error // Custom function for generating HCL syntax for the resource

	// Hints for adding default value to HCL representation for attributes not found in resource discovery
	DefaultValuesForMissingAttributes map[string]interface{}

	// Hints for adding resource attributes to `ignore_changes` in HCL representation
	// This is added to avoid plan failure/diff for attributes that service does not return in read response
	// The attributes references are interpolated in case of nested attributes
	IgnorableRequiredMissingAttributes map[string]bool
}
type ProcessOCIResourcesFunc func(*ResourceDiscoveryContext, []*OCIResource) ([]*OCIResource, error)
type OCIResource struct {
	TerraformResource
	CompartmentId    string
	RawResource      interface{}
	SourceAttributes map[string]interface{}
	GetHclStringFn   func(*strings.Builder, *OCIResource, map[string]string) error
	Parent           *OCIResource
	IsErrorResource  bool
}
type TfHclVersion11 struct {
	Value TfVersionEnum
}
type TfVersionEnum string

// Wrapper around string value to differentiate strings from interpolations
// Differentiation needed to write oci_resource.resource_name vs "oci_resource.resource_name" for v0.12
type InterpolationString struct {
	ResourceReference string
	Interpolation     string
	Value             string
}

/*  ctxLock is the common lock for the whole struct
WARN: Make sure NOT to pass ResourceDiscoveryContext as value,
as that would copy the struct and locks should not be copied
*/
type ResourceDiscoveryContext struct {
	CtxLock                     sync.Mutex // common lock for the whole context, make sure to acquire the lock before modifying any field in the ResourceDiscoveryContext
	TerraformProviderBinaryPath string
	TerraformCLIPath            string
	Terraform                   *tfexec.Terraform
	Clients                     *tf_client.OracleClients
	ExpectedResourceIds         map[string]bool
	TenancyOcid                 string
	DiscoveredResources         []*OCIResource
	SummaryStatements           []string
	TargetSpecificResources     bool
	ResourceHintsLookup         map[string]*TerraformResourceHints
	*ExportCommandArgs
	ErrorList                    ErrorList
	MissingAttributesPerResource map[string][]string
	IsImportError                bool // flag indicates if there was an import failure and if reference map needs to be updated
	State                        interface{}
	TimeTakenToDiscover          time.Duration
	TimeTakenToGenerateState     time.Duration
	TimeTakenForEntireExport     time.Duration
}
type TerraformResource struct {
	Id                         string
	ImportId                   string
	TerraformClass             string
	TerraformName              string
	TerraformReferenceIdString string // syntax independent interpolation- `resource_type.resource_name.id`
	TerraformTypeInfo          *TerraformResourceHints
	OmitFromExport             bool
}
type TfHclVersion interface {
	ToString() string
	GetReference(reference string) string
	GetVarHclString(string) string
	GetDataSourceHclString(string, string) string
	GetSingleExpHclString(string) string
	GetDoubleExpHclString(string, string) string
}
type ExportCommandArgs struct {
	CompartmentId                *string
	CompartmentName              *string
	IDs                          []string
	Services                     []string
	OutputDir                    *string
	GenerateState                bool
	TFVersion                    *TfHclVersion
	RetryTimeout                 *string
	ExcludeServices              []string
	IsExportWithRelatedResources bool
	Parallelism                  int
	VarsExportResourceLevel      []string
	VarExportGlobalLevel         []string
}
type ErrorList struct {
	Errors []*ResourceDiscoveryError
}
type ResourceDiscoveryError struct {
	ResourceType   string
	ParentResource string
	Error          error
	ResourceGraph  *TerraformResourceGraph
}

var TfHclVersionvar TfHclVersion
var GetHclStringFromGenericMap = func(builder *strings.Builder, ociRes *OCIResource, interpolationMap map[string]string) error {
	resourceSchema := ResourcesMap[ociRes.TerraformClass]

	builder.WriteString(fmt.Sprintf("resource %s %s {\n", ociRes.TerraformClass, ociRes.TerraformName))
	// Generate variable file from user input
	if err := exportAttributeAsVariable(ociRes.SourceAttributes, ociRes.TerraformClass, ociRes.TerraformName, interpolationMap); err != nil {
		return err
	}
	utils.Debugf("getHCLStringFromMap for resource %s", ociRes.TerraformName)
	if err := GetHCLStringFromMap(builder, ociRes.SourceAttributes, resourceSchema, interpolationMap, ociRes, ""); err != nil {
		return err
	}

	if ociRes.TerraformTypeInfo != nil && len(ociRes.TerraformTypeInfo.IgnorableRequiredMissingAttributes) > 0 {
		builder.WriteString("\n# Required attributes that were not found in discovery have been added to " +
			"lifecycle ignore_changes")
		builder.WriteString("\n# This is done to avoid terraform plan failure for the existing infrastructure")
		builder.WriteString("\nlifecycle {\n" +
			"ignore_changes = [")

		missingAttributes := make([]string, 0, len(ociRes.TerraformTypeInfo.IgnorableRequiredMissingAttributes))

		for attribute := range ociRes.TerraformTypeInfo.IgnorableRequiredMissingAttributes {
			missingAttributes = append(missingAttributes, TfHclVersionvar.GetReference(attribute))
		}
		builder.WriteString(strings.Join(missingAttributes, ","))

		builder.WriteString("]\n" +
			"}\n")
	}
	builder.WriteString("}\n\n")

	return nil
}
var ResourcesMap map[string]*schema.Resource
var CompartmentScopeServices []string
var IsMissingRequiredAttributes bool
var TenancyScopeServices []string
var FailedResourceReferenceSet map[string]bool // stores the terraform reference name for failed resources, used to remove InterpolationString type values if a resource failed to import
// This function should only be used to escape TF-characters in strings
var CompartmentResourceGraphs = make(map[string]TerraformResourceGraph)
var TenancyResourceGraphs = map[string]TerraformResourceGraph{}
var RefMapLock sync.Mutex
var resourceNameCountLock sync.Mutex
var ReferenceMap map[string]string                              //	stores references to replace the ocids in config
var LoadBalancerCertificateNameMap map[string]map[string]string // helper map to generate references for certificate names, stores certificate name to certificate name interpolation
var ResourceNameCount map[string]int

var DatasourcesMap map[string]*schema.Resource
var Vars map[string]string
var ExportRelatedResourcesGraph TerraformResourceGraph = make(map[string][]TerraformResourceAssociation)

var availabilityDomainResourceGraph TerraformResourceGraph = make(map[string][]TerraformResourceAssociation)

var VarsExportForResourceLevel map[string][]string // store resource type and attribute from customer input to be converted in var file for resource level
var VarsExportForGlobalLevel []string              // store attributes list from customer input to be converted in var file for global level

// Tags to filter resources
const OkeTagValue = "oke"
const ResourceCreatedByInstancePool = "oci:compute:instancepool"
const OracleTagsCreatedBy = "Oracle-Tags.CreatedBy"
