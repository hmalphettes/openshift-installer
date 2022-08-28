package installconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pborman/uuid"
	utilrand "k8s.io/apimachinery/pkg/util/rand"

	"github.com/openshift/installer/pkg/asset"
)

const (
	randomLen = 5
)

// ClusterID is the unique ID of the cluster, immutable during the cluster's life
type ClusterID struct {
	// UUID is a globally unique identifier.
	UUID string

	// InfraID is an identifier for the cluster that is more human friendly.
	// This does not have
	InfraID string
}

var (
	// replace all characters that are not `alphanum` or `-` with `-`
	re = regexp.MustCompile("[^A-Za-z0-9-]")
	// replace all multiple dashes in a sequence with single one.
	re2 = regexp.MustCompile(`-{2,}`)
)

var _ asset.Asset = (*ClusterID)(nil)

// Dependencies returns install-config.
func (a *ClusterID) Dependencies() []asset.Asset {
	return []asset.Asset{
		&InstallConfig{},
	}
}

// Generate generates a new ClusterID
func (a *ClusterID) Generate(dep asset.Parents) error {
	ica := &InstallConfig{}
	dep.Get(ica)

	// resource using InfraID usually have suffixes like `[-/_][a-z]{3,4}` eg. `_int`, `-ext` or `-ctlp`
	// and the maximum length for most resources is approx 32.
	maxLen := 27

	// add random chars to the end to randomize
	a.InfraID = generateInfraID(ica.Config.ObjectMeta.Name, maxLen)
	a.UUID = uuid.New()
	return nil
}

// Name returns the human-friendly name of the asset.
func (a *ClusterID) Name() string {
	return "Cluster ID"
}

// All characters that are not alphanum are replaced by '-'
// consecutive '-' are collapsed and '-' are trimmed from the right
func normalizeString(raw string) string {
	normalized := re.ReplaceAllString(raw, "-")
	normalized = re2.ReplaceAllString(normalized, "-")
	if len(normalized) == 0 || normalized == "-" {
		panic(fmt.Sprintf("Invalid input string \"%s\". It must contain at least 1 alphanum character", raw))
	}
	normalized = strings.TrimRight(normalized, "-")
	return normalized
}

func lookupPredefinedValue(envName string) string {
	value := os.Getenv(envName)
	if value == "" {
		dotFileName := fmt.Sprintf(".%s", strings.ToLower(envName))
		if _, err := os.Stat(dotFileName); err == nil {
			raw, _ := ioutil.ReadFile(dotFileName)
			value = string(raw)
		}
	}
	if value == "" || strings.ToLower(value) == "false" {
		return ""
	}
	if value == "." {
		value, _ = os.Getwd()
		value = filepath.Base(value)
	}
	value = normalizeString(value)
	return value
}

func truncate(value string, maxLen int) string {
	// truncate to maxBaseLen
	if len(value) > maxLen {
		value = value[:maxLen]
	}
	return strings.TrimRight(value, "-")
}

// generateInfraID take base and returns a ID that
// - is of length maxLen
// - only contains `alphanum` or `-`
func generateInfraID(base string, maxLen int) string {
	preGeneratedInfraID := lookupPredefinedValue("INFRA_ID")
	if preGeneratedInfraID != "" {
		infraID := normalizeString(preGeneratedInfraID)
		os.Setenv("INFRA_ID", infraID)
		return infraID
	}

	// normalize early to maximum the number of meaningful characters extracted
	base = normalizeString(base)
	maxBaseLen := maxLen - (randomLen + 1)

	// truncate to maxBaseLen
	base = truncate(base, maxBaseLen)

	suffix := lookupPredefinedValue("INFRA_ID_SUFFIX")
	if suffix != "" {
		infraID := fmt.Sprintf("%s-%s", base, suffix)
		infraID = truncate(infraID, maxLen)
		os.Setenv("INFRA_ID", infraID)
		return infraID
	} else {
		suffix = utilrand.String(randomLen)
	}

	// add random chars to the end to randomize
	return fmt.Sprintf("%s-%s", base, suffix)
}
