package upstream

import (
	"context"

	"github.com/sagernet/sing-box/option"
	"resty.dev/v3"
)

type UpstreamSubscriber interface {
	Name() string
	UserAgent() string
	Profile(ctx context.Context, client *resty.Client, url string) (string, error)
	Outboards(ctx context.Context, client *resty.Client, url string) ([]ProxyOutbound, error)
}

type ProxyOutbound interface {
	ToOutbound() (option.Outbound, error)
	Name() string
}
