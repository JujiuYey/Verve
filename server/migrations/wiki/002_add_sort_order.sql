-- ============================================
-- Wiki 文件夹与文档排序字段
-- ============================================

ALTER TABLE wiki_folders
ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0;

ALTER TABLE wiki_documents
ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_folders_parent_sort
ON wiki_folders(parent_id, sort_order);

CREATE INDEX IF NOT EXISTS idx_documents_folder_sort
ON wiki_documents(folder_id, sort_order);

COMMENT ON COLUMN wiki_folders.sort_order IS '同级文件夹排序权重，升序展示';
COMMENT ON COLUMN wiki_documents.sort_order IS '同一文件夹内文档排序权重，升序展示';
