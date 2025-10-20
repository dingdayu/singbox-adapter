package token

import (
	"fmt"

	"github.com/spf13/viper"
)

type Token struct {
	Token    string   `yaml:"token"`
	Keywords []string `yaml:"keywords"`
}

func GetToken(tks string) (Token, error) {
	var tokens []Token
	if err := viper.UnmarshalKey("tokens", &tokens); err != nil {
		return Token{}, err
	}
	for _, tk := range tokens {
		if tk.Token == tks {
			return tk, nil
		}
	}
	return Token{}, fmt.Errorf("not found token: %s", tks)
}
