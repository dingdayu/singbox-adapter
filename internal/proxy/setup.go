package proxy

import (
	"context"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dingdayu/go-project-template/internal/upstream"
	"github.com/sagernet/sing-box/include"
	"github.com/spf13/viper"
	"resty.dev/v3"
)

var client = resty.New()

// store holds the aggregated outbounds for lock-free reads via atomic.Value.
// Convention: the stored slice must never be mutated after Store.
var store atomic.Value // of type []upstream.ProxyOutbound

// perUpstream holds the latest outbounds per upstream URL; guarded by perMu.
var (
	perMu       sync.Mutex
	perUpstream map[string][]upstream.ProxyOutbound
)

// master cancel to stop all tickers during reload.
var (
	masterMu     sync.Mutex
	masterCancel context.CancelFunc
)

type Upstream struct {
	URL          string   `mapstructure:"url"`
	Timeout      int      `mapstructure:"timeout"`
	Retry        int      `mapstructure:"retry"`
	Interval     int      `mapstructure:"interval"`
	NodeKeywords []string `mapstructure:"node_keywords"`
}

func Setup() error {
	var upstreams []Upstream

	// 🔑 关键：只解析 `upstreams` 键
	if err := viper.UnmarshalKey("upstreams", &upstreams); err != nil {
		log.Fatalf("Unable to decode 'upstreams' into struct: %v", err)
	}

	// initialize atomic store if not set
	store.Store(make([]upstream.ProxyOutbound, 0))
	return reloadUpstreams(upstreams)
}

// Reload reads the latest upstreams from viper and restarts the background fetch tickers.
func Reload() error {
	var upstreams []Upstream
	if err := viper.UnmarshalKey("upstreams", &upstreams); err != nil {
		log.Printf("proxy.Reload: unable to decode 'upstreams': %v", err)
		return err
	}
	return reloadUpstreams(upstreams)
}

func reloadUpstreams(upstreams []Upstream) error {
	// cancel previous master
	masterMu.Lock()
	if masterCancel != nil {
		masterCancel()
	}
	// reset per-upstream cache
	perMu.Lock()
	perUpstream = make(map[string][]upstream.ProxyOutbound)
	perMu.Unlock()

	ctx := include.Context(context.Background())
	// new master context
	ctx, cancel := context.WithCancel(ctx)
	masterCancel = cancel
	masterMu.Unlock()

	// Start new tickers
	for _, u := range upstreams {
		if u.Interval <= 0 {
			u.Interval = 300
		}
		startUpstream(ctx, u)
	}
	return nil
}

func startUpstream(ctx context.Context, u Upstream) {
	go func() {
		// immediate fetch
		updatePerAndAggregate(u.URL, FetchUpstreams(ctx, u))

		ticker := time.NewTicker(time.Duration(u.Interval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				updatePerAndAggregate(u.URL, FetchUpstreams(ctx, u))
			case <-ctx.Done():
				log.Printf("proxy: ticker stopped for %s", u.URL)
				return
			}
		}
	}()
}

func FetchUpstreams(ctx context.Context, up Upstream) []upstream.ProxyOutbound {
	var result []upstream.ProxyOutbound

	ups := []upstream.UpstreamSubscriber{
		upstream.ClashVergeSubscriber{},
		upstream.SingBoxSubscriber{},
	}

	for _, subscriber := range ups {
		rs, err := subscriber.Outboards(ctx, client, up.URL)
		if err != nil {
			log.Printf("Failed to fetch from %s: %v", subscriber.Name(), err)
			continue
		}

		// 如果没有获取到任何出站，继续尝试下一个订阅源
		if len(rs) == 0 {
			continue
		}

		var filtered []upstream.ProxyOutbound
		if up.NodeKeywords != nil && len(up.NodeKeywords) > 0 {
			// filter rs by keywords

			for _, r := range rs {
				if AnyContained(r.Name(), up.NodeKeywords) {
					filtered = append(filtered, r)
				}
			}
		} else {
			filtered = rs
		}
		if len(filtered) > 0 {
			result = append(result, filtered...)
			break
		}
	}

	return result
}

// updateStore replaces the global OutboundsStore content atomically.
func updateStore(items []upstream.ProxyOutbound) {
	// Store new aggregated slice atomically
	store.Store(items)
}

// GetOutbounds returns a snapshot of current outbounds safely.
func GetOutbounds[T upstream.ProxyOutbound]() []T {
	// lock-free read snapshot
	v := store.Load()
	if v == nil {
		return nil
	}
	current := v.([]upstream.ProxyOutbound)
	// Build result with type assertion to T; skip non-matching types
	out := make([]T, 0, len(current))
	for _, it := range current {
		if cast, ok := it.(T); ok {
			out = append(out, cast)
		}
	}
	return out
}

// updatePerAndAggregate replaces perUpstream[url] and atomically updates aggregated store.
func updatePerAndAggregate(url string, items []upstream.ProxyOutbound) {
	perMu.Lock()
	defer perMu.Unlock()

	if perUpstream == nil {
		perUpstream = make(map[string][]upstream.ProxyOutbound)
	}

	// store a copy to avoid external mutation
	perUpstream[url] = append([]upstream.ProxyOutbound(nil), items...)

	// aggregate all slices
	total := 0
	for _, list := range perUpstream {
		total += len(list)
	}
	agg := make([]upstream.ProxyOutbound, 0, total)
	for _, list := range perUpstream {
		agg = append(agg, list...)
	}

	updateStore(agg)
}

func AnyContained(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
