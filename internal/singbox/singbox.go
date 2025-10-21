package singbox

import (
	"net/netip"
	"time"

	"github.com/dingdayu/go-project-template/internal/proxy"
	"github.com/dingdayu/go-project-template/internal/upstream"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
)

var dnsServers = []option.DNSServerOptions{
	{
		Type: C.DNSTypeHTTPS,
		Tag:  "google-doh",
		Options: option.RemoteHTTPSDNSServerOptions{
			RemoteTLSDNSServerOptions: option.RemoteTLSDNSServerOptions{
				RemoteDNSServerOptions: option.RemoteDNSServerOptions{
					DNSServerAddressOptions: option.DNSServerAddressOptions{
						Server:     "8.8.8.8",
						ServerPort: 443,
					},
					LocalDNSServerOptions: option.LocalDNSServerOptions{
						DialerOptions: option.DialerOptions{
							Detour: "auto-out",
						},
					},
				},

				OutboundTLSOptionsContainer: option.OutboundTLSOptionsContainer{
					TLS: &option.OutboundTLSOptions{
						ServerName: "dns.google",
					},
				},
			},
			Path: "/dns-query",
		},
	},
	{
		Type: C.DNSTypeHTTPS,
		Tag:  "alidns",
		Options: option.RemoteHTTPSDNSServerOptions{
			RemoteTLSDNSServerOptions: option.RemoteTLSDNSServerOptions{
				RemoteDNSServerOptions: option.RemoteDNSServerOptions{
					DNSServerAddressOptions: option.DNSServerAddressOptions{
						Server:     "223.5.5.5",
						ServerPort: 443,
					},
					LocalDNSServerOptions: option.LocalDNSServerOptions{
						DialerOptions: option.DialerOptions{
							Detour: "direct-out",
						},
					},
				},
				OutboundTLSOptionsContainer: option.OutboundTLSOptionsContainer{
					TLS: &option.OutboundTLSOptions{
						ServerName: "dns.alidns.com",
					},
				},
			},
			Path: "/dns-query",
		},
	},
	{
		Type: C.DNSTypeHTTPS,
		Tag:  "cloudflare-doh",
		Options: option.RemoteHTTPSDNSServerOptions{
			RemoteTLSDNSServerOptions: option.RemoteTLSDNSServerOptions{
				RemoteDNSServerOptions: option.RemoteDNSServerOptions{
					DNSServerAddressOptions: option.DNSServerAddressOptions{
						Server:     "1.1.1.1",
						ServerPort: 443,
					},
					LocalDNSServerOptions: option.LocalDNSServerOptions{
						DialerOptions: option.DialerOptions{
							Detour: "auto-out",
						},
					},
				},
				OutboundTLSOptionsContainer: option.OutboundTLSOptionsContainer{
					TLS: &option.OutboundTLSOptions{
						ServerName: "cloudflare-dns.com",
					},
				},
			},
			Path: "/dns-query",
		},
	},
}

var dnsRules = []option.DNSRule{
	// 1) adblock 优先且直接拒绝
	{
		Type: C.RuleTypeDefault,
		DefaultOptions: option.DefaultDNSRule{
			RawDefaultDNSRule: option.RawDefaultDNSRule{
				RuleSet: []string{"adblock"},
			},
			DNSRuleAction: option.DNSRuleAction{
				Action:        C.RuleActionTypeReject,
				RejectOptions: option.RejectActionOptions{},
			},
		},
	},
	// 保留原有按域名走 alidns 的特例
	{
		Type: C.RuleTypeDefault,
		DefaultOptions: option.DefaultDNSRule{
			RawDefaultDNSRule: option.RawDefaultDNSRule{
				DomainSuffix: []string{
					"onmicrosoft.cn",
					"s4b4.com",
					"github.com",
					"raw.githubusercontent.com",
				},
			},
			DNSRuleAction: option.DNSRuleAction{
				Action: C.RuleActionTypeRoute,
				RouteOptions: option.DNSRouteActionOptions{
					Server: "alidns",
				},
			},
		},
	},
	// 国内直连优先使用 alidns 解析
	{
		Type: C.RuleTypeDefault,
		DefaultOptions: option.DefaultDNSRule{
			RawDefaultDNSRule: option.RawDefaultDNSRule{
				RuleSet:   []string{"geosite-cn", "geoip-cn"},
				ClashMode: C.RuleActionTypeDirect,
			},
			DNSRuleAction: option.DNSRuleAction{
				Action: C.RuleActionTypeRoute,
				RouteOptions: option.DNSRouteActionOptions{
					Server: "alidns",
				},
			},
		},
	},
	// 全局兜底
	{
		Type: C.RuleTypeDefault,
		DefaultOptions: option.DefaultDNSRule{
			RawDefaultDNSRule: option.RawDefaultDNSRule{
				ClashMode: "global",
			},
			DNSRuleAction: option.DNSRuleAction{
				Action: C.RuleActionTypeRoute,
				RouteOptions: option.DNSRouteActionOptions{
					Server: "alidns",
				},
			},
		},
	},
}

