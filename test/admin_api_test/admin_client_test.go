package user_api_test

import (
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/admin/admin_api/admin_api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client/rest_api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/gin-gonic/gin"
)

const BaseUrl = "http://localhost"

func initClient(t *testing.T, g *gin.Engine, config ...string) (app_context.Context, api_client.Client, *admin_api_client.AdminClient) {
	app := test_utils.InitAppContextNoDb(t, testDir, utils.OptionalArg("admin_api_client.jsonc", config...))

	restApiClient := test_utils.RestApiTestClient(t, g, BaseUrl)
	client := rest_api_client.New(restApiClient)

	adminClient := admin_api_client.NewAdminClient(client)

	return app, client, adminClient
}

func TestInitClient(t *testing.T) {
	serverApp, _, server := initServer(t)
	clientApp, _, _ := initClient(t, test_utils.BBGinEngine(t, server))
	serverApp.Close()
	clientApp.Close()
}
