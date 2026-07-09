package service

import (
	"context"
	"fmt"
	"strings"

	wiki_db "verve/app/wiki/models/db"
)

type folderFinder interface {
	FindOne(ctx context.Context, id string) (*wiki_db.Folder, error)
}

type FolderScope struct {
	RootFolderID string
	FolderPath   string
}

type RootResolver struct {
	folders folderFinder
}

func NewRootResolver(folders folderFinder) *RootResolver {
	return &RootResolver{folders: folders}
}

func (r *RootResolver) Resolve(ctx context.Context, folderID string) (FolderScope, error) {
	if strings.TrimSpace(folderID) == "" {
		return FolderScope{}, fmt.Errorf("folder id is required")
	}
	seen := map[string]bool{}
	path := make([]string, 0, 4)
	currentID := folderID
	var rootID string

	for currentID != "" {
		if seen[currentID] {
			return FolderScope{}, fmt.Errorf("folder cycle detected at %s", currentID)
		}
		seen[currentID] = true
		folder, err := r.folders.FindOne(ctx, currentID)
		if err != nil {
			return FolderScope{}, err
		}
		path = append([]string{folder.Name}, path...)
		rootID = folder.ID
		if folder.ParentID == nil || strings.TrimSpace(*folder.ParentID) == "" {
			break
		}
		currentID = *folder.ParentID
	}

	return FolderScope{
		RootFolderID: rootID,
		FolderPath:   strings.Join(path, "/"),
	}, nil
}
