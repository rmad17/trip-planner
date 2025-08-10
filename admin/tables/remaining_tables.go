package tables

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

// GetTripHopTable returns the trip hop table configuration
func GetTripHopTable(ctx *context.Context) table.Table {
	tripHopTable := table.NewDefaultTable(table.DefaultConfig())

	info := tripHopTable.GetInfo().HideFilterArea()
	info.AddField("ID", "id", db.Text)
	info.AddField("Name", "name", db.Varchar)
	info.AddField("City", "city", db.Varchar)
	info.AddField("Country", "country", db.Varchar)
	info.AddField("Start Date", "start_date", db.Date)
	info.AddField("End Date", "end_date", db.Date)

	info.SetTable("trip_hops").SetTitle("Trip Hops").SetDescription("Trip Destinations")

	formList := tripHopTable.GetForm()
	formList.AddField("Name", "name", db.Varchar, form.Text)
	formList.AddField("City", "city", db.Varchar, form.Text)
	formList.AddField("Country", "country", db.Varchar, form.Text)
	formList.SetTable("trip_hops")

	return tripHopTable
}

// GetTripDayTable returns the trip day table configuration
func GetTripDayTable(ctx *context.Context) table.Table {
	tripDayTable := table.NewDefaultTable(table.DefaultConfig())

	info := tripDayTable.GetInfo().HideFilterArea()
	info.AddField("ID", "id", db.Text)
	info.AddField("Date", "date", db.Date)
	info.AddField("Day Number", "day_number", db.Int)
	info.AddField("Title", "title", db.Varchar)
	info.AddField("Day Type", "day_type", db.Varchar)

	info.SetTable("trip_days").SetTitle("Trip Days").SetDescription("Daily Itinerary")

	formList := tripDayTable.GetForm()
	formList.AddField("Date", "date", db.Date, form.Datetime)
	formList.AddField("Title", "title", db.Varchar, form.Text)
	formList.SetTable("trip_days")

	return tripDayTable
}

// GetActivityTable returns the activity table configuration
func GetActivityTable(ctx *context.Context) table.Table {
	activityTable := table.NewDefaultTable(table.DefaultConfig())

	info := activityTable.GetInfo().HideFilterArea()
	info.AddField("ID", "id", db.Text)
	info.AddField("Name", "name", db.Varchar)
	info.AddField("Activity Type", "activity_type", db.Varchar)
	info.AddField("Start Time", "start_time", db.Timestamp)
	info.AddField("Location", "location", db.Varchar)

	info.SetTable("activities").SetTitle("Activities").SetDescription("Trip Activities")

	formList := activityTable.GetForm()
	formList.AddField("Name", "name", db.Varchar, form.Text)
	formList.AddField("Activity Type", "activity_type", db.Varchar, form.Text)
	formList.SetTable("activities")

	return activityTable
}

// GetStayTable returns the stay table configuration
func GetStayTable(ctx *context.Context) table.Table {
	stayTable := table.NewDefaultTable(table.DefaultConfig())

	info := stayTable.GetInfo().HideFilterArea()
	info.AddField("ID", "id", db.Text)
	info.AddField("Stay Type", "stay_type", db.Varchar)
	info.AddField("Start Date", "start_date", db.Date)
	info.AddField("End Date", "end_date", db.Date)
	info.AddField("Is Prepaid", "is_prepaid", db.Bool)

	info.SetTable("stays").SetTitle("Stays").SetDescription("Accommodations")

	formList := stayTable.GetForm()
	formList.AddField("Stay Type", "stay_type", db.Varchar, form.Text)
	formList.SetTable("stays")

	return stayTable
}

// GetDocumentTable returns the document table configuration
func GetDocumentTable(ctx *context.Context) table.Table {
	documentTable := table.NewDefaultTable(table.DefaultConfig())

	info := documentTable.GetInfo().HideFilterArea()
	info.AddField("ID", "id", db.Text)
	info.AddField("Name", "name", db.Varchar)
	info.AddField("Original Name", "original_name", db.Varchar)
	info.AddField("Category", "category", db.Varchar)
	info.AddField("File Size", "file_size", db.Bigint)
	info.AddField("Uploaded At", "uploaded_at", db.Timestamp)

	info.SetTable("documents").SetTitle("Documents").SetDescription("Uploaded Files")

	formList := documentTable.GetForm()
	formList.AddField("Name", "name", db.Varchar, form.Text)
	formList.AddField("Category", "category", db.Varchar, form.Text)
	formList.SetTable("documents")

	return documentTable
}

// GetDocumentShareTable returns the document share table configuration
func GetDocumentShareTable(ctx *context.Context) table.Table {
	shareTable := table.NewDefaultTable(table.DefaultConfig())

	info := shareTable.GetInfo().HideFilterArea()
	info.AddField("ID", "id", db.Text)
	info.AddField("Permission", "permission", db.Varchar)
	info.AddField("Is Active", "is_active", db.Bool)
	info.AddField("Expires At", "expires_at", db.Timestamp)

	info.SetTable("document_shares").SetTitle("Document Shares").SetDescription("Shared Documents")

	formList := shareTable.GetForm()
	formList.AddField("Permission", "permission", db.Varchar, form.Text)
	formList.SetTable("document_shares")

	return shareTable
}

// GetExpenseSplitTable returns the expense split table configuration
func GetExpenseSplitTable(ctx *context.Context) table.Table {
	splitTable := table.NewDefaultTable(table.DefaultConfig())

	info := splitTable.GetInfo().HideFilterArea()
	info.AddField("ID", "id", db.Text)
	info.AddField("Amount", "amount", db.Decimal)
	info.AddField("Percentage", "percentage", db.Decimal)
	info.AddField("Shares", "shares", db.Int)
	info.AddField("Is Paid", "is_paid", db.Bool)

	info.SetTable("expense_splits").SetTitle("Expense Splits").SetDescription("Expense Sharing")

	formList := splitTable.GetForm()
	formList.AddField("Amount", "amount", db.Decimal, form.Currency)
	formList.SetTable("expense_splits")

	return splitTable
}

// GetExpenseSettlementTable returns the expense settlement table configuration
func GetExpenseSettlementTable(ctx *context.Context) table.Table {
	settlementTable := table.NewDefaultTable(table.DefaultConfig())

	info := settlementTable.GetInfo().HideFilterArea()
	info.AddField("ID", "id", db.Text)
	info.AddField("Amount", "amount", db.Decimal)
	info.AddField("Currency", "currency", db.Varchar)
	info.AddField("Status", "status", db.Varchar)
	info.AddField("Settled At", "settled_at", db.Timestamp)

	info.SetTable("expense_settlements").SetTitle("Expense Settlements").SetDescription("Payment Settlements")

	formList := settlementTable.GetForm()
	formList.AddField("Amount", "amount", db.Decimal, form.Currency)
	formList.AddField("Status", "status", db.Varchar, form.Text)
	formList.SetTable("expense_settlements")

	return settlementTable
}