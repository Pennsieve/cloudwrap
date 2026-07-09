// Package awsssm wraps the AWS SDK for Go v2 SSM client with the two paginated
// operations cloudwrap needs: fetching parameter values and describing
// parameter metadata for a config path.
package awsssm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"

	"github.com/pennsieve/cloudwrap/internal/config"
	"github.com/pennsieve/cloudwrap/internal/params"
)

// region is fixed to match the original tool's behavior (SsmClient::default()
// used us-east-1).
const region = "us-east-1"

// api is the narrow slice of the SSM client that Client depends on. It exists
// so tests can substitute a mock.
type api interface {
	GetParametersByPath(ctx context.Context, in *ssm.GetParametersByPathInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error)
	DescribeParameters(ctx context.Context, in *ssm.DescribeParametersInput, optFns ...func(*ssm.Options)) (*ssm.DescribeParametersOutput, error)
}

// Client fetches parameters from AWS SSM Parameter Store.
type Client struct {
	ssm api
}

// New builds a Client using the default AWS credential chain, pinned to
// us-east-1.
func New(ctx context.Context) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return &Client{ssm: ssm.NewFromConfig(cfg)}, nil
}

// GetParameters returns all parameter values under the config path,
// recursively and decrypted, following pagination to completion.
func (c *Client) GetParameters(ctx context.Context, cfg config.Config) ([]params.Parameter, error) {
	var out []params.Parameter
	var token *string

	for {
		res, err := c.ssm.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
			Path:           aws.String(cfg.Path()),
			Recursive:      aws.Bool(true),
			WithDecryption: aws.Bool(true),
			NextToken:      token,
		})
		if err != nil {
			return nil, err
		}

		for _, p := range res.Parameters {
			out = append(out, params.Parameter{
				Name:    aws.ToString(p.Name),
				Value:   aws.ToString(p.Value),
				Version: p.Version,
			})
		}

		if res.NextToken == nil {
			return out, nil
		}
		token = res.NextToken
	}
}

// DescribeParameters returns metadata for all parameters whose name begins with
// the config path, following pagination to completion.
func (c *Client) DescribeParameters(ctx context.Context, cfg config.Config) ([]params.Metadata, error) {
	var out []params.Metadata
	var token *string

	filter := types.ParameterStringFilter{
		Key:    aws.String("Name"),
		Option: aws.String("BeginsWith"),
		Values: []string{cfg.Path()},
	}

	for {
		res, err := c.ssm.DescribeParameters(ctx, &ssm.DescribeParametersInput{
			ParameterFilters: []types.ParameterStringFilter{filter},
			NextToken:        token,
		})
		if err != nil {
			return nil, err
		}

		for _, p := range res.Parameters {
			m := params.Metadata{
				Name:             aws.ToString(p.Name),
				Version:          p.Version,
				LastModifiedUser: aws.ToString(p.LastModifiedUser),
			}
			if p.LastModifiedDate != nil {
				m.LastModifiedDate = *p.LastModifiedDate
			}
			out = append(out, m)
		}

		if res.NextToken == nil {
			return out, nil
		}
		token = res.NextToken
	}
}
