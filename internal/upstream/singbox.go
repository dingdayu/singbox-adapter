package upstream

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sagernet/sing-box/option"
	sjson "github.com/sagernet/sing/common/json"
	"resty.dev/v3"
)

type SingBoxSubscriber struct{}

func (c SingBoxSubscriber) Name() string {
	return "Sing Box"
}

func (c SingBoxSubscriber) UserAgent() string {
	return "SFM/1.12.9 (Build 1; sing-box 1.12.9; language zh_CN)"
}

func (c SingBoxSubscriber) Profile(ctx context.Context, client *resty.Client, url string) (string, error) {
	resp, err := client.R().SetContext(ctx).SetHeader("User-Agent", c.UserAgent()).Get(url)
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

func (c SingBoxSubscriber) Outboards(ctx context.Context, client *resty.Client, url string) ([]ProxyOutbound, error) {
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
	err := sjson.UnmarshalContext(ctx, in, &profile)
	if err != nil {
		return nil, fmt.Errorf("unmarshal profile failed: %w", err)
	}

	if len(profile.Outbounds) == 0 {
		return nil, fmt.Errorf("no outbounds found in profile")
	}
	result := make([]ProxyOutbound, 0, len(profile.Outbounds))
	for _, p := range profile.Outbounds {
		if p.Type == "direct" || p.Type == "selector" || p.Type == "dns" || p.Type == "urltest" {
			continue
		}
		result = append(result, p)
	}
	return result, nil
}

type SingBoxProfile struct {
	Outbounds []SingBoxOutbound `json:"outbounds"`
}

type SingBoxOutbound struct {
	option.Outbound
}

func (p SingBoxOutbound) ToOutbound() (option.Outbound, error) {
	return p.Outbound, nil
}

func (p SingBoxOutbound) Name() string {
	return p.Outbound.Tag
}

func (p SingBoxOutbound) FilterOutboundsByKeywords(keywords []string) (option.Outbound, error) {
	if len(keywords) > 0 {
		found := false
		for _, kw := range keywords {
			if strings.Contains(p.Tag, kw) {
				found = true
				break
			}
		}
		if !found {
			return option.Outbound{}, fmt.Errorf("proxy %s does not match keywords", p.Tag)
		}
	}
	return p.ToOutbound()
}
