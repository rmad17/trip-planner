package tables

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

// GetTravellerTable returns the traveller table configuration
func GetTravellerTable(ctx *context.Context) table.Table {
	travellerTable := table.NewDefaultTable(ctx)

	info := travellerTable.GetInfo().HideFilterArea()

	info.AddField("ID", "id", db.Text).
		FieldSortable()
	info.AddField("First Name", "first_name", db.Varchar).
		FieldFilterable()
	info.AddField("Last Name", "last_name", db.Varchar).
		FieldFilterable()
	info.AddField("Email", "email", db.Varchar).
		FieldFilterable()
	info.AddField("Phone", "phone", db.Varchar)
	info.AddField("Nationality", "nationality", db.Varchar)
	info.AddField("Role", "role", db.Varchar)
	info.AddField("Is Active", "is_active", db.Bool).
		FieldDisplay(func(value types.FieldModel) interface{} {
			if value.Value == "true" {
				return "Yes"
			}
			return "No"
		})
	info.AddField("Joined At", "joined_at", db.Timestamp).
		FieldSortable()

	info.SetTable("travellers").SetTitle("Travellers").SetDescription("Trip Participants")

	formList := travellerTable.GetForm()

	formList.AddField("First Name", "first_name", db.Varchar, form.Text).
		FieldMust()
	formList.AddField("Last Name", "last_name", db.Varchar, form.Text).
		FieldMust()
	formList.AddField("Email", "email", db.Varchar, form.Email)
	formList.AddField("Phone", "phone", db.Varchar, form.Text).
		FieldHelpMsg("Include country code")
	formList.AddField("Date of Birth", "date_of_birth", db.Date, form.Datetime)
	formList.AddField("Nationality", "nationality", db.Varchar, form.Text).
		FieldHelpMsg("ISO country code")
	formList.AddField("Passport Number", "passport_number", db.Varchar, form.Text)
	formList.AddField("Passport Expiry", "passport_expiry", db.Date, form.Datetime)
	formList.AddField("Emergency Contact", "emergency_contact", db.Text, form.TextArea)
	formList.AddField("Dietary Restrictions", "dietary_restrictions", db.Text, form.TextArea)
	formList.AddField("Medical Notes", "medical_notes", db.Text, form.TextArea)
	formList.AddField("Role", "role", db.Varchar, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "Organizer", Value: "organizer"},
			{Text: "Participant", Value: "participant"},
			{Text: "Guest", Value: "guest"},
		}).
		FieldDefault("participant")
	formList.AddField("Is Active", "is_active", db.Bool, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: "Yes", Value: "true"},
			{Text: "No", Value: "false"},
		}).
		FieldDefault("true")
	formList.AddField("Notes", "notes", db.Text, form.TextArea)

	formList.SetTable("travellers").SetTitle("Travellers").SetDescription("Trip Participants")

	return travellerTable
}