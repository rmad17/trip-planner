package tables

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

// GetUserTable returns the user table configuration
func GetUserTable(ctx *context.Context) table.Table {
	userTable := table.NewDefaultTable(ctx)

	info := userTable.GetInfo().HideFilterArea()

	info.AddField("ID", "id", db.Text).
		FieldSortable()
	info.AddField("Email", "email", db.Varchar).
		FieldFilterable()
	info.AddField("First Name", "first_name", db.Varchar).
		FieldFilterable()
	info.AddField("Last Name", "last_name", db.Varchar).
		FieldFilterable()
	info.AddField("Verified", "is_verified", db.Bool).
		FieldDisplay(func(value types.FieldModel) interface{} {
			if value.Value == "true" {
				return "Yes"
			}
			return "No"
		})
	info.AddField("Email Verified At", "email_verified_at", db.Timestamp).
		FieldSortable()
	info.AddField("Created At", "created_at", db.Timestamp).
		FieldSortable()

	info.SetTable("users").SetTitle("Users").SetDescription("System Users")

	formList := userTable.GetForm()

	formList.AddField("Email", "email", db.Varchar, form.Email).
		FieldMust().
		FieldHelpMsg("User's email address")
	formList.AddField("First Name", "first_name", db.Varchar, form.Text).
		FieldMust()
	formList.AddField("Last Name", "last_name", db.Varchar, form.Text).
		FieldMust()
	formList.AddField("Password", "password_hash", db.Varchar, form.Password).
		FieldHelpMsg("Leave blank to keep current password")
	formList.AddField("Verified", "is_verified", db.Bool, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: "Yes", Value: "true"},
			{Text: "No", Value: "false"},
		}).
		FieldDefault("false")

	formList.SetTable("users").SetTitle("Users").SetDescription("System Users")

	return userTable
}