package payload

type EnsureAgentInstanceRequest struct {
	RootFolderID string  `json:"root_folder_id"`
	Name         *string `json:"name"`
	Description  *string `json:"description"`
}
