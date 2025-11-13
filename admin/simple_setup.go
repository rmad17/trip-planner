package admin

import (
	"os"

	"github.com/GoAdminGroup/go-admin/engine"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/chartjs"
	"github.com/GoAdminGroup/themes/adminlte"
	"github.com/gin-gonic/gin"

	_ "github.com/GoAdminGroup/go-admin/adapter/gin"
	_ "github.com/GoAdminGroup/go-admin/modules/db/drivers/postgres"
)

func SimpleSetupGoAdmin(r *gin.Engine) {
	eng := engine.Default()

	adminConfig := config.Config{
		Databases: config.DatabaseList{
			"default": {
				Host:       "localhost",
				Port:       "5432",
				User:       "postgres",
				Pwd:        "postgres",
				Name:       "trip",
				MaxIdleCon: 50,
				MaxOpenCon: 150,
				Driver:     "postgres",
			},
		},
		UrlPrefix: "admin",
		Store: config.Store{
			Path:   "./uploads",
			Prefix: "uploads",
		},
		Language:    language.EN,
		IndexUrl:    "/",
		LoginUrl:    "/login",
		Debug:       os.Getenv("APP_ENV") == "development",
		ColorScheme: adminlte.ColorschemeSkinBlue,
		Title:       "Trip Planner Admin",
		Logo:        "<b>Trip</b>Planner",
		MiniLogo:    "TP",
		Theme:       "adminlte",
		LoginTitle:  "Trip Planner Admin",
		LoginLogo:   "<b>Trip</b>Planner Admin",
		AuthUserTable: "goadmin_users",
	}

	template.AddComp(chartjs.NewChart())

	if err := eng.AddConfig(&adminConfig).Use(r); err != nil {
		panic(err)
	}
}