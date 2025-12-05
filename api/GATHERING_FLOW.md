# Simplified Gathering & Voting Flow

## Overview
The gathering/voting system has been simplified to reduce complexity while maintaining concurrency control through `unit_slots`.

## Key Changes

### 1. Joint Ownership Support
- **Removed**: Unique constraint on `is_voting` per unit
- **Now**: All active owners of a unit can see it in eligible voters
- **Concurrency**: `unit_slots` prevents double-voting (first owner to vote wins)
- **Flexibility**: Different owners can vote in different gatherings

### 2. Simplified Ballot Submission
- **Before**: Separate steps to add participant, then submit ballot
- **Now**: Single API call that auto-creates participant and submits ballot
- **Route**: `POST /v1/api/associations/{associationId}/gatherings/{gatheringId}/ballot`

### 3. New Eligible Voters Endpoint
- **Route**: `GET /v1/api/associations/{associationId}/gatherings/{gatheringId}/eligible-voters`
- **Returns**: Owners grouped with their available units
- **Shows**: Which units are available vs. already voted

## Complete Flow

### 1. Create Gathering (unchanged)
```
POST /v1/api/associations/{associationId}/gatherings
{
  "title": "Annual General Meeting 2024",
  "gathering_type": "repeated",
  "gathering_date": "2024-12-15T10:00:00Z",
  "location": "Community Hall",
  "qualification_unit_types": ["apartment"],
  "qualification_floors": [1, 2, 3]
}
```

**Backend automatically**:
- Computes qualified units stats
- Creates `unit_slots` for all qualified units (with `participant_id=NULL`)

### 2. Create Voting Matters (unchanged)
```
POST /v1/api/associations/{associationId}/gatherings/{gatheringId}/matters
{
  "order_index": 1,
  "title": "Approve 2025 Budget",
  "matter_type": "budget",
  "voting_config": {
    "type": "yes_no",
    "required_majority": "simple",
    "quorum": 50,
    "allow_abstention": true
  }
}
```

### 3. Get Eligible Voters (new)
```
GET /v1/api/associations/{associationId}/gatherings/{gatheringId}/eligible-voters
```

**Response**:
```json
[
  {
    "owner": {
      "id": 123,
      "name": "John & Jane Doe",
      "identification_number": "1234567890",
      "contact_email": "john@example.com"
    },
    "units": [
      {
        "id": 10,
        "unit_number": "A1",
        "floor": 1,
        "area": 50.0,
        "voting_weight": 0.05,
        "is_available": true
      },
      {
        "id": 11,
        "unit_number": "P1",
        "area": 10.0,
        "voting_weight": 0.01,
        "is_available": true
      }
    ],
    "total_available_weight": 0.06,
    "total_available_area": 60.0,
    "has_available_units": true,
    "available_units_count": 2
  }
]
```

### 4. Submit Ballot (simplified)
```
POST /v1/api/associations/{associationId}/gatherings/{gatheringId}/ballot
{
  "voter_type": "owner",
  "owner_id": 123,
  "unit_ids": [10, 11],
  "ballot_content": {
    "1": {"vote_value": "yes"},
    "2": {"vote_value": "no"},
    "3": {"option_id": "option_a"}
  }
}
```

**For delegate voting**:
```json
{
  "voter_type": "delegate",
  "owner_id": 456,
  "delegating_owner_id": 123,
  "delegation_document_ref": "POA-2024-001",
  "unit_ids": [10],
  "ballot_content": {...}
}
```

**Backend automatically**:
1. Validates gathering is "active"
2. Validates all units are owned by the owner
3. Validates all units are available (not already voted)
4. Creates `gathering_participant` record
5. Assigns `unit_slots` to this participant
6. Creates ballot with hash
7. Updates gathering stats
8. Updates vote tallies

**Response**:
```json
{
  "status": "ballot_submitted",
  "ballot_hash": "a1b2c3d4...",
  "ballot_id": 42,
  "participant_id": 7
}
```

