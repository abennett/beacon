package providers

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

type AWSProvider struct {
	client *route53.Client
	zoneID string
}

type AWSCredentials struct {
	AccessKeyID     string `toml:"access_key_id"`
	SecretAccessKey string `toml:"secret_access_key"`
	HostedZoneId    string `toml:"hosted_zone_id"`
}

func NewAWSProvider(accessKey, secretKey, zoneID string) (*AWSProvider, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     accessKey,
				SecretAccessKey: secretKey,
			},
		}),
		config.WithRegion("us-east-1"))
	if err != nil {
		return nil, err
	}
	client := route53.NewFromConfig(cfg)
	return &AWSProvider{
		client: client,
		zoneID: zoneID,
	}, nil
}

func (ap *AWSProvider) SetRecord(ctx context.Context, domain string, ip string, ttl int) error {
	domainPtr := aws.String(domain)
	change := types.Change{
		Action: types.ChangeActionUpsert,
		ResourceRecordSet: &types.ResourceRecordSet{
			Name: domainPtr,
			Type: types.RRTypeA,
			TTL:  aws.Int64(int64(ttl)),
			ResourceRecords: []types.ResourceRecord{
				{
					Value: aws.String(ip),
				},
			},
		},
	}
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{change},
		},
		HostedZoneId: aws.String(ap.zoneID),
	}
	_, err := ap.client.ChangeResourceRecordSets(ctx, input)
	return err
}
