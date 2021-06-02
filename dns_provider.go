package main

import "context"

type DNSProvider interface {
	SetRecord(ctx context.Context, domain string, ip string, ttl int) error
}

type ProviderCredentials interface {
	NewClient() (DNSProvider, error)
}
