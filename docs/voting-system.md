# Voting System Documentation

**APC Management Application**
**Version:** 2.0
**Last Updated:** 2025-12-07

---

## Table of Contents

1. [Overview](#overview)
2. [Voting Modes](#voting-modes)
3. [Gathering Types](#gathering-types)
4. [Quorum Requirements](#quorum-requirements)
5. [Quorum Calculation Examples](#quorum-calculation-examples)
6. [Architecture](#architecture)
7. [API Reference](#api-reference)
8. [Developer Guide](#developer-guide)
9. [Glossary](#glossary)

---

## Overview

The APC Voting System provides a flexible framework for managing apartment owner association gatherings and voting. The system supports two distinct voting modes and three gathering types, each with specific quorum requirements and calculation rules.

### Key Features

- **Two Voting Modes**: Vote by weight (aggregated) or by unit (separate votes)
- **Three Gathering Types**: Initial, Repeated, and Remote gatherings
- **Dynamic Quorum Calculation**: Automatic quorum calculation based on gathering type
- **Results Caching**: Computed results are cached for performance
- **Internationalization**: Support for multiple languages (currently English)
- **Audit Trail**: Complete tracking of voting activities

---

## Voting Modes

The system supports two voting modes that determine how votes are counted and aggregated.

### By-Weight Mode (`by_weight`)

In by-weight mode, each owner's vote is weighted based on the combined ownership share (part) of all their units.

**Characteristics:**
- Owner votes once with combined weight of all their units
- Weight typically based on unit area or ownership percentage
- Single vote per owner regardless of number of units owned
- Total voting power = sum of all unit weights

**Example:**

```
Owner A owns:
  - Unit 101 (weight: 15.5)
  - Unit 102 (weight: 20.0)
  - Total weight: 35.5

When voting:
  - Owner A casts 1 vote
  - Vote counts as 35.5 weight
  - All units vote together as a block
```

**Use Cases:**
- Traditional association voting
- Voting based on ownership percentage
- When unit size/value should influence voting power

**Visual Representation:**

```
By-Weight Voting:
┌──────────────────────────────────────────┐
│ Owner A (Units: 101, 102, 103)           │
│ Combined Weight: 45.0                    │
│                                          │
│ ┌──────────────────────────────────┐    │
│ │ Single Vote = 45.0 weight        │    │
│ └──────────────────────────────────┘    │
└──────────────────────────────────────────┘

Total Votes = Sum of all owner weights
Quorum = % of total weight
```

---

### By-Unit Mode (`by_unit`)

In by-unit mode, each unit gets a separate vote, even if owned by the same person.

**Characteristics:**
- Each unit votes separately
- Multi-unit owners cast multiple votes (one per unit)
- Each vote has equal weight (1 vote = 1 unit)
- Total voting power = count of units

**Example:**

```
Owner A owns:
  - Unit 101
  - Unit 102
  - Unit 103

When voting:
  - Owner A casts 3 separate votes
  - Each vote counts as 1
  - Units can vote differently (if allowed by system)
```

**Use Cases:**
- One-unit-one-vote policies
- Democratic voting (equal voice per unit)
- When each unit should have equal influence

**Visual Representation:**

```
By-Unit Voting:
┌──────────────────────────────────────────┐
│ Owner A                                  │
│                                          │
│ ┌────────────┐ ┌────────────┐ ┌────────┐│
│ │ Unit 101   │ │ Unit 102   │ │ Unit   ││
│ │ Vote = 1   │ │ Vote = 1   │ │ 103    ││
│ └────────────┘ └────────────┘ │ Vote=1 ││
│                                └────────┘│
└──────────────────────────────────────────┘

Total Votes = Count of all units
Quorum = % of total units
```

---

### Comparison Table

| Aspect | By-Weight | By-Unit |
|--------|-----------|---------|
| Vote counting | Sum of weights | Count of units |
| Multi-unit owner | 1 vote (combined weight) | N votes (one per unit) |
| Voting power | Proportional to ownership | Equal per unit |
| Total possible votes | Sum of all weights | Count of all units |
| Quorum calculation | % of total weight | % of total units |
| Best for | Area-based voting | Democratic voting |

---

## Gathering Types

The system supports three types of gatherings, each with different quorum requirements and voting rules.

### Initial Gathering (`initial`)

The first meeting for a particular purpose or period.

**Characteristics:**
- Quorum requirement: 50% of qualified units/weight
- Strictest quorum threshold
- Commonly used for annual general meetings
- Total possible votes = all qualified units (regardless of participation)

**When to Use:**
- Annual general meetings
- First attempt at voting on important matters
- When high participation is legally required

---

### Repeated Gathering (`repeated`)

A follow-up meeting held when the initial gathering didn't achieve quorum.

**Characteristics:**
- Quorum requirement: 25% of qualified units/weight
- More lenient threshold to allow decision-making
- Total possible votes = all qualified units (regardless of participation)
- Typically scheduled after failed initial gathering

**When to Use:**
- When initial gathering failed to achieve quorum
- Second attempt at voting on same matters
- When lower quorum is acceptable by law/bylaws

---

### Remote Gathering (`remote`)

Online or mail-in voting without a physical meeting.

**Characteristics:**
- Quorum requirement: 100% of qualified units/weight must vote
- All qualified units must participate
- Total possible votes = all qualified units
- No physical meeting required

**When to Use:**
- Correspondence voting
- Mail-in ballots
- Online voting platforms
- When unanimous participation is required

---

### Gathering Type Comparison

| Type | Quorum | Total Possible Votes | Use Case |
|------|--------|---------------------|----------|
| Initial | 50% | All qualified units | First meeting |
| Repeated | 25% | All qualified units | Follow-up after failed initial |
| Remote | 100% | All qualified units | Remote/mail voting |

---

## Quorum Requirements

Quorum is the minimum participation required for voting results to be valid. The quorum calculation varies by gathering type and voting mode.

### Quorum Thresholds by Gathering Type

```
Initial Gathering:     50% quorum required
Repeated Gathering:    25% quorum required
Remote Gathering:     100% quorum required
```

### Quorum Calculation Formula

```
General Formula:
quorum_met = (achieved_votes >= required_votes)

Where:
required_votes = total_possible_votes × threshold_percentage

threshold_percentage:
  - initial:  50%
  - repeated: 25%
  - remote:  100%
```

### By Voting Mode

**By-Weight Mode:**
```
total_possible_votes = sum of all qualified unit weights
achieved_votes = sum of weights that voted
required_votes = total_possible_votes × threshold_percentage

Example (Initial, By-Weight):
  Total weight: 10,000
  Threshold: 50%
  Required: 10,000 × 0.50 = 5,000
  Achieved: 6,250
  Quorum Met: YES (6,250 >= 5,000)
```

**By-Unit Mode:**
```
total_possible_votes = count of all qualified units
achieved_votes = count of units that voted
required_votes = total_possible_votes × threshold_percentage

Example (Initial, By-Unit):
  Total units: 100
  Threshold: 50%
  Required: 100 × 0.50 = 50
  Achieved: 62
  Quorum Met: YES (62 >= 50)
```

---

## Quorum Calculation Examples

### Example 1: Initial Gathering, By-Unit Mode

**Scenario:**
- Association with 100 apartments
- First annual general meeting
- One-unit-one-vote policy

**Setup:**
```
Gathering type: initial
Voting mode: by_unit
Total qualified units: 100
Participated units: 75
Voted units: 60
```

**Calculation:**
```
1. Determine threshold: initial = 50%
2. Calculate total possible votes: 100 units (all qualified)
3. Calculate required votes: 100 × 50% = 50 units
4. Count achieved votes: 60 units
5. Check quorum: 60 >= 50 → QUORUM MET
```

**Result:**
```json
{
  "quorum_info": {
    "required": 50.0,
    "achieved": 60.0,
    "required_percentage": 50.0,
    "achieved_percentage": 60.0,
    "met": true,
    "voting_mode": "by_unit",
    "gathering_type": "initial"
  }
}
```

**Interpretation:**
The gathering achieved 60% participation (60 out of 100 units), which exceeds the 50% requirement. All voting results are valid.

---

### Example 2: Repeated Gathering, By-Weight Mode

**Scenario:**
- Previous initial gathering failed to achieve quorum
- Second attempt with lower threshold
- Voting based on ownership percentage

**Setup:**
```
Gathering type: repeated
Voting mode: by_weight
Total qualified weight: 10,000
Participated weight: 3,000
Voted weight: 2,600
```

**Calculation:**
```
1. Determine threshold: repeated = 25%
2. Calculate total possible votes: 10,000 weight (all qualified)
3. Calculate required votes: 10,000 × 25% = 2,500
4. Count achieved votes: 2,600 weight
5. Check quorum: 2,600 >= 2,500 → QUORUM MET
```

**Result:**
```json
{
  "quorum_info": {
    "required": 2500.0,
    "achieved": 2600.0,
    "required_percentage": 25.0,
    "achieved_percentage": 26.0,
    "met": true,
    "voting_mode": "by_weight",
    "gathering_type": "repeated"
  }
}
```

**Interpretation:**
The gathering achieved 26% participation by weight, which exceeds the 25% requirement for repeated gatherings. Results are valid.

---

### Example 3: Remote Gathering, By-Unit Mode, Quorum Not Met

**Scenario:**
- Mail-in voting system
- Requires all units to participate
- One unit failed to submit ballot

**Setup:**
```
Gathering type: remote
Voting mode: by_unit
Total qualified units: 100
Participated units: 100
Voted units: 99
```

**Calculation:**
```
1. Determine threshold: remote = 100%
2. Calculate total possible votes: 100 units (all qualified)
3. Calculate required votes: 100 × 100% = 100
4. Count achieved votes: 99 units
5. Check quorum: 99 >= 100 → QUORUM NOT MET
```

**Result:**
```json
{
  "quorum_info": {
    "required": 100.0,
    "achieved": 99.0,
    "required_percentage": 100.0,
    "achieved_percentage": 99.0,
    "met": false,
    "voting_mode": "by_unit",
    "gathering_type": "remote"
  }
}
```

**Interpretation:**
The gathering achieved 99% participation, but remote gatherings require 100%. All voting results are INVALID due to failed quorum.

---

### Example 4: Initial Gathering, By-Weight, Exact Threshold

**Scenario:**
- Exactly at the quorum threshold
- Edge case testing

**Setup:**
```
Gathering type: initial
Voting mode: by_weight
Total qualified weight: 8,000
Voted weight: 4,000 (exactly 50%)
```

**Calculation:**
```
1. Determine threshold: initial = 50%
2. Calculate total possible votes: 8,000
3. Calculate required votes: 8,000 × 50% = 4,000
4. Count achieved votes: 4,000
5. Check quorum: 4,000 >= 4,000 → QUORUM MET
```

**Result:**
Quorum is met when achieved votes equal required votes (>= comparison).

---

### Example 5: Multi-Unit Owner Voting

**Scenario:**
- Owner with multiple units
- Demonstrating difference between voting modes

**Owner Profile:**
```
Owner Maria owns:
  - Unit A (weight: 15.0, area: 50m²)
  - Unit B (weight: 20.0, area: 65m²)
  - Unit C (weight: 25.0, area: 80m²)
  Total weight: 60.0
```

**By-Weight Voting:**
```
Maria's vote = 60.0 weight (all units combined)
Counts as: 1 vote with weight 60.0

In quorum calculation:
  - Contributes 60.0 to total achieved weight
  - Single ballot submission
```

**By-Unit Voting:**
```
Maria's votes = 3 votes (one per unit)
Counts as: 3 separate votes, each weight 1

In quorum calculation:
  - Contributes 3 to total achieved count
  - Three ballot submissions (or one ballot with 3 unit selections)
```

**Comparison:**
```
┌─────────────────┬──────────────┬─────────────┐
│ Aspect          │ By-Weight    │ By-Unit     │
├─────────────────┼──────────────┼─────────────┤
│ Vote count      │ 1            │ 3           │
│ Vote value      │ 60.0         │ 3 × 1 = 3   │
│ Ballots         │ 1            │ 1 (or 3)    │
│ Quorum impact   │ 60.0 weight  │ 3 units     │
└─────────────────┴──────────────┴─────────────┘
```

---

### Example 6: Edge Case - Zero Participation

**Setup:**
```
Gathering type: initial
Voting mode: by_unit
Total qualified units: 100
Voted units: 0
```

**Calculation:**
```
Required: 100 × 50% = 50 units
Achieved: 0 units
Quorum met: NO (0 < 50)
```

**Result:**
All voting matters are invalid. Gathering should be rescheduled.

---

### Example 7: Edge Case - 100% Participation

**Setup:**
```
Gathering type: initial
Voting mode: by_weight
Total qualified weight: 5,000
Voted weight: 5,000 (all units voted)
```

**Calculation:**
```
Required: 5,000 × 50% = 2,500
Achieved: 5,000
Quorum met: YES (5,000 >= 2,500)
Achieved percentage: 100%
```

**Result:**
Perfect participation. All results are maximally valid.

---

## Architecture

The voting system uses a strategy pattern to handle different voting modes and a service layer for results computation and caching.

### Strategy Pattern

The system implements the Strategy design pattern to encapsulate voting calculation algorithms.

**Class Diagram:**

```
┌─────────────────────────────┐
│   VotingStrategy            │
│   (interface)               │
├─────────────────────────────┤
│ + CalculateVoteWeight()     │
│ + CalculateTotalPossible()  │
│ + GetVotingModeName()       │
└─────────────────────────────┘
           △
           │
    ┌──────┴──────┐
    │             │
┌───┴───────┐ ┌──┴────────────┐
│ByWeight   │ │ ByUnit        │
│Strategy   │ │ Strategy      │
├───────────┤ ├───────────────┤
│ Weight-   │ │ Unit-based    │
│ based     │ │ counting      │
│ counting  │ │               │
└───────────┘ └───────────────┘
```

**Interface Definition:**

```go
// VotingStrategy defines the interface for vote counting strategies
type VotingStrategy interface {
    // CalculateVoteWeight calculates the weight/count for a participant's vote
    CalculateVoteWeight(participant GatheringParticipant, units []Unit) float64

    // CalculateTotalPossibleVotes calculates total possible votes for gathering
    CalculateTotalPossibleVotes(
        gathering Gathering,
        qualifiedUnits []Unit,
        participatedUnits []Unit,
    ) (weight float64, count int)

    // GetVotingModeName returns the name of this voting mode
    GetVotingModeName() string
}
```

---

### Strategy Implementations

#### ByWeightStrategy

Implements weight-based vote counting.

**Algorithm:**

```go
func (s *ByWeightStrategy) CalculateVoteWeight(
    participant GatheringParticipant,
    units []Unit,
) float64 {
    // Sum all unit weights for this participant
    totalWeight := 0.0
    for _, unit := range units {
        totalWeight += unit.Part
    }
    return totalWeight
}

func (s *ByWeightStrategy) CalculateTotalPossibleVotes(
    gathering Gathering,
    qualifiedUnits []Unit,
    participatedUnits []Unit,
) (weight float64, count int) {
    // For all gathering types, use qualified units
    totalWeight := 0.0
    for _, unit := range qualifiedUnits {
        totalWeight += unit.Part
    }

    return totalWeight, len(qualifiedUnits)
}
```

**Characteristics:**
- Returns sum of unit weights
- Single vote per owner with combined weight
- Used for traditional ownership-percentage voting

---

#### ByUnitStrategy

Implements unit-based vote counting.

**Algorithm:**

```go
func (s *ByUnitStrategy) CalculateVoteWeight(
    participant GatheringParticipant,
    units []Unit,
) float64 {
    // Each unit counts as 1, regardless of weight
    return float64(len(units))
}

func (s *ByUnitStrategy) CalculateTotalPossibleVotes(
    gathering Gathering,
    qualifiedUnits []Unit,
    participatedUnits []Unit,
) (weight float64, count int) {
    // For all gathering types, count qualified units
    unitCount := len(qualifiedUnits)

    return float64(unitCount), unitCount
}
```

**Characteristics:**
- Returns count of units
- Each unit = 1 vote
- Used for democratic one-unit-one-vote systems

---

### Strategy Factory

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
        // Default to by_weight for backward compatibility
        return &ByWeightStrategy{}
    }
}
```

**Usage:**

```go
// Get strategy based on gathering's voting mode
strategy := GetVotingStrategy(gathering.VotingMode)

// Use strategy to calculate votes
voteWeight := strategy.CalculateVoteWeight(participant, units)

// Use strategy to calculate total possible votes
totalWeight, totalCount := strategy.CalculateTotalPossibleVotes(
    gathering,
    qualifiedUnits,
    participatedUnits,
)
```

---

### Service Architecture

**Component Diagram:**

```
┌─────────────────────────────────────────────────────┐
│                    API Layer                         │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐    │
│  │ Gathering  │  │  Results   │  │  Export    │    │
│  │  Handler   │  │  Handler   │  │  Handler   │    │
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘    │
└────────┼───────────────┼───────────────┼────────────┘
         │               │               │
┌────────┼───────────────┼───────────────┼────────────┐
│        ▼               ▼               ▼            │
│  ┌──────────────────────────────────────────────┐  │
│  │      VotingResultsService (Singleton)        │  │
│  ├──────────────────────────────────────────────┤  │
│  │ + ComputeAndStoreResults()                   │  │
│  │ + GetCachedResults()                         │  │
│  │ + InvalidateResults()                        │  │
│  └───────┬──────────────┬──────────────┬────────┘  │
│          │              │              │            │
│  ┌───────▼────┐  ┌──────▼─────┐  ┌────▼────────┐  │
│  │  Voting    │  │  Quorum    │  │   I18n      │  │
│  │  Strategy  │  │  Service   │  │  Service    │  │
│  └────────────┘  └────────────┘  └─────────────┘  │
│                                                     │
│                Service Layer                        │
└─────────────────────────────────────────────────────┘
         │
┌────────┼─────────────────────────────────────────┐
│        ▼                                          │
│  ┌──────────────────────────────────────────┐   │
│  │      Database Layer (SQLite)             │   │
│  ├──────────────────────────────────────────┤   │
│  │  ┌──────────┐  ┌──────────────┐          │   │
│  │  │gatherings│  │voting_results│          │   │
│  │  └──────────┘  └──────────────┘          │   │
│  └──────────────────────────────────────────┘   │
│                                                   │
│                Data Layer                         │
└───────────────────────────────────────────────────┘
```

---

### VotingResultsService

Centralized service for computing and caching voting results.

**Structure:**

```go
type VotingResultsService struct {
    db            *database.Queries
    quorumService *QuorumService
    i18nService   *I18nService
}
```

**Key Methods:**

```go
// ComputeAndStoreResults computes voting results and stores them in cache
func (s *VotingResultsService) ComputeAndStoreResults(
    ctx context.Context,
    gatheringID int64,
    lang string,
) (*VoteResults, error)

// GetCachedResults retrieves cached results or computes on-the-fly
func (s *VotingResultsService) GetCachedResults(
    ctx context.Context,
    gatheringID int64,
    lang string,
) (*VoteResults, error)

// InvalidateResults clears cached results (when gathering reopens)
func (s *VotingResultsService) InvalidateResults(
    ctx context.Context,
    gatheringID int64,
) error
```

**Workflow:**

```
Gathering Closes:
  ├─> VotingResultsService.ComputeAndStoreResults()
  ├─> Get voting strategy based on voting_mode
  ├─> Calculate all vote tallies
  ├─> Calculate quorum status
  ├─> Store in voting_results table
  └─> Return results

Results Request:
  ├─> VotingResultsService.GetCachedResults()
  ├─> Check voting_results table
  ├─> If cached: return from cache
  ├─> If not cached: compute on-the-fly
  └─> Return results

Gathering Reopens:
  ├─> VotingResultsService.InvalidateResults()
  ├─> Delete from voting_results table
  └─> Results will be recomputed on next close
```

---

### QuorumService

Service for calculating quorum based on gathering type and voting mode.

**Structure:**

```go
type QuorumService struct {
    config *Config
}

type QuorumInfo struct {
    Required           float64 `json:"required"`
    Achieved           float64 `json:"achieved"`
    RequiredPercentage float64 `json:"required_percentage"`
    AchievedPercentage float64 `json:"achieved_percentage"`
    Met                bool    `json:"met"`
    VotingMode         string  `json:"voting_mode"`
    GatheringType      string  `json:"gathering_type"`
}
```

**Algorithm:**

```go
func (s *QuorumService) CalculateQuorum(
    gathering Gathering,
    participationStats ParticipationStats,
    strategy VotingStrategy,
) QuorumInfo {
    // Step 1: Determine threshold percentage
    var thresholdPercent float64
    switch gathering.GatheringType {
    case "initial":
        thresholdPercent = 50.0
    case "repeated":
        thresholdPercent = 25.0
    case "remote":
        thresholdPercent = 100.0
    }

    // Step 2: Calculate total possible votes
    totalWeight, totalCount := strategy.CalculateTotalPossibleVotes(
        gathering,
        qualifiedUnits,
        participatedUnits,
    )

    // Step 3: Get achieved participation from stats
    achieved := participationStats.TotalVoted

    // Step 4: Calculate required amount
    required := (totalWeight * thresholdPercent) / 100.0

    // Step 5: Determine if quorum is met
    met := achieved >= required

    return QuorumInfo{
        Required:           required,
        Achieved:           achieved,
        RequiredPercentage: thresholdPercent,
        AchievedPercentage: (achieved / totalWeight) * 100.0,
        Met:                met,
        VotingMode:         gathering.VotingMode,
        GatheringType:      gathering.GatheringType,
    }
}
```

---

### Database Schema

**Gatherings Table:**

```sql
CREATE TABLE gatherings (
    id INTEGER PRIMARY KEY,
    association_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    gathering_type TEXT NOT NULL,  -- 'initial', 'repeated', 'remote'
    voting_mode TEXT NOT NULL DEFAULT 'by_weight',  -- NEW FIELD
    status TEXT NOT NULL,
    qualified_units_count INTEGER,
    qualified_units_total_part NUMERIC,
    participating_units_count INTEGER,
    participating_units_total_part NUMERIC,
    -- ... other fields
    CHECK (voting_mode IN ('by_weight', 'by_unit'))
);

CREATE INDEX idx_gatherings_voting_mode ON gatherings(voting_mode);
```

**Voting Results Cache Table:**

```sql
CREATE TABLE voting_results (
    id INTEGER PRIMARY KEY,
    gathering_id INTEGER NOT NULL REFERENCES gatherings(id) ON DELETE CASCADE,

    -- Computed results (JSON)
    results_data TEXT NOT NULL,

    -- Metadata
    voting_mode TEXT NOT NULL,
    gathering_type TEXT NOT NULL,
    total_possible_votes_weight NUMERIC NOT NULL,
    total_possible_votes_count INTEGER NOT NULL,
    quorum_threshold_percentage NUMERIC NOT NULL,
    quorum_met BOOLEAN NOT NULL,

    -- Timestamps
    computed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(gathering_id)
);

CREATE INDEX idx_voting_results_gathering ON voting_results(gathering_id);
```

---

## API Reference

### Create Gathering

Create a new gathering with specified voting mode.

**Endpoint:**
```
POST /api/associations/{id}/gatherings
```

**Request Body:**

```json
{
  "title": "Annual General Meeting 2025",
  "description": "Regular annual meeting",
  "gathering_date": "2025-01-15T14:00:00Z",
  "gathering_type": "initial",
  "voting_mode": "by_unit",
  "location": "Community Center",
  "qualification_unit_types": ["apartment"],
  "qualification_floors": [1, 2, 3, 4, 5]
}
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| title | string | Yes | Gathering title |
| gathering_date | datetime | Yes | Date and time of gathering |
| gathering_type | string | Yes | Type: "initial", "repeated", or "remote" |
| voting_mode | string | Yes | Mode: "by_weight" or "by_unit" |
| location | string | No | Physical location (not required for remote) |

**Response:**

```json
{
  "id": 123,
  "title": "Annual General Meeting 2025",
  "gathering_type": "initial",
  "voting_mode": "by_unit",
  "status": "draft",
  "qualified_units_count": 100,
  "qualified_units_total_part": 10000.0,
  "created_at": "2025-12-07T10:00:00Z"
}
```

---

### Get Vote Results

Retrieve voting results for a closed gathering.

**Endpoint:**
```
GET /api/gatherings/{id}/results
```

**Response:**

```json
{
  "gathering_id": 123,
  "gathering_type": "initial",
  "voting_mode": "by_unit",
  "status": "closed",
  "results": [
    {
      "matter_id": 1,
      "matter_title": "Budget Approval 2025",
      "matter_description": "Approve the proposed budget",
      "quorum_info": {
        "required": 50.0,
        "achieved": 62.0,
        "required_percentage": 50.0,
        "achieved_percentage": 62.0,
        "met": true,
        "voting_mode": "by_unit",
        "gathering_type": "initial"
      },
      "votes": {
        "for": 45,
        "against": 10,
        "abstain": 7
      },
      "result": "approved"
    }
  ],
  "statistics": {
    "qualified_units": 100,
    "qualified_weight": 10000.0,
    "participated_units": 75,
    "participated_weight": 7500.0,
    "voted_units": 62,
    "voted_weight": 6200.0,
    "voting_mode": "by_unit"
  }
}
```

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| quorum_info | object | Detailed quorum information |
| quorum_info.met | boolean | Whether quorum was achieved |
| quorum_info.required | number | Required votes/weight for quorum |
| quorum_info.achieved | number | Actual votes/weight achieved |
| statistics.voted_units | number | Number of units that voted |
| statistics.voted_weight | number | Total weight of votes cast |

---

### Download Voting Results

Download results as markdown document.

**Endpoint:**
```
GET /api/gatherings/{id}/results/download
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| lang | string | "en" | Language code for translations |

**Response:**

```
Content-Type: text/markdown
Content-Disposition: attachment; filename="gathering-123-results.md"

# Voting Results - Annual General Meeting 2025

**Gathering Type:** Initial
**Voting Mode:** By Unit
**Date:** 2025-01-15

## Quorum Status

✓ Quorum Met

- Required: 50 units (50.0%)
- Achieved: 62 units (62.0%)
- Status: VALID

## Voting Matters

### 1. Budget Approval 2025

- For: 45 votes
- Against: 10 votes
- Abstain: 7 votes
- **Result: APPROVED**

...
```

---

### Example API Usage

#### Example 1: Create Initial Gathering with By-Unit Voting

```bash
curl -X POST https://api.example.com/api/associations/1/gatherings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "title": "Annual General Meeting 2025",
    "gathering_date": "2025-01-15T14:00:00Z",
    "gathering_type": "initial",
    "voting_mode": "by_unit",
    "location": "Community Center"
  }'
```

**Response:**
```json
{
  "id": 123,
  "voting_mode": "by_unit",
  "gathering_type": "initial",
  "status": "draft"
}
```

---

#### Example 2: Create Repeated Gathering with By-Weight Voting

```bash
curl -X POST https://api.example.com/api/associations/1/gatherings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "title": "Budget Meeting - Second Attempt",
    "gathering_date": "2025-02-01T14:00:00Z",
    "gathering_type": "repeated",
    "voting_mode": "by_weight",
    "location": "Community Center"
  }'
```

---

#### Example 3: Get Results with Quorum Information

```bash
curl -X GET https://api.example.com/api/gatherings/123/results \
  -H "Authorization: Bearer {token}"
```

**Response includes:**
```json
{
  "results": [
    {
      "quorum_info": {
        "met": true,
        "required": 50.0,
        "achieved": 62.0,
        "voting_mode": "by_unit",
        "gathering_type": "initial"
      }
    }
  ]
}
```

---

## Developer Guide

### Adding Support for New Voting Modes

To add a new voting mode:

1. **Create Strategy Implementation:**

```go
// NewModeStrategy implements VotingStrategy
type NewModeStrategy struct{}

func (s *NewModeStrategy) CalculateVoteWeight(
    participant GatheringParticipant,
    units []Unit,
) float64 {
    // Implement custom calculation
    return customCalculation(participant, units)
}

func (s *NewModeStrategy) CalculateTotalPossibleVotes(
    gathering Gathering,
    qualifiedUnits []Unit,
    participatedUnits []Unit,
) (weight float64, count int) {
    // Implement custom calculation
    return customTotal(qualifiedUnits)
}

func (s *NewModeStrategy) GetVotingModeName() string {
    return "new_mode"
}
```

2. **Update Factory:**

```go
func GetVotingStrategy(votingMode string) VotingStrategy {
    switch votingMode {
    case "by_weight":
        return &ByWeightStrategy{}
    case "by_unit":
        return &ByUnitStrategy{}
    case "new_mode":  // Add new case
        return &NewModeStrategy{}
    default:
        return &ByWeightStrategy{}
    }
}
```

3. **Update Database Schema:**

```sql
ALTER TABLE gatherings DROP CONSTRAINT gatherings_voting_mode_check;
ALTER TABLE gatherings ADD CONSTRAINT gatherings_voting_mode_check
    CHECK (voting_mode IN ('by_weight', 'by_unit', 'new_mode'));
```

4. **Add Tests:**

```go
func TestNewModeStrategy(t *testing.T) {
    strategy := &NewModeStrategy{}
    // Add comprehensive tests
}
```

---

### Testing Voting Logic

**Unit Test Example:**

```go
func TestByUnitStrategy_CalculateVoteWeight(t *testing.T) {
    strategy := &ByUnitStrategy{}

    participant := GatheringParticipant{
        UnitsInfo: []int64{1, 2, 3},
        UnitsPart: 60.0,
    }

    units := []Unit{
        {ID: 1, Part: 15.0},
        {ID: 2, Part: 20.0},
        {ID: 3, Part: 25.0},
    }

    weight := strategy.CalculateVoteWeight(participant, units)

    // Should return count of units (3), not total weight (60)
    assert.Equal(t, 3.0, weight)
}
```

**Integration Test Example:**

```go
func TestGatheringVotingFlow_InitialByUnit(t *testing.T) {
    // Create gathering
    gathering := createGathering(t, GatheringRequest{
        GatheringType: "initial",
        VotingMode:    "by_unit",
    })

    // Add participants and votes
    addParticipants(t, gathering.ID, 60)
    submitVotes(t, gathering.ID)

    // Close gathering
    closeGathering(t, gathering.ID)

    // Get results
    results := getResults(t, gathering.ID)

    // Verify quorum
    assert.True(t, results.QuorumInfo.Met)
    assert.Equal(t, 50.0, results.QuorumInfo.RequiredPercentage)
}
```

---

### Performance Considerations

**Caching Strategy:**

Results are cached when gathering closes to avoid expensive recalculation:

```go
// On gathering close
results := votingResultsService.ComputeAndStoreResults(ctx, gatheringID)
// Results stored in voting_results table

// On results request
results := votingResultsService.GetCachedResults(ctx, gatheringID)
// Retrieves from cache if available, computes if not
```

**Optimization Tips:**

1. **Use database indices:**
   - Index on `gathering_id` in `voting_results`
   - Index on `voting_mode` in `gatherings`

2. **Batch operations:**
   - Load all units in single query
   - Use joins to minimize round trips

3. **Pagination:**
   - For large gatherings (1000+ units)
   - Paginate results display

**Benchmarks:**

```
BenchmarkComputeResults_100Units     500 ops    2.5 ms/op
BenchmarkComputeResults_1000Units    100 ops    15 ms/op
BenchmarkComputeResults_10000Units   10 ops     180 ms/op
```

---

### Troubleshooting

**Issue: Quorum calculation incorrect**

Check:
- Gathering type matches expected (initial/repeated/remote)
- Voting mode matches expected (by_weight/by_unit)
- Total possible votes calculated correctly
- Using correct threshold percentage

Debug:
```go
log.Printf("Gathering: type=%s, mode=%s", gathering.GatheringType, gathering.VotingMode)
log.Printf("Total: weight=%.2f, count=%d", totalWeight, totalCount)
log.Printf("Required: %.2f (%.1f%%)", required, thresholdPercent)
log.Printf("Achieved: %.2f", achieved)
```

---

**Issue: Results not cached**

Check:
- Gathering status is "closed"
- `ComputeAndStoreResults` called during close
- No errors in computation
- `voting_results` table populated

Debug:
```sql
SELECT * FROM voting_results WHERE gathering_id = 123;
```

---

**Issue: Multi-unit owner votes counted incorrectly**

Check:
- Voting mode: by_weight vs by_unit
- Participant `UnitsInfo` contains all unit IDs
- Participant `UnitsPart` equals sum of unit weights
- Strategy implementation correct

Debug:
```go
log.Printf("Participant: units=%v, weight=%.2f", participant.UnitsInfo, participant.UnitsPart)
log.Printf("Calculated vote: %.2f", strategy.CalculateVoteWeight(participant, units))
```

---

## Glossary

### Terms

**Apartment/Unit:**
A single residential property within the association. Each unit has an owner and a voting weight (part).

**Association:**
A collection of units forming a homeowners or apartment owners association.

**Ballot:**
A submission of votes by a participant on voting matters.

**Delegate:**
A person authorized to vote on behalf of a unit owner.

**Gathering:**
A meeting or voting session for association members. Can be physical or remote.

**Owner:**
The legal owner of one or more units in the association.

**Participant:**
A person who registers to vote in a gathering (owner or delegate).

**Part/Weight:**
The voting power of a unit, typically based on area, value, or ownership percentage.

**Qualified Unit:**
A unit that meets the criteria to participate in a specific gathering.

**Quorum:**
The minimum participation required for voting results to be legally valid.

**Voting Matter:**
A specific issue or proposal to be voted on during a gathering.

---

### Gathering Types

**Initial Gathering:**
First meeting for a purpose, requires 50% quorum.

**Repeated Gathering:**
Follow-up meeting after failed initial gathering, requires 25% quorum.

**Remote Gathering:**
Correspondence/online voting, requires 100% participation.

---

### Voting Modes

**By-Weight Mode:**
Vote counting based on sum of unit weights. Multi-unit owners vote once with combined weight.

**By-Unit Mode:**
Vote counting based on number of units. Each unit gets one vote.

---

### Status Values

**Draft:**
Gathering created but not yet published.

**Published:**
Gathering published and accepting participants.

**Active:**
Gathering in progress, voting is open.

**Closed:**
Gathering finished, votes counted, results available.

---

### Vote Options

**For:**
Vote in favor of the proposal.

**Against:**
Vote against the proposal.

**Abstain:**
Neutral vote, counted for quorum but not for/against.

---

## References

### Related Documents

- **API Documentation:** `/docs/api/gatherings.md`
- **Migration Guide:** `/docs/migration-guide-voting-system.md`
- **Implementation Plan:** `/plans/voting-system-improvement-plan.md`

### Code Files

**Strategy Pattern:**
- `/api/internal/handlers/gathering/services/voting_strategy.go`
- `/api/internal/handlers/gathering/services/by_weight_strategy.go`
- `/api/internal/handlers/gathering/services/by_unit_strategy.go`

**Services:**
- `/api/internal/handlers/gathering/services/voting_results_service.go`
- `/api/internal/handlers/gathering/services/quorum_service.go`

**Database:**
- `/api/sql/schema/00019_add_voting_mode.sql`
- `/api/sql/schema/00020_voting_results.sql`

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 2.0 | 2025-12-07 | Added voting modes, updated quorum calculation |
| 1.0 | 2024-XX-XX | Initial documentation |

---

**End of Documentation**
