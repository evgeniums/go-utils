package sms_test

import (
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context/app_default"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/sms/providers/gatewayapi"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
)

func TestGatewayapi(t *testing.T) {

	configFile := "gatewayapi_config.json"

	if !test_utils.ExternalConfigFileExists(configFile) || test_utils.Phone == "" {
		t.Skip("Skip TestGatewayapi because external config or phone not defined")
	}

	configFile = test_utils.ExternalConfigFilePath(configFile)
	phone := test_utils.Phone

	app := app_default.New(nil)
	err := app.Init(configFile)
	if err != nil {
		t.Fatalf("Failed to init application context: %s", err)
	}

	sender := gatewayapi.New()
	err = sender.Init(app.Cfg(), app.Logger(), app.Validator())
	if err == nil {
		t.Fatalf("expected failure, got passed")
	}
	err = sender.Init(app.Cfg(), app.Logger(), app.Validator(), "gatewayapi")
	if err != nil {
		t.Fatalf("Failed to init gatewayapi module: %s", err)
	}

	opCtx := &op_context.ContextBase{}
	opCtx.Init(app, app.Logger(), app.DB())

	resp, err := sender.Send(opCtx, "Test message", phone)
	opCtx.Close()
	if resp != nil {
		t.Logf("Response: %+v", resp)
	}
	if err != nil {
		t.Fatalf("Failed to send SMS: %s", err)
	}
}
