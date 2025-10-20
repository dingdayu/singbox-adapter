package upstream

import (
	"context"
	"fmt"
	"os"
	"strings"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"go.yaml.in/yaml/v2"
	"resty.dev/v3"
)

type ClashVergeSubscriber struct{}

func (c ClashVergeSubscriber) Name() string {
	return "Clash Verge"
}

func (c ClashVergeSubscriber) UserAgent() string {
	return "clash-verge/v2.4.2"
}

func (c ClashVergeSubscriber) Profile(ctx context.Context, client *resty.Client, url string) (string, error) {
	resp, err := client.R().SetContext(ctx).SetHeader("User-Agent", c.UserAgent()).Get(url)
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

func (c ClashVergeSubscriber) Outboards(ctx context.Context, client *resty.Client, url string) ([]ProxyOutbound, error) {
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

	var profile ClashVergeProfile
	if err := yaml.Unmarshal(in, &profile); err != nil {
		return nil, fmt.Errorf("unmarshal profile failed: %w", err)
	}

	if len(profile.Proxies) == 0 {
		return nil, fmt.Errorf("no proxies found in profile")
	}
	// convert to []ProxyOutbound
	result := make([]ProxyOutbound, 0, len(profile.Proxies))
	for _, p := range profile.Proxies {
		// use value receiver impl
		cp := p
		result = append(result, cp)
	}
	return result, nil
}

// mapToPluginOpts converts a map to a semicolon-separated string for plugin options.
func mapToPluginOpts(m map[string]any) string {
	if len(m) == 0 {
		return ""
	}
	var opts []string
	for k, v := range m {
		switch k {
		case "mode":
			k = "obfs"
		case "host":
			k = "obfs-host"
		case "path":
			k = "obfs-path"
		}

		switch value := v.(type) {
		case string:
			opts = append(opts, k+"="+value)
		case []any:
			var s []string
			for _, item := range value {
				s = append(s, fmt.Sprint(item))
			}
			opts = append(opts, k+"="+strings.Join(s, ","))
		default:
			// 使用 fmt.Sprint 以避免类型断言 panic（支持 int/bool 等）
			opts = append(opts, k+"="+fmt.Sprint(v))
		}
	}
	return strings.Join(opts, ";")
}

type ClashVergeProfile struct {
	Proxies []ClashVergeProxy `yaml:"proxies"`
}

type ClashVergeProxy struct {
	// 重命名原来的 Name 字段以避免与 Name() 方法冲突
	NameRaw string `yaml:"name" json:"name"`
	Type    string `yaml:"type" json:"type"` // ss / trojan / vmess ...
	Server  string `yaml:"server" json:"server"`
	Port    int    `yaml:"port" json:"port"`

	// ---- Shadowsocks 通用
	Cipher   string `yaml:"cipher,omitempty" json:"cipher,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`

	// ---- SIP003 插件
	Plugin     string         `yaml:"plugin,omitempty" json:"plugin,omitempty"`           // obfs / v2ray-plugin ...
	PluginOpts map[string]any `yaml:"plugin-opts,omitempty" json:"plugin-opts,omitempty"` // 注意: 这是一个对象

	// ---- 其他常见开关
	UDP bool `yaml:"udp,omitempty" json:"udp,omitempty"`
	TFO bool `yaml:"tfo,omitempty" json:"tfo,omitempty"` // tcp fast open
}

// 将方法改为返回字段 NameRaw，避免自引用
func (p ClashVergeProxy) Name() string {
	return p.NameRaw
}

func (p ClashVergeProxy) ToOutbound() (option.Outbound, error) {
	switch p.Type {
	case "ss", "shadowsocks":
		return p.toShadowsocksOutbound(), nil
	case "trojan":
		return p.toTrojanOutbound(), nil
	case "hysteria2":
		return p.toHysteria2Outbound(), nil
	case "vmess":
		return p.toVmessOutbound(), nil
	case "vless":
		return p.toVlessOutbound(), nil
	// case "socks5", "socks":
	// 	return p.toSocks5Outbound(), nil
	// case "http", "https":
	// 	return p.toHTTPOutbound(), nil
	default:
		return option.Outbound{}, fmt.Errorf("not yet supported proxy type: %s", p.Type)
	}
}

func (p ClashVergeProxy) toShadowsocksOutbound() option.Outbound {
	var plugin string

	switch strings.ToLower(p.Plugin) {
	case "obfs":
		plugin = "obfs-local"
	case "v2ray-plugin":
		plugin = "v2ray-plugin"
	default:
		plugin = ""
	}

	return option.Outbound{
		Tag:  p.NameRaw,
		Type: C.TypeShadowsocks,

		Options: option.ShadowsocksOutboundOptions{
			ServerOptions: option.ServerOptions{
				Server:     p.Server,
				ServerPort: uint16(p.Port),
			},
			Method:        p.Cipher,
			Password:      p.Password,
			Plugin:        plugin,
			PluginOptions: mapToPluginOpts(p.PluginOpts),
		},
	}
}

func (p ClashVergeProxy) toTrojanOutbound() option.Outbound {
	return option.Outbound{
		Tag:  p.NameRaw,
		Type: C.TypeTrojan,

		Options: option.TrojanOutboundOptions{
			ServerOptions: option.ServerOptions{
				Server:     p.Server,
				ServerPort: uint16(p.Port),
			},
			Password: p.Password,
		},
	}
}

func (p ClashVergeProxy) toHysteria2Outbound() option.Outbound {
	return option.Outbound{
		Tag:  p.NameRaw,
		Type: C.TypeHysteria2,

		Options: option.Hysteria2OutboundOptions{
			ServerOptions: option.ServerOptions{
				Server:     p.Server,
				ServerPort: uint16(p.Port),
			},
		},
	}
}

func (p ClashVergeProxy) toVmessOutbound() option.Outbound {
	return option.Outbound{
		Tag:  p.NameRaw,
		Type: C.TypeVMess,

		Options: option.VMessOutboundOptions{
			ServerOptions: option.ServerOptions{
				Server:     p.Server,
				ServerPort: uint16(p.Port),
			},
			UUID: p.Password,
		},
	}
}

func (p ClashVergeProxy) toVlessOutbound() option.Outbound {
	return option.Outbound{
		Tag:  p.NameRaw,
		Type: C.TypeVLESS,

		Options: option.VLESSOutboundOptions{
			ServerOptions: option.ServerOptions{
				Server:     p.Server,
				ServerPort: uint16(p.Port),
			},
			UUID: p.Password,
		},
	}
}