// Create a netip.Addr variable to take its address (cannot take address of temporary value)
var (
	loopback     = netip.MustParseAddr("127.0.0.1")
	loopbackAddr = badoption.Addr(loopback)
)

var inbounds = []option.Inbound{
	{
		Type: "tun",
		Tag:  "tun-in",
		Options: option.TunInboundOptions{
			AutoRoute: true,
			Address: badoption.Listable[netip.Prefix]{
				netip.MustParsePrefix("172.19.0.1/30"),
				netip.MustParsePrefix("2001:0470:f9da:fdfa::1/64"),
			},
			MTU:         9000,
			StrictRoute: true, // 强制严格路由规则，可能导致部分流量无法访问，建议开启后测试可用性
		},
	},
	{
		Type: "socks",
		Tag:  "socks-in",
		Options: option.SocksInboundOptions{
			ListenOptions: option.ListenOptions{
				Listen:     &loopbackAddr,
				ListenPort: 2333,
			},
		},
	},
	{
		Type: "mixed",
		Tag:  "mixed-in",
		Options: option.HTTPMixedInboundOptions{
			ListenOptions: option.ListenOptions{
				Listen:     &loopbackAddr,
				ListenPort: 2334,
			},
		},
	},
}

var ruleSet = []option.RuleSet{
	{
		Type:   C.RuleSetTypeRemote,
		Tag:    "geosite-cn",
		Format: C.RuleSetFormatBinary,
		RemoteOptions: option.RemoteRuleSet{
			URL:            "https://jsd.onmicrosoft.cn/gh/SagerNet/sing-geosite@rule-set/geosite-cn.srs",
			DownloadDetour: "direct-out",
		},
	},
	{
		Type:   C.RuleSetTypeRemote,
		Tag:    "geoip-cn",
		Format: C.RuleSetFormatBinary,
		RemoteOptions: option.RemoteRuleSet{
			URL:            "https://jsd.onmicrosoft.cn/gh/SagerNet/sing-geoip@rule-set/geoip-cn.srs",
			DownloadDetour: "direct-out",
		},
	},
	{
		Type:   C.RuleSetTypeRemote,
		Tag:    "adblock",
		Format: C.RuleSetFormatBinary,
		RemoteOptions: option.RemoteRuleSet{
			URL:            "https://jsd.onmicrosoft.cn/gh/SagerNet/sing-geosite@rule-set/geosite-adblock.srs",
			DownloadDetour: "direct-out",
		},
	},
}

var rules = []option.Rule{
	// 2) adblock 路由层直接拒绝，优先级最高
	{
		Type: C.RuleTypeDefault,
		DefaultOptions: option.DefaultRule{
			RawDefaultRule: option.RawDefaultRule{
				RuleSet: []string{"adblock"},
			},
			RuleAction: option.RuleAction{
				Action:        C.RuleActionTypeReject,
				RejectOptions: option.RejectActionOptions{},
			},
		},
	},
	// 内网直连 + 国内直连
	{
		Type: C.RuleTypeDefault,
		DefaultOptions: option.DefaultRule{
			RawDefaultRule: option.RawDefaultRule{
				ClashMode:   C.RuleActionTypeDirect,
				IPIsPrivate: true,
				RuleSet:     []string{"geosite-cn", "geoip-cn"},
			},
			RuleAction: option.RuleAction{
				Action: C.RuleActionTypeRoute,
				RouteOptions: option.RouteActionOptions{
					Outbound: "direct-out",
				},
			},
		},
	},
	// 全局兜底 -> auto-out
	{
		Type: C.RuleTypeDefault,
		DefaultOptions: option.DefaultRule{
			RawDefaultRule: option.RawDefaultRule{
				ClashMode: "global",
			},
			RuleAction: option.RuleAction{
				Action: C.RuleActionTypeRoute,
				RouteOptions: option.RouteActionOptions{
					Outbound: "auto-out",
				},
			},
		},
	},
}

