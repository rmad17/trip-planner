package tables

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

// GetContent returns the dashboard content
func GetContent(ctx *context.Context) (types.Panel, error) {
	col1 := `
		<div class="box box-primary">
			<div class="box-header with-border">
				<h3 class="box-title">Welcome to Trip Planner Admin</h3>
			</div>
			<div class="box-body">
				<p>Manage your trip planning system from this admin panel.</p>
				<ul>
					<li><strong>Users:</strong> Manage system users and their permissions</li>
					<li><strong>Trip Plans:</strong> View and manage all trip plans</li>
					<li><strong>Travellers:</strong> Manage trip participants</li>
					<li><strong>Expenses:</strong> Track and manage trip expenses</li>
					<li><strong>Documents:</strong> Handle uploaded documents</li>
				</ul>
			</div>
		</div>
	`

	return types.Panel{
		Content: types.HTML(col1),
		Title:   "Trip Planner Dashboard",
		Description: "Administrative dashboard for the trip planning system",
	}, nil
}

// Generators contains all table generators for the admin panel
var Generators = map[string]table.Generator{
	"users":              GetUserTable,
	"trip_plans":         GetTripPlanTable,
	"trip_hops":          GetTripHopTable,
	"trip_days":          GetTripDayTable,
	"activities":         GetActivityTable,
	"travellers":         GetTravellerTable,
	"stays":              GetStayTable,
	"documents":          GetDocumentTable,
	"document_shares":    GetDocumentShareTable,
	"expenses":           GetExpenseTable,
	"expense_splits":     GetExpenseSplitTable,
	"expense_settlements": GetExpenseSettlementTable,
}