# Implementation Plan: Voting System Improvements

**Project:** APC Management Application - Voting System Enhancement
**Version:** 1.0
**Date:** 2025-12-07
**Status:** Draft

---

## Executive Summary

This implementation plan addresses critical issues in the voting system and introduces two voting modes (by-weight and by-unit) for apartment owner gatherings. The changes affect vote counting, quorum calculations, total possible votes determination, and results computation.

**Key Goals:**
1. Add voting mode attribute to gatherings (by-weight vs by-unit)
2. Fix voting matter vote count calculations
3. Implement proper voting strategy pattern
4. Unify results calculation across download ballots and download results
5. Add internationalization support
6. Correct quorum calculations based on gathering type and voting mode
7. Fix total possible votes calculation per gathering type

**Estimated Duration:** 3-4 weeks
**Risk Level:** High (Core business logic changes)

---

## Table of Contents

1. [Requirements Analysis](#1-requirements-analysis)
2. [Database Schema Changes](#2-database-schema-changes)
3. [Architecture & Design](#3-architecture--design)
4. [Implementation Roadmap](#4-implementation-roadmap)
5. [Testing Strategy](#5-testing-strategy)
6. [Risk Assessment](#6-risk-assessment)
7. [Rollout Plan](#7-rollout-plan)
8. [Implementation Tickets](#8-implementation-tickets)

---

## 1. Requirements Analysis

### 1.1 Core Requirements Breakdown

#### Req 1: Gathering Voting Mode Attribute

**Technical Components:**
- Database schema: Add `voting_mode` enum column to `gatherings` table
- Domain model: Add `VotingMode` field to `Gathering` struct
- API: Update create/update gathering endpoints to accept voting_mode
- Validation: Ensure voting_mode is required on creation

**Affected Files:**
- `/api/sql/schema/00019_add_voting_mode.sql` (new)
- `/api/sql/queries/gatherings.sql`
- `/api/internal/database/gatherings.sql.go` (generated)
- `/api/internal/handlers/gathering/domain/models.go`
- `/api/internal/handlers/gathering/handlers/gathering_handler.go`

**Dependencies:** None (foundational change)

---

#### Req 2: Voting Calculations - Support Both Modes

**Technical Components:**
- Strategy pattern implementation for vote counting
- VotingStrategy interface with two implementations:
  - `ByWeightStrategy`: Counts based on combined unit weights per owner
  - `ByUnitStrategy`: Counts based on separate votes per unit owned
- Update tally service to use strategy pattern
- Update results handler to use strategy pattern

**Affected Files:**
- `/api/internal/handlers/gathering/services/voting_strategy.go` (new)
- `/api/internal/handlers/gathering/services/by_weight_strategy.go` (new)
- `/api/internal/handlers/gathering/services/by_unit_strategy.go` (new)
- `/api/internal/handlers/gathering/services/tally_service.go`
- `/api/internal/handlers/gathering/handlers/results_handler.go`
- `/api/internal/handlers/gathering/handlers/export_handler.go`

**Dependencies:** Req 1 (voting mode must exist)

**Current Issues to Address:**
- Tally service currently only counts by weight (participant.UnitsPart)
- No distinction between owner-level vs unit-level voting
- Participation tracking conflates units and participants

---

#### Req 3: Total Possible Votes Calculation

**Technical Components:**
- New service method: `CalculateTotalPossibleVotes(gathering, votingMode)`
- Logic varies by gathering type:
  - **Initial/Repeated**: Total = sum of participated units
  - **Remote**: Total = sum of ALL qualified units
- Calculation varies by voting mode:
  - **By weight**: Sum of unit weights (part)
  - **By unit**: Count of units

**Affected Files:**
- `/api/internal/handlers/gathering/services/quorum_service.go`
- `/api/internal/handlers/gathering/services/stats_service.go`
- `/api/internal/handlers/gathering/handlers/results_handler.go`

**Dependencies:** Req 1, Req 2

**Current Issues:**
- Results handler uses `gathering.QualifiedUnitsTotalPart` for all gathering types
- Doesn't distinguish between participated vs qualified units
- No concept of "total possible votes" separate from qualified units

---

#### Req 4: Results Calculation Unification

**Technical Components:**
- New `VotingResultsService` that computes results once
- Store computed results in `voting_results` table when voting closes
- Cache mechanism for results
- Both download endpoints use same service
- Internationalization framework integration

**Affected Files:**
- `/api/sql/schema/00020_voting_results.sql` (new)
- `/api/sql/queries/voting_results.sql` (new)
- `/api/internal/handlers/gathering/services/voting_results_service.go` (new)
- `/api/internal/handlers/gathering/services/i18n_service.go` (new)
- `/api/internal/handlers/gathering/handlers/results_handler.go`
- `/api/internal/handlers/gathering/handlers/export_handler.go`
- `/api/internal/handlers/gathering/handlers/gathering_handler.go` (close gathering)

**Dependencies:** Req 2, Req 3

**Current Issues:**
- Results computed in 3 places with different logic:
  - `results_handler.go` HandleGetVoteResults (lines 106-268)
  - `export_handler.go` HandleDownloadVotingResults (lines 112-227)
  - `tally_service.go` UpdateVoteTallies (lines 62-144)
- No i18n support
- Performance: recalculation on every request

---

#### Req 5: Quorum Calculation Rules

**Technical Components:**
- Update `QuorumService.CalculateIfPassed()` method
- Implement quorum threshold by gathering type:
  - Initial: 50%
  - Repeated: 25%
  - Remote: 0% we consider it ok disregarding total number of votes.
- Quorum base varies by voting mode:
  - By weight: percentage of total weight
  - By unit: percentage of total units

**Affected Files:**
- `/api/internal/handlers/gathering/services/quorum_service.go`
- `/api/internal/handlers/gathering/domain/models.go` (add QuorumInfo struct)

**Dependencies:** Req 1, Req 3

**Current Issues:**
- Current quorum check (quorum_service.go line 30-36) uses config.Quorum directly
- Doesn't vary by gathering type
- Always uses weight-based calculation
- Compares participated vs qualified instead of using gathering type rules

---

### 1.2 Edge Cases & Special Scenarios

1. **Owner with multiple units, voting by-weight:**
   - Owner with Unit A (weight: 15) + Unit B (weight: 20)
   - Single participant record with total weight: 35
   - Single vote counts as 35 weight

2. **Owner with multiple units, voting by-unit:**
   - Owner with Unit A + Unit B
   - Single participant record but votes count separately per unit
   - Need to track unit-level votes, not just participant-level

3. **Remote gathering with partial participation:**
   - No quorum requirements for partial paticipation.

4. **Repeated gathering with low quorum:**
   - Only needs 25% of qualified units
   - Total possible votes = participated units (not qualified)
   - More lenient for decision-making

5. **Abstentions in different voting modes:**
   - By weight: abstention weight = participant's total weight
   - By unit: abstention count = number of units abstaining

6. **Migration of existing gatherings:**
   - Default to "by_weight" to maintain backward compatibility
   - Historical data remains valid

---

## 2. Database Schema Changes

### 2.1 Migration 00019: Add Voting Mode to Gatherings

**File:** `/api/sql/schema/00019_add_voting_mode.sql`

```sql
-- +goose Up
-- +goose StatementBegin
SELECT 'Adding voting_mode to gatherings table';

-- Add voting_mode column with default 'by_weight' for backward compatibility
ALTER TABLE gatherings
ADD COLUMN voting_mode TEXT NOT NULL DEFAULT 'by_weight'
CHECK (voting_mode IN ('by_weight', 'by_unit'));

CREATE INDEX idx_gatherings_voting_mode ON gatherings(voting_mode);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'Removing voting_mode from gatherings table';

DROP INDEX IF EXISTS idx_gatherings_voting_mode;
ALTER TABLE gatherings DROP COLUMN voting_mode;

-- +goose StatementEnd
```

**Impact:**
- Non-breaking: Default value ensures existing rows work
- Existing gatherings default to "by_weight" (current behavior)
- New gatherings must specify voting_mode

---

### 2.2 Migration 00020: Voting Results Cache Table

**File:** `/api/sql/schema/00020_voting_results.sql`

```sql
-- +goose Up
-- +goose StatementBegin
SELECT 'Creating voting_results cache table';

-- Table to store computed voting results (generated when gathering closes)
CREATE TABLE voting_results (
    id INTEGER PRIMARY KEY,
    gathering_id INTEGER NOT NULL REFERENCES gatherings(id) ON DELETE CASCADE,

    -- Computed results (JSON)
    results_data TEXT NOT NULL, -- Complete VoteResults struct as JSON

    -- Metadata
    voting_mode TEXT NOT NULL,
    gathering_type TEXT NOT NULL,
    total_possible_votes_weight NUMERIC NOT NULL,
    total_possible_votes_count INTEGER NOT NULL,
    quorum_threshold_percentage NUMERIC NOT NULL,
    quorum_met BOOLEAN NOT NULL,

    -- Timestamps
    computed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Ensure one result per gathering
    UNIQUE(gathering_id)
);

CREATE INDEX idx_voting_results_gathering ON voting_results(gathering_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'Dropping voting_results table';

DROP TABLE IF EXISTS voting_results;

-- +goose StatementEnd
```

**Impact:**
- New table for caching results
- Results computed once when gathering closes
- Eliminates redundant calculations
- Supports audit trail (computed_at timestamp)

---

### 2.3 Updated Query Files

#### Update: `/api/sql/queries/gatherings.sql`

Add voting_mode to create and update queries:

```sql
-- name: CreateGathering :one
INSERT INTO gatherings (association_id, title, description, intent, location, gathering_date,
                        gathering_type, voting_mode, status, qualification_unit_types,
                        qualification_floors, qualification_entrances,
                        qualification_custom_rule, qualified_units_count, qualified_units_total_part,
                        qualified_units_total_area)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateGathering :one
UPDATE gatherings
SET title                          = ?,
    description                    = ?,
    intent                         = ?,
    location                       = ?,
    gathering_date                 = ?,
    gathering_type                 = ?,
    voting_mode                    = ?,
    qualification_unit_types       = ?,
    qualification_floors           = ?,
    qualification_entrances        = ?,
    qualification_custom_rule      = ?,
    qualified_units_count          = ?,
    qualified_units_total_part     = ?,
    qualified_units_total_area     = ?,
    participating_units_count      = ?,
    participating_units_total_part = ?,
    participating_units_total_area = ?,
    updated_at                     = CURRENT_TIMESTAMP
WHERE id = ?
  AND association_id = ? RETURNING *;
```

#### New: `/api/sql/queries/voting_results.sql`

```sql
-- name: GetVotingResults :one
SELECT * FROM voting_results
WHERE gathering_id = ?;

-- name: CreateVotingResults :one
INSERT INTO voting_results (
    gathering_id,
    results_data,
    voting_mode,
    gathering_type,
    total_possible_votes_weight,
    total_possible_votes_count,
    quorum_threshold_percentage,
    quorum_met
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateVotingResults :one
UPDATE voting_results
SET results_data = ?,
    total_possible_votes_weight = ?,
    total_possible_votes_count = ?,
    quorum_threshold_percentage = ?,
    quorum_met = ?,
    computed_at = CURRENT_TIMESTAMP
WHERE gathering_id = ? RETURNING *;

-- name: DeleteVotingResults :exec
DELETE FROM voting_results
WHERE gathering_id = ?;
```

---

## 3. Architecture & Design

### 3.1 Strategy Pattern for Voting Calculations

**Interface:**

```go
// VotingStrategy defines the interface for vote counting strategies
type VotingStrategy interface {
    // CalculateVoteWeight calculates the weight/count for a participant's vote
    CalculateVoteWeight(participant GatheringParticipant, units []Unit) float64

    // CalculateTotalPossibleVotes calculates total possible votes for gathering
    CalculateTotalPossibleVotes(gathering Gathering, qualifiedUnits []Unit, participatedUnits []Unit) (weight float64, count int)

    // GetVotingModeName returns the name of this voting mode
    GetVotingModeName() string
}
```

**Implementations:**

1. **ByWeightStrategy:**
   - CalculateVoteWeight: Returns sum of all unit weights for participant
   - CalculateTotalPossibleVotes:
     - Initial/Repeated: Sum weights of participated units
     - Remote: Sum weights of all qualified units
   - GetVotingModeName: Returns "by_weight"

2. **ByUnitStrategy:**
   - CalculateVoteWeight: Returns count of units (each unit = 1 vote)
   - CalculateTotalPossibleVotes:
     - Initial/Repeated: Count of participated units
     - Remote: Count of all qualified units
   - GetVotingModeName: Returns "by_unit"

**Factory Pattern:**

```go
// GetVotingStrategy returns the appropriate strategy based on voting mode
func GetVotingStrategy(votingMode string) VotingStrategy {
    switch votingMode {
    case "by_weight":
        return &ByWeightStrategy{}
    case "by_unit":
        return &ByUnitStrategy{}
    default:
        return &ByWeightStrategy{} // Default fallback
    }
}
```

---

### 3.2 VotingResults Service Architecture

**Service Structure:**

```go
type VotingResultsService struct {
    db             *database.Queries
    quorumService  *QuorumService
    i18nService    *I18nService
}

// Main methods:
// 1. ComputeAndStoreResults(gatheringID, lang) - Compute and cache results
// 2. GetCachedResults(gatheringID, lang) - Retrieve cached results
// 3. InvalidateResults(gatheringID) - Clear cache when gathering reopens
// 4. GenerateMarkdownReport(results, format) - Export to markdown
```

**Workflow:**

1. When gathering closes (status: active → closed):
   - Call `ComputeAndStoreResults(gatheringID, "en")`
   - Calculate all vote tallies using appropriate strategy
   - Calculate quorum status
   - Store in `voting_results` table

2. When downloading results or ballots:
   - Call `GetCachedResults(gatheringID, lang)`
   - If not cached, compute on-the-fly (backward compatibility)
   - Return unified results structure

3. When reopening gathering:
   - Call `InvalidateResults(gatheringID)`
   - Delete cached results
   - Recalculate when closing again

---

### 3.3 Quorum Calculation Logic

**QuorumInfo Structure:**

```go
type QuorumInfo struct {
    Required           float64 `json:"required"`           // Required threshold (weight or count)
    Achieved           float64 `json:"achieved"`           // Achieved participation (weight or count)
    RequiredPercentage float64 `json:"required_percentage"` // Threshold percentage (25, 50, 100)
    AchievedPercentage float64 `json:"achieved_percentage"` // Actual participation percentage
    Met                bool    `json:"met"`                // Whether quorum is met
    VotingMode         string  `json:"voting_mode"`        // by_weight or by_unit
    GatheringType      string  `json:"gathering_type"`     // initial, repeated, remote
}
```

**Calculation Algorithm:**

```
function CalculateQuorum(gathering, participationStats, strategy):
    // Step 1: Determine threshold percentage based on gathering type
    thresholdPercent = switch gathering.GatheringType:
        case "initial":  return 50.0
        case "repeated": return 25.0
        case "remote":   return 100.0

    // Step 2: Calculate total possible votes based on gathering type
    if gathering.GatheringType == "remote":
        totalPossible = strategy.CalculateTotalPossibleVotes(
            gathering,
            qualifiedUnits,
            qualifiedUnits  // Remote uses ALL qualified units
        )
    else:
        totalPossible = strategy.CalculateTotalPossibleVotes(
            gathering,
            qualifiedUnits,
            participatedUnits  // Initial/Repeated uses participated units
        )

    // Step 3: Calculate achieved participation
    achieved = participationStats.TotalVoted  // From strategy

    // Step 4: Calculate required amount
    required = (totalPossible * thresholdPercent) / 100

    // Step 5: Determine if quorum is met
    met = achieved >= required

    return QuorumInfo{
        Required: required,
        Achieved: achieved,
        RequiredPercentage: thresholdPercent,
        AchievedPercentage: (achieved / totalPossible) * 100,
        Met: met,
        VotingMode: gathering.VotingMode,
        GatheringType: gathering.GatheringType
    }
```

---

### 3.4 Domain Model Updates

**Updated Gathering Model:**

```go
type Gathering struct {
    // ... existing fields ...
    VotingMode string `json:"voting_mode"` // "by_weight" or "by_unit"
}
```

**Updated VoteMatterResult:**

```go
type VoteMatterResult struct {
    // ... existing fields ...
    QuorumInfo QuorumInfo `json:"quorum_info"` // Detailed quorum information
}
```

**New VotingResultsCached:**

```go
type VotingResultsCached struct {
    ID                         int64           `json:"id"`
    GatheringID                int64           `json:"gathering_id"`
    ResultsData                VoteResults     `json:"results_data"`
    VotingMode                 string          `json:"voting_mode"`
    GatheringType              string          `json:"gathering_type"`
    TotalPossibleVotesWeight   float64         `json:"total_possible_votes_weight"`
    TotalPossibleVotesCount    int             `json:"total_possible_votes_count"`
    QuorumThresholdPercentage  float64         `json:"quorum_threshold_percentage"`
    QuorumMet                  bool            `json:"quorum_met"`
    ComputedAt                 time.Time       `json:"computed_at"`
}
```

---

### 3.5 API Contract Changes

**CreateGathering Request:**

```json
{
  "title": "Annual General Meeting 2025",
  "description": "...",
  "gathering_date": "2025-01-15T14:00:00Z",
  "gathering_type": "initial",
  "voting_mode": "by_unit",  // NEW REQUIRED FIELD
  // ... other fields
}
```

**GetVoteResults Response (Enhanced):**

```json
{
  "gathering_id": 123,
  "results": [
    {
      "matter_id": 1,
      "matter_title": "Budget Approval",
      "quorum_info": {  // NEW
        "required": 50.0,
        "achieved": 62.5,
        "required_percentage": 50.0,
        "achieved_percentage": 62.5,
        "met": true,
        "voting_mode": "by_unit",
        "gathering_type": "initial"
      },
      // ... existing fields
    }
  ],
  "statistics": {
    "qualified_units": 100,
    "qualified_weight": 10000.0,
    "participated_units": 75,  // NEW - clarified
    "participated_weight": 7500.0,
    "voted_units": 62,  // NEW - units that actually voted
    "voted_weight": 6250.0,
    "voting_mode": "by_unit"  // NEW
  }
}
```

**Backward Compatibility:**
- Existing API requests without `voting_mode` default to "by_weight"
- Existing responses remain valid (new fields are additive)

---

## 4. Implementation Roadmap

### Phase 1: Database & Core Models (Week 1)

**Milestone:** Database schema updated, domain models extended

#### Task 1.1: Database Migrations
- Create and test migration 00019 (voting_mode)
- Create and test migration 00020 (voting_results)
- Run migrations on dev database
- Verify rollback scripts work

**Files:**
- `/api/sql/schema/00019_add_voting_mode.sql`
- `/api/sql/schema/00020_voting_results.sql`

**Success Criteria:**
- Migrations run without errors
- Rollback succeeds
- Existing gatherings have voting_mode = 'by_weight'

---

#### Task 1.2: Update SQL Queries
- Add voting_mode to CreateGathering query
- Add voting_mode to UpdateGathering query
- Create voting_results queries
- Regenerate sqlc code: `sqlc generate`

**Files:**
- `/api/sql/queries/gatherings.sql`
- `/api/sql/queries/voting_results.sql` (new)
- `/api/internal/database/gatherings.sql.go` (generated)
- `/api/internal/database/voting_results.sql.go` (generated, new)

**Success Criteria:**
- sqlc generates without errors
- New methods available in database.Queries

---

#### Task 1.3: Update Domain Models
- Add VotingMode field to Gathering struct
- Add QuorumInfo struct
- Add VotingResultsCached struct
- Update mapper functions

**Files:**
- `/api/internal/handlers/gathering/domain/models.go`

**Success Criteria:**
- Code compiles
- JSON serialization works correctly
- Backward compatibility maintained (omitempty tags where needed)

---

### Phase 2: Strategy Pattern Implementation (Week 1-2)

**Milestone:** Voting strategies implemented and tested

#### Task 2.1: Create Voting Strategy Interface & Implementations
- Define VotingStrategy interface
- Implement ByWeightStrategy
- Implement ByUnitStrategy
- Implement strategy factory
- Write unit tests for each strategy

**Files:**
- `/api/internal/handlers/gathering/services/voting_strategy.go` (new)
- `/api/internal/handlers/gathering/services/by_weight_strategy.go` (new)
- `/api/internal/handlers/gathering/services/by_unit_strategy.go` (new)
- `/api/internal/handlers/gathering/services/voting_strategy_test.go` (new)

**Success Criteria:**
- All strategy methods implemented
- Unit tests pass with 100% coverage
- Edge cases handled (zero units, single unit, multiple units)

---

#### Task 2.2: Update Tally Service to Use Strategy
- Refactor UpdateVoteTallies to accept strategy
- Use strategy for weight calculations
- Maintain backward compatibility

**Files:**
- `/api/internal/handlers/gathering/services/tally_service.go`

**Success Criteria:**
- Tally calculations correct for both modes
- Existing tests still pass
- No performance regression

---

#### Task 2.3: Update Quorum Service
- Implement CalculateQuorum with gathering type logic
- Update CalculateIfPassed to use new quorum calculation
- Add QuorumInfo return type

**Files:**
- `/api/internal/handlers/gathering/services/quorum_service.go`

**Success Criteria:**
- Quorum calculated correctly for all 6 combinations (3 types × 2 modes)
- Edge cases covered (0 participation, 100% participation)
- Tests pass

---

### Phase 3: Results Service & Caching (Week 2)

**Milestone:** Unified results calculation with caching

#### Task 3.1: Create Voting Results Service
- Implement VotingResultsService
- Implement ComputeAndStoreResults
- Implement GetCachedResults
- Implement InvalidateResults
- Add comprehensive logging

**Files:**
- `/api/internal/handlers/gathering/services/voting_results_service.go` (new)
- `/api/internal/handlers/gathering/services/voting_results_service_test.go` (new)

**Success Criteria:**
- Results computed correctly
- Caching works properly
- Cache invalidation successful
- Performance benchmarks met

---

#### Task 3.2: Create I18n Service (Basic)
- Implement basic I18nService structure
- Add English translations
- Prepare for future language expansion
- Translation keys for vote results

**Files:**
- `/api/internal/handlers/gathering/services/i18n_service.go` (new)
- `/api/locales/en.json` (new)

**Success Criteria:**
- Basic i18n framework in place
- English translations complete
- Extensible for future languages

---

#### Task 3.3: Update Gathering Handler (Close Gathering)
- Add results computation when closing gathering
- Update status transition logic
- Add error handling for failed computation

**Files:**
- `/api/internal/handlers/gathering/handlers/gathering_handler.go`

**Success Criteria:**
- Results computed automatically on close
- Errors logged and handled gracefully
- Can still close gathering if computation fails (fallback)

---

### Phase 4: Handler Updates (Week 3)

**Milestone:** All handlers use unified results service

#### Task 4.1: Update Results Handler
- Refactor HandleGetVoteResults to use VotingResultsService
- Remove duplicated calculation logic
- Add quorum info to response
- Update statistics calculation

**Files:**
- `/api/internal/handlers/gathering/handlers/results_handler.go`

**Success Criteria:**
- Uses cached results when available
- Falls back to computation if cache miss
- Response includes all new fields
- Backward compatible response

---

#### Task 4.2: Update Export Handler
- Refactor HandleDownloadVotingResults to use VotingResultsService
- Refactor HandleDownloadVotingBallots to use VotingResultsService
- Remove duplicated calculation logic
- Add i18n support to markdown generation
- Add quorum information to exports

**Files:**
- `/api/internal/handlers/gathering/handlers/export_handler.go`

**Success Criteria:**
- Both exports use same calculation
- Markdown includes quorum info
- i18n ready (English first)
- Consistent with results handler

---

#### Task 4.3: Update Gathering Handler (Create/Update)
- Add voting_mode to create request validation
- Add voting_mode to update request
- Update request/response models

**Files:**
- `/api/internal/handlers/gathering/handlers/gathering_handler.go`

**Success Criteria:**
- voting_mode required on create
- voting_mode updatable (before publishing)
- Validation prevents invalid values

---

### Phase 5: Testing & Quality Assurance (Week 3-4)

**Milestone:** Comprehensive test coverage, bugs fixed

#### Task 5.1: Unit Tests
- Test all strategies
- Test quorum service
- Test results service
- Test i18n service
- Target: 85%+ coverage

**Files:**
- Multiple `*_test.go` files

**Success Criteria:**
- Coverage > 85%
- All edge cases tested
- Mock data realistic

---

#### Task 5.2: Integration Tests
- Test complete voting flow (create → vote → close → results)
- Test both voting modes
- Test all gathering types
- Test quorum scenarios (met/not met)

**Files:**
- `/api/internal/handlers/gathering/integration_test.go` (new)

**Success Criteria:**
- Full flow works for all combinations
- Database state correct
- Cache behaves correctly

---

#### Task 5.3: Migration Testing
- Test migration on copy of production data
- Verify backward compatibility
- Test rollback procedures

**Success Criteria:**
- Existing gatherings migrated correctly
- No data loss
- Performance acceptable

---

### Phase 6: Documentation & Deployment (Week 4)

**Milestone:** Production ready

#### Task 6.1: Documentation
- Update API documentation
- Add voting mode examples
- Document quorum calculation
- Update README

**Files:**
- `/docs/api/gatherings.md` (new/update)
- `/docs/voting-system.md` (new)
- `/README.md`

**Success Criteria:**
- Complete API examples
- Diagrams for voting modes
- Developer onboarding docs

---

#### Task 6.2: Performance Testing
- Load test with 1000 units
- Benchmark results computation
- Profile memory usage
- Optimize bottlenecks

**Success Criteria:**
- Results computation < 2 seconds for 1000 units
- Memory usage reasonable
- No N+1 queries

---

#### Task 6.3: Deployment Preparation
- Create deployment checklist
- Prepare rollback plan
- Database backup procedures
- Monitoring setup

**Success Criteria:**
- Deployment runbook complete
- Rollback tested
- Monitoring dashboards ready

---

## 5. Testing Strategy

### 5.1 Unit Test Requirements

#### Voting Strategy Tests

**ByWeightStrategy:**
- Single unit owner: weight calculated correctly
- Multi-unit owner: weights summed correctly
- Edge case: zero weight
- Edge case: very large weight (precision)

**ByUnitStrategy:**
- Single unit owner: count = 1
- Multi-unit owner: count = number of units
- Edge case: 100 units per owner

**Test Data:**
```go
testCases := []struct{
    name           string
    participant    GatheringParticipant
    units          []Unit
    expectedWeight float64  // for by-weight
    expectedCount  int      // for by-unit
}{
    {
        name: "single unit owner",
        participant: GatheringParticipant{UnitsInfo: []int64{1}, UnitsPart: 10.5},
        units: []Unit{{ID: 1, Part: 10.5}},
        expectedWeight: 10.5,
        expectedCount: 1,
    },
    {
        name: "multi-unit owner",
        participant: GatheringParticipant{UnitsInfo: []int64{1,2}, UnitsPart: 25.5},
        units: []Unit{{ID: 1, Part: 10.5}, {ID: 2, Part: 15.0}},
        expectedWeight: 25.5,
        expectedCount: 2,
    },
    // ... more cases
}
```

---

#### Quorum Service Tests

**Test Matrix:**

| Gathering Type | Voting Mode | Participation | Quorum Met | Test Name |
|---------------|-------------|---------------|------------|-----------|
| Initial | By Weight | 60% | Yes | initial_by_weight_quorum_met |
| Initial | By Weight | 40% | No | initial_by_weight_quorum_not_met |
| Initial | By Unit | 60% | Yes | initial_by_unit_quorum_met |
| Initial | By Unit | 40% | No | initial_by_unit_quorum_not_met |
| Repeated | By Weight | 30% | Yes | repeated_by_weight_quorum_met |
| Repeated | By Weight | 20% | No | repeated_by_weight_quorum_not_met |
| Repeated | By Unit | 30% | Yes | repeated_by_unit_quorum_met |
| Repeated | By Unit | 20% | No | repeated_by_unit_quorum_not_met |
| Remote | By Weight | 100% | Yes | remote_by_weight_quorum_met |
| Remote | By Weight | 99% | No | remote_by_weight_quorum_not_met |
| Remote | By Unit | 100% | Yes | remote_by_unit_quorum_met |
| Remote | By Unit | 99% | No | remote_by_unit_quorum_not_met |

**Minimum 12 test cases**

---

#### Results Service Tests

- Compute results for gathering with votes
- Cache results correctly
- Retrieve cached results
- Invalidate cache
- Handle missing cache (fallback)
- Concurrent access safety

---

### 5.2 Integration Test Scenarios

#### Scenario 1: Initial Gathering, By Unit Mode

```
1. Create gathering (initial, by_unit, 100 qualified units)
2. Add 3 voting matters
3. Create 60 participants (60 units)
4. Submit 50 ballots (50 units voted)
5. Close gathering
6. Verify:
   - Quorum: required = 50%, achieved = 50/100 = 50% → MET
   - Results cached in voting_results table
   - Download results shows correct counts
   - Download ballots shows correct data
   - Both downloads match
```

---

#### Scenario 2: Repeated Gathering, By Weight Mode

```
1. Create gathering (repeated, by_weight, total weight = 10000)
2. Add 2 voting matters
3. Create 30 participants (total weight = 3000)
4. Submit 25 ballots (total weight = 2600)
5. Close gathering
6. Verify:
   - Quorum: required = 25%, achieved = 2600/10000 = 26% → MET
   - Results show weight-based calculations
   - Multi-unit owners counted by combined weight
```

---

#### Scenario 3: Remote Gathering, By Unit Mode, Quorum Not Met

```
1. Create gathering (remote, by_unit, 100 qualified units)
2. Add 1 voting matter
3. Create 99 participants (99 units)
4. Submit 99 ballots
5. Close gathering
6. Verify:
   - Quorum: required = 100%, achieved = 99/100 = 99% → NOT MET
   - All matters marked as failed due to quorum
   - Results still computed and cached
```

---

### 5.3 Edge Case Testing

1. **Zero participation:** Gathering with 0 votes
2. **100% participation:** All qualified units vote
3. **Exact quorum threshold:** Participation exactly at threshold
4. **Large gathering:** 1000+ units
5. **Single unit association:** Only 1 qualified unit
6. **Unanimous votes:** All votes same option
7. **Split votes:** Exactly 50/50 split
8. **All abstentions:** Every ballot abstains
9. **Mixed participant types:** Owners and delegates
10. **Reopened gathering:** Close → reopen → close again

---

### 5.4 Performance Testing

**Benchmarks:**

```go
func BenchmarkComputeResults_100Units(b *testing.B)
func BenchmarkComputeResults_1000Units(b *testing.B)
func BenchmarkComputeResults_10000Units(b *testing.B)
```

**Targets:**
- 100 units: < 100ms
- 1000 units: < 2s
- 10000 units: < 30s

**Load Testing:**
- 100 concurrent result requests
- Response time < 500ms (with cache)
- Response time < 5s (without cache, 1000 units)

---

## 6. Risk Assessment

### 6.1 High-Risk Areas

#### Risk 1: Breaking Changes to Vote Counting

**Description:** New vote counting logic may produce different results than current system

**Impact:** High - Core business logic
**Probability:** Medium
**Severity:** Critical

**Mitigation:**
1. Run parallel calculations (old vs new) on test data
2. Create comparison report showing differences
3. Validate with stakeholders before deployment
4. Keep old calculation in separate function for reference
5. Feature flag for gradual rollout

**Rollback Plan:**
- Revert migration 00019
- Revert code changes
- Old behavior restored

---

#### Risk 2: Data Migration Issues

**Description:** Existing gatherings may have inconsistent data when adding voting_mode

**Impact:** Medium - Affects historical data
**Probability:** Low
**Severity:** High

**Mitigation:**
1. Default all existing gatherings to "by_weight" (current behavior)
2. Test migration on production data copy
3. Add data validation script post-migration
4. Create manual override for edge cases

**Rollback Plan:**
- Migration 00019 has tested rollback script
- No data loss (just column removal)

---

#### Risk 3: Performance Degradation

**Description:** Results computation may be slow for large gatherings

**Impact:** Medium - User experience
**Probability:** Medium
**Severity:** Medium

**Mitigation:**
1. Cache results in database
2. Compute asynchronously when closing gathering
3. Use database indices effectively
4. Profile and optimize hot paths
5. Load test with realistic data volumes

**Rollback Plan:**
- Disable caching if issues arise
- Fall back to on-the-fly computation

---

### 6.2 Medium-Risk Areas

#### Risk 4: Quorum Calculation Errors

**Description:** Complex quorum rules may have implementation bugs

**Impact:** High - Legal/compliance issue
**Probability:** Low
**Severity:** High

**Mitigation:**
1. Extensive test matrix (12+ test cases)
2. Manual verification by domain expert
3. Add detailed logging for quorum calculations
4. Include quorum calculation in audit log

**Detection:**
- Comprehensive unit tests
- Integration tests with known outcomes
- Manual review of test results

---

#### Risk 5: Backward Compatibility

**Description:** API changes may break existing clients

**Impact:** Medium - Client integration
**Probability:** Low
**Severity:** Medium

**Mitigation:**
1. Default voting_mode to "by_weight"
2. Make new fields optional/additive only
3. Version API if needed (/v2 endpoints)
4. Document breaking changes clearly

**Detection:**
- API contract testing
- Backward compatibility test suite

---

### 6.3 Low-Risk Areas

#### Risk 6: Cache Staleness

**Description:** Cached results may become stale if gathering is modified

**Impact:** Low - Results recomputed on close
**Probability:** Low
**Severity:** Low

**Mitigation:**
1. Clear cache when reopening gathering
2. Add cache timestamp validation
3. TTL on cached results

---

#### Risk 7: Internationalization Issues

**Description:** i18n implementation may have bugs or missing translations

**Impact:** Low - English fallback works
**Probability:** Medium
**Severity:** Low

**Mitigation:**
1. Start with English only
2. Use translation keys with fallbacks
3. Add more languages incrementally

---

### 6.4 Risk Mitigation Summary

| Risk | Mitigation Strategy | Success Metric |
|------|---------------------|----------------|
| Vote Counting Changes | Parallel testing, comparison report | 100% match on test data |
| Data Migration | Default values, validation script | 0 data loss, 0 invalid rows |
| Performance | Caching, profiling, optimization | < 2s for 1000 units |
| Quorum Errors | Test matrix, manual verification | All 12+ tests pass |
| Backward Compatibility | Optional fields, defaults | 0 breaking changes |
| Cache Staleness | Invalidation logic | Cache always fresh |
| i18n Issues | English-first, fallbacks | 100% English coverage |

---

## 7. Rollout Plan

### 7.1 Pre-Deployment Checklist

- [ ] All unit tests passing (coverage > 85%)
- [ ] All integration tests passing
- [ ] Performance benchmarks met
- [ ] Database migrations tested on production data copy
- [ ] Rollback procedures documented and tested
- [ ] API documentation updated
- [ ] Monitoring dashboards configured
- [ ] Deployment runbook reviewed
- [ ] Stakeholder sign-off on vote counting changes
- [ ] Database backup completed

---

### 7.2 Deployment Sequence

#### Step 1: Database Migration (Maintenance Window)

**Time:** 5 minutes
**Downtime:** None (migration runs live)

```bash
# Run migrations
cd /home/alexm/projects/apc/apc/api
goose -dir sql/schema sqlite3 database.db up

# Verify migration
sqlite3 database.db "SELECT COUNT(*) FROM gatherings WHERE voting_mode IS NULL;"
# Should return 0

sqlite3 database.db "SELECT COUNT(*) FROM voting_results;"
# Should return 0 (new table, empty)
```

**Rollback:**
```bash
goose -dir sql/schema sqlite3 database.db down
```

---

#### Step 2: Code Deployment

**Time:** 10 minutes
**Downtime:** 30 seconds (rolling restart)

```bash
# Build new version
go build -o apc-api ./cmd/api

# Stop old version
systemctl stop apc-api

# Deploy new version
cp apc-api /usr/local/bin/

# Start new version
systemctl start apc-api

# Verify health
curl http://localhost:8080/health
```

**Verification:**
- Health check passes
- Can create gathering with voting_mode
- Can retrieve existing gatherings
- Logs show no errors

**Rollback:**
```bash
# Restore old binary
cp /usr/local/bin/apc-api.backup /usr/local/bin/apc-api
systemctl restart apc-api
```

---

#### Step 3: Smoke Tests (Production)

**Time:** 15 minutes

1. Create test gathering (by_unit mode)
2. Add voting matter
3. Submit test ballot
4. Close gathering
5. Download results
6. Download ballots
7. Verify results match
8. Delete test gathering

**Success Criteria:**
- All operations succeed
- Results computation < 2s
- No errors in logs

---

#### Step 4: Monitoring & Observation

**Time:** 24 hours

Monitor:
- API response times (p50, p95, p99)
- Error rates
- Database query performance
- Cache hit rate
- Quorum calculation logs

**Alerts:**
- Error rate > 1%
- Response time p95 > 5s
- Database CPU > 80%

---

### 7.3 Feature Flag Strategy

**Optional:** Gradual rollout using feature flags

```go
// Feature flag check
func (h *GatheringHandler) HandleCreateGathering() {
    if !featureFlags.IsEnabled("voting_mode_selection") {
        // Force by_weight for now
        req.VotingMode = "by_weight"
    }
    // ... rest of handler
}
```

**Rollout Stages:**
1. **Week 1:** Internal testing only (flag: off)
2. **Week 2:** Enable for 10% of associations (flag: 10%)
3. **Week 3:** Enable for 50% of associations (flag: 50%)
4. **Week 4:** Enable for 100% (flag: 100%)

**Kill Switch:**
- If issues arise, disable flag immediately
- Revert to by_weight mode automatically

---

### 7.4 Rollback Procedures

#### Scenario 1: Critical Bug in Vote Counting

**Detection:** Incorrect results reported by user

**Action:**
1. Immediately stop creating new gatherings
2. Analyze reported issue
3. If confirmed critical:
   - Revert code to previous version
   - Run migration rollback
   - Notify users of issue
4. Fix bug in dev environment
5. Redeploy with fix

**Timeline:** 2 hours max

---

#### Scenario 2: Performance Issues

**Detection:** Response times exceed SLA

**Action:**
1. Enable results computation async if not already
2. Increase cache TTL
3. Add database indices if missing
4. If still issues:
   - Disable results caching
   - Fall back to on-the-fly computation
5. Optimize in dev environment
6. Redeploy with optimizations

**Timeline:** 4 hours max

---

#### Scenario 3: Data Integrity Issues

**Detection:** Inconsistent data in voting_results table

**Action:**
1. Clear voting_results table
2. Recompute all results from scratch
3. Verify integrity
4. If issues persist:
   - Disable caching entirely
   - Use on-the-fly computation only

**Timeline:** 6 hours max

---

### 7.5 Post-Deployment Validation

**Day 1:**
- Monitor error logs continuously
- Check results computation performance
- Verify cache hit rates
- Review user feedback

**Week 1:**
- Compare old vs new vote counts (parallel calculation)
- Validate quorum calculations manually
- Performance metrics review
- User acceptance testing

**Week 4:**
- Comprehensive system review
- Performance optimization
- Documentation updates
- Lessons learned session

**Success Criteria:**
- 0 critical bugs
- < 5 minor bugs
- Performance within SLA
- Positive user feedback
- Successful quorum calculations on 10+ gatherings

---

## 8. Implementation Tickets

### Ticket 1: Database Schema - Add Voting Mode

**Epic:** Voting System Improvements
**Priority:** High
**Estimate:** 3 days
**Dependencies:** None

**Description:**
Add voting_mode column to gatherings table and create voting_results cache table.

**Acceptance Criteria:**
- [ ] Migration 00019 created and tested
- [ ] Migration 00020 created and tested
- [ ] Migrations run successfully on dev database
- [ ] Rollback scripts tested and working
- [ ] Existing gatherings default to "by_weight"
- [ ] voting_results table created with correct schema
- [ ] Indices created for performance

**Technical Tasks:**
1. Create migration 00019_add_voting_mode.sql
2. Create migration 00020_voting_results.sql
3. Test migrations on local database
4. Test migrations on copy of production data
5. Test rollback scripts
6. Document migration procedure

**Files Changed:**
- `/api/sql/schema/00019_add_voting_mode.sql` (new)
- `/api/sql/schema/00020_voting_results.sql` (new)

**Testing:**
- Unit: N/A (SQL only)
- Integration: Migration test script
- Manual: Visual inspection of migrated data

---

### Ticket 2: Database Queries - Update SQLC Queries

**Epic:** Voting System Improvements
**Priority:** High
**Estimate:** 2 days
**Dependencies:** Ticket 1

**Description:**
Update SQL queries to include voting_mode and create queries for voting_results table.

**Acceptance Criteria:**
- [ ] gatherings.sql updated with voting_mode
- [ ] voting_results.sql created with CRUD queries
- [ ] sqlc generation successful
- [ ] Generated code compiles
- [ ] No breaking changes to existing queries

**Technical Tasks:**
1. Update CreateGathering query
2. Update UpdateGathering query
3. Create voting_results.sql with queries
4. Run sqlc generate
5. Fix any compilation errors
6. Update tests to use new generated code

**Files Changed:**
- `/api/sql/queries/gatherings.sql`
- `/api/sql/queries/voting_results.sql` (new)
- `/api/internal/database/gatherings.sql.go` (generated)
- `/api/internal/database/voting_results.sql.go` (generated, new)

**Testing:**
- Unit: Test generated query methods
- Integration: Test queries against database
- Manual: Execute queries in SQL console

---

### Ticket 3: Domain Models - Add Voting Mode

**Epic:** Voting System Improvements
**Priority:** High
**Estimate:** 2 days
**Dependencies:** Ticket 2

**Description:**
Update domain models to include voting mode and quorum information structures.

**Acceptance Criteria:**
- [ ] Gathering struct has VotingMode field
- [ ] QuorumInfo struct created
- [ ] VotingResultsCached struct created
- [ ] JSON tags correct
- [ ] Mapper functions updated
- [ ] Backward compatibility maintained

**Technical Tasks:**
1. Add VotingMode to Gathering struct
2. Create QuorumInfo struct
3. Create VotingResultsCached struct
4. Update DBGatheringToResponse mapper
5. Add JSON serialization tests
6. Update API examples in code comments

**Files Changed:**
- `/api/internal/handlers/gathering/domain/models.go`

**Testing:**
- Unit: JSON serialization/deserialization tests
- Integration: API request/response tests
- Manual: Check Swagger/API docs

---

### Ticket 4: Voting Strategy - Interface and Implementations

**Epic:** Voting System Improvements
**Priority:** High
**Estimate:** 5 days
**Dependencies:** Ticket 3

**Description:**
Implement strategy pattern for voting calculations with by-weight and by-unit strategies.

**Acceptance Criteria:**
- [ ] VotingStrategy interface defined
- [ ] ByWeightStrategy implemented
- [ ] ByUnitStrategy implemented
- [ ] Strategy factory function created
- [ ] Unit tests for all strategies (100% coverage)
- [ ] Edge cases handled

**Technical Tasks:**
1. Define VotingStrategy interface
2. Implement ByWeightStrategy
3. Implement ByUnitStrategy
4. Create GetVotingStrategy factory
5. Write comprehensive unit tests
6. Test edge cases (0 units, 100 units, etc.)
7. Benchmark performance

**Files Changed:**
- `/api/internal/handlers/gathering/services/voting_strategy.go` (new)
- `/api/internal/handlers/gathering/services/by_weight_strategy.go` (new)
- `/api/internal/handlers/gathering/services/by_unit_strategy.go` (new)
- `/api/internal/handlers/gathering/services/voting_strategy_test.go` (new)

**Testing:**
- Unit: 20+ test cases covering all scenarios
- Integration: Test with real gathering data
- Performance: Benchmark with 1000 units

---

### Ticket 5: Quorum Service - Update Calculation Logic

**Epic:** Voting System Improvements
**Priority:** High
**Estimate:** 4 days
**Dependencies:** Ticket 4

**Description:**
Update quorum service to calculate quorum based on gathering type and voting mode.

**Acceptance Criteria:**
- [ ] CalculateQuorum method implemented
- [ ] Gathering type logic correct (25%, 50%, 100%)
- [ ] Voting mode logic correct (weight vs unit)
- [ ] CalculateIfPassed updated
- [ ] 12+ test cases pass
- [ ] QuorumInfo returned with detailed breakdown

**Technical Tasks:**
1. Implement CalculateQuorum method
2. Update CalculateIfPassed to use new logic
3. Add QuorumInfo struct population
4. Write test matrix (3 types × 2 modes × 2 outcomes)
5. Test edge cases (exact threshold, 0%, 100%)
6. Add detailed logging

**Files Changed:**
- `/api/internal/handlers/gathering/services/quorum_service.go`

**Testing:**
- Unit: 12 test cases minimum
- Integration: Test with real gatherings
- Manual: Verify quorum calculations with stakeholder

---

### Ticket 6: Voting Results Service - Computation and Caching

**Epic:** Voting System Improvements
**Priority:** High
**Estimate:** 5 days
**Dependencies:** Ticket 5

**Description:**
Create VotingResultsService to compute results once and cache them in database.

**Acceptance Criteria:**
- [ ] VotingResultsService created
- [ ] ComputeAndStoreResults implemented
- [ ] GetCachedResults implemented
- [ ] InvalidateResults implemented
- [ ] Results computed correctly using strategy pattern
- [ ] Caching works properly
- [ ] Performance acceptable (< 2s for 1000 units)

**Technical Tasks:**
1. Create VotingResultsService struct
2. Implement ComputeAndStoreResults
3. Implement GetCachedResults with fallback
4. Implement InvalidateResults
5. Integrate with voting strategies
6. Add comprehensive logging
7. Write unit tests
8. Write integration tests
9. Performance benchmark

**Files Changed:**
- `/api/internal/handlers/gathering/services/voting_results_service.go` (new)
- `/api/internal/handlers/gathering/services/voting_results_service_test.go` (new)

**Testing:**
- Unit: Test computation, caching, invalidation
- Integration: Full flow with database
- Performance: Benchmark with varying sizes

---

### Ticket 7: I18n Service - Basic Framework

**Epic:** Voting System Improvements
**Priority:** Medium
**Estimate:** 3 days
**Dependencies:** None (parallel with others)

**Description:**
Create basic internationalization service and English translations.

**Acceptance Criteria:**
- [ ] I18nService created
- [ ] Translation key structure defined
- [ ] English translations complete
- [ ] Fallback mechanism works
- [ ] Extensible for future languages

**Technical Tasks:**
1. Create I18nService struct
2. Define translation key constants
3. Create en.json with English translations
4. Implement translation lookup with fallback
5. Write unit tests
6. Document how to add new languages

**Files Changed:**
- `/api/internal/handlers/gathering/services/i18n_service.go` (new)
- `/api/locales/en.json` (new)
- `/api/locales/README.md` (new)

**Testing:**
- Unit: Test translations, fallbacks
- Manual: Verify English text correctness

---

### Ticket 8: Gathering Handler - Add Voting Mode to Create/Update

**Epic:** Voting System Improvements
**Priority:** High
**Estimate:** 3 days
**Dependencies:** Ticket 3

**Description:**
Update gathering handler to accept and validate voting_mode in create and update operations.

**Acceptance Criteria:**
- [ ] CreateGathering accepts voting_mode
- [ ] UpdateGathering accepts voting_mode
- [ ] Validation ensures valid values
- [ ] Default to "by_weight" if not provided (backward compat)
- [ ] API tests updated

**Technical Tasks:**
1. Update CreateGatheringRequest struct
2. Add voting_mode validation
3. Update create handler logic
4. Update update handler logic
5. Update API tests
6. Update API documentation

**Files Changed:**
- `/api/internal/handlers/gathering/handlers/gathering_handler.go`
- `/api/internal/handlers/gathering/domain/models.go`

**Testing:**
- Unit: Test request validation
- Integration: Test create/update with voting_mode
- Manual: Test via API client

---

### Ticket 9: Gathering Handler - Results Computation on Close

**Epic:** Voting System Improvements
**Priority:** High
**Estimate:** 3 days
**Dependencies:** Ticket 6

**Description:**
Update gathering close logic to compute and cache results automatically.

**Acceptance Criteria:**
- [ ] Results computed when gathering closes
- [ ] Results cached in voting_results table
- [ ] Error handling for failed computation
- [ ] Gathering still closes even if computation fails
- [ ] Audit log entry created

**Technical Tasks:**
1. Update close gathering handler
2. Call VotingResultsService.ComputeAndStoreResults
3. Add error handling and logging
4. Create audit log entry
5. Write tests for success and failure cases

**Files Changed:**
- `/api/internal/handlers/gathering/handlers/gathering_handler.go`

**Testing:**
- Unit: Test close with mock results service
- Integration: Test full close flow
- Error: Test computation failure handling

---

### Ticket 10: Results Handler - Use Voting Results Service

**Epic:** Voting System Improvements
**Priority:** High
**Estimate:** 4 days
**Dependencies:** Ticket 6

**Description:**
Refactor results handler to use VotingResultsService instead of computing inline.

**Acceptance Criteria:**
- [ ] HandleGetVoteResults uses VotingResultsService
- [ ] Cached results used when available
- [ ] Falls back to computation if cache miss
- [ ] Response includes QuorumInfo
- [ ] Response includes updated statistics
- [ ] Backward compatible response format

**Technical Tasks:**
1. Refactor HandleGetVoteResults
2. Remove inline computation logic
3. Use VotingResultsService.GetCachedResults
4. Add QuorumInfo to response
5. Update statistics calculation
6. Update tests

**Files Changed:**
- `/api/internal/handlers/gathering/handlers/results_handler.go`

**Testing:**
- Unit: Test with cached and non-cached results
- Integration: Test full API flow
- Backward Compat: Test with old gatherings

---

### Ticket 11: Export Handler - Use Voting Results Service

**Epic:** Voting System Improvements
**Priority:** High
**Estimate:** 4 days
**Dependencies:** Ticket 6, Ticket 7

**Description:**
Refactor export handler to use VotingResultsService and add i18n support.

**Acceptance Criteria:**
- [ ] HandleDownloadVotingResults uses VotingResultsService
- [ ] HandleDownloadVotingBallots uses VotingResultsService
- [ ] Both exports show consistent data
- [ ] Quorum info included in markdown
- [ ] i18n support for English
- [ ] Markdown formatting improved

**Technical Tasks:**
1. Refactor HandleDownloadVotingResults
2. Refactor HandleDownloadVotingBallots
3. Remove inline computation logic
4. Add QuorumInfo to markdown output
5. Integrate I18nService
6. Update markdown templates
7. Update tests

**Files Changed:**
- `/api/internal/handlers/gathering/handlers/export_handler.go`

**Testing:**
- Unit: Test markdown generation
- Integration: Test downloads with cached results
- Manual: Verify markdown format and content

---

### Ticket 12: Tally Service - Use Voting Strategy

**Epic:** Voting System Improvements
**Priority:** Medium
**Estimate:** 3 days
**Dependencies:** Ticket 4

**Description:**
Update tally service to use voting strategy pattern.

**Acceptance Criteria:**
- [ ] UpdateVoteTallies accepts voting strategy
- [ ] Uses strategy for vote weight calculation
- [ ] Backward compatible with existing tallies
- [ ] Tests updated

**Technical Tasks:**
1. Add strategy parameter to UpdateVoteTallies
2. Use strategy.CalculateVoteWeight
3. Update tests
4. Verify tallies match expected values

**Files Changed:**
- `/api/internal/handlers/gathering/services/tally_service.go`

**Testing:**
- Unit: Test with both strategies
- Integration: Test tally updates
- Regression: Test with existing data

---

### Ticket 13: Integration Testing - Complete Voting Flow

**Epic:** Voting System Improvements
**Priority:** High
**Estimate:** 5 days
**Dependencies:** All previous tickets

**Description:**
Create comprehensive integration tests for complete voting flow covering all scenarios.

**Acceptance Criteria:**
- [ ] 3 main scenarios tested (initial by-unit, repeated by-weight, remote by-unit)
- [ ] All gathering types tested
- [ ] Both voting modes tested
- [ ] Quorum met and not met tested
- [ ] Cache behavior tested
- [ ] Edge cases tested

**Technical Tasks:**
1. Create integration test framework
2. Implement Scenario 1: Initial + By Unit
3. Implement Scenario 2: Repeated + By Weight
4. Implement Scenario 3: Remote + By Unit + No Quorum
5. Add edge case tests
6. Add cache invalidation tests
7. Document test data setup

**Files Changed:**
- `/api/internal/handlers/gathering/integration_test.go` (new)
- `/api/internal/handlers/gathering/testdata/` (new directory)

**Testing:**
- Integration: All scenarios pass
- Performance: Tests complete in < 30s

---

### Ticket 14: Performance Testing and Optimization

**Epic:** Voting System Improvements
**Priority:** High
**Estimate:** 4 days
**Dependencies:** Ticket 13

**Description:**
Performance test the system and optimize bottlenecks.

**Acceptance Criteria:**
- [ ] Benchmarks created for key operations
- [ ] Load test with 1000 units passes
- [ ] Results computation < 2s for 1000 units
- [ ] Cache hit rate > 90%
- [ ] No N+1 query issues
- [ ] Database indices optimal

**Technical Tasks:**
1. Create benchmark tests
2. Run load tests
3. Profile CPU and memory
4. Identify bottlenecks
5. Optimize hot paths
6. Add/optimize database indices
7. Verify improvements

**Files Changed:**
- `/api/internal/handlers/gathering/benchmark_test.go` (new)
- Various service files (optimizations)

**Testing:**
- Performance: All benchmarks meet targets
- Load: 100 concurrent requests handled

---

### Ticket 15: Documentation and Deployment Prep

**Epic:** Voting System Improvements
**Priority:** Medium
**Estimate:** 3 days
**Dependencies:** Ticket 14

**Description:**
Complete documentation and prepare for deployment.

**Acceptance Criteria:**
- [ ] API documentation updated
- [ ] Voting system guide created
- [ ] Migration guide created
- [ ] Deployment runbook created
- [ ] Rollback procedures documented
- [ ] Monitoring dashboards configured

**Technical Tasks:**
1. Update API documentation with examples
2. Create voting system guide with diagrams
3. Create migration guide
4. Create deployment runbook
5. Document rollback procedures
6. Set up monitoring dashboards
7. Create deployment checklist

**Files Changed:**
- `/docs/api/gatherings.md` (new/update)
- `/docs/voting-system.md` (new)
- `/docs/deployment/voting-system-deployment.md` (new)
- `/README.md`

**Testing:**
- Manual: Follow deployment runbook in staging
- Manual: Test rollback procedures

---

## Appendix A: File Change Summary

### New Files (15)

1. `/api/sql/schema/00019_add_voting_mode.sql`
2. `/api/sql/schema/00020_voting_results.sql`
3. `/api/sql/queries/voting_results.sql`
4. `/api/internal/database/voting_results.sql.go` (generated)
5. `/api/internal/handlers/gathering/services/voting_strategy.go`
6. `/api/internal/handlers/gathering/services/by_weight_strategy.go`
7. `/api/internal/handlers/gathering/services/by_unit_strategy.go`
8. `/api/internal/handlers/gathering/services/voting_results_service.go`
9. `/api/internal/handlers/gathering/services/i18n_service.go`
10. `/api/internal/handlers/gathering/services/voting_strategy_test.go`
11. `/api/internal/handlers/gathering/services/voting_results_service_test.go`
12. `/api/internal/handlers/gathering/integration_test.go`
13. `/api/internal/handlers/gathering/benchmark_test.go`
14. `/api/locales/en.json`
15. `/docs/voting-system.md`

### Modified Files (8)

1. `/api/sql/queries/gatherings.sql`
2. `/api/internal/database/gatherings.sql.go` (generated)
3. `/api/internal/handlers/gathering/domain/models.go`
4. `/api/internal/handlers/gathering/services/tally_service.go`
5. `/api/internal/handlers/gathering/services/quorum_service.go`
6. `/api/internal/handlers/gathering/handlers/gathering_handler.go`
7. `/api/internal/handlers/gathering/handlers/results_handler.go`
8. `/api/internal/handlers/gathering/handlers/export_handler.go`

**Total:** 23 files (15 new, 8 modified)

---

## Appendix B: Database Schema Diagrams

### Before Changes

```
gatherings
├── id
├── association_id
├── title
├── gathering_type (initial, repeated, remote)
├── status
├── qualified_units_count
├── qualified_units_total_part
├── participating_units_count
├── participating_units_total_part
└── ...

vote_tallies
├── gathering_id
├── voting_matter_id
└── tally_data (JSON)
```

### After Changes

```
gatherings
├── id
├── association_id
├── title
├── gathering_type (initial, repeated, remote)
├── voting_mode (by_weight, by_unit) [NEW]
├── status
├── qualified_units_count
├── qualified_units_total_part
├── participating_units_count
├── participating_units_total_part
└── ...

vote_tallies
├── gathering_id
├── voting_matter_id
└── tally_data (JSON)

voting_results [NEW TABLE]
├── id
├── gathering_id (FK)
├── results_data (JSON)
├── voting_mode
├── gathering_type
├── total_possible_votes_weight
├── total_possible_votes_count
├── quorum_threshold_percentage
├── quorum_met
└── computed_at
```

---

## Appendix C: Quorum Calculation Examples

### Example 1: Initial Gathering, By Unit Mode

**Setup:**
- Gathering type: Initial
- Voting mode: By unit
- Total qualified units: 100
- Participated units: 75
- Voted units: 60

**Calculation:**
```
Threshold: 50% (initial gathering)
Total possible votes: 100 (all qualified units for initial)
Required: 100 × 50% = 50 units
Achieved: 60 units
Quorum met: 60 >= 50 → YES
```

---

### Example 2: Repeated Gathering, By Weight Mode

**Setup:**
- Gathering type: Repeated
- Voting mode: By weight
- Total qualified weight: 10,000
- Participated weight: 3,000
- Voted weight: 2,600

**Calculation:**
```
Threshold: 25% (repeated gathering)
Total possible votes: 10,000 (all qualified weight for repeated)
Required: 10,000 × 25% = 2,500
Achieved: 2,600
Quorum met: 2,600 >= 2,500 → YES
```

---

### Example 3: Remote Gathering, By Unit Mode

**Setup:**
- Gathering type: Remote
- Voting mode: By unit
- Total qualified units: 100
- Participated units: 100 (all must participate for remote)
- Voted units: 99

**Calculation:**
```
Threshold: 100% (remote gathering)
Total possible votes: 100 (all qualified units for remote)
Required: 100 × 100% = 100 units
Achieved: 99 units
Quorum met: 99 >= 100 → NO
```

---

## Appendix D: Glossary

**Gathering Types:**
- **Initial Gathering:** First meeting, requires 50% quorum
- **Repeated Gathering:** Follow-up meeting, requires 25% quorum (more lenient)
- **Remote Gathering:** Online/mail voting, requires 100% quorum (all qualified must vote)

**Voting Modes:**
- **By Weight:** Each owner's vote counts based on combined weight of their units
- **By Unit:** Each owner casts separate votes for each unit they own

**Key Metrics:**
- **Qualified Units:** Total units eligible to vote based on gathering criteria
- **Participated Units:** Units whose owners attended/registered
- **Voted Units:** Units whose owners actually submitted ballots

**Quorum:** Minimum participation required for voting results to be valid

**Weight/Part:** Voting power of a unit, typically based on area or ownership percentage

---

## Appendix E: Success Metrics

### Technical Metrics

1. **Code Quality:**
   - Test coverage > 85%
   - 0 critical bugs
   - < 5 minor bugs in first month

2. **Performance:**
   - Results computation < 2s for 1000 units
   - API response time p95 < 500ms
   - Cache hit rate > 90%

3. **Reliability:**
   - 0 data loss incidents
   - 0 incorrect quorum calculations
   - Uptime > 99.9%

### Business Metrics

1. **Adoption:**
   - 10+ gatherings use new voting modes within 1 month
   - 0 rollbacks required
   - Positive user feedback

2. **Compliance:**
   - 100% accurate quorum calculations
   - Audit trail complete
   - Legal requirements met

3. **Efficiency:**
   - 50% reduction in results computation time (via caching)
   - 0 duplicate calculation logic
   - Single source of truth for results

---

## Document Change Log

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-12-07 | Implementation Planner | Initial comprehensive plan |

---

**End of Implementation Plan**
