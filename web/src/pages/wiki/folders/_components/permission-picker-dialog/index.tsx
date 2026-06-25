import { SearchIcon } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

import { departmentApi } from "@/api/system/department";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";

import { BreadcrumbNav } from "./_components/breadcrumb-nav";
import { DepartmentRow } from "./_components/department-row";
import { SelectedItemRow } from "./_components/selected-item-row";
import { UserRow } from "./_components/user-row";
import type { Department, SelectedItem, User } from "./_shared/types";

interface ContactPickerDialogProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  onConfirm?: (selected: SelectedItem[]) => void;
  defaultSelected?: SelectedItem[];
}

export function PermissionPickerDialog({
  open,
  onOpenChange,
  onConfirm,
  defaultSelected = [],
}: ContactPickerDialogProps) {
  const [selected, setSelected] = useState<SelectedItem[]>(defaultSelected);
  const [breadcrumb, setBreadcrumb] = useState<{ id: string; name: string }[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [departmentTree, setDepartmentTree] = useState<Department[]>([]);
  const [loadedUsers, setLoadedUsers] = useState<Record<string, User[]>>({});
  const [loading, setLoading] = useState(false);
  const [searchResults, setSearchResults] = useState<{
    departments: Department[];
    users: User[];
  } | null>(null);
  const [searching, setSearching] = useState(false);

  // Load department tree
  useEffect(() => {
    if (open) {
      departmentApi
        .tree()
        .then((data) => {
          setDepartmentTree(data);
          setLoading(false);
        })
        .catch((error) => {
          console.error("Failed to load department tree:", error);
          toast.error("加载部门树失败");
          setLoading(false);
        });
    }
  }, [open]);

  const handleOpenChange = useCallback(
    (value: boolean) => {
      if (value) {
        setSelected(defaultSelected);
        setBreadcrumb([]);
        setSearchQuery("");
        setLoadedUsers({});
        setLoading(true);
      }
      onOpenChange?.(value);
    },
    [defaultSelected, onOpenChange],
  );

  const currentItems = useMemo(() => {
    let items = departmentTree;
    for (const crumb of breadcrumb) {
      const found = items.find((d) => d.id === crumb.id);
      if (found?.children) {
        items = found.children;
      } else {
        items = [];
        break;
      }
    }
    return items;
  }, [breadcrumb, departmentTree]);

  const currentUsers = useMemo(() => {
    if (breadcrumb.length === 0) return [];
    const lastCrumb = breadcrumb[breadcrumb.length - 1];
    return loadedUsers[lastCrumb.id] || [];
  }, [breadcrumb, loadedUsers]);

  // Search functionality
  useEffect(() => {
    if (!searchQuery.trim()) {
      // Clear search results when query is empty
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setSearchResults(null);
      return;
    }

    const timer = setTimeout(() => {
      setSearching(true);
      departmentApi
        .search(searchQuery.trim())
        .then((data) => {
          setSearchResults({
            departments: data.departments,
            users: data.users,
          });
          setSearching(false);
        })
        .catch((error) => {
          console.error("Search failed:", error);
          toast.error("搜索失败");
          setSearching(false);
        });
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery]);

  const isSelected = useCallback(
    (id: string) => {
      return selected.some((item) => item.id === id);
    },
    [selected],
  );

  const toggleSelect = useCallback((item: SelectedItem) => {
    setSelected((prev) => {
      if (prev.some((s) => s.id === item.id)) {
        return prev.filter((s) => s.id !== item.id);
      }
      return [...prev, item];
    });
  }, []);

  const removeSelected = useCallback((id: string) => {
    setSelected((prev) => prev.filter((s) => s.id !== id));
  }, []);

  const navigateToChild = useCallback((dept: Department) => {
    setBreadcrumb((prev) => [...prev, { id: dept.id, name: dept.name }]);

    // Load users for this department if not already loaded
    setLoadedUsers((prevLoaded) => {
      if (!prevLoaded[dept.id]) {
        departmentApi
          .findUsers(dept.id)
          .then((users) => {
            setLoadedUsers((current) => ({
              ...current,
              [dept.id]: users,
            }));
          })
          .catch((error) => {
            console.error("Failed to load users:", error);
            toast.error("加载用户失败");
          });
      }
      return prevLoaded;
    });
  }, []);

  const navigateToBreadcrumb = useCallback((index: number) => {
    if (index < 0) {
      setBreadcrumb([]);
    } else {
      setBreadcrumb((prev) => prev.slice(0, index + 1));
    }
  }, []);

  const hasChildren = useCallback((dept: Department): boolean => {
    return dept.children !== undefined && dept.children.length > 0;
  }, []);

  const getUserDisplayName = (user: User) => {
    return user.full_name || user.username;
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-3xl p-0 gap-0">
        <DialogHeader className="p-6 pb-0">
          <DialogTitle>设置权限</DialogTitle>
        </DialogHeader>

        <div className="flex h-105 border-t border-b mt-4">
          {/* Left Panel */}
          <div className="flex-1 flex flex-col border-r">
            <div className="p-3">
              <div className="relative">
                <SearchIcon className="absolute left-2.5 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
                <Input
                  placeholder="搜索用户、部门..."
                  className="pl-8 h-8 text-sm"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                />
              </div>
            </div>

            {!searchResults && (
              <BreadcrumbNav breadcrumb={breadcrumb} onNavigate={navigateToBreadcrumb} />
            )}

            <div className="flex-1 overflow-y-auto px-1">
              {loading || searching ? (
                <div className="py-8 text-center text-sm text-muted-foreground">加载中...</div>
              ) : searchResults ? (
                <div>
                  {searchResults.departments.map((dept) => (
                    <DepartmentRow
                      key={dept.id}
                      dept={dept}
                      checked={isSelected(dept.id)}
                      onCheck={() =>
                        toggleSelect({ id: dept.id, name: dept.name, type: "department" })
                      }
                      onNavigate={() => {
                        setSearchQuery("");
                      }}
                      hasChildren={hasChildren(dept)}
                    />
                  ))}
                  {searchResults.users.map((user) => (
                    <UserRow
                      key={user.id}
                      user={{ id: user.id, name: getUserDisplayName(user), avatar: user.avatar }}
                      checked={isSelected(user.id)}
                      onCheck={() =>
                        toggleSelect({ id: user.id, name: getUserDisplayName(user), type: "user" })
                      }
                      subtitle={user.department_path}
                    />
                  ))}
                  {searchResults.departments.length === 0 && searchResults.users.length === 0 && (
                    <div className="py-8 text-center text-sm text-muted-foreground">
                      未找到匹配结果
                    </div>
                  )}
                </div>
              ) : (
                <div>
                  {currentItems.map((dept) => (
                    <DepartmentRow
                      key={dept.id}
                      dept={dept}
                      checked={isSelected(dept.id)}
                      onCheck={() =>
                        toggleSelect({ id: dept.id, name: dept.name, type: "department" })
                      }
                      onNavigate={() => navigateToChild(dept)}
                      hasChildren={hasChildren(dept)}
                    />
                  ))}
                  {currentUsers.map((user) => (
                    <UserRow
                      key={user.id}
                      user={{ id: user.id, name: getUserDisplayName(user), avatar: user.avatar }}
                      checked={isSelected(user.id)}
                      onCheck={() =>
                        toggleSelect({ id: user.id, name: getUserDisplayName(user), type: "user" })
                      }
                    />
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Right Panel */}
          <div className="w-84 flex flex-col">
            <div className="p-3 text-sm text-muted-foreground">
              已选：
              {selected.length} 个
            </div>
            <div className="flex-1 overflow-y-auto px-3">
              {selected.map((item) => (
                <SelectedItemRow key={item.id} item={item} onRemove={removeSelected} />
              ))}
            </div>
          </div>
        </div>

        <DialogFooter className="p-4">
          <Button variant="outline" onClick={() => handleOpenChange(false)}>
            取消
          </Button>
          <Button onClick={() => onConfirm?.(selected)}>确认</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
