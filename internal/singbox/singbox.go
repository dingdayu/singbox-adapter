package singbox

import (
	"net/netip"
	"strings"
	"time"

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
	{
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
	},
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

func defaultOptionsTags(aiOutbounds []string, testOutbounds []string) []option.Outbound {
	var ots []option.Outbound

	ots = append(ots, []option.Outbound{
		{
			Tag:     "direct-out",
			Type:    C.TypeDirect,
			Options: option.DirectOutboundOptions{},
		},
	}...)

	if len(aiOutbounds) > 0 {
		ots = append(ots, option.Outbound{
			Tag:  "ai-proxy",
			Type: C.TypeSelector,
			Options: option.SelectorOutboundOptions{
				Outbounds: aiOutbounds,
			},
		})
	}
	if len(testOutbounds) > 0 {
		ots = append(ots, option.Outbound{
			Tag:  "auto-out",
			Type: "urltest",
			Options: option.URLTestOutboundOptions{
				URL:       "https://www.google.com/generate_204",
				Interval:  badoption.Duration(300 * time.Second),
				Tolerance: 50,
				Outbounds: testOutbounds,
			},
		})
	}
	return ots
}

func OutboundToProfile[T upstream.ProxyOutbound](ots []T) (option.Options, error) {
	var opts option.Options

	var aiOutboundTags []string
	var testOutboundTags []string

	for _, ot := range ots {
		ot, err := ot.ToOutbound()
		if err != nil || ot.Tag == "" {
			continue
		}

		testOutboundTags = append(testOutboundTags, ot.Tag)
		if strings.Contains(strings.ToLower(ot.Tag), "台北") || strings.Contains(strings.ToLower(ot.Tag), "jp") || strings.Contains(strings.ToLower(ot.Tag), "us") || strings.Contains(strings.ToLower(ot.Tag), "sg") || strings.Contains(strings.ToLower(ot.Tag), "de") || strings.Contains(strings.ToLower(ot.Tag), "tw") || strings.Contains(strings.ToLower(ot.Tag), "tr") || strings.Contains(strings.ToLower(ot.Tag), "kr") || strings.Contains(strings.ToLower(ot.Tag), "gb") {
			aiOutboundTags = append(aiOutboundTags, ot.Tag)
		}
	}

	outbounds := defaultOptionsTags(aiOutboundTags, testOutboundTags)
	for _, ot := range ots {
		if to, err := ot.ToOutbound(); err == nil {
			outbounds = append(outbounds, to)
		}
	}
	opts = option.Options{
		Log: &option.LogOptions{
			Level:     "info",
			Timestamp: true,
		},
		DNS: &option.DNSOptions{
			RawDNSOptions: option.RawDNSOptions{
				Servers: dnsServers,
				Rules:   dnsRules,
				Final:   "alidns",
			},
		},
		Inbounds: inbounds,
		Route: &option.RouteOptions{
			AutoDetectInterface: true,
			RuleSet:             ruleSet,
			Rules:               rules,
		},
		Outbounds: outbounds,
	}

	return opts, nil
}
