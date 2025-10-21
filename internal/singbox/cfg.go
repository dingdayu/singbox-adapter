package singbox

import (
	"fmt"

	"github.com/spf13/viper"
)

type SelectorItem struct {
	Name      string   `mapstructure:"name"`
	Type      string   `mapstructure:"type"`
	Upstreams []string `mapstructure:"upstreams"`
	Keywords  []string `mapstructure:"keywords"`
}

func GetSelectors() ([]SelectorItem, error) {
	var selectors []SelectorItem
	if err := viper.UnmarshalKey("selector", &selectors); err != nil {
		return nil, fmt.Errorf("singbox.GetSelectors: unable to decode 'selector' into struct: %v", err)
	}
	return selectors, nil
}
