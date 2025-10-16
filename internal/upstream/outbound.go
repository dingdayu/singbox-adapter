package upstream

import (
	"context"

	"github.com/sagernet/sing-box/option"
	"resty.dev/v3"
)

type UpstreamSubscriber[T ProxyOutbound] interface {
	Name() string
	UserAgent() string
	Profile(ctx context.Context, client *resty.Client) string
	Outboards() ([]T, error)
}

type ProxyOutbound interface {
	ToOutbound() option.Outbound
}
