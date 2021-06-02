package main

import (
	"errors"
	"reflect"
	"strings"

	"github.com/abennett/beacon/providers"
	toml "github.com/pelletier/go-toml/v2"
)

var (
	ErrNoConfig              = errors.New("no config block found")
	ErrOneCredentialRequired = errors.New("only one credential can be provided")
)

type Config struct {
	Domain         string                    `toml:"domain"`
	TTL            int                       `toml:"ttl"`
	AWSCredentials *providers.AWSCredentials `toml:"aws"`
}

func LoadConfig(b []byte) (*Config, error) {
	var config Config
	if err := toml.Unmarshal(b, &config); err != nil {
		return nil, err
	}
	if !hasOneCredential(config) {
		return &config, ErrOneCredentialRequired
	}
	return &config, nil
}

func hasOneCredential(c Config) bool {
	v := reflect.ValueOf(c)
	t := reflect.TypeOf(c)
	var totalCredentials int
	for x := 0; x < v.NumField(); x++ {
		isCred := strings.HasSuffix(t.Field(x).Name, "Credentials")
		if isCred {
			if !v.Field(x).IsNil() {
				totalCredentials++
			}
		}
	}
	return totalCredentials == 1
}
