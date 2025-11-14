package tables

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

// GetTripPlanTable returns the trip plan table configuration
func GetTripPlanTable(ctx *context.Context) table.Table {
	tripPlanTable := table.NewDefaultTable(ctx)

	info := tripPlanTable.GetInfo().HideFilterArea()

	info.AddField("ID", "id", db.Text).
		FieldSortable()
	info.AddField("Name", "name", db.Varchar).
		FieldFilterable()
	info.AddField("Description", "description", db.Text).
		FieldLimit(100)
	info.AddField("Start Date", "start_date", db.Date).
		FieldSortable()
	info.AddField("End Date", "end_date", db.Date).
		FieldSortable()
	info.AddField("Budget", "budget", db.Decimal).
		FieldDisplay(func(value types.FieldModel) interface{} {
			if value.Value != "" {
				return "$" + value.Value
			}
			return "N/A"
		})
	info.AddField("Currency", "currency", db.Varchar)
	info.AddField("Status", "status", db.Varchar).
		FieldFilterable()
	info.AddField("Is Public", "is_public", db.Bool).
		FieldDisplay(func(value types.FieldModel) interface{} {
			if value.Value == "true" {
				return "Yes"
			}
			return "No"
		})
	info.AddField("Created At", "created_at", db.Timestamp).
		FieldSortable()

	info.SetTable("trip_plans").SetTitle("Trip Plans").SetDescription("All Trip Plans")

	formList := tripPlanTable.GetForm()

	formList.AddField("Name", "name", db.Varchar, form.Text).
		FieldMust()
	formList.AddField("Description", "description", db.Text, form.TextArea)
	formList.AddField("Start Date", "start_date", db.Date, form.Datetime)
	formList.AddField("End Date", "end_date", db.Date, form.Datetime)
	formList.AddField("Min Days", "min_days", db.Int, form.Number)
	formList.AddField("Max Days", "max_days", db.Int, form.Number)
	formList.AddField("Travel Mode", "travel_mode", db.Varchar, form.Text)
	formList.AddField("Trip Type", "trip_type", db.Varchar, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "Leisure", Value: "leisure"},
			{Text: "Business", Value: "business"},
			{Text: "Adventure", Value: "adventure"},
			{Text: "Family", Value: "family"},
			{Text: "Cultural", Value: "cultural"},
		})
	formList.AddField("Budget", "budget", db.Decimal, form.Currency)
	formList.AddField("Currency", "currency", db.Varchar, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "USD", Value: "USD"},
			{Text: "EUR", Value: "EUR"},
			{Text: "GBP", Value: "GBP"},
			{Text: "INR", Value: "INR"},
			{Text: "CAD", Value: "CAD"},
			{Text: "AUD", Value: "AUD"},
			{Text: "JPY", Value: "JPY"},
			{Text: "Other", Value: "OTHER"},
		}).
		FieldDefault("USD")
	formList.AddField("Status", "status", db.Varchar, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "Planning", Value: "planning"},
			{Text: "Confirmed", Value: "confirmed"},
			{Text: "In Progress", Value: "in_progress"},
			{Text: "Completed", Value: "completed"},
			{Text: "Cancelled", Value: "cancelled"},
		}).
		FieldDefault("planning")
	formList.AddField("Is Public", "is_public", db.Bool, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: "Yes", Value: "true"},
			{Text: "No", Value: "false"},
		}).
		FieldDefault("false")
	formList.AddField("Notes", "notes", db.Text, form.TextArea)

	formList.SetTable("trip_plans").SetTitle("Trip Plans").SetDescription("All Trip Plans")

	return tripPlanTable
}
