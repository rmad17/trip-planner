package notifications

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTemplateTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&NotificationTemplate{})
	require.NoError(t, err)

	return db
}

func TestTemplateService_RenderText(t *testing.T) {
	db := setupTemplateTestDB(t)
	service := NewTemplateService(db)

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "simple text",
			template: "Hello World",
			data:     map[string]interface{}{},
			expected: "Hello World",
			wantErr:  false,
		},
		{
			name:     "simple variable substitution",
			template: "Hello {{.Name}}",
			data:     map[string]interface{}{"Name": "John"},
			expected: "Hello John",
			wantErr:  false,
		},
		{
			name:     "multiple variables",
			template: "Hello {{.FirstName}} {{.LastName}}",
			data:     map[string]interface{}{"FirstName": "John", "LastName": "Doe"},
			expected: "Hello John Doe",
			wantErr:  false,
		},
		{
			name:     "with conditional",
			template: "Hello{{if .Name}} {{.Name}}{{end}}",
			data:     map[string]interface{}{"Name": "John"},
			expected: "Hello John",
			wantErr:  false,
		},
		{
			name:     "with range",
			template: "Items: {{range .Items}}{{.}}, {{end}}",
			data:     map[string]interface{}{"Items": []string{"A", "B", "C"}},
			expected: "Items: A, B, C, ",
			wantErr:  false,
		},
		{
			name:     "invalid template syntax",
			template: "Hello {{.Name",
			data:     map[string]interface{}{},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.renderText(tt.template, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestTemplateService_RenderHTML(t *testing.T) {
	db := setupTemplateTestDB(t)
	service := NewTemplateService(db)

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "simple HTML",
			template: "<h1>Hello World</h1>",
			data:     map[string]interface{}{},
			expected: "<h1>Hello World</h1>",
			wantErr:  false,
		},
		{
			name:     "HTML with variable",
			template: "<h1>Hello {{.Name}}</h1>",
			data:     map[string]interface{}{"Name": "John"},
			expected: "<h1>Hello John</h1>",
			wantErr:  false,
		},
		{
			name:     "HTML with nested tags",
			template: "<div><p>Hello {{.Name}}</p></div>",
			data:     map[string]interface{}{"Name": "John"},
			expected: "<div><p>Hello John</p></div>",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.renderHTML(tt.template, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestTemplateService_ValidateTemplate(t *testing.T) {
	db := setupTemplateTestDB(t)
	service := NewTemplateService(db)

	tests := []struct {
		name     string
		template *NotificationTemplate
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid template",
			template: &NotificationTemplate{
				Name:    "test",
				Content: "Hello {{.Name}}",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			template: &NotificationTemplate{
				Content: "Hello",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing content",
			template: &NotificationTemplate{
				Name: "test",
			},
			wantErr: true,
			errMsg:  "content is required",
		},
		{
			name: "invalid content syntax",
			template: &NotificationTemplate{
				Name:    "test",
				Content: "Hello {{.Name",
			},
			wantErr: true,
			errMsg:  "invalid content template syntax",
		},
		{
			name: "invalid HTML syntax",
			template: &NotificationTemplate{
				Name:        "test",
				Content:     "Hello {{.Name}}",
				ContentHTML: "<div>{{.Name",
			},
			wantErr: true,
			errMsg:  "invalid HTML template syntax",
		},
		{
			name: "valid with HTML",
			template: &NotificationTemplate{
				Name:        "test",
				Content:     "Hello {{.Name}}",
				ContentHTML: "<div>{{.Name}}</div>",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateTemplate(tt.template)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTemplateService_Render(t *testing.T) {
	db := setupTemplateTestDB(t)
	service := NewTemplateService(db)
	ctx := context.Background()

	template := &NotificationTemplate{
		Name:        "test",
		Content:     "Hello {{.Name}}, welcome to {{.AppName}}!",
		ContentHTML: "<h1>Hello {{.Name}}</h1><p>Welcome to {{.AppName}}!</p>",
	}

	data := map[string]interface{}{
		"Name":    "John Doe",
		"AppName": "Trip Planner",
	}

	content, contentHTML, err := service.Render(ctx, template, data)
	require.NoError(t, err)

	assert.Equal(t, "Hello John Doe, welcome to Trip Planner!", content)
	assert.Equal(t, "<h1>Hello John Doe</h1><p>Welcome to Trip Planner!</p>", contentHTML)
}

func TestTemplateService_RenderComplexData(t *testing.T) {
	db := setupTemplateTestDB(t)
	service := NewTemplateService(db)
	ctx := context.Background()

	template := &NotificationTemplate{
		Name: "trip_invitation",
		Content: `Hello {{.RecipientName}},

{{.InviterName}} has invited you to join their trip to {{.Destination}}.

Trip Details:
- Dates: {{.StartDate}} to {{.EndDate}}
- Duration: {{.Duration}} days
- Travelers: {{.TravelerCount}}

{{if .Message}}Personal message: {{.Message}}{{end}}`,
	}

	data := map[string]interface{}{
		"RecipientName": "Alice",
		"InviterName":   "Bob",
		"Destination":   "Paris, France",
		"StartDate":     "June 1, 2024",
		"EndDate":       "June 10, 2024",
		"Duration":      9,
		"TravelerCount": 4,
		"Message":       "Can't wait to explore Paris with you!",
	}

	content, _, err := service.Render(ctx, template, data)
	require.NoError(t, err)
	assert.Contains(t, content, "Alice")
	assert.Contains(t, content, "Bob")
	assert.Contains(t, content, "Paris, France")
	assert.Contains(t, content, "Can't wait to explore Paris with you!")
}

func TestTemplateService_RenderWithMissingData(t *testing.T) {
	db := setupTemplateTestDB(t)
	service := NewTemplateService(db)
	ctx := context.Background()

	template := &NotificationTemplate{
		Name:    "test",
		Content: "Hello {{.Name}}, you have {{.Count}} items.",
	}

	// Missing 'Count' field
	data := map[string]interface{}{
		"Name": "John",
	}

	// Should not error, but missing fields will be empty
	content, _, err := service.Render(ctx, template, data)
	require.NoError(t, err)
	assert.Contains(t, content, "John")
	// Count will be rendered as "<no value>"
	assert.Contains(t, content, "you have")
}

func TestTemplateService_ListWithFilters(t *testing.T) {
	db := setupTemplateTestDB(t)
	service := NewTemplateService(db)
	ctx := context.Background()

	// Create test templates
	templates := []*NotificationTemplate{
		{
			Name:     "email_welcome",
			Type:     TypeTransactional,
			Channel:  ChannelEmail,
			Content:  "Welcome!",
			IsActive: true,
		},
		{
			Name:     "sms_reminder",
			Type:     TypeReminder,
			Channel:  ChannelSMS,
			Content:  "Reminder",
			IsActive: true,
		},
		{
			Name:     "email_marketing",
			Type:     TypeMarketing,
			Channel:  ChannelEmail,
			Content:  "Special offer",
			IsActive: false,
		},
	}

	for _, tpl := range templates {
		err := service.Create(ctx, tpl)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		filters       TemplateFilters
		expectedCount int
		checkFunc     func(t *testing.T, results []*NotificationTemplate)
	}{
		{
			name:          "no filters",
			filters:       TemplateFilters{},
			expectedCount: 3,
		},
		{
			name: "filter by channel",
			filters: TemplateFilters{
				Channel: func() *NotificationChannel { c := ChannelEmail; return &c }(),
			},
			expectedCount: 2,
			checkFunc: func(t *testing.T, results []*NotificationTemplate) {
				for _, r := range results {
					assert.Equal(t, ChannelEmail, r.Channel)
				}
			},
		},
		{
			name: "filter by type",
			filters: TemplateFilters{
				Type: func() *NotificationType { t := TypeTransactional; return &t }(),
			},
			expectedCount: 1,
		},
		{
			name: "filter by active status",
			filters: TemplateFilters{
				IsActive: func() *bool { b := true; return &b }(),
			},
			expectedCount: 2,
		},
		{
			name: "with limit",
			filters: TemplateFilters{
				Limit: 2,
			},
			expectedCount: 2,
		},
		{
			name: "combined filters",
			filters: TemplateFilters{
				Channel:  func() *NotificationChannel { c := ChannelEmail; return &c }(),
				IsActive: func() *bool { b := true; return &b }(),
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := service.List(ctx, tt.filters)
			require.NoError(t, err)
			assert.Len(t, results, tt.expectedCount)

			if tt.checkFunc != nil {
				tt.checkFunc(t, results)
			}
		})
	}
}

func TestTemplateService_Update(t *testing.T) {
	db := setupTemplateTestDB(t)
	service := NewTemplateService(db)
	ctx := context.Background()

	// Create template
	template := &NotificationTemplate{
		Name:    "test",
		Content: "Original content",
		Version: 1,
	}

	err := service.Create(ctx, template)
	require.NoError(t, err)

	originalID := template.ID

	// Update template
	template.Content = "Updated content"
	err = service.Update(ctx, template)
	require.NoError(t, err)

	// Verify version incremented
	assert.Equal(t, 2, template.Version)
	assert.Equal(t, originalID, template.ID) // ID should not change

	// Retrieve and verify
	retrieved, err := service.Get(ctx, template.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated content", retrieved.Content)
	assert.Equal(t, 2, retrieved.Version)
}

func TestTemplateService_Delete(t *testing.T) {
	db := setupTemplateTestDB(t)

	// Also need Notification table for foreign key check
	err := db.AutoMigrate(&Notification{})
	require.NoError(t, err)

	service := NewTemplateService(db)
	ctx := context.Background()

	// Create template
	template := &NotificationTemplate{
		Name:    "test",
		Content: "Test content",
	}

	err = service.Create(ctx, template)
	require.NoError(t, err)

	templateID := template.ID

	// Delete template
	err = service.Delete(ctx, templateID)
	require.NoError(t, err)

	// Verify deletion
	_, err = service.Get(ctx, templateID)
	assert.Error(t, err)
}

func TestTemplateService_DeleteInUse(t *testing.T) {
	db := setupTemplateTestDB(t)

	// Need Notification table
	err := db.AutoMigrate(&Notification{})
	require.NoError(t, err)

	service := NewTemplateService(db)
	ctx := context.Background()

	// Create template
	template := &NotificationTemplate{
		Name:    "test",
		Content: "Test content",
	}

	err = service.Create(ctx, template)
	require.NoError(t, err)

	// Create notification using this template
	notification := &Notification{
		TemplateID: &template.ID,
		Content:    "Test",
		Channel:    ChannelEmail,
	}
	err = db.Create(notification).Error
	require.NoError(t, err)

	// Try to delete template - should fail
	err = service.Delete(ctx, template.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "in use")
}