func defaultOptionsTags[T upstream.ProxyOutbound](ots []T) []option.Outbound {
	var otd []option.Outbound

	otd = append(otd, []option.Outbound{
		{
			Tag:     "direct-out",
			Type:    C.TypeDirect,
			Options: option.DirectOutboundOptions{},
		},
	}...)

	selectors, _ := GetSelectors()
	for _, sel := range selectors {
		switch sel.Name {
		case "auto-proxy":
			autoOutbounds := []string{}
			seen := make(map[string]bool)
			for _, ot := range ots {
				to, err := ot.ToOutbound()
				if err != nil || to.Tag == "" {
					continue
				}
				if len(sel.Keywords) == 0 || proxy.AnyContained(ot.Name(), sel.Keywords) {
					if !seen[to.Tag] {
						autoOutbounds = append(autoOutbounds, to.Tag)
						seen[to.Tag] = true
					}
				}
			}
			otd = append(otd, option.Outbound{
				Tag:  "auto-out",
				Type: "urltest",
				Options: option.URLTestOutboundOptions{
					URL:       "https://www.google.com/generate_204",
					Interval:  badoption.Duration(300 * time.Second),
					Tolerance: 50,
					Outbounds: autoOutbounds,
				},
			})
		case "ai-proxy":
			if len(sel.Keywords) == 0 {
				continue
			}
			var aiOutbounds []string
			for _, ot := range ots {
				if proxy.AnyContained(ot.Name(), sel.Keywords) {
					if to, err := ot.ToOutbound(); err == nil {
						aiOutbounds = append(aiOutbounds, to.Tag)
					}
				}
			}
			if len(aiOutbounds) > 0 {
				otd = append(otd, option.Outbound{
					Tag:  "ai-proxy",
					Type: C.TypeSelector,
					Options: option.SelectorOutboundOptions{
						Outbounds: aiOutbounds,
					},
				})
			}
		}
	}

	return otd
}

func OutboundToProfile[T upstream.ProxyOutbound](ots []T) (option.Options, error) {
	var opts option.Options

	outbounds := defaultOptionsTags(ots)
	for _, ot := range ots {
		if to, err := ot.ToOutbound(); err == nil {
			outbounds = append(outbounds, to)
		}
	}

	cRule := rules
	cRuleSet := ruleSet
	cDNSRules := dnsRules
	// 3) 仅当存在 ai-proxy 出站时，才追加 AI 相关路由与规则集
	if hasAIProxySelector(outbounds) {
		cRule = append(cRule, option.Rule{
			Type: C.RuleTypeDefault,
			DefaultOptions: option.DefaultRule{
				RawDefaultRule: option.RawDefaultRule{
					DomainSuffix: []string{"openai.com", "oaistatic.com", "oaiusercontent.com"},
					RuleSet:      []string{"openai", "gemini"},
					PackageName:  []string{"com.openai.chatgpt", "com.google.android.apps.bard", "com.google.bard"},
				},
				RuleAction: option.RuleAction{
					Action: C.RuleActionTypeRoute,
					RouteOptions: option.RouteActionOptions{
						Outbound: "ai-proxy",
					},
				},
			},
		})
		cRuleSet = append(cRuleSet, []option.RuleSet{
			{
				Type:   C.RuleSetTypeRemote,
				Tag:    "openai",
				Format: C.RuleSetFormatBinary,
				RemoteOptions: option.RemoteRuleSet{
					URL:            "https://jsd.onmicrosoft.cn/gh/SagerNet/sing-geosite@rule-set/geosite-openai.srs",
					DownloadDetour: "direct-out",
				},
			},
			{
				Type:   C.RuleSetTypeRemote,
				Tag:    "gemini",
				Format: C.RuleSetFormatBinary,
				RemoteOptions: option.RemoteRuleSet{
					URL:            "https://jsd.onmicrosoft.cn/gh/SagerNet/sing-geosite@rule-set/geosite-google-gemini.srs",
					DownloadDetour: "direct-out",
				},
			},
		}...)

		cDNSRules = append(cDNSRules, []option.DNSRule{
			// AI 站点优先走 cloudflare-doh
			{
				Type: C.RuleTypeDefault,
				DefaultOptions: option.DefaultDNSRule{
					RawDefaultDNSRule: option.RawDefaultDNSRule{
						RuleSet: []string{"openai", "gemini"},
					},
					DNSRuleAction: option.DNSRuleAction{
						Action: C.RuleActionTypeRoute,
						RouteOptions: option.DNSRouteActionOptions{
							Server: "cloudflare-doh",
						},
					},
				},
			},
		}...)
	}

	opts = option.Options{
		Log: &option.LogOptions{
			Level:     "info",
			Timestamp: true,
		},
		DNS: &option.DNSOptions{
			RawDNSOptions: option.RawDNSOptions{
				Servers: dnsServers,
				Rules:   cDNSRules,
				Final:   "alidns",
			},
		},
		Inbounds: inbounds,
		Route: &option.RouteOptions{
			AutoDetectInterface: true,
			// 4) 使用 cRuleSet（包含动态追加的规则集）
			RuleSet: cRuleSet,
			Rules:   cRule,
		},
		Outbounds: outbounds,
	}

	return opts, nil
}

// hasAIProxySelector checks if the outbounds contain an outbound with tag "ai-proxy"
func hasAIProxySelector(ots []option.Outbound) bool {
	for _, ot := range ots {
		if ot.Tag == "ai-proxy" {
			return true
		}
	}
	return false
}
