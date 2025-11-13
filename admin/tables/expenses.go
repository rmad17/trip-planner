package tables

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

// GetExpenseTable returns the expense table configuration
func GetExpenseTable(ctx *context.Context) table.Table {
	expenseTable := table.NewDefaultTable(ctx)

	info := expenseTable.GetInfo().HideFilterArea()

	info.AddField("ID", "id", db.Text).
		FieldSortable()
	info.AddField("Title", "title", db.Varchar).
		FieldFilterable()
	info.AddField("Amount", "amount", db.Decimal).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return "$" + value.Value
		})
	info.AddField("Currency", "currency", db.Varchar)
	info.AddField("Category", "category", db.Varchar).
		FieldFilterable()
	info.AddField("Date", "date", db.Timestamp).
		FieldSortable()
	info.AddField("Location", "location", db.Varchar)
	info.AddField("Payment Method", "payment_method", db.Varchar)
	info.AddField("Split Method", "split_method", db.Varchar)
	info.AddField("Created At", "created_at", db.Timestamp).
		FieldSortable()

	info.SetTable("expenses").SetTitle("Expenses").SetDescription("Trip Expenses")

	formList := expenseTable.GetForm()

	formList.AddField("Title", "title", db.Varchar, form.Text).
		FieldMust()
	formList.AddField("Description", "description", db.Text, form.TextArea)
	formList.AddField("Amount", "amount", db.Decimal, form.Currency).
		FieldMust()
	formList.AddField("Currency", "currency", db.Varchar, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "USD", Value: "USD"},
			{Text: "EUR", Value: "EUR"},
			{Text: "GBP", Value: "GBP"},
			{Text: "INR", Value: "INR"},
		}).
		FieldDefault("USD")
	formList.AddField("Category", "category", db.Varchar, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "Accommodation", Value: "accommodation"},
			{Text: "Transportation", Value: "transportation"},
			{Text: "Food", Value: "food"},
			{Text: "Activities", Value: "activities"},
			{Text: "Shopping & Gifts", Value: "shopping_gifts"},
			{Text: "Insurance", Value: "insurance"},
			{Text: "Visas & Fees", Value: "visas_fees"},
			{Text: "Medical", Value: "medical"},
			{Text: "Communication", Value: "communication"},
			{Text: "Miscellaneous", Value: "miscellaneous"},
			{Text: "Other", Value: "other"},
		})
	formList.AddField("Other Category", "other_category", db.Varchar, form.Text).
		FieldHelpMsg("Used when category is 'other'")
	formList.AddField("Date", "date", db.Timestamp, form.Datetime).
		FieldMust()
	formList.AddField("Location", "location", db.Varchar, form.Text)
	formList.AddField("Vendor", "vendor", db.Varchar, form.Text)
	formList.AddField("Payment Method", "payment_method", db.Varchar, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "Cash", Value: "cash"},
			{Text: "Card", Value: "card"},
			{Text: "Digital Pay", Value: "digital_pay"},
			{Text: "Bank Transfer", Value: "bank_transfer"},
			{Text: "Cheque", Value: "cheque"},
			{Text: "Other", Value: "other"},
		}).
		FieldDefault("card")
	formList.AddField("Split Method", "split_method", db.Varchar, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "Equal", Value: "equal"},
			{Text: "Exact", Value: "exact"},
			{Text: "Percentage", Value: "percentage"},
			{Text: "Shares", Value: "shares"},
			{Text: "Paid By", Value: "paid_by"},
		}).
		FieldDefault("equal")
	formList.AddField("Receipt URL", "receipt_url", db.Varchar, form.Url)
	formList.AddField("Notes", "notes", db.Text, form.TextArea)
	formList.AddField("Is Recurring", "is_recurring", db.Bool, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: "Yes", Value: "true"},
			{Text: "No", Value: "false"},
		}).
		FieldDefault("false")

	formList.SetTable("expenses").SetTitle("Expenses").SetDescription("Trip Expenses")

	return expenseTable
}