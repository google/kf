package buildpackapplifecycle

import "strings"

const (
	DetectFailMsg          = "None of the buildpacks detected a compatible application"
	CompileFailMsg         = "Failed to compile droplet"
	ReleaseFailMsg         = "Failed to build droplet release"
	SupplyFailMsg          = "Failed to run all supply scripts"
	NoSupplyScriptFailMsg  = "Error: one of the buildpacks chosen to supply dependencies does not support multi-buildpack apps"
	MissingFinalizeWarnMsg = "Warning: the last buildpack is not compatible with multi-buildpack apps and cannot make use of any dependencies supplied by the buildpacks specified before it"
	FinalizeFailMsg        = "Failed to run finalize script"
	DETECT_FAIL_CODE       = 222
	COMPILE_FAIL_CODE      = 223
	RELEASE_FAIL_CODE      = 224
	SUPPLY_FAIL_CODE       = 225
	FINALIZE_FAIL_CODE     = 226
)

func ExitCodeFromError(err error) int {
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, DetectFailMsg):
		return DETECT_FAIL_CODE
	case strings.Contains(errMsg, CompileFailMsg):
		return COMPILE_FAIL_CODE
	case strings.Contains(errMsg, ReleaseFailMsg):
		return RELEASE_FAIL_CODE
	case strings.Contains(errMsg, SupplyFailMsg):
		return SUPPLY_FAIL_CODE
	case strings.Contains(errMsg, NoSupplyScriptFailMsg):
		return SUPPLY_FAIL_CODE
	case strings.Contains(errMsg, FinalizeFailMsg):
		return FINALIZE_FAIL_CODE
	default:
		return 1
	}
}

type LifecycleMetadata struct {
	BuildpackKey      string              `json:"buildpack_key,omitempty"`
	DetectedBuildpack string              `json:"detected_buildpack"`
	Buildpacks        []BuildpackMetadata `json:"buildpacks"`
}

type BuildpackMetadata struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type ProcessTypes map[string]string

type Sidecar struct {
	Name         string   `yaml:"name" json:"name"`
	ProcessTypes []string `yaml:"process_types" json:"process_types"`
	Command      string   `yaml:"command" json:"command"`
	Memory       int      `yaml:"memory,omitempty" json:"memory,omitempty"`
}

type Process struct {
	Type    string `yaml:"type" json:"type"`
	Command string `yaml:"command" json:"command"`
}

type StagingResult struct {
	LifecycleMetadata `json:"lifecycle_metadata"`
	ProcessTypes      `json:"process_types"`
	ProcessList       []Process `json:"processes, omitempty"`
	Sidecars          []Sidecar `json:"sidecars, omitempty"`
	ExecutionMetadata string    `json:"execution_metadata"`
	LifecycleType     string    `json:"lifecycle_type"`
}

func UpdateStagingResult(result StagingResult, lifeMeta LifecycleMetadata) StagingResult {
	result.LifecycleMetadata = lifeMeta
	result.LifecycleType = "buildpack"
	result.ExecutionMetadata = ""
	return result
}

func NewStagingResult(procTypes ProcessTypes, lifeMeta LifecycleMetadata) StagingResult {
	return StagingResult{
		LifecycleType:     "buildpack",
		LifecycleMetadata: lifeMeta,
		ProcessTypes:      procTypes,
		ExecutionMetadata: "",
	}
}
