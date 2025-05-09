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

export interface UnitUpdateRequest {
  unit_number?: string;
  address?: string;
  entrance?: number;
  unit_type?: string;
  floor?: number;
  room_count?: number;
}

export interface UnitReportDetails {
  unit_details: {
    id: number;
    cadastral_number: string;
    unit_number: string;
    address: string;
    entrance: number;
    area: number;
    part: number;
    unit_type: string;
    floor: number;
    room_count: number;
    created_at: string;
    updated_at: string;
  };
  building_details: {
    id: number;
    name: string;
    address: string;
    cadastral_number: string;
    total_area: number;
  };
  current_owners: Array<{
    id: number;
    name: string;
    identification_number: string;
    contact_phone: string;
    contact_email: string;
    is_active: boolean;
  }>;
  ownership_history: Array<{
    id: number;
    owner_id: number;
    owner_name: string;
    start_date: string;
    end_date: string | null;
    is_active: boolean;
    registration_document: string;
    registration_date: string;
  }>;
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
  category_type: string;
  category_family: string;
  category_name: string;
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

// Ownership Related Types
export interface Ownership {
  id: number;
  unit_id: number;
  owner_id: number;
  owner_name: string;
  owner_normalized_name: string;
  owner_identification_number: string;
  association_id: number;
  start_date: string;
  end_date: string | null;
  is_active: boolean;
  registration_document: string;
  registration_date: string;
  created_at: string;
  updated_at: string;
}

export interface OwnershipCreateRequest {
  owner_id: number;
  start_date: string;
  end_date?: string | null;
  registration_document: string;
  registration_date: string;
  is_exclusive?: boolean; // Whether this ownership deactivates all others
}

export interface OwnerCreateRequest {
  name: string;
  identification_number: string;
  contact_phone: string;
  contact_email: string;
}

export interface OwnershipDisableRequest {
  end_date?: string;
  disable_reason?: string;
}

// Expense Distribution Types
export interface ExpenseDistributionResponse {
  start_date: string;
  end_date: string;
  unit_type: string | null;
  category_type: string | null;
  category_family: string | null;
  category_id: number | null;
  distribution_method: 'area' | 'count' | 'equal';
  total_units: number;
  total_area: number;
  category_totals: Record<string, CategoryTotal>;
  total_expenses: number;
  unit_distributions: UnitDistribution[];
}

export interface CategoryTotal {
  amount: number;
  id: number;
  type: string;
  family: string;
}

export interface UnitDistribution {

  id: number;
  building_id: number;
  unit_number: string;
  building_name: string;
  building_address: string;
  unit_type: string;
  area: number;
  part: number;
  distribution_factor: number;
  expenses_share: Record<string, number>;
  total_share: number;
  detailed_expenses?: Record<number, ExpenseShare>;
}

export interface ExpenseShare {
  expense_id: number;
  description: string;
  date: string;
  category_name: string;
  category_type: string;
  category_family: string;
  category_id: number;
  total_amount: number;
  unit_share: number;
}

// Owner Report Types
export interface OwnerReportItem {
  owner: Owner;
  co_owners?: OwnerCoOwner[];
  units?: OwnerUnit[];
  statistics: OwnerStats;
}

export interface OwnerCoOwner {
  id: number;
  name: string;
  identification_number: string;
  contact_phone: string;
  contact_email: string;
  shared_unit_ids: number[];
  shared_unit_nums: string[];
}

export interface OwnerUnit {
  unit_id: number;
  unit_number: string;
  building_name: string;
  building_address: string;
  area: number;
  part: number;
  unit_type: string;
}

export interface OwnerStats {
  total_units: number;
  total_area: number;
  total_condo_part: number;
}

export interface VotingOwner {
  owner_id: number;
  name: string;
  identification_number: string;
  contact_phone: string;
  contact_email: string;
  units: Array<VotingUnit>;
  total_units: number;
  total_area: number;
  total_condo_part: number;
}

export interface VotingUnit {

  unit_id: number;
  unit_number: string;
  building_id: number;
  building_name: string;
  area: number;
  part: number;
  unit_type: string;

}

export interface VotingOwnersResponse {
  voting_owners: VotingOwner[];
  total_units: number;
  total_area: number;
}

export enum UnitType {
  Apartment = 'apartment',
  Commercial = 'commercial',
  Office = 'office',
  Parking = 'parking',
  Storage = 'storage'
}


