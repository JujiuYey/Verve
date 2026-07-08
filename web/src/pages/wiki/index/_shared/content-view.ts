import type { Document } from "@/api/wiki/document";
import type { Folder } from "@/api/wiki/folder";

interface FolderContentViewInput {
  folders: Folder[];
  documents: Document[];
  searchKeyword: string;
}

interface FolderContentView {
  folders: Folder[];
  documents: Document[];
}

function normalizeKeyword(value: string) {
  return value.trim().toLowerCase();
}

function matchesFolder(folder: Folder, keyword: string) {
  return (
    folder.name.toLowerCase().includes(keyword) ||
    (folder.description || "").toLowerCase().includes(keyword)
  );
}

function matchesDocument(document: Document, keyword: string) {
  return document.filename.toLowerCase().includes(keyword);
}

export function getFolderContentView({
  folders,
  documents,
  searchKeyword,
}: FolderContentViewInput): FolderContentView {
  const keyword = normalizeKeyword(searchKeyword);

  const filteredFolders = keyword
    ? folders.filter((folder) => matchesFolder(folder, keyword))
    : folders;

  const filteredDocuments = keyword
    ? documents.filter((document) => matchesDocument(document, keyword))
    : documents;

  return {
    folders: filteredFolders,
    documents: filteredDocuments,
  };
}
