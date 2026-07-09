package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	wiki_db "verve/app/wiki/models/db"
)

const (
	AgentKeyWikiLearning = "wiki_learning"
	AgentStatusActive    = "active"
)

type AgentInstanceRepository struct {
	db *bun.DB
}

func NewAgentInstanceRepository(db *bun.DB) *AgentInstanceRepository {
	return &AgentInstanceRepository{db: db}
}

func (r *AgentInstanceRepository) FindByRoot(ctx context.Context, userID, rootFolderID string) (*wiki_db.AgentInstance, error) {
	instance := new(wiki_db.AgentInstance)
	err := r.db.NewSelect().
		Model(instance).
		Relation("RootFolder").
		Where("wai.user_id = ?", strings.TrimSpace(userID)).
		Where("wai.root_folder_id = ?", strings.TrimSpace(rootFolderID)).
		Scan(ctx)
	return instance, err
}

func (r *AgentInstanceRepository) EnsureByRoot(ctx context.Context, userID string, rootFolder *wiki_db.Folder, name *string, description *string) (*wiki_db.AgentInstance, error) {
	displayName := defaultAgentName(rootFolder.Name, name)
	instance := &wiki_db.AgentInstance{
		ID:           compactAgentInstanceID(),
		UserID:       strings.TrimSpace(userID),
		RootFolderID: rootFolder.ID,
		AgentKey:     AgentKeyWikiLearning,
		Name:         displayName,
		Description:  normalizeOptionalString(description),
		Status:       AgentStatusActive,
	}
	_, err := r.db.NewInsert().
		Model(instance).
		On("CONFLICT (user_id, root_folder_id) DO UPDATE").
		Set("status = ?", AgentStatusActive).
		Set("updated_at = ?", time.Now()).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return r.FindByRoot(ctx, userID, rootFolder.ID)
}

func defaultAgentName(rootFolderName string, custom *string) string {
	if custom != nil && strings.TrimSpace(*custom) != "" {
		return strings.TrimSpace(*custom)
	}
	name := strings.TrimSpace(rootFolderName)
	if name == "" {
		return "Wiki 学习 Agent"
	}
	return name + " 学习 Agent"
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func compactAgentInstanceID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
