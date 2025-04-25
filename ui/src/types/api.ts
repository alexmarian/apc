// API Response Types
export interface ApiResponse<T> {
  data: T;
  message?: string;
  status: number;
}

// Account Related Types
export interface Account {
  id: number;
  number: string;
  destination: string;
  description: string;
  association_id: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface AccountCreateRequest {
  number: string;
  destination: string;
  description: string;
}

export interface AccountUpdateRequest {
  number: string;
  destination: string;
  description: string;
}

// Association Related Types
export interface Association {
  id: number;
  name: string;
  address: string;
  administrator: string;
  createdAt: string;
  updatedAt: string;
}

// Building Related Types
export interface Building {
  id: number;
  name: string;
  address: string;
  cadastral_number: string;
  total_area: number;
  createdAt: string;
  updatedAt: string;
}

// Unit Related Types
export interface Unit {
  id: number;
  cadastral_number: string;
  building_id: number;
  unit_number: string;
  address: string;
  entrance: number;
  area: number;
  part: number;
  unit_type: string;
  floor: number;
  room_count: number;
  createdAt: string;
  updatedAt: string;
}

// Owner Related Types
export interface Owner {
  id: number;
  name: string;
  normalized_name: string;
  identification_number: string;
  contact_phone: string;
  contact_email: string;
  first_detected_at: string;
  created_at: string;
  updated_at: string;
}

// Expense Related Types
export interface Expense {
  id: number;
  amount: number;
  description: string;
  destination: string;
  date: string;
  month: number;
  year: number;
  category_id: number;
  category_type?: string;
  category_family?: string;
  category_name?: string;
  account_id: number;
  account_number?: string;
  account_name?: string;
  created_at?: string;
  updated_at?: string;
}

export interface ExpenseCreateRequest {
  amount: number;
  description: string;
  destination: string;
  date: string; // ISO format
  category_id: number;
  account_id: number;
}

// Category Related Types
export interface Category {
  id: number;
  type: string;
  family: string;
  name: string;
  is_deleted: boolean;
  association_id: number;
  created_at: string;
  updated_at: string;
}

export interface CategoryCreateRequest {
  type: string;
  family: string;
  name: string;
}

// Auth Related Types
export interface LoginRequest {
  login: string;
  password: string;
  totp: string;
  expires_in_seconds?: number;
}

export interface LoginResponse {
  login: string;
  token: string;
  refresh_token: string;
}
