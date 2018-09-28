package client

const (
	ContainerType                          = "container"
	ContainerFieldArgs                     = "args"
	ContainerFieldCommand                  = "command"
	ContainerFieldEnv                      = "env"
	ContainerFieldEnvFrom                  = "envFrom"
	ContainerFieldImage                    = "image"
	ContainerFieldImagePullPolicy          = "imagePullPolicy"
	ContainerFieldLifecycle                = "lifecycle"
	ContainerFieldLivenessProbe            = "livenessProbe"
	ContainerFieldName                     = "name"
	ContainerFieldPorts                    = "ports"
	ContainerFieldReadinessProbe           = "readinessProbe"
	ContainerFieldResources                = "resources"
	ContainerFieldSecurityContext          = "securityContext"
	ContainerFieldStdin                    = "stdin"
	ContainerFieldStdinOnce                = "stdinOnce"
	ContainerFieldTTY                      = "tty"
	ContainerFieldTerminationMessagePath   = "terminationMessagePath"
	ContainerFieldTerminationMessagePolicy = "terminationMessagePolicy"
	ContainerFieldVolumeDevices            = "volumeDevices"
	ContainerFieldVolumeMounts             = "volumeMounts"
	ContainerFieldWorkingDir               = "workingDir"
)

type Container struct {
	Args                     []string              `json:"args,omitempty" yaml:"args,omitempty"`
	Command                  []string              `json:"command,omitempty" yaml:"command,omitempty"`
	Env                      []EnvVar              `json:"env,omitempty" yaml:"env,omitempty"`
	EnvFrom                  []EnvFromSource       `json:"envFrom,omitempty" yaml:"envFrom,omitempty"`
	Image                    string                `json:"image,omitempty" yaml:"image,omitempty"`
	ImagePullPolicy          string                `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy,omitempty"`
	Lifecycle                *Lifecycle            `json:"lifecycle,omitempty" yaml:"lifecycle,omitempty"`
	LivenessProbe            *Probe                `json:"livenessProbe,omitempty" yaml:"livenessProbe,omitempty"`
	Name                     string                `json:"name,omitempty" yaml:"name,omitempty"`
	Ports                    []ContainerPort       `json:"ports,omitempty" yaml:"ports,omitempty"`
	ReadinessProbe           *Probe                `json:"readinessProbe,omitempty" yaml:"readinessProbe,omitempty"`
	Resources                *ResourceRequirements `json:"resources,omitempty" yaml:"resources,omitempty"`
	SecurityContext          *SecurityContext      `json:"securityContext,omitempty" yaml:"securityContext,omitempty"`
	Stdin                    bool                  `json:"stdin,omitempty" yaml:"stdin,omitempty"`
	StdinOnce                bool                  `json:"stdinOnce,omitempty" yaml:"stdinOnce,omitempty"`
	TTY                      bool                  `json:"tty,omitempty" yaml:"tty,omitempty"`
	TerminationMessagePath   string                `json:"terminationMessagePath,omitempty" yaml:"terminationMessagePath,omitempty"`
	TerminationMessagePolicy string                `json:"terminationMessagePolicy,omitempty" yaml:"terminationMessagePolicy,omitempty"`
	VolumeDevices            []VolumeDevice        `json:"volumeDevices,omitempty" yaml:"volumeDevices,omitempty"`
	VolumeMounts             []VolumeMount         `json:"volumeMounts,omitempty" yaml:"volumeMounts,omitempty"`
	WorkingDir               string                `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
}
