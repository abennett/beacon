package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var testConfig = `
domain	 = "home.bennett.life"
ttl		 = 300

[aws]
access_key_id		= "meep"
secret_access_key	= "mop"
`

func TestLoadConfig(t *testing.T) {
	_, err := LoadConfig([]byte(testConfig))
	require.NoError(t, err)
}
