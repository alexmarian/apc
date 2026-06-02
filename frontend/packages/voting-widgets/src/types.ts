export interface VotingOption {
  id: string
  text: string
}

export interface VotingConfig {
  type: 'yes_no' | 'single_choice' | 'multiple_choice' | 'ranking'
  options?: VotingOption[]
  allow_abstention: boolean
}

export interface MatterInfo {
  id: number
  title: string
  title_ru: string
  description: string
  description_ru: string
  matter_type: string
  order_index: number
  voting_config: VotingConfig
  is_informative: boolean
}

export interface BallotVote {
  matter_id: number
  values: string[]
}

export interface GatheringInfo {
  id: number
  title: string
  description: string
  status: string
  voting_mode: string
  qualified_units_count: number
  qualified_units_total_part: number
}

export interface OwnerInfo {
  id: number
  name: string
  identification: string
}

export interface UnitInfo {
  unit_id: number
  unit_number: string
  building_name: string
  area: number
  voting_weight: number
  is_available: boolean
}

export interface BallotInfo {
  ballot_id: number
  ballot_hash: string
  submitted_at: string | null
  ballot_content: Record<string, BallotVote>
  is_valid: boolean
}

export interface VoteResult {
  choice: string
  vote_count: number
  weight_sum: number
  percentage: number
  weight_percentage: number
}

export interface QuorumInfo {
  required: number
  achieved: number
  required_percentage: number
  achieved_percentage: number
  met: boolean
}

export interface MatterStatistics {
  total_participants: number
  total_votes: number
  total_weight: number
  abstentions: number
  participation_rate: number
}

export interface MatterResult {
  matter_id: number
  matter_title: string
  matter_type: string
  voting_config: VotingConfig
  votes: VoteResult[]
  statistics: MatterStatistics
  quorum_info?: QuorumInfo
  result: string
  is_passed: boolean
}

export interface GatheringSummary {
  qualified_units: number
  qualified_weight: number
  participating_units: number
  participating_weight: number
  voted_units: number
  voted_weight: number
  participation_rate: number
  voting_completion_rate: number
  voting_mode: string
}

export interface VoteResults {
  gathering_id: number
  results: MatterResult[]
  statistics: GatheringSummary
  generated_at: string
}

export interface MemberContext {
  gathering: GatheringInfo
  owner: OwnerInfo
  units: UnitInfo[]
  matters: MatterInfo[]
  ballot: BallotInfo | null
  results: VoteResults | null
}

export interface BallotReceipt {
  ballot_id: number
  ballot_hash: string
  submitted_at: string | null
  ballot_content?: Record<string, BallotVote>
}

export class HttpError extends Error {
  constructor(
    message: string,
    public readonly status: number
  ) {
    super(message)
    this.name = 'HttpError'
  }
}

export interface VotingService {
  getContext(): Promise<MemberContext>
  submitBallot(content: Record<string, BallotVote>): Promise<BallotReceipt>
}

export interface VotingWidgetProps {
  service: VotingService
  initialContext?: MemberContext
}
