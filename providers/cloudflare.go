package providers

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
)

type CloudflareProvider struct {
	api      *cloudflare.API
	zoneID   string
	recordID string
}

type CloudflareCredentials struct {
	Key      string
	Email    string
	ZoneID   string
	RecordID string
}

func NewCloudflareProvider(cfc *CloudflareCredentials) (*CloudflareProvider, error) {
	api, err := cloudflare.New(cfc.Key, cfc.Email)
	if err != nil {
		return nil, err
	}
	return &CloudflareProvider{
		api:      api,
		zoneID:   cfc.ZoneID,
		recordID: cfc.RecordID,
	}, nil
}

func (cf *CloudflareProvider) SetRecord(ctx context.Context, domain string, ip string, ttl int) error {
	rr := cloudflare.DNSRecord{
		Type:    "A",
		Name:    domain,
		TTL:     ttl,
		Content: ip,
	}
	return cf.api.UpdateDNSRecord(ctx, cf.zoneID, cf.recordID, rr)
}
