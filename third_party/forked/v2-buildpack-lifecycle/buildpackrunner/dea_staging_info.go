package buildpackrunner

// Used to generate YAML file read by the DEA
type DeaStagingInfo struct {
	DetectedBuildpack string `json:"detected_buildpack" yaml:"detected_buildpack"`
	StartCommand      string `json:"start_command" yaml:"start_command"`
}
