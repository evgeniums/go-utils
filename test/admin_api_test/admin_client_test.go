package admin_api_test

import (
	"testing"

	"github.com/evgeniums/go-utils/pkg/admin/admin_api/admin_api_client"
	"github.com/evgeniums/go-utils/pkg/api/api_client/rest_api_client"
	"github.com/evgeniums/go-utils/pkg/app_context"
	"github.com/evgeniums/go-utils/pkg/test_utils"
	"github.com/evgeniums/go-utils/pkg/utils"
	"github.com/gin-gonic/gin"
)

const BaseUrl = "http://localhost/api/1.0.0"

func initClient(t *testing.T, g *gin.Engine, config ...string) (app_context.Context, *admin_api_client.AdminClient) {
	app := test_utils.InitDefaultAppContextNoDb(t, testDir, utils.OptionalArg("admin_api_client.jsonc", config...))

	opCtx := test_utils.SimpleOpContext(app, "prepare")
	restApiClient := test_utils.RestApiTestClient(t, g, BaseUrl)
	restApiClient.Prepare(opCtx)

	client := rest_api_client.New(restApiClient)

	adminClient := admin_api_client.NewAdminClient(client)

	return app, adminClient
}

func TestInitClient(t *testing.T) {
	serverApp, _, server := initServer(t)
	clientApp, _ := initClient(t, test_utils.BBGinEngine(t, server))
	serverApp.Close()
	clientApp.Close()
}
