package admin

import (
	"os"

	"github.com/GoAdminGroup/go-admin/engine"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/chartjs"
	"github.com/GoAdminGroup/themes/adminlte"
	"github.com/gin-gonic/gin"

	_ "github.com/GoAdminGroup/go-admin/adapter/gin"
	_ "github.com/GoAdminGroup/go-admin/modules/db/drivers/postgres"

	"triplanner/admin/tables"
)

func SetupGoAdmin(r *gin.Engine) {
	eng := engine.Default()

	adminConfig := engine.Config{
		Databases: engine.DatabaseList{
			"default": {
				Host:       os.Getenv("DB_HOST"),
				Port:       os.Getenv("DB_PORT"),
				User:       os.Getenv("DB_USER"),
				Pwd:        os.Getenv("DB_PASSWORD"),
				Name:       os.Getenv("DB_NAME"),
				MaxIdleCon: 50,
				MaxOpenCon: 150,
				Driver:     "postgres",
			},
		},
		UrlPrefix: "admin",
		Store: engine.Store{
			Path:   "./uploads",
			Prefix: "uploads",
		},
		Language:    engine.EN,
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

	if err := eng.AddConfig(&adminConfig).
		AddGenerators(tables.Generators).
		Use(r); err != nil {
		panic(err)
	}

	eng.HTML("GET", "/admin", tables.GetContent)
}