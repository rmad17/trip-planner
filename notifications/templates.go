package notifications

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	textTemplate "text/template"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TemplateServiceImpl implements the TemplateService interface
type TemplateServiceImpl struct {
	db *gorm.DB
}

// NewTemplateService creates a new template service
func NewTemplateService(db *gorm.DB) *TemplateServiceImpl {
	return &TemplateServiceImpl{db: db}
}

// Create creates a new template
func (ts *TemplateServiceImpl) Create(ctx context.Context, tpl *NotificationTemplate) error {
	// Validate template
	if err := ts.validateTemplate(tpl); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	// Check for duplicate name
	var count int64
	if err := ts.db.WithContext(ctx).Model(&NotificationTemplate{}).
		Where("name = ?", tpl.Name).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check template name: %w", err)
	}

	if count > 0 {
		return errors.New("template with this name already exists")
	}

	// Create template
	if err := ts.db.WithContext(ctx).Create(tpl).Error; err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	return nil
}

// Update updates an existing template
func (ts *TemplateServiceImpl) Update(ctx context.Context, tpl *NotificationTemplate) error {
	// Validate template
	if err := ts.validateTemplate(tpl); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	// Increment version
	tpl.Version++

	// Update template
	if err := ts.db.WithContext(ctx).Save(tpl).Error; err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	return nil
}

// Get gets a template by ID
func (ts *TemplateServiceImpl) Get(ctx context.Context, templateID uuid.UUID) (*NotificationTemplate, error) {
	var tpl NotificationTemplate
	if err := ts.db.WithContext(ctx).First(&tpl, "id = ?", templateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return &tpl, nil
}

// GetByName gets a template by name
func (ts *TemplateServiceImpl) GetByName(ctx context.Context, name string) (*NotificationTemplate, error) {
	var tpl NotificationTemplate
	if err := ts.db.WithContext(ctx).Where("name = ?", name).First(&tpl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return &tpl, nil
}

// List lists all templates with optional filters
func (ts *TemplateServiceImpl) List(ctx context.Context, filters TemplateFilters) ([]*NotificationTemplate, error) {
	query := ts.db.WithContext(ctx).Model(&NotificationTemplate{})

	if filters.Channel != nil {
		query = query.Where("channel = ?", *filters.Channel)
	}

	if filters.Type != nil {
		query = query.Where("type = ?", *filters.Type)
	}

	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}

	if filters.Locale != "" {
		query = query.Where("locale = ?", filters.Locale)
	}

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var templates []*NotificationTemplate
	if err := query.Order("created_at DESC").Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	return templates, nil
}

// Delete deletes a template
func (ts *TemplateServiceImpl) Delete(ctx context.Context, templateID uuid.UUID) error {
	// Check if template is in use
	var count int64
	if err := ts.db.WithContext(ctx).Model(&Notification{}).
		Where("template_id = ?", templateID).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check template usage: %w", err)
	}

	if count > 0 {
		return errors.New("cannot delete template: it is being used by notifications")
	}

	// Delete template
	if err := ts.db.WithContext(ctx).Delete(&NotificationTemplate{}, "id = ?", templateID).Error; err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}

// Render renders a template with data
func (ts *TemplateServiceImpl) Render(ctx context.Context, tpl *NotificationTemplate, data map[string]interface{}) (string, string, error) {
	var content, contentHTML string
	var err error

	// Render plain text content
	if tpl.Content != "" {
		content, err = ts.renderText(tpl.Content, data)
		if err != nil {
			return "", "", fmt.Errorf("failed to render content: %w", err)
		}
	}

	// Render HTML content
	if tpl.ContentHTML != "" {
		contentHTML, err = ts.renderHTML(tpl.ContentHTML, data)
		if err != nil {
			return "", "", fmt.Errorf("failed to render HTML content: %w", err)
		}
	}

	return content, contentHTML, nil
}

// Helper methods

func (ts *TemplateServiceImpl) validateTemplate(tpl *NotificationTemplate) error {
	if tpl.Name == "" {
		return errors.New("template name is required")
	}

	if tpl.Content == "" {
		return errors.New("template content is required")
	}

	// Validate template syntax
	if _, err := textTemplate.New("test").Parse(tpl.Content); err != nil {
		return fmt.Errorf("invalid content template syntax: %w", err)
	}

	if tpl.ContentHTML != "" {
		if _, err := template.New("test").Parse(tpl.ContentHTML); err != nil {
			return fmt.Errorf("invalid HTML template syntax: %w", err)
		}
	}

	return nil
}

func (ts *TemplateServiceImpl) renderText(templateStr string, data map[string]interface{}) (string, error) {
	tmpl, err := textTemplate.New("content").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func (ts *TemplateServiceImpl) renderHTML(templateStr string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("html").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return buf.String(), nil
}