### 5. Get Results
```
GET /v1/api/associations/{associationId}/gatherings/{gatheringId}/results
```

**Response includes**:
- Vote tallies per matter (yes/no/abstain counts with weights)
- Pass/fail verdict per matter
- Participation stats:
  - Qualified units count & weight
  - Participating units count & weight (who voted)
  - Participation rate
  - Voting completion rate

## Joint Ownership Scenarios

### Scenario 1: Unit owned by husband and wife
**Database**:
```
ownerships:
  unit_id=1, owner_id=123 (husband), is_active=TRUE, is_voting=TRUE
  unit_id=1, owner_id=124 (wife), is_active=TRUE, is_voting=TRUE

unit_slots:
  unit_id=1, participant_id=NULL  ← Available
```

**Eligible voters**:
- Both husband (123) and wife (124) see Unit 1 as available
- Both have `total_available_weight = 0.05` (unit's full weight)

**Husband votes**:
- Ballot submission succeeds
- `unit_slots.participant_id` set to husband's participant ID
- Unit 1 now shows `is_available=false` for both owners

**Wife tries to vote**:
- Ballot submission fails: "Unit 1 is not available (already assigned)"

### Scenario 2: Multiple gatherings over time
**Gathering 1 (January)**:
- Husband attends and votes with Unit 1
- Wife cannot vote with Unit 1 (already assigned)

**Gathering 2 (July)**:
- New gathering created → new `unit_slots` created (all NULL)
- Both husband and wife see Unit 1 as available again
- Wife attends and votes with Unit 1
- Husband cannot vote with Unit 1 (already assigned)

## Database Tables

### `unit_slots` (concurrency control)
```sql
CREATE TABLE unit_slots (
    id INTEGER PRIMARY KEY,
    gathering_id INTEGER NOT NULL,
    unit_id INTEGER NOT NULL,
    participant_id INTEGER NULL,  -- NULL = available, set = voted
    UNIQUE (gathering_id, unit_id)
);
```

**Purpose**: Ensures each unit can only vote once per gathering

### `gathering_participants` (tracking)
```sql
CREATE TABLE gathering_participants (
    id INTEGER PRIMARY KEY,
    gathering_id INTEGER NOT NULL,
    participant_type TEXT NOT NULL,  -- 'owner' or 'delegate'
    owner_id INTEGER,
    delegating_owner_id INTEGER,
    units_info TEXT NOT NULL,  -- JSON array of unit IDs
    units_area NUMERIC NOT NULL,
    units_part NUMERIC NOT NULL
);
```

**Purpose**: Track who voted (automatically created from ballots)

### `voting_ballots`
```sql
CREATE TABLE voting_ballots (
    id INTEGER PRIMARY KEY,
    gathering_id INTEGER NOT NULL,
    participant_id INTEGER NOT NULL,
    ballot_content TEXT NOT NULL,  -- JSON with all votes
    ballot_hash TEXT NOT NULL,
    UNIQUE (gathering_id, participant_id)
);
```

**Purpose**: Store actual votes (one ballot per participant)

## Migration Applied

**File**: `sql/schema/00018_remove_voting_owner_constraint.sql`

**Changes**:
- Dropped `idx_voting_owner_per_unit` unique index
- Set `is_voting = TRUE` for all active ownerships
- Allows multiple owners per unit to be voting-eligible

## Benefits

1. **Simpler API**: One call instead of multiple steps
2. **Flexible joint ownership**: Different owners can vote in different gatherings
3. **Race condition protection**: `unit_slots` prevents double-voting
4. **Better UX**: Frontend shows exactly who can vote with which units
5. **Automatic tracking**: Participants auto-created from ballots
6. **Comprehensive stats**: All participation metrics calculated automatically

## Deprecated Endpoints

The following endpoints still work but are **optional**:

- `POST /gatherings/{id}/participants` - No longer needed (auto-created on ballot submission)
- `POST /gatherings/{id}/participants/{id}/checkin` - Optional (can track attendance separately)

The old ballot submission route that required `participant_id` in the path has been removed.
