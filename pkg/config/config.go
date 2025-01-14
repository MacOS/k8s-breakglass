package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

type Transition struct {
	From                     string   `yaml:"from" json:"from"`
	To                       string   `yaml:"to" json:"to"`
	Duration                 int64    `yaml:"duration" json:"duration"`
	SelfApproval             bool     `yaml:"selfApproval" json:"selfApproval"`
	ApprovalGroups           []string `yaml:"approvalGroups" json:"approvalGroups"`
	GlobalBreakglassExcluded bool     `yaml:"globalBreakglassExcluded" json:"-"`
}

type Keycloak struct {
	Url          string
	ClientID     string `yaml:"clientID"`
	ClientSecret string `yaml:"clientSecret"`
	LoginRealm   string
	ManagedRealm string
}

type Frontend struct {
	OIDCAuthority string `yaml:"oidcAuthority"`
	OIDCClientID  string `yaml:"oidcClientID"`
}

type JWT struct {
	JWTPrivateKey string
	JWTPublicKey  string
	Expiry        int64
	Issuer        string
}

type Mail struct {
	Host               string
	Port               int
	User               string
	Password           string
	InsecureSkipVerify bool `yaml:"insecureSkipVerify"`
}

type Server struct {
	ListenAddress string `yaml:"listenAddress"`
	TLSCertFile   string `yaml:"tlsCertFile"`
	TLSKeyFile    string `yaml:"tlsKeyFile"`
	BaseURL       string `yaml:"baseURL"`
}

type ClusterAccess struct {
	FrontendPage  string   `yaml:"frontentPage"`
	ClusterGroups []string `yaml:"clusterGroups"`
	Approvers     []string `yaml:"approvers"`
}

type Config struct {
	Server                 Server
	PossibleTransitions    []Transition
	GlobalBreakglassGroups []string `yaml:"globalBreakglassGroups"`
	Keycloak               Keycloak
	BreakglassJWT          JWT
	Mail                   Mail
	Frontend               Frontend
	ClusterAccess          ClusterAccess `yaml:"clusterAccess"`
}

func Load() (Config, error) {
	var config Config

	configPath := os.Getenv("BREAKGLASS_CONFIG_PATH")
	if len(configPath) == 0 {
		configPath = "./config.yaml"
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("trying to open breakglass config file %s: %v", configPath, err)
	}

	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return config, fmt.Errorf("error unmarshaling YAML %s: %v", configPath, err)
	}
	return config, nil
}

// Sets default values for configuration fields where it makes sense.
func (c *Config) Defaults() {
	if c.ClusterAccess.FrontendPage == "" {
		c.ClusterAccess.FrontendPage = "http://localhost:5173"
	}
}

// Validates semantically configuration fields.
func (c Config) Validate() error {
	if len(c.ClusterAccess.Approvers) == 0 {
		return errors.New("ClusterAccess requires at least single approver")
	}
	return nil
}

func (a Transition) Equal(b Transition) bool {
	if a.From != b.From {
		return false
	}
	if a.To != b.To {
		return false
	}
	if a.Duration != b.Duration {
		return false
	}
	if a.SelfApproval != b.SelfApproval {
		return false
	}
	if !cmp.Equal(a.ApprovalGroups, b.ApprovalGroups) {
		return false
	}
	return true
}
