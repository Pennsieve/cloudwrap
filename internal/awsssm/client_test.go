package awsssm

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"

	"github.com/pennsieve/cloudwrap/internal/config"
)

// mockAPI drives Client through canned, paginated responses.
type mockAPI struct {
	getPages      []*ssm.GetParametersByPathOutput
	getCalls      int
	getSeenTokens []string

	descPages []*ssm.DescribeParametersOutput
	descCalls int
}

func (m *mockAPI) GetParametersByPath(_ context.Context, in *ssm.GetParametersByPathInput, _ ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error) {
	m.getSeenTokens = append(m.getSeenTokens, aws.ToString(in.NextToken))
	out := m.getPages[m.getCalls]
	m.getCalls++
	return out, nil
}

func (m *mockAPI) DescribeParameters(_ context.Context, _ *ssm.DescribeParametersInput, _ ...func(*ssm.Options)) (*ssm.DescribeParametersOutput, error) {
	out := m.descPages[m.descCalls]
	m.descCalls++
	return out, nil
}

func TestGetParametersFollowsPagination(t *testing.T) {
	m := &mockAPI{
		getPages: []*ssm.GetParametersByPathOutput{
			{
				Parameters: []types.Parameter{
					{Name: aws.String("/env/svc/one-key"), Value: aws.String("v1"), Version: 1},
				},
				NextToken: aws.String("page2"),
			},
			{
				Parameters: []types.Parameter{
					{Name: aws.String("/env/svc/two"), Value: aws.String("v2"), Version: 3},
				},
				NextToken: nil,
			},
		},
	}
	c := &Client{ssm: m}

	got, err := c.GetParameters(context.Background(), config.New("env", "svc"))
	if err != nil {
		t.Fatal(err)
	}
	if m.getCalls != 2 {
		t.Fatalf("expected 2 API calls (pagination), got %d", m.getCalls)
	}
	// First call has no token; second carries the token from page 1.
	if m.getSeenTokens[0] != "" || m.getSeenTokens[1] != "page2" {
		t.Fatalf("unexpected pagination tokens: %v", m.getSeenTokens)
	}
	if len(got) != 2 || got[0].Name != "/env/svc/one-key" || got[1].Value != "v2" {
		t.Fatalf("unexpected params: %+v", got)
	}
}

func TestDescribeParametersFollowsPagination(t *testing.T) {
	ts := time.Unix(1524598562, 0)
	m := &mockAPI{
		descPages: []*ssm.DescribeParametersOutput{
			{
				Parameters: []types.ParameterMetadata{
					{Name: aws.String("/env/svc/one-key"), Version: 1, LastModifiedUser: aws.String("u"), LastModifiedDate: &ts},
				},
				NextToken: aws.String("next"),
			},
			{
				Parameters: []types.ParameterMetadata{
					{Name: aws.String("/env/svc/two"), Version: 2},
				},
			},
		},
	}
	c := &Client{ssm: m}

	got, err := c.DescribeParameters(context.Background(), config.New("env", "svc"))
	if err != nil {
		t.Fatal(err)
	}
	if m.descCalls != 2 {
		t.Fatalf("expected 2 API calls, got %d", m.descCalls)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 metadata rows, got %d", len(got))
	}
	if got[0].LastModifiedUser != "u" || !got[0].LastModifiedDate.Equal(ts) {
		t.Fatalf("metadata not mapped correctly: %+v", got[0])
	}
}
