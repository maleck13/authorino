package metadata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/3scale-labs/authorino/pkg/config/common"
	"github.com/3scale-labs/authorino/pkg/config/identity"
)

type UserInfo struct {
	OIDC         string `yaml:"oidc,omitempty"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

func (self *UserInfo) Call(ctx common.AuthContext) (interface{}, error) {
	// find oidc config and the userinfo endpoint
	idConfig, _ := ctx.FindIdentityByName(self.OIDC)

	if idConfig == nil {
		return nil, fmt.Errorf("Null OIDC object for config %v. Skipping related UserInfo metadata.", self.OIDC)
	}

	idConfigStruct := idConfig.(*identity.OIDC)
	provider, _ := idConfigStruct.NewProvider(ctx)
	var providerClaims map[string]interface{}
	_ = provider.Claims(&providerClaims)
	userInfoURL, _ := url.Parse(providerClaims["introspection_endpoint"].(string))
	userInfoURL.User = url.UserPassword(self.ClientID, self.ClientSecret)

	// extract access token
	accessToken, _ := ctx.AuthorizationToken()

	// fetch user info
	formData := url.Values{
		"token":           {accessToken},
		"token_type_hint": {"requesting_party_token"},
	}
	resp, err := http.PostForm(userInfoURL.String(), formData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// parse the response
	var claims map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&claims)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
