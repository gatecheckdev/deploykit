package deploy

type GitHubAction struct {
	Name        string                       `yaml:"name"`
	Description string                       `yaml:"description"`
	Inputs      map[string]GitHubActionInput `yaml:"inputs"`
	Runs        GitHubActionRuns             `yaml:"runs"`
}

type GitHubActionInput struct {
	Description string `yaml:"description"`
	Default     any    `yaml:"default,omitempty"`
	Required    bool   `yaml:"required,omitempty"`
}

type GitHubActionRuns struct {
	Using string            `yaml:"using"`
	Image string            `yaml:"image"`
	Env   map[string]string `yaml:"env"`
}
