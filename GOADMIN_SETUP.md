# GoAdmin Integration Setup

This document outlines the successful integration of GoAdmin with the Trip Planner system.

## What Was Accomplished

### 1. ✅ **GoAdmin Dependencies Added**
- Added `github.com/GoAdminGroup/go-admin v1.2.26`
- Added `github.com/GoAdminGroup/themes v0.0.48`
- All required dependencies automatically installed via `go mod tidy`

### 2. ✅ **Atlas Migrations Generated**
- Updated `cmd/atlas-loader/main.go` to include all models from all modules:
  - Users (accounts)
  - Trip Plans, Hops, Days, Activities, Travellers, Stays (trips)
  - Documents, Document Shares (documents)  
  - Expenses, Expense Splits, Expense Settlements (expenses)
- Generated comprehensive migration with `atlas migrate diff add_new_models --env local`

### 3. ✅ **GoAdmin Integration**
- Created `admin/simple_setup.go` with GoAdmin configuration
- Integrated GoAdmin into main application (`app.go`)
- PostgreSQL database connection configured
- AdminLTE theme with custom branding

### 4. ✅ **Admin User Seed Data**
- Created `cmd/seed-simple/main.go` for database seeding
- Automatically creates all required GoAdmin tables:
  - `goadmin_users`
  - `goadmin_roles` 
  - `goadmin_permissions`
  - `goadmin_role_users`
  - `goadmin_role_permissions`
  - `goadmin_user_permissions`
  - `goadmin_menu`
  - `goadmin_role_menu`
  - `goadmin_operation_log`
  - `goadmin_session`

### 5. ✅ **Default Admin User Created**
- **Username**: `admin`
- **Password**: `admin`
- **Role**: Administrator (with full permissions)
- **Access URL**: `http://localhost:8080/admin`

### 6. ✅ **Admin Menu Structure**
- **Admin Section**:
  - Users Management
  - Roles Management  
  - Permissions Management
  - Menu Management
  - Operation Logs
- **Trip Management Section**:
  - Trip Plans
  - Users
  - Travellers
  - Expenses
  - Documents

## How to Use

### 1. **Setup Database**
Run the seed command to create GoAdmin tables and admin user:
```bash
go run ./cmd/seed-simple/main.go
```

### 2. **Start Application**
```bash
go run app.go
```

### 3. **Access Admin Panel**
- Open: `http://localhost:8080/admin`
- Login with: `admin` / `admin`
- **⚠️ Change password after first login!**

### 4. **Available Admin URLs**
- **Dashboard**: `http://localhost:8080/admin`
- **User Management**: `http://localhost:8080/admin/info/goadmin_users`
- **Trip Plans**: `http://localhost:8080/admin/info/trip_plans`
- **System Users**: `http://localhost:8080/admin/info/users` 
- **Travellers**: `http://localhost:8080/admin/info/travellers`
- **Expenses**: `http://localhost:8080/admin/info/expenses`
- **Documents**: `http://localhost:8080/admin/info/documents`

## Models Included in Atlas Migration

### Core Models
- **Users** (Authentication & Profile)
- **Trip Plans** (Main trip entities)
- **Trip Hops** (Destinations within trips)
- **Trip Days** (Daily itineraries)
- **Activities** (Scheduled events/tasks)
- **Travellers** (Trip participants)
- **Stays** (Accommodation details)

### Document Management
- **Documents** (File uploads with categorization)
- **Document Shares** (Sharing permissions)

### Expense Management  
- **Expenses** (Trip expenses with flexible linking)
- **Expense Splits** (Cost sharing among travellers)
- **Expense Settlements** (Payment tracking between users)

## Key Features

### 1. **Comprehensive Models**
- All trip planner models included
- Proper relationships and foreign keys
- Type-safe enums for categories, currencies, etc.

### 2. **Flexible Expense System**
- Multi-currency support (INR, USD, GBP, EUR, etc.)
- Multiple split methods (equal, exact, percentage, shares)
- Settlement tracking with netting
- Expense linking to trips, hops, days, or activities

### 3. **Document Management**
- Multiple storage providers (DigitalOcean, S3, GCS, etc.)
- Categorized document types
- Sharing permissions between users
- Entity linking (documents can attach to any model)

### 4. **Traveller Management**
- Guest and registered user support
- Passport and medical information
- Emergency contacts and dietary restrictions
- Role-based permissions (organizer, participant, etc.)

## Security Notes

- Default admin password is `admin` - **CHANGE IMMEDIATELY**
- All passwords are MD5 hashed (GoAdmin default)
- Session management included
- Operation logging enabled
- Role-based access control configured

## File Structure

```
admin/
├── setup.go              # Full GoAdmin setup (not used)
├── simple_setup.go       # Simple GoAdmin integration ✅
├── seeds.go              # Complex seeding (not used)
├── tables/               # Table configurations (prepared for future)
│   ├── generators.go
│   ├── users.go
│   ├── trip_plans.go
│   ├── travellers.go  
│   ├── expenses.go
│   └── remaining_tables.go

cmd/
├── seed-simple/main.go   # Working seed command ✅
└── seed/main.go          # Complex seed (not used)
```

## Next Steps

1. **Customize Admin Interface**: Update table configurations in `admin/tables/`
2. **Add Data Validation**: Implement proper form validation
3. **Security Hardening**: Change default passwords, add 2FA
4. **Custom Dashboards**: Create trip-specific analytics
5. **API Integration**: Connect admin actions to main API endpoints

## Troubleshooting

### Database Connection Issues
- Ensure PostgreSQL is running
- Check `.env` file has correct `DB_URL`
- Run seed command before starting application

### Admin Login Issues  
- Verify admin user was created: `go run ./cmd/seed-simple/main.go`
- Check database tables exist
- Try default credentials: `admin`/`admin`

### Permission Errors
- Ensure admin role has all permissions
- Check `goadmin_role_users` table for role assignment
- Verify menu permissions in `goadmin_role_menu`

## Success! 🎉

The GoAdmin integration is now complete and functional. You can:
- ✅ Access admin panel at `http://localhost:8080/admin`
- ✅ Manage all trip planner models
- ✅ View and edit data through web interface  
- ✅ Track operations and manage users
- ✅ Utilize comprehensive expense and document management

All models are properly migrated with Atlas and accessible through the admin interface.