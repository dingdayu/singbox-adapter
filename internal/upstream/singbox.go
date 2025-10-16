package upstream

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sagernet/sing-box/option"
	"go.yaml.in/yaml/v2"
	"resty.dev/v3"
)

type SingBoxSubscriber struct{}

func (c SingBoxSubscriber) Name() string {
	return "Sing Box"
}

func (c SingBoxSubscriber) UserAgent() string {
	return "sing-box/v1.12.9"
}

func (c SingBoxSubscriber) Profile(ctx context.Context, client *resty.Client, url string) (string, error) {
	resp, err := client.R().SetContext(ctx).SetHeader("User-Agent", c.UserAgent()).Get(url)
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

func (c SingBoxSubscriber) Outboards(ctx context.Context, client *resty.Client, url string) ([]SingBoxOutbound, error) {
	var in []byte

	if strings.HasPrefix(url, "file://") {
		p, err := os.ReadFile(strings.TrimPrefix(url, "file://"))
		if err != nil {
			return nil, err
		}
		in = p
	} else {
		resp, err := client.R().SetContext(ctx).SetHeader("User-Agent", c.UserAgent()).Get(url)
		if err != nil {
			return nil, err
		}
		in = resp.Bytes()
	}

	if len(in) == 0 {
		return nil, fmt.Errorf("fetch proxy for upstream failed: %s", url)
	}

	// ---- For debug purpose ----

	var profile SingBoxProfile
	if err := yaml.Unmarshal(in, &profile); err != nil {
		return nil, fmt.Errorf("unmarshal profile failed: %w", err)
	}

	if len(profile.Outbounds) == 0 {
		return nil, fmt.Errorf("no outbounds found in profile")
	}
	return profile.Outbounds, nil
}

type SingBoxProfile struct {
	Outbounds []SingBoxOutbound `json:"outbounds"`
}

type SingBoxOutbound struct {
	option.Outbound
}

func (p SingBoxOutbound) ToOutbound() option.Outbound {
	return p.Outbound
}
