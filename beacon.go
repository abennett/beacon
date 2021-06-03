package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/abennett/beacon/providers"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

type Beacon struct {
	config   *Config
	logger   *zap.Logger
	provider DNSProvider
	errorCh  chan error
}

type DNSRecord struct {
	Domain string
	TTL    int
	IP     string
}

func New(c *Config, l *zap.Logger) (*Beacon, error) {
	provider, err := newProvider(c)
	if err != nil {
		return nil, err
	}
	errCh := make(chan error, 1)
	b := &Beacon{
		config:   c,
		logger:   l,
		errorCh:  errCh,
		provider: provider,
	}
	return b, nil
}

func newProvider(c *Config) (DNSProvider, error) {
	var provider DNSProvider
	switch {
	case c.AWSCredentials != nil:
		p, err := providers.NewAWSProvider(
			c.AWSCredentials.AccessKeyID,
			c.AWSCredentials.SecretAccessKey,
			c.AWSCredentials.HostedZoneId,
		)
		if err != nil {
			return nil, err
		}
		provider = p
	case c.CloudflareCredentials != nil:
		p, err := providers.NewCloudflareProvider(c.CloudflareCredentials)
		if err != nil {
			return nil, err
		}
		provider = p
	}
	return provider, nil
}

func (b *Beacon) Start(ctx context.Context) error {
	return b.loop(ctx)
}

func (b *Beacon) loop(ctx context.Context) error {
	timer := time.NewTimer(0)
	for {
		var dnsIP string
		var ttl int
		b.logger.Info("checking")
		record, ip, same, err := b.Check()
		if err != nil {
			return err
		}
		if record != nil {
			dnsIP = record.IP
			ttl = record.TTL
		}
		if !same {
			b.logger.Info("dns record needs updated",
				zap.String("domain", b.config.Domain),
				zap.String("real_ip", ip),
				zap.String("dns_ip", dnsIP),
				zap.Int("ttl", ttl),
			)
			err = b.provider.SetRecord(ctx, b.config.Domain, ip, b.config.TTL)
			if err != nil {
				return err
			}
		} else {
			b.logger.Info("dns records are correct")
		}
		waitTime := time.Second * (time.Duration(b.config.TTL) + 1)
		b.logger.Info("waiting", zap.Duration("wait_time", waitTime))
		if !timer.Stop() {
			<-timer.C
		}
		timer.Reset(waitTime)
		select {
		case <-timer.C:
			break
		case <-ctx.Done():
			return nil
		}
	}
}

func (b *Beacon) Check() (*DNSRecord, string, bool, error) {
	ip, err := LookupIP()
	if err != nil {
		return nil, "", false, err
	}
	record, err := LookupDomain(b.config.Domain)
	if err != nil {
		return nil, "", false, err
	}
	if record == nil {
		return nil, ip, false, nil
	}
	return record, ip, ip == record.IP, nil
}

func LookupDomain(domain string) (*DNSRecord, error) {
	fqdn := dns.Fqdn(domain)
	m := new(dns.Msg)
	m.SetQuestion(fqdn, dns.TypeA)
	r, err := dns.Exchange(m, "1.1.1.1:53")
	if len(r.Answer) == 0 {
		return nil, nil
	}
	aRecord, ok := r.Answer[0].(*dns.A)
	if !ok {
		return nil, fmt.Errorf("could not cast: %T", r)
	}
	record := &DNSRecord{
		Domain: aRecord.Hdr.Name,
		TTL:    int(aRecord.Hdr.Ttl),
		IP:     aRecord.A.String(),
	}
	return record, err
}

func LookupIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid status code: %s", resp.Status)
	}
	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(rb), nil
}
