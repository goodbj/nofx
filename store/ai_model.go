package store

import (
	"errors"
	"fmt"
	"nofx/crypto"
	"nofx/logger"
	"strings"
	"time"

	"gorm.io/gorm"
)

// AIModelStore AI model storage
type AIModelStore struct {
	db *gorm.DB
}

// AIModel AI model configuration
type AIModel struct {
	ID                   string                 `gorm:"primaryKey" json:"id"`
	UserID               string                 `gorm:"column:user_id;not null;default:default;index" json:"user_id"`
	Name                 string                 `gorm:"not null" json:"name"`
	Provider             string                 `gorm:"not null" json:"provider"`
	Enabled              bool                   `gorm:"default:false" json:"enabled"`
	APIKey               crypto.EncryptedString `gorm:"column:api_key;default:''" json:"apiKey"`
	CustomAPIURL         string                 `gorm:"column:custom_api_url;default:''" json:"customApiUrl"`
	CustomModelName      string                 `gorm:"column:custom_model_name;default:''" json:"customModelName"`
	UseBrowserAutomation bool                   `gorm:"column:use_browser_automation;default:false" json:"useBrowserAutomation"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

func (AIModel) TableName() string { return "ai_models" }

// NewAIModelStore creates a new AIModelStore
func NewAIModelStore(db *gorm.DB) *AIModelStore {
	return &AIModelStore{db: db}
}

func (s *AIModelStore) initTables() error {
	// For PostgreSQL with existing table, we need to ensure the new column is added
	if s.db.Dialector.Name() == "postgres" {
		// Check if table exists
		var tableExists int64
		s.db.Raw(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'ai_models'`).Scan(&tableExists)
		if tableExists > 0 {
			// Check if the use_browser_automation column exists
			var columnExists int64
			s.db.Raw(`SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'ai_models' AND column_name = 'use_browser_automation'`).Scan(&columnExists)
			if columnExists == 0 {
				// Add the new column if it doesn't exist
				if err := s.db.Exec("ALTER TABLE ai_models ADD COLUMN use_browser_automation BOOLEAN DEFAULT FALSE").Error; err != nil {
					logger.Errorf("Failed to add use_browser_automation column: %v", err)
					return err
				}
			}
			return nil
		}
	}
	return s.db.AutoMigrate(&AIModel{})
}

func (s *AIModelStore) initDefaultData() error {
	// No longer pre-populate AI models - create on demand when user configures
	return nil
}

// List retrieves user's AI model list
func (s *AIModelStore) List(userID string) ([]*AIModel, error) {
	var models []*AIModel
	err := s.db.Where("user_id = ?", userID).Order("id").Find(&models).Error
	if err != nil {
		return nil, err
	}
	return models, nil
}

// Get retrieves a single AI model
func (s *AIModelStore) Get(userID, modelID string) (*AIModel, error) {
	if modelID == "" {
		return nil, fmt.Errorf("model ID cannot be empty")
	}

	candidates := []string{}
	if userID != "" {
		candidates = append(candidates, userID)
	}
	if userID != "default" {
		candidates = append(candidates, "default")
	}
	if len(candidates) == 0 {
		candidates = append(candidates, "default")
	}

	for _, uid := range candidates {
		var model AIModel
		err := s.db.Where("user_id = ? AND id = ?", uid, modelID).First(&model).Error
		if err == nil {
			return &model, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// GetByID retrieves an AI model by ID only (for debate engine)
func (s *AIModelStore) GetByID(modelID string) (*AIModel, error) {
	if modelID == "" {
		return nil, fmt.Errorf("model ID cannot be empty")
	}

	var model AIModel
	err := s.db.Where("id = ?", modelID).First(&model).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// GetDefault retrieves the default enabled AI model
func (s *AIModelStore) GetDefault(userID string) (*AIModel, error) {
	if userID == "" {
		userID = "default"
	}
	model, err := s.firstEnabled(userID)
	if err == nil {
		return model, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if userID != "default" {
		return s.firstEnabled("default")
	}
	return nil, fmt.Errorf("please configure an available AI model in the system first")
}

func (s *AIModelStore) firstEnabled(userID string) (*AIModel, error) {
	var model AIModel
	err := s.db.Where("user_id = ? AND enabled = ?", userID, true).
		Order("updated_at DESC, id ASC").
		First(&model).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// Update updates AI model, creates if not exists
// IMPORTANT: If apiKey is empty string, the existing API key will be preserved (not overwritten)
func (s *AIModelStore) Update(userID, id string, enabled bool, apiKey, customAPIURL, customModelName string, useBrowserAutomation bool) error {
	// Try exact ID match first
	var existingModel AIModel
	err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&existingModel).Error
	if err == nil {
		// Update existing model
		updates := map[string]interface{}{
			"enabled":                enabled,
			"custom_api_url":         customAPIURL,
			"custom_model_name":      customModelName,
			"use_browser_automation": useBrowserAutomation,
			"updated_at":             time.Now().UTC(),
		}
		// If apiKey is not empty, update it (encryption handled by crypto.EncryptedString)
		if apiKey != "" {
			updates["api_key"] = crypto.EncryptedString(apiKey)
		}
		return s.db.Model(&existingModel).Updates(updates).Error
	}

	// Create new record
	provider := id
	if id != "" {
		parts := strings.Split(id, "_")
		if len(parts) >= 2 {
			provider = parts[len(parts)-1]
		} else {
			provider = id
		}
	}

	// Generate name based on provider
	var name string
	if provider == "deepseek" {
		name = "DeepSeek AI"
	} else if provider == "qwen" {
		name = "Qwen AI"
	} else {
		name = provider + " AI"
	}

	// Find the next available number for this name
	counter := 1
	for {
		checkName := name
		if counter > 1 {
			checkName = fmt.Sprintf("%s %d", name, counter)
		}
		var existingCount int64
		s.db.Model(&AIModel{}).Where("user_id = ? AND name = ?", userID, checkName).Count(&existingCount)
		if existingCount == 0 {
			if counter > 1 {
				name = fmt.Sprintf("%s %d", name, counter)
			}
			break
		}
		counter++
	}

	logger.Infof("✓ Creating new AI model configuration: ID=%s, Provider=%s, Name=%s", id, provider, name)
	newModel := &AIModel{
		ID:                   id,
		UserID:               userID,
		Name:                 name,
		Provider:             provider,
		Enabled:              enabled,
		APIKey:               crypto.EncryptedString(apiKey),
		CustomAPIURL:         customAPIURL,
		CustomModelName:      customModelName,
		UseBrowserAutomation: useBrowserAutomation,
	}
	return s.db.Create(newModel).Error
}

// Create creates an AI model
func (s *AIModelStore) Create(userID, id, name, provider string, enabled bool, apiKey, customAPIURL string) error {
	model := &AIModel{
		ID:           id,
		UserID:       userID,
		Name:         name,
		Provider:     provider,
		Enabled:      enabled,
		APIKey:       crypto.EncryptedString(apiKey),
		CustomAPIURL: customAPIURL,
	}
	// Use FirstOrCreate to ignore if already exists
	return s.db.Where("id = ?", id).FirstOrCreate(model).Error
}

// Delete removes an AI model by ID
func (s *AIModelStore) Delete(userID, modelID string) error {
	// Find the model to ensure it belongs to the user
	var model AIModel
	err := s.db.Where("user_id = ? AND id = ?", userID, modelID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("model with ID %s not found for user %s", modelID, userID)
		}
		return err
	}

	// Delete the model
	return s.db.Delete(&model).Error
}
