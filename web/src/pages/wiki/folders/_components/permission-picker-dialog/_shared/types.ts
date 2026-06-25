export interface Department {
  id: string;
  name: string;
  children?: Department[];
  users?: User[];
}

export interface User {
  id: string;
  username: string;
  full_name?: string;
  avatar?: string;
  department_path?: string;
}

export interface SelectedItem {
  id: string;
  name: string;
  type: "department" | "user";
}
