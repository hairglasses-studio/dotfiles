package secrets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

// Identity represents the current user/service identity.
type Identity struct {
	// User info
	User    string `json:"user"`
	UserID  string `json:"user_id,omitempty"`
	HomeDir string `json:"home_dir,omitempty"`
	Shell   string `json:"shell,omitempty"`

	// AWS info
	AWSProfile string `json:"aws_profile,omitempty"`
	AWSAccount string `json:"aws_account,omitempty"`
	AWSRegion  string `json:"aws_region,omitempty"`
	AWSUserARN string `json:"aws_user_arn,omitempty"`

	// Git info
	GitHubUser string `json:"github_user,omitempty"`
	GitEmail   string `json:"git_email,omitempty"`
	GitName    string `json:"git_name,omitempty"`

	// Kubernetes info
	Cluster     string `json:"cluster,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	KubeContext string `json:"kube_context,omitempty"`

	// 1Password info
	OPAccount string `json:"op_account,omitempty"`
	OPVault   string `json:"op_vault,omitempty"`

	// Environment info
	Hostname string `json:"hostname,omitempty"`
	OS       string `json:"os,omitempty"`
	Arch     string `json:"arch,omitempty"`

	// Permissions/Roles
	Roles []string `json:"roles,omitempty"`
}

// Whoami returns identity information for the current user/service.
func Whoami(ctx context.Context) (*Identity, error) {
	id := &Identity{}

	// Basic user info
	if u, err := user.Current(); err == nil {
		id.User = u.Username
		id.UserID = u.Uid
		id.HomeDir = u.HomeDir
	} else {
		id.User = os.Getenv("USER")
		id.HomeDir = os.Getenv("HOME")
	}

	// Shell
	id.Shell = os.Getenv("SHELL")

	// Hostname
	if hostname, err := os.Hostname(); err == nil {
		id.Hostname = hostname
	}

	// OS info from environment (typically set by uname or similar)
	id.OS = os.Getenv("OSTYPE")
	if id.OS == "" {
		// Try to detect
		switch {
		case fileExists("/etc/os-release"):
			id.OS = "linux"
		case fileExists("/System/Library"):
			id.OS = "darwin"
		}
	}

	// AWS info
	id.AWSProfile = os.Getenv("AWS_PROFILE")
	id.AWSRegion = os.Getenv("AWS_REGION")
	if id.AWSRegion == "" {
		id.AWSRegion = os.Getenv("AWS_DEFAULT_REGION")
	}

	// Get AWS caller identity if available
	if awsInfo := getAWSIdentity(ctx); awsInfo != nil {
		id.AWSAccount = awsInfo.Account
		id.AWSUserARN = awsInfo.Arn
	}

	// Git info
	id.GitEmail = getGitConfig("user.email")
	id.GitName = getGitConfig("user.name")
	id.GitHubUser = os.Getenv("GITHUB_USER")
	if id.GitHubUser == "" {
		id.GitHubUser = getGitHubUser(ctx)
	}

	// Kubernetes info
	id.KubeContext = os.Getenv("KUBECTL_CONTEXT")
	if id.KubeContext == "" {
		id.KubeContext = getKubeContext()
	}
	id.Namespace = os.Getenv("KUBECTL_NAMESPACE")
	if id.Namespace == "" {
		id.Namespace = getKubeNamespace()
	}

	// 1Password info
	id.OPAccount = os.Getenv("OP_ACCOUNT")
	id.OPVault = os.Getenv("OP_VAULT")

	return id, nil
}

// AWSIdentity holds AWS identity info.
type AWSIdentity struct {
	Account string
	Arn     string
	UserID  string
}

// getAWSIdentity fetches AWS caller identity.
func getAWSIdentity(ctx context.Context) *AWSIdentity {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "aws", "sts", "get-caller-identity", "--output", "json")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil
	}

	var result struct {
		Account string `json:"Account"`
		Arn     string `json:"Arn"`
		UserID  string `json:"UserId"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil
	}

	return &AWSIdentity{
		Account: result.Account,
		Arn:     result.Arn,
		UserID:  result.UserID,
	}
}

// getGitConfig fetches a git config value.
func getGitConfig(key string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "config", "--get", key)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return ""
	}

	return strings.TrimSpace(stdout.String())
}

// getGitHubUser fetches the current GitHub user.
func getGitHubUser(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gh", "api", "user", "--jq", ".login")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return ""
	}

	return strings.TrimSpace(stdout.String())
}

// getKubeContext fetches the current kubectl context.
func getKubeContext() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "kubectl", "config", "current-context")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return ""
	}

	return strings.TrimSpace(stdout.String())
}

// getKubeNamespace fetches the current kubectl namespace.
func getKubeNamespace() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "kubectl", "config", "view", "--minify", "--output", "jsonpath={..namespace}")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return ""
	}

	ns := strings.TrimSpace(stdout.String())
	if ns == "" {
		return "default"
	}
	return ns
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// String returns a formatted string representation of the identity.
func (id *Identity) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("User: %s\n", id.User))

	if id.Hostname != "" {
		b.WriteString(fmt.Sprintf("Hostname: %s\n", id.Hostname))
	}

	if id.AWSProfile != "" || id.AWSAccount != "" {
		b.WriteString(fmt.Sprintf("AWS: profile=%s account=%s region=%s\n",
			id.AWSProfile, id.AWSAccount, id.AWSRegion))
	}

	if id.GitHubUser != "" {
		b.WriteString(fmt.Sprintf("GitHub: %s\n", id.GitHubUser))
	}

	if id.KubeContext != "" {
		b.WriteString(fmt.Sprintf("Kubernetes: context=%s namespace=%s\n",
			id.KubeContext, id.Namespace))
	}

	if id.OPAccount != "" {
		b.WriteString(fmt.Sprintf("1Password: account=%s vault=%s\n",
			id.OPAccount, id.OPVault))
	}

	return b.String()
}

// ToMap converts the identity to a map for JSON serialization.
func (id *Identity) ToMap() map[string]any {
	return map[string]any{
		"user":         id.User,
		"user_id":      id.UserID,
		"home_dir":     id.HomeDir,
		"shell":        id.Shell,
		"hostname":     id.Hostname,
		"os":           id.OS,
		"aws_profile":  id.AWSProfile,
		"aws_account":  id.AWSAccount,
		"aws_region":   id.AWSRegion,
		"aws_user_arn": id.AWSUserARN,
		"github_user":  id.GitHubUser,
		"git_email":    id.GitEmail,
		"git_name":     id.GitName,
		"kube_context": id.KubeContext,
		"namespace":    id.Namespace,
		"op_account":   id.OPAccount,
		"op_vault":     id.OPVault,
		"roles":        id.Roles,
	}
}

// HasAWSAccess returns true if AWS credentials are available.
func (id *Identity) HasAWSAccess() bool {
	return id.AWSAccount != "" && id.AWSUserARN != ""
}

// HasGitHubAccess returns true if GitHub CLI access is available.
func (id *Identity) HasGitHubAccess() bool {
	return id.GitHubUser != ""
}

// HasKubernetesAccess returns true if kubectl access is available.
func (id *Identity) HasKubernetesAccess() bool {
	return id.KubeContext != ""
}

// Has1PasswordAccess returns true if 1Password CLI access is available.
func (id *Identity) Has1PasswordAccess() bool {
	return id.OPAccount != ""
}
