// Copyright (c) 2026 REXI-hub team
//
// This file is part of REXI-hub v2.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"rexi-hub-v2/common"
	"rexi-hub-v2/model"
	"rexi-hub-v2/setting"
	"rexi-hub-v2/setting/system_setting"

	"github.com/gin-gonic/gin"
)

type OidcResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type OidcUser struct {
	OpenID            string `json:"sub"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Picture           string `json:"picture"`
}

func getOidcUserInfoByCode(code string) (*OidcUser, error) {
	if code == "" {
		return nil, errors.New("Invalid parameters")
	}

	values := url.Values{}
	values.Set("client_id", system_setting.GetOIDCSettings().ClientId)
	values.Set("client_secret", system_setting.GetOIDCSettings().ClientSecret)
	values.Set("code", code)
	values.Set("grant_type", "authorization_code")
	values.Set("redirect_uri", fmt.Sprintf("%s/oauth/oidc", setting.ServerAddress))
	formData := values.Encode()
	req, err := http.NewRequest("POST", system_setting.GetOIDCSettings().TokenEndpoint, strings.NewReader(formData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		common.SysLog(err.Error())
		return nil, errors.New("Unable to connect to the OIDC server, please try again later!")
	}
	defer res.Body.Close()
	var oidcResponse OidcResponse
	err = json.NewDecoder(res.Body).Decode(&oidcResponse)
	if err != nil {
		return nil, err
	}

	if oidcResponse.AccessToken == "" {
		common.SysError("OIDC failed to get token, please check the settings!")
		return nil, errors.New("OIDC failed to get token, please check the settings!")
	}

	req, err = http.NewRequest("GET", system_setting.GetOIDCSettings().UserInfoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+oidcResponse.AccessToken)
	res2, err := client.Do(req)
	if err != nil {
		common.SysLog(err.Error())
		return nil, errors.New("Unable to connect to the OIDC server, please try again later!")
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		common.SysError("OIDC failed to get user info! Please check the settings!")
		return nil, errors.New("OIDC failed to get user info! Please check the settings!")
	}

	var oidcUser OidcUser
	err = json.NewDecoder(res2.Body).Decode(&oidcUser)
	if err != nil {
		return nil, err
	}
	if oidcUser.OpenID == "" || oidcUser.Email == "" {
		common.SysError("OIDC returned empty user info! Please check the settings!")
		return nil, errors.New("OIDC returned empty user info! Please check the settings!")
	}
	return &oidcUser, nil
}

func OidcAuth(c *gin.Context) {
	code := c.Query("code")
	oidcUser, err := getOidcUserInfoByCode(code)
	if err != nil {
		respondWithError(c, http.StatusOK, err.Error())
		return
	}

	oauthUser := &OAuthUser{
		ID:          oidcUser.OpenID,
		Username:    oidcUser.PreferredUsername,
		DisplayName: oidcUser.Name,
		Email:       oidcUser.Email,
		Provider:    ProviderOIDC,
	}

	oidcSettings := system_setting.GetOIDCSettings()
	config := &OAuthConfig{
		Enabled:      oidcSettings.Enabled,
		ClientID:     oidcSettings.ClientId,
		ClientSecret: oidcSettings.ClientSecret,
	}

	handleOAuthLogin(c, oauthUser, config,
		model.IsOidcIdAlreadyTaken,
		func(user *model.User) error {
			user.OidcId = oauthUser.ID
			return user.FillUserByOidcId()
		},
		createOIDCUser,
	)
}

func OidcBind(c *gin.Context) {
	code := c.Query("code")
	oidcUser, err := getOidcUserInfoByCode(code)
	if err != nil {
		respondWithError(c, http.StatusOK, err.Error())
		return
	}

	oauthUser := &OAuthUser{
		ID:          oidcUser.OpenID,
		Username:    oidcUser.PreferredUsername,
		DisplayName: oidcUser.Name,
		Email:       oidcUser.Email,
		Provider:    ProviderOIDC,
	}

	oidcSettings := system_setting.GetOIDCSettings()
	config := &OAuthConfig{
		Enabled:      oidcSettings.Enabled,
		ClientID:     oidcSettings.ClientId,
		ClientSecret: oidcSettings.ClientSecret,
	}

	handleOAuthBind(c, oauthUser, config,
		model.IsOidcIdAlreadyTaken,
		func(user *model.User) error {
			user.OidcId = oauthUser.ID
			return user.FillUserByOidcId()
		},
	)
}
