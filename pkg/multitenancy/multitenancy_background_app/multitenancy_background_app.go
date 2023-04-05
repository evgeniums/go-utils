package multitenancy_background_app

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context/app_default"
	"github.com/evgeniums/go-backend-helpers/pkg/db/db_gorm"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/multitenancy/app_with_multitenancy"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	finish "github.com/evgeniums/go-finish-service"
)

type MainRunner interface {
	Run(fin *finish.Finisher)
}

type BuildMainRunner = func(app app_with_multitenancy.AppWithMultitenancy, opCtx op_context.Context) (MainRunner, error)

func Exec(buildConfig *app_context.BuildConfig, tenancyDbModels *multitenancy.TenancyDbModels, buildMainRunner BuildMainRunner, appConfig ...app_with_multitenancy.AppConfigI) {

	appPath := app_default.Application()
	configPath := fmt.Sprintf("%s.jsonc", appPath[:len(appPath)-len(filepath.Ext(appPath))])

	// get name of configuration file
	configFile := flag.String("config", configPath, "Configuration file")
	flag.Parse()
	fmt.Printf("Using config file %v\n", *configFile)

	// init db gorm models
	db_gorm.NewModelStore(true)

	// init app context
	app := app_with_multitenancy.NewApp(buildConfig, tenancyDbModels)
	initOpCtx, err := app.InitWithArgs(*configFile, flag.Args())
	if err != nil {
		initOpCtx.Close()
		app_context.AbortFatal(app, "failed to init application context", err)
	}
	defer app.Close()

	// create main runner
	runner, err := buildMainRunner(app, initOpCtx)
	initOpCtx.Close()
	if err != nil {
		app_context.AbortFatal(app, "failed to init main runner", err)
	}

	// run
	startedMsg := fmt.Sprintf("%s started", app.Application())
	app.Logger().Info(startedMsg)
	fmt.Println(startedMsg)
	fin := finish.New()
	runner.Run(fin)
	fin.Add(app.Pubsub())

	// wait for signals
	fin.Wait()

	// done
	finishedMsg := fmt.Sprintf("%s finished", app.Application())
	app.Logger().Info(finishedMsg)
	fmt.Println(finishedMsg)
}
