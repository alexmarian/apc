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
  document_ref?: string;
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
  document_ref: string;
}

// Category Related Types
export interface Category {
  id: number;
  type: string;
  family: string;
  name: string;
  is_deleted: boolean;
  association_id: number;
  original_labels?: Record<string, string>;
  created_at: string;
  updated_at: string;
}

export interface CategoryCreateRequest {
  type: string;
  family: string;
  name: string;
  original_labels?: Record<string, string>;
}

export interface CategoryUpdateRequest {
  type: string;
  family: string;
  name: string;
  original_labels?: Record<string, string>;
}

export interface CategoryUsageResponse {
  category_id: number;
  usage_count: number;
  last_used_at: string | null;
  recent_expenses: Array<{
    id: number;
    description: string;
    amount: number;
    date: string;
  }>;
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
  unit_cadastral_number: string;
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

export interface ResetPasswordResponse {
  qrCode?: string;
}

export interface ResetPasswordRequest {
  token: string;
  new_password: string;
  reset_totp_secret: boolean;
}

// Gathering Related Types
export interface Gathering {
  id: number;
  association_id: number;
  title: string;
  description: string;
  location: string;
  scheduled_date: string;
  status: GatheringStatus;
  type: GatheringType;
  voting_mode: 'by_weight' | 'by_unit';
  qualification_criteria: QualificationCriteria;
  qualified_units: number;
  qualified_area: number;
  qualified_weight: number;
  participating_units: number;
  participating_area: number;
  participating_weight: number;
  created_at: string;
  updated_at: string;
}

export enum GatheringStatus {
  Draft = 'draft',
  Published = 'published',
  Active = 'active',
  Closed = 'closed',
  Tallied = 'tallied'
}

export enum GatheringType {
  Initial = 'initial',
  Repeated = 'repeated'
}

export interface QualificationCriteria {
  unit_types?: string[];
  floors?: number[];
  entrances?: number[];
}

export interface GatheringCreateRequest {
  title: string;
  description: string;
  intent: string;
  location: string;
  gathering_date: string;
  gathering_type: string;
  voting_mode?: 'by_weight' | 'by_unit';
  qualification_unit_types: string[];
  qualification_floors: number[];
  qualification_entrances: number[];
  qualification_custom_rule: string;
}

export interface GatheringUpdateRequest {
  title?: string;
  description?: string;
  intent?: string;
  location?: string;
  gathering_date?: string;
  gathering_type?: string;
  voting_mode?: 'by_weight' | 'by_unit';
  qualification_unit_types?: string[];
  qualification_floors?: number[];
  qualification_entrances?: number[];
  qualification_custom_rule?: string;
}

export interface GatheringStatusUpdateRequest {
  status: GatheringStatus;
}

// Voting Matter Related Types
export interface VotingMatter {
  id: number;
  gathering_id: number;
  title: string;
  description: string;
  matter_type: VotingMatterType;
  order_index: number;
  voting_config: VotingConfig;
  created_at: string;
  updated_at: string;
}

export enum VotingMatterType {
  Budget = 'budget',
  Election = 'election',
  Policy = 'policy',
  Poll = 'poll',
  Extraordinary = 'extraordinary'
}

export interface VotingConfig {
  type: VotingType;
  options?: VotingOption[];
  required_majority: MajorityType;
  required_majority_value?: number;
  quorum: number;
  is_anonymous: boolean;
  allow_abstention: boolean;
}

export interface VotingOption {
  id: string;
  text: string;
}

export enum VotingType {
  YesNo = 'yes_no',
  MultipleChoice = 'multiple_choice',
  SingleChoice = 'single_choice',
  Ranking = 'ranking'
}

export enum MajorityType {
  Simple = 'simple',
  Absolute = 'absolute',
  Qualified = 'qualified',
  Unanimous = 'unanimous'
}

export interface VotingMatterCreateRequest {
  title: string;
  description: string;
  matter_type: VotingMatterType;
  order_index: number;
  voting_config: VotingConfig;
}

export interface VotingMatterUpdateRequest {
  title?: string;
  description?: string;
  matter_type?: VotingMatterType;
  order_index?: number;
  voting_config?: VotingConfig;
}

// Participant Related Types
export interface GatheringParticipant {
  id: number;
  gathering_id: number;
  type: ParticipantType;
  owner_id: number;
  owner_name: string;
  contact_phone: string;
  contact_email: string;
  unit_ids: number[];
  delegation_document?: string;
  delegate_name?: string;
  delegate_contact?: string;
  checked_in_at?: string;
  created_at: string;
  updated_at: string;
}

export enum ParticipantType {
  Owner = 'owner',
  Delegate = 'delegate'
}

export interface ParticipantCreateRequest {
  participant_type: ParticipantType;
  owner_id?: number;
  unit_ids: number[];
  delegating_owner_id?: number;
  delegation_document_ref?: string;
  delegate_name?: string;
  delegate_contact?: string;
}

export interface ParticipantCheckInRequest {
  checked_in_at: string;
}

// Ballot and Voting Related Types
export interface Ballot {
  gathering_id: number;
  participant_id: number;
  votes: Vote[];
  submitted_at: string;
  submitted_from_ip: string;
  submitted_user_agent: string;
}

export interface Vote {
  matter_id: number;
  choice: string | string[];
  weight: number;
}

export interface BallotSubmissionRequest {
  votes: Vote[];
}

// Results and Statistics Types
export interface VotingResults {
  gathering_id: number;
  results: MatterResult[];
  statistics: GatheringStatistics;
  generated_at: string;
}

export interface MatterResult {
  matter_id: number;
  matter_title: string;
  matter_type: VotingMatterType;
  voting_config: VotingConfig;
  votes: VoteResult[];
  statistics: MatterStatistics;
  quorum_info?: QuorumInfo;
  result: string;
  is_passed: boolean;
}

export interface VoteResult {
  choice: string;
  vote_count: number;
  weight_sum: number;
  percentage: number;
  weight_percentage: number;
}

export interface MatterStatistics {
  total_participants: number;
  total_votes: number;
  total_weight: number;
  abstentions: number;
  abstention_weight: number;
  participation_rate: number;
  weight_participation_rate: number;
}

export interface GatheringStatistics {
  qualified_units: number;
  qualified_area: number;
  qualified_weight: number;
  participating_units: number;
  participating_area: number;
  participating_weight: number;
  voted_units: number;
  voted_weight: number;
  voted_area: number;
  participation_rate: number;        // participating / qualified (by count)
  voting_completion_rate: number;    // voted / participating (by count)
  voting_mode: string;
}

export interface QuorumInfo {
  required: number;
  achieved: number;
  required_percentage: number;
  achieved_percentage: number;
  met: boolean;
  voting_mode: string;
  gathering_type: string;
}

// Qualified Units Types
export interface QualifiedUnit {
  id: number;
  unit_number: string;
  cadastral_number: string;
  floor: number;
  entrance: number;
  area: number;
  part: number;
  unit_type: string;
  building_name: string;
  building_address: string;
  is_participating: boolean;
  owner_id: number;
  owner_name: string;
}

// Non-participating Owners Types
export interface NonParticipatingOwner {
  owner_id: number;
  owner_name: string;
  identification_number: string;
  contact_phone: string;
  contact_email: string;
  unit_ids: number[];
  unit_numbers: string[];
  total_units: number;
  total_area: number;
  total_weight: number;
}

// Audit and Notification Types
export interface VotingAuditLog {
  id: number;
  gathering_id: number;
  user_id: number;
  action: string;
  entity_type: string;
  entity_id: number;
  changes: Record<string, any>;
  ip_address: string;
  user_agent: string;
  created_at: string;
}

export interface VotingNotification {
  id: number;
  gathering_id: number;
  recipient_type: string;
  recipient_ids: number[];
  message: string;
  sent_at: string;
  delivery_status: string;
  created_at: string;
}

export interface NotificationCreateRequest {
  recipient_type: string;
  recipient_ids: number[];
  message: string;
}


