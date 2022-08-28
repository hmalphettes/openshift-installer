package installconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_generateInfraID(t *testing.T) {
	tests := []struct {
		input string

		expLen       int
		expNonRand   string
		envVars      map[string]string
		expClusterID string
	}{{
		input:        "qwertyuiop",
		expLen:       10 + randomLen + 1,
		expNonRand:   "qwertyuiop",
		envVars:      map[string]string{},
		expClusterID: "",
	}, {
		input:        "qwertyuiopasdfghjklzxcvbnm",
		expLen:       27,
		expNonRand:   "qwertyuiopasdfghjklzx",
		envVars:      map[string]string{},
		expClusterID: "",
	}, {
		input:        "qwertyuiopasdfghjklz-cvbnm",
		expLen:       26,
		expNonRand:   "qwertyuiopasdfghjklz",
		envVars:      map[string]string{},
		expClusterID: "",
	}, {
		input:        "qwe.rty.@iop!",
		expLen:       11 + randomLen + 1,
		expNonRand:   "qwe-rty-iop",
		envVars:      map[string]string{},
		expClusterID: "",
	}, {
		input:      "poc",
		expLen:     -1,
		expNonRand: "",
		envVars: map[string]string{
			"INFRA_ID_SUFFIX": "001",
		},
		expClusterID: "poc-001",
	}, {
		input:      "poc",
		expLen:     -1,
		expNonRand: "",
		envVars: map[string]string{
			"INFRA_ID_SUFFIX": ".",
		},
		expClusterID: "poc-installconfig",
	}, {
		input:      "not-applicable",
		expLen:     -1,
		expNonRand: "",
		envVars: map[string]string{
			"INFRA_ID": "myclusterid",
		},
		expClusterID: "myclusterid",
	}, {
		input:      "not-applicable",
		expLen:     -1,
		expNonRand: "",
		envVars: map[string]string{
			"INFRA_ID": ".",
		},
		expClusterID: "installconfig",
	}, {
		input:      "not-applicable",
		expLen:     -1,
		expNonRand: "",
		envVars: map[string]string{
			"INFRA_ID": "mycluster-id@.strange",
		},
		expClusterID: "mycluster-id-strange",
	}}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			for k, v := range test.envVars {
				os.Setenv(k, v)
			}
			got := generateInfraID(test.input, 27)
			for k := range test.envVars {
				os.Setenv(k, "")
			}
			t.Log("InfraID", got)
			if test.expClusterID != "" {
				assert.Equal(t, test.expClusterID, got)
			} else {
				assert.Equal(t, test.expLen, len(got))
				assert.Equal(t, test.expNonRand, got[:len(got)-randomLen-1])
			}
		})
	}
}
