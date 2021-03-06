package authmethods_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/authmethods"
	"github.com/hashicorp/boundary/api/authtokens"
	"github.com/hashicorp/boundary/internal/servers/controller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticate(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	tc := controller.NewTestController(t, nil)
	defer tc.Shutdown()

	client := tc.Client()
	methods := authmethods.NewClient(client)

	tok, err := methods.Authenticate(tc.Context(), tc.Server().DevAuthMethodId, "login", map[string]interface{}{"login_name": "user", "password": "passpass"})
	require.NoError(err)
	assert.NotNil(tok)

	_, err = methods.Authenticate(tc.Context(), tc.Server().DevAuthMethodId, "login", map[string]interface{}{"login_name": "user", "password": "wrong"})
	require.Error(err)
	apiErr := api.AsServerError(err)
	require.NotNil(apiErr)
	assert.EqualValuesf(http.StatusUnauthorized, apiErr.Response().StatusCode(), "Expected unauthorized, got %q", apiErr.Message)

	// Also ensure that, for now, using "credentials" still works, as well as no command.
	reqBody := map[string]interface{}{
		"credentials": map[string]interface{}{"login_name": "user", "password": "passpass"},
	}
	req, err := client.NewRequest(tc.Context(), "POST", fmt.Sprintf("auth-methods/%s:authenticate", tc.Server().DevAuthMethodId), reqBody)
	require.NoError(err)
	resp, err := client.Do(req)
	require.NoError(err)

	target := new(authtokens.AuthTokenReadResult)
	target.Item = new(authtokens.AuthToken)
	apiErr, err = resp.Decode(target.Item)
	require.NoError(err)
	require.Nil(apiErr)
	require.NotNil(target.GetItem())
	require.NotEmpty(target.GetItem().(*authtokens.AuthToken).Token)
}
