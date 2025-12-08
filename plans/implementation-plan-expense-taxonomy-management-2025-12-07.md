# Implementation Plan: Expense Taxonomy Management

**Project:** APC (Apartment Owners Association) Management Application
**Feature:** Expense Taxonomy Management System
**Version:** 1.0
**Date:** 2025-12-07
**Status:** Ready for Development
**PRD Reference:** `/prds/prd-expense-taxonomy-management-2025-12-07.md`

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Technical Architecture Breakdown](#technical-architecture-breakdown)
3. [Development Phases and Milestones](#development-phases-and-milestones)
4. [Resource Allocation and Estimates](#resource-allocation-and-estimates)
5. [Risk Assessment and Mitigation](#risk-assessment-and-mitigation)
6. [Dependencies and Blockers](#dependencies-and-blockers)
7. [Testing Strategy](#testing-strategy)
8. [Detailed Implementation Checklist](#detailed-implementation-checklist)
9. [Migration and Deployment](#migration-and-deployment)
10. [Success Metrics](#success-metrics)

---

## Executive Summary

### Project Overview
This plan details the technical implementation of a comprehensive Expense Taxonomy Management system that enables association administrators to manage their three-level expense categorization hierarchy (type → family → name) through an intuitive admin interface. Currently, expense categories are hardcoded values; this feature will provide full CRUD capabilities while preserving data integrity through soft deletion.

### Strategic Value
- **Flexibility**: Associations customize expense categories to match their accounting practices
- **Autonomy**: Reduces dependency on technical team for category management
- **Data Integrity**: Soft deletion ensures historical expense data remains intact
- **Scalability**: New associations start with sensible defaults and adjust as needed

### Timeline and Team
- **Total Duration**: 4 weeks (with 1 week buffer = 5 weeks total)
- **Team Composition**:
  - 1 Backend Developer (Go/sqlc expertise)
  - 1 Frontend Developer (Vue 3/TypeScript expertise)
  - 1 QA Engineer (part-time, weeks 3-5)
  - 1 Tech Lead (oversight and code review)
- **Target Completion**: End of Week 5
- **MVP Delivery**: End of Week 3

### Key Success Criteria
- ✅ 100% of existing expenses maintain valid category references post-migration
- ✅ Association admins can create/modify categories without technical assistance
- ✅ Zero data loss from category deletions (soft delete only)
- ✅ Page load time under 2 seconds for category management interface
- ✅ 80%+ test coverage for critical paths

### High-Level Risks
1. **Data Migration Complexity** (High Impact, Medium Probability) - Mitigated by extensive testing and rollback plans
2. **Performance with Large Datasets** (Medium Impact, Medium Probability) - Mitigated by database indexes and pagination
3. **User Confusion About Soft Delete** (Medium Impact, High Probability) - Mitigated by clear UX and documentation

---

## Technical Architecture Breakdown

### System Overview

The Expense Taxonomy Management feature integrates into the existing APC application architecture as a new administrative module following established patterns from the accounts and expenses pages.

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend (Vue 3 + TypeScript)           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐   │
│  │ /categories  │  │  /expenses   │  │    /accounts       │   │
│  │   index.vue  │  │  index.vue   │  │    index.vue       │   │
│  └──────────────┘  └──────────────┘  └────────────────────┘   │
│         │                 │                      │              │
│         ├─────────────────┴──────────────────────┘              │
│         │                                                        │
│  ┌──────▼───────────────────────────────────────────────────┐  │
│  │         Association Selector (existing)                   │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌───────────────────┐  ┌──────────────────┐  ┌────────────┐  │
│  │ CategoriesList    │  │  CategoryForm    │  │ CategoryTree│  │
│  │ (new component)   │  │  (new component) │  │ (new comp.) │  │
│  └───────────────────┘  └──────────────────┘  └────────────┘  │
│         │                        │                    │         │
│         └────────────────────────┴────────────────────┘         │
│                                 │                               │
│  ┌──────────────────────────────▼────────────────────────────┐ │
│  │           categoryApi (services/api.ts)                    │ │
│  │  - getCategories(), createCategory(), updateCategory(),   │ │
│  │  - deactivateCategory(), reactivateCategory(),            │ │
│  │  - getCategoryUsage(), bulkDeactivate()                   │ │
│  └───────────────────────────────────────────────────────────┘ │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTP/REST API
┌──────────────────────────▼──────────────────────────────────────┐
│                    Backend (Go + sqlc)                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │           HTTP Router (main.go)                           │ │
│  │  - Middleware: Auth, Association Resource Validation      │ │
│  └───────────────────────────────────────────────────────────┘ │
│         │                                                        │
│  ┌──────▼──────────────────────────────────────────────────┐   │
│  │       Category Handlers (handlers/categories.go)        │   │
│  │  Existing:                      New:                    │   │
│  │  - HandleGetActiveCategories    - HandleGetAllCategories│   │
│  │  - HandleGetCategory            - HandleUpdateCategory  │   │
│  │  - HandleCreateCategory         - HandleReactivateCategory│ │
│  │  - HandleDeactivateCategory     - HandleGetCategoryUsage│   │
│  │                                 - HandleBulkDeactivate   │   │
│  │                                 - HandleBulkReactivate   │   │
│  └─────────────────────────────────────────────────────────┘   │
│         │                                                        │
│  ┌──────▼──────────────────────────────────────────────────┐   │
│  │      Database Layer (database/categories.sql.go)        │   │
│  │  - Type-safe query methods generated by sqlc            │   │
│  └─────────────────────────────────────────────────────────┘   │
└──────────────────────────┬──────────────────────────────────────┘
                           │ SQL
┌──────────────────────────▼──────────────────────────────────────┐
│                     SQLite Database                              │
├─────────────────────────────────────────────────────────────────┤
│  ┌───────────────┐  ┌─────────────┐  ┌──────────────────────┐  │
│  │  categories   │  │  expenses   │  │   associations       │  │
│  │  - id         │  │  - id       │  │   - id               │  │
│  │  - type       │  │  - amount   │  │   - name             │  │
│  │  - family     │  │  - category ├─►│                      │  │
│  │  - name       │  │    _id (FK) │  │                      │  │
│  │  - is_deleted │  └─────────────┘  └──────────────────────┘  │
│  │  - assoc_id───┼─────────────────►                           │
│  └───────────────┘                                              │
│                                                                  │
│  New Indexes:                                                   │
│  - idx_categories_association_deleted (association_id, is_deleted)│
│  - idx_categories_lookup (association_id, type, family, name)  │
└─────────────────────────────────────────────────────────────────┘
```

### Database Schema Changes

#### Existing Schema (No Changes)
The current `categories` table structure remains unchanged:

```sql
CREATE TABLE categories (
    id             INTEGER PRIMARY KEY,
    type           TEXT    NOT NULL,
    family         TEXT    NOT NULL,
    name           TEXT    NOT NULL,
    is_deleted     BOOLEAN NOT NULL DEFAULT FALSE,
    association_id INTEGER NOT NULL REFERENCES associations (id),
    created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### New Indexes Required

**Migration File**: `api/sql/schema/00021_add_category_indexes.sql`

```sql
-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - adding category performance indexes';

-- Index for filtering active/inactive categories per association
CREATE INDEX IF NOT EXISTS idx_categories_association_deleted
ON categories(association_id, is_deleted);

-- Composite index for uniqueness checks and efficient lookups
CREATE INDEX IF NOT EXISTS idx_categories_lookup
ON categories(association_id, type, family, name)
WHERE is_deleted = FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - removing category indexes';

DROP INDEX IF EXISTS idx_categories_association_deleted;
DROP INDEX IF EXISTS idx_categories_lookup;

-- +goose StatementEnd
```

**Rationale**:
- `idx_categories_association_deleted`: Optimizes queries filtering by association and deletion status
- `idx_categories_lookup`: Partial index supports uniqueness validation and prevents duplicate active categories

### API Endpoints

#### Existing Endpoints (Maintain)
| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/v1/api/associations/{associationId}/categories` | HandleGetActiveCategories | Get active categories only |
| GET | `/v1/api/associations/{associationId}/categories/{categoryId}` | HandleGetCategory | Get single category |
| POST | `/v1/api/associations/{associationId}/categories` | HandleCreateCategory | Create new category |
| PUT | `/v1/api/associations/{associationId}/categories/{categoryId}/deactivate` | HandleDeactivateCategory | Soft delete category |

#### New Endpoints (Implement)
| Method | Endpoint | Handler | Description | Priority |
|--------|----------|---------|-------------|----------|
| GET | `/v1/api/associations/{associationId}/categories/all?include_inactive=true` | HandleGetAllCategories | Get all categories with optional inactive filter | Must Have |
| PUT | `/v1/api/associations/{associationId}/categories/{categoryId}` | HandleUpdateCategory | Update category details | Must Have |
| PUT | `/v1/api/associations/{associationId}/categories/{categoryId}/reactivate` | HandleReactivateCategory | Reactivate soft-deleted category | Should Have |
| GET | `/v1/api/associations/{associationId}/categories/{categoryId}/usage` | HandleGetCategoryUsage | Get usage statistics | Should Have |
| POST | `/v1/api/associations/{associationId}/categories/bulk-deactivate` | HandleBulkDeactivateCategories | Bulk soft delete | Could Have |
| POST | `/v1/api/associations/{associationId}/categories/bulk-reactivate` | HandleBulkReactivateCategories | Bulk reactivate | Could Have |

**Authorization**: All endpoints require:
1. Valid JWT token (via `MiddlewareAuth`)
2. Association membership validation (via `MiddlewareAssociationResource`)

### New SQL Queries

**File**: `api/sql/queries/categories.sql` (extend existing)

```sql
-- name: GetAllCategories :many
SELECT *
FROM categories
WHERE association_id = ?
  AND (? = FALSE OR is_deleted = FALSE)  -- include_inactive parameter
ORDER BY is_deleted ASC, type, family, name;

-- name: GetAllCategoriesWithFilter :many
SELECT *
FROM categories
WHERE association_id = ?
  AND (? = TRUE OR is_deleted = FALSE)  -- include_inactive_param
ORDER BY is_deleted ASC, type, family, name;

-- name: UpdateCategory :one
UPDATE categories
SET type = ?,
    family = ?,
    name = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND association_id = ?
RETURNING *;

-- name: ReactivateCategory :exec
UPDATE categories
SET is_deleted = FALSE,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND association_id = ?;

-- name: GetCategoryUsageCount :one
SELECT COUNT(*) as usage_count
FROM expenses
WHERE category_id = ?;

-- name: GetCategoryUsageDetails :many
SELECT id, description, amount, date, created_at
FROM expenses
WHERE category_id = ?
ORDER BY date DESC
LIMIT 10;

-- name: BulkDeactivateCategories :exec
UPDATE categories
SET is_deleted = TRUE,
    updated_at = CURRENT_TIMESTAMP
WHERE id IN (sqlc.slice('category_ids'))
  AND association_id = ?;

-- name: BulkReactivateCategories :exec
UPDATE categories
SET is_deleted = FALSE,
    updated_at = CURRENT_TIMESTAMP
WHERE id IN (sqlc.slice('category_ids'))
  AND association_id = ?;

-- name: CheckCategoryUniqueness :one
SELECT COUNT(*) as count
FROM categories
WHERE association_id = ?
  AND type = ?
  AND family = ?
  AND name = ?
  AND is_deleted = FALSE
  AND (? = 0 OR id != ?);  -- exclude_id parameter for updates
```

### Frontend Architecture

#### New Files

```
ui/src/
├── pages/
│   └── categories/
│       └── index.vue                    # Main category management page
├── components/
│   ├── CategoriesList.vue              # Table/List view component
│   ├── CategoryForm.vue                # Modal form for create/edit
│   └── CategoryTree.vue                # Tree visualization (optional)
```

#### Component Hierarchy

**`/pages/categories/index.vue`** (Main Page)
- Pattern: Follow `/pages/accounts/index.vue`
- State Management: Vue 3 Composition API with `ref`, `reactive`, `computed`
- Responsibilities:
  - Association selector integration
  - Modal state management (create/edit)
  - Coordinate between list and form components
  - Handle route guards (auth required)

**`CategoriesList.vue`** (List Component)
- Pattern: Follow `AccountsList.vue` structure
- UI: Naive UI `NDataTable` or `NTree`
- Features:
  - Display categories in hierarchical or flat view
  - Search and filter controls
  - Inline actions: Edit, Deactivate/Reactivate
  - Usage count badges
  - Loading states and empty states
  - Pagination (100 items per page)

**`CategoryForm.vue`** (Form Modal)
- Pattern: Follow `AccountForm.vue` structure
- UI: Naive UI `NModal` + `NForm`
- Features:
  - Three input fields: Type, Family, Name
  - Validation (required fields, max length, uniqueness)
  - Support create and edit modes
  - i18n integration for labels
  - Error handling and display

**`CategoryTree.vue`** (Optional Tree View)
- UI: Naive UI `NTree` component
- Features:
  - Collapsible/expandable nodes
  - Type → Family → Name hierarchy visualization
  - Inline action buttons per node
  - Status indicators (active/inactive)

#### API Service Extensions

**File**: `ui/src/services/api.ts` (extend existing)

```typescript
// Add to existing categoryApi or create if doesn't exist
export const categoryApi = {
  // Existing (if present)
  getActiveCategories: (associationId: number) =>
    axios.get(`/v1/api/associations/${associationId}/categories`),

  getCategory: (associationId: number, categoryId: number) =>
    axios.get(`/v1/api/associations/${associationId}/categories/${categoryId}`),

  createCategory: (associationId: number, data: CategoryCreateRequest) =>
    axios.post(`/v1/api/associations/${associationId}/categories`, data),

  // New methods
  getAllCategories: (associationId: number, includeInactive: boolean = false) =>
    axios.get(`/v1/api/associations/${associationId}/categories/all`, {
      params: { include_inactive: includeInactive }
    }),

  updateCategory: (associationId: number, categoryId: number, data: CategoryUpdateRequest) =>
    axios.put(`/v1/api/associations/${associationId}/categories/${categoryId}`, data),

  deactivateCategory: (associationId: number, categoryId: number) =>
    axios.put(`/v1/api/associations/${associationId}/categories/${categoryId}/deactivate`),

  reactivateCategory: (associationId: number, categoryId: number) =>
    axios.put(`/v1/api/associations/${associationId}/categories/${categoryId}/reactivate`),

  getCategoryUsage: (associationId: number, categoryId: number) =>
    axios.get(`/v1/api/associations/${associationId}/categories/${categoryId}/usage`),

  bulkDeactivate: (associationId: number, categoryIds: number[]) =>
    axios.post(`/v1/api/associations/${associationId}/categories/bulk-deactivate`, {
      ids: categoryIds
    }),

  bulkReactivate: (associationId: number, categoryIds: number[]) =>
    axios.post(`/v1/api/associations/${associationId}/categories/bulk-reactivate`, {
      ids: categoryIds
    })
}
```

#### TypeScript Type Definitions

**File**: `ui/src/types/api.ts` (extend existing)

```typescript
// Extend existing types
export interface Category {
  id: number
  type: string
  family: string
  name: string
  is_deleted: boolean
  association_id: number
  created_at: string
  updated_at: string
}

export interface CategoryCreateRequest {
  type: string
  family: string
  name: string
}

export interface CategoryUpdateRequest {
  type: string
  family: string
  name: string
}

export interface CategoryUsageResponse {
  category_id: number
  usage_count: number
  last_used_at: string | null
  recent_expenses: Array<{
    id: number
    description: string
    amount: number
    date: string
  }>
}

export interface CategoryTreeNode {
  key: string
  label: string
  children?: CategoryTreeNode[]
  category?: Category
  isType?: boolean
  isFamily?: boolean
  isName?: boolean
}
```

### i18n Integration

**Files**: `ui/src/i18n/locales/en.json` and `ro.json`

```json
{
  "categories": {
    "title": "Expense Categories",
    "createNew": "Create New Category",
    "editCategory": "Edit Category",
    "type": "Type",
    "family": "Family",
    "name": "Name",
    "status": "Status",
    "active": "Active",
    "inactive": "Inactive",
    "usageCount": "Usage Count",
    "lastUsed": "Last Used",
    "actions": "Actions",
    "noCategories": "No categories found",
    "searchPlaceholder": "Search categories...",
    "filterByStatus": "Filter by Status",
    "includeInactive": "Include Inactive",
    "confirmDeactivate": "Are you sure you want to deactivate this category?",
    "confirmBulkDeactivate": "Are you sure you want to deactivate {count} categories?",
    "categoryInUse": "This category is used by {count} expenses",
    "categoryCreated": "Category created successfully",
    "categoryUpdated": "Category updated successfully",
    "categoryDeactivated": "Category deactivated successfully",
    "categoryReactivated": "Category reactivated successfully",
    "duplicateCategory": "A category with these values already exists",
    "validation": {
      "typeRequired": "Type is required",
      "familyRequired": "Family is required",
      "nameRequired": "Name is required",
      "maxLength": "Maximum {max} characters allowed"
    },
    "types": {
      "Budgeted": "Budgeted",
      "Maintenance": "Maintenance",
      "Improvement": "Improvement"
    },
    "families": {
      "personnel": "Personnel Expenses",
      "security": "Security and Safety",
      "utilities": "Utilities"
    },
    "names": {
      "salary_janitor": "Salary Janitor",
      "internet": "Internet Services"
    }
  }
}
```

**Romanian translations** (ro.json):

```json
{
  "categories": {
    "title": "Categorii de Cheltuieli",
    "createNew": "Crează Categorie Nouă",
    "editCategory": "Editează Categoria",
    "type": "Tip",
    "family": "Familie",
    "name": "Nume",
    "status": "Stare",
    "active": "Activă",
    "inactive": "Inactivă",
    "usageCount": "Număr Utilizări",
    "lastUsed": "Ultima Utilizare",
    "actions": "Acțiuni",
    "noCategories": "Nu există categorii",
    "searchPlaceholder": "Caută categorii...",
    "filterByStatus": "Filtrează după Stare",
    "includeInactive": "Include Inactive",
    "confirmDeactivate": "Sigur doriți să dezactivați această categorie?",
    "confirmBulkDeactivate": "Sigur doriți să dezactivați {count} categorii?",
    "categoryInUse": "Această categorie este folosită de {count} cheltuieli",
    "categoryCreated": "Categorie creată cu succes",
    "categoryUpdated": "Categorie actualizată cu succes",
    "categoryDeactivated": "Categorie dezactivată cu succes",
    "categoryReactivated": "Categorie reactivată cu succes",
    "duplicateCategory": "O categorie cu aceste valori există deja",
    "types": {
      "Budgeted": "Bugetate",
      "Maintenance": "Întreținere",
      "Improvement": "Îmbunătățiri"
    },
    "families": {
      "personnel": "Cheltuieli de Personal",
      "security": "Securitate",
      "utilities": "Utilități"
    },
    "names": {
      "salary_janitor": "Salariu Maturator",
      "internet": "Servicii Internet"
    }
  }
}
```

### Security Considerations

1. **Authentication & Authorization**
   - All endpoints protected by JWT middleware
   - Association membership validated via `MiddlewareAssociationResource`
   - No cross-association data leakage

2. **Input Validation**
   - SQL injection prevention via sqlc prepared statements
   - XSS protection through Vue's automatic escaping
   - Max length validation: 100 characters per field
   - Allowed characters: alphanumeric, spaces, underscore, hyphen

3. **Business Logic Validation**
   - Uniqueness check before create/update
   - Prevent deactivation of last active category
   - Association ID immutability (cannot change category ownership)

4. **Audit Trail**
   - `created_at` and `updated_at` timestamps automatically maintained
   - Soft delete preserves full history

### Performance Considerations

1. **Database Optimization**
   - Composite indexes for frequent queries
   - Partial index for uniqueness checks (active categories only)
   - LIMIT clauses on usage detail queries

2. **Frontend Optimization**
   - Pagination: 100 categories per page
   - Lazy loading of usage statistics (on-demand)
   - Debounced search input (300ms delay)
   - Virtual scrolling for tree view if needed

3. **Caching Strategy**
   - Browser caching for category lists (Cache-Control headers)
   - Vue component-level caching for static data
   - No server-side caching initially (future enhancement)

4. **Performance Targets**
   - Page load: < 2 seconds
   - Search response: < 500ms
   - API response: < 1 second
   - Bulk operations: < 3 seconds for 100 categories

---

## Development Phases and Milestones

### Overview

The implementation is divided into 4 main development phases plus a deployment phase, with clear deliverables and success criteria for each. Total estimated duration is 4 weeks of active development plus 1 week buffer.

```
Week 1: Phase 1 (Database & Backend Foundation)
Week 2: Phase 2 (Core Frontend UI)
Week 3: Phase 3 (Advanced Features & Integration)
Week 4: Phase 4 (Testing & Polish)
Week 5: Phase 5 (Migration & Deployment)
```

---

### Phase 1: Database & Backend Foundation
**Duration**: 5 days (Week 1)
**Team**: Backend Developer (primary), Tech Lead (review)

#### Objectives
- Set up database infrastructure (indexes)
- Implement all backend API endpoints
- Achieve 80%+ unit test coverage for backend logic
- Prepare sqlc-generated code for frontend consumption

#### Deliverables

1. **Database Migration** (`00021_add_category_indexes.sql`)
   - Composite indexes for performance
   - Tested on production database copy
   - Migration rollback script prepared

2. **SQL Queries** (`api/sql/queries/categories.sql`)
   - 6 new sqlc queries (GetAllCategories, UpdateCategory, ReactivateCategory, GetCategoryUsageCount, GetCategoryUsageDetails, BulkDeactivate, BulkReactivate, CheckCategoryUniqueness)
   - All queries parameterized and association-scoped
   - sqlc regeneration successful

3. **Backend Handlers** (`api/internal/handlers/categories.go`)
   - HandleGetAllCategories (with include_inactive query param)
   - HandleUpdateCategory (with uniqueness validation)
   - HandleReactivateCategory
   - HandleGetCategoryUsage
   - HandleBulkDeactivateCategories
   - HandleBulkReactivateCategories
   - Error handling for all edge cases

4. **Router Updates** (`api/main.go`)
   - 6 new routes registered
   - Middleware applied (Auth + AssociationResource)
   - Route documentation comments

5. **Unit Tests** (`api/internal/handlers/categories_test.go`)
   - Test coverage > 80%
   - Happy path and error scenarios
   - Association scoping validation
   - Uniqueness constraint tests

#### Success Criteria
- ✅ All API endpoints functional via Postman/curl
- ✅ Unit tests pass with >80% coverage
- ✅ sqlc generates code without errors
- ✅ No breaking changes to existing expense endpoints
- ✅ Database migration runs successfully on test environment

#### Dependencies
- None (can start immediately)

---

### Phase 2: Core Frontend UI
**Duration**: 5 days (Week 2)
**Team**: Frontend Developer (primary), Tech Lead (review)
**Prerequisites**: Phase 1 complete (API endpoints functional)

#### Objectives
- Build category management page UI
- Implement basic CRUD operations via UI
- Integrate with existing authentication and association selector
- Establish i18n translations

#### Deliverables

1. **Categories Page** (`ui/src/pages/categories/index.vue`)
   - Page structure following `/accounts` pattern
   - Association selector integration
   - Modal state management
   - Create and Edit workflows

2. **Categories List Component** (`ui/src/components/CategoriesList.vue`)
   - NDataTable displaying categories
   - Columns: Type, Family, Name, Status, Usage Count, Actions
   - Active/Inactive visual distinction
   - Edit and Deactivate action buttons
   - Empty state with helpful message

3. **Category Form Component** (`ui/src/components/CategoryForm.vue`)
   - NModal with NForm for create/edit
   - Three input fields: Type, Family, Name
   - Form validation (required, max length)
   - Error message display
   - Save and Cancel actions

4. **API Service Layer** (`ui/src/services/api.ts`)
   - categoryApi methods implemented
   - TypeScript types defined (`ui/src/types/api.ts`)
   - Axios interceptors configured

5. **i18n Translations** (`ui/src/i18n/locales/*.json`)
   - English translations complete
   - Romanian translations complete
   - All UI strings internationalized

6. **Router Integration**
   - `/categories` route added
   - Auth guard applied
   - Navigation menu link added (if applicable)

#### Success Criteria
- ✅ Can create new category via UI and see it in expense forms
- ✅ Can edit existing category and changes reflect immediately
- ✅ Can deactivate category and it disappears from expense dropdowns
- ✅ Association selector filters categories correctly
- ✅ No console errors or warnings
- ✅ Responsive design works on desktop (1024px+)

#### Dependencies
- Phase 1 complete (backend APIs functional)

---

### Phase 3: Advanced Features & Integration
**Duration**: 5 days (Week 3)
**Team**: Frontend Developer (primary), Backend Developer (support), QA Engineer (starts testing)

#### Objectives
- Implement advanced UI features (search, filter, bulk ops, usage stats)
- Add reactivation functionality
- Enhance UX with loading states and confirmations
- Begin integration testing

#### Deliverables

1. **Search and Filter**
   - Search box with debounced input (300ms)
   - Status filter (All / Active / Inactive)
   - Type filter dropdown
   - Real-time result updates

2. **Category Tree View** (`ui/src/components/CategoryTree.vue`) - Optional
   - NTree component integration
   - Hierarchical Type → Family → Name visualization
   - Collapsible nodes
   - Inline action buttons

3. **Bulk Operations**
   - Checkbox selection in table/tree
   - Bulk Deactivate button with confirmation dialog
   - Bulk Reactivate button with confirmation dialog
   - Display count of selected items

4. **Usage Statistics**
   - Display usage count next to each category
   - Show warning when deactivating category in use
   - "Last used" date display
   - Optional: Click to view expenses using this category

5. **Reactivation Flow**
   - Filter to show only inactive categories
   - Reactivate button for inactive items
   - Confirmation before reactivation
   - Success message on completion

6. **UX Enhancements**
   - Loading spinners during API calls
   - Skeleton screens for initial load
   - Empty states for search results
   - Toast notifications for success/error
   - Confirmation dialogs for destructive actions

7. **Integration Testing**
   - End-to-end test scenarios (Playwright/Cypress)
   - Create → Edit → Deactivate → Reactivate workflow
   - Verify category appears in expense form
   - Verify soft delete preserves expense references

#### Success Criteria
- ✅ All "Must Have" and "Should Have" user stories completed
- ✅ Search returns results within 500ms
- ✅ Bulk operations work for up to 100 categories
- ✅ Integration tests pass for critical workflows
- ✅ No regressions in existing expense management
- ✅ Accessibility checklist verified (keyboard nav, ARIA labels)

#### Dependencies
- Phase 2 complete (core UI functional)

---

### Phase 4: Testing, Polish & Documentation
**Duration**: 5 days (Week 4)
**Team**: QA Engineer (primary), Frontend/Backend Developers (bug fixes), Tech Lead (review)

#### Objectives
- Comprehensive testing across all scenarios
- Performance optimization
- Security audit
- User documentation
- Bug fixes and refinements

#### Deliverables

1. **Testing**
   - Unit test coverage report (target: 80%+ backend, 70%+ frontend)
   - Integration test suite (critical paths)
   - E2E test scenarios (happy path + edge cases)
   - Performance testing (1000+ categories dataset)
   - Security testing (SQL injection, XSS, auth bypass attempts)
   - Browser compatibility testing (Chrome, Firefox, Safari, Edge)

2. **Performance Optimization**
   - Identify and fix slow queries
   - Optimize frontend bundle size
   - Implement pagination if needed
   - Add loading indicators for slow operations

3. **Bug Fixes**
   - Address all Critical and High severity bugs
   - Triage Medium/Low bugs for post-launch
   - Regression testing after fixes

4. **Documentation**
   - User guide for association admins (with screenshots)
   - API documentation (endpoint specs, examples)
   - Developer README for category management module
   - Deployment runbook

5. **Code Review & Refactoring**
   - Tech lead review of all code
   - Refactor based on feedback
   - Code cleanup (remove console.logs, commented code)
   - Consistent formatting and linting

6. **User Acceptance Testing Prep**
   - UAT test plan document
   - Test data seeded in staging environment
   - UAT guide for stakeholders

#### Success Criteria
- ✅ All critical bugs fixed
- ✅ Test coverage targets met
- ✅ Performance metrics achieved (page load <2s, search <500ms)
- ✅ Security audit passed (no High/Critical vulnerabilities)
- ✅ Documentation complete and reviewed
- ✅ UAT readiness approved by Product Manager

#### Dependencies
- Phase 3 complete (all features implemented)

---

### Phase 5: Migration, Deployment & Rollout
**Duration**: 5 days (Week 5)
**Team**: Backend Developer (migration scripts), DevOps/Tech Lead (deployment), QA Engineer (validation)

#### Objectives
- Migrate existing categories data
- Deploy to production with zero downtime
- Monitor system health
- Provide post-launch support

#### Deliverables

1. **Data Migration Script**
   - Backfill script for existing categories
   - Dry-run mode for validation
   - Rollback script prepared
   - Migration tested on production copy

2. **Default Category Seeding**
   - Default Romanian categories SQL file
   - Seed script for new associations
   - Integration with association creation workflow

3. **Deployment**
   - Staging deployment and smoke testing
   - Production deployment during maintenance window
   - Database migration executed
   - Post-deployment verification

4. **Monitoring & Alerts**
   - Application logs reviewed
   - Error rate monitoring (target: <1%)
   - Performance metrics tracking
   - User feedback collection mechanism

5. **Rollback Plan**
   - Rollback procedure documented
   - Database restoration tested
   - Feature flag to disable new UI if needed

6. **Post-Launch Support**
   - On-call support for first 48 hours
   - Bug hotfix process established
   - User feedback triage

#### Success Criteria
- ✅ Zero downtime deployment
- ✅ 100% of existing expenses retain valid category references
- ✅ Migration completes without errors
- ✅ All associations receive default categories
- ✅ Monitoring alerts configured and tested
- ✅ No critical bugs reported in first 48 hours
- ✅ User adoption metrics tracking begins

#### Dependencies
- Phase 4 complete (all testing passed, UAT approved)

---

## Resource Allocation and Estimates

### Team Structure

| Role | Allocation | Primary Responsibilities | Weeks |
|------|------------|-------------------------|-------|
| **Backend Developer** | Full-time | Database, API handlers, SQL queries, unit tests, migration | 1-5 |
| **Frontend Developer** | Full-time | Vue components, TypeScript, API integration, UI tests | 2-5 |
| **QA Engineer** | Part-time (50%) | Test planning, integration tests, E2E tests, UAT support | 3-5 |
| **Tech Lead** | Part-time (25%) | Code review, architecture decisions, risk mitigation, deployment | 1-5 |

**Total Effort**: ~15 person-weeks (3.75 weeks with 4-person team)

### Effort Distribution by Phase

| Phase | Backend | Frontend | QA | Tech Lead | Total |
|-------|---------|----------|-----|-----------|-------|
| Phase 1: Database & Backend | 5 days | - | - | 1 day | 6 days |
| Phase 2: Core Frontend | 1 day | 5 days | - | 1 day | 7 days |
| Phase 3: Advanced Features | 1 day | 5 days | 2 days | 1 day | 9 days |
| Phase 4: Testing & Polish | 2 days | 2 days | 5 days | 2 days | 11 days |
| Phase 5: Deployment | 3 days | 1 day | 2 days | 3 days | 9 days |
| **Total** | **12 days** | **13 days** | **9 days** | **8 days** | **42 days** |

### Timeline with Buffer

```
Week 1: [████████████████████] Phase 1 Complete
Week 2: [████████████████████] Phase 2 Complete
Week 3: [████████████████████] Phase 3 Complete
Week 4: [████████████████████] Phase 4 Complete
Week 5: [████████████████████] Phase 5 Complete + Buffer
```

**Total Timeline**: 5 weeks (4 weeks active + 1 week buffer = 25% contingency)

### Critical Path

The critical path determines the minimum project duration:

```
Phase 1 (Backend) → Phase 2 (Frontend) → Phase 3 (Features) → Phase 4 (Testing) → Phase 5 (Deploy)
```

**Critical Path Duration**: 4 weeks (assuming no blockers)

**Parallel Work Opportunities**:
- Week 3: Frontend and Backend can work on separate features simultaneously
- Week 4: Bug fixes and documentation can be parallelized
- Week 5: Migration testing can run in parallel with deployment prep

### Required Skills

| Skill | Importance | Team Member | Proficiency Required |
|-------|------------|-------------|---------------------|
| Go Programming | Critical | Backend Dev | Expert |
| sqlc & SQLite | Critical | Backend Dev | Advanced |
| Vue 3 Composition API | Critical | Frontend Dev | Expert |
| TypeScript | Critical | Frontend Dev | Advanced |
| Naive UI Library | High | Frontend Dev | Intermediate |
| REST API Design | High | Backend Dev | Advanced |
| vue-i18n | Medium | Frontend Dev | Intermediate |
| Testing (Jest/Vitest) | High | Frontend/QA | Intermediate |
| E2E Testing (Playwright) | Medium | QA | Intermediate |
| Database Migration (goose) | Medium | Backend Dev | Intermediate |

### Assumptions

1. Team members are already familiar with the existing codebase
2. Development environment is set up and functional
3. No major external blockers (e.g., infrastructure issues)
4. Code review turnaround time: 1 business day
5. QA can start testing as features are completed (shift-left testing)

---

## Risk Assessment and Mitigation

### Risk Matrix

| Risk # | Risk Description | Impact | Probability | Severity | Mitigation Strategy |
|--------|------------------|--------|-------------|----------|---------------------|
| R1 | Data Migration Failure | High | Medium | **CRITICAL** | Extensive testing, rollback plan, dry-run validation |
| R2 | Performance Degradation | Medium | Medium | **HIGH** | Database indexes, pagination, load testing |
| R3 | User Confusion (Soft Delete) | Medium | High | **HIGH** | Clear UX, tooltips, documentation, training |
| R4 | Existing Expense Forms Break | High | Low | **HIGH** | Integration tests, no changes to CategorySelector |
| R5 | Translation Key Management | Low | High | **MEDIUM** | Auto-generate keys, fallback to raw values |
| R6 | Scope Creep | Medium | High | **MEDIUM** | PRD adherence, change request process |
| R7 | Cross-Browser Compatibility | Low | Low | **LOW** | Use Naive UI components, browser testing |
| R8 | Accidental Bulk Deletion | High | Low | **HIGH** | Confirmation dialogs, prevent last category deletion |

---

### Detailed Risk Analysis

#### R1: Data Migration Failure
**Description**: Migration script fails to preserve existing expense-category relationships, causing data loss or broken reports.

**Impact**: High - Could result in data loss, broken expense reports, user trust issues
**Probability**: Medium - Complex data transformations always carry risk
**Mitigation**:
1. **Pre-Migration**:
   - Create full database backup before migration
   - Test migration on production database copy (3 iterations minimum)
   - Implement dry-run mode that validates without making changes
   - Write validation queries to verify relationship integrity
   - Prepare rollback SQL script

2. **During Migration**:
   - Run validation queries before and after migration
   - Log all changes for audit trail
   - Monitor for errors in real-time
   - Execute during low-traffic maintenance window

3. **Post-Migration**:
   - Run integrity checks (see validation queries in PRD Appendix E)
   - Verify sample expenses still have valid categories
   - Monitor error logs for 48 hours
   - Keep rollback script ready for 1 week

**Contingency Plan**: If migration fails, rollback to previous database state, investigate root cause, fix script, retry during next maintenance window.

---

#### R2: Performance Degradation with Large Datasets
**Description**: Associations with 500+ categories experience slow page loads, sluggish search, or timeout errors.

**Impact**: Medium - Degrades user experience, may make feature unusable for large associations
**Probability**: Medium - Current design hasn't been tested at scale
**Mitigation**:
1. **Database Optimization**:
   - Implement composite indexes (Phase 1)
   - Use partial indexes for active category queries
   - Add EXPLAIN QUERY PLAN analysis to all queries
   - Set LIMIT clauses on usage detail queries

2. **Frontend Optimization**:
   - Implement pagination (100 items per page)
   - Add virtual scrolling for tree view
   - Debounce search input (300ms)
   - Lazy load usage statistics on-demand
   - Use Vue's `v-memo` for expensive renders

3. **Performance Testing**:
   - Create test dataset with 1000+ categories
   - Run load tests with 50 concurrent users
   - Measure and baseline critical metrics
   - Set performance budgets (page load <2s, search <500ms)

**Contingency Plan**: If performance targets not met, implement server-side pagination, add Redis caching layer, or simplify tree view to flat table.

---

#### R3: User Confusion About Soft Delete Behavior
**Description**: Users don't understand why "deleted" categories still appear in historical reports or why they can't permanently delete categories.

**Impact**: Medium - Increases support burden, user frustration, potential data mismanagement
**Probability**: High - Soft delete is non-intuitive to non-technical users
**Mitigation**:
1. **UX Design**:
   - Use "Deactivate" instead of "Delete" in UI
   - Add tooltip: "Deactivated categories remain visible in past expenses"
   - Visual distinction: Gray out inactive categories with strikethrough
   - Show status badge: "Inactive" in red

2. **User Education**:
   - In-app first-time-use tooltip explaining soft delete
   - Help documentation with screenshots
   - FAQ section: "Why can't I delete this category?"
   - Training session for association admins during rollout

3. **Confirmation Dialogs**:
   - Before deactivation: "This category will be hidden from new expenses but remain in history"
   - Show usage count: "Used by 24 expenses"
   - Offer "Reactivate" option prominently

**Contingency Plan**: If support tickets remain high, create more prominent educational materials, consider adding a "What is soft delete?" info modal.

---

#### R4: Breaking Existing Expense Forms
**Description**: Changes to category management inadvertently break expense creation/editing forms.

**Impact**: High - Core functionality broken, blocks expense entry operations
**Probability**: Low - Existing components unchanged, but integration risk exists
**Mitigation**:
1. **Isolation**:
   - No changes to `CategorySelector.vue` component
   - No changes to `LocalizedCategoryDisplay.vue` component
   - Existing API endpoint `/categories` returns active categories only
   - New endpoint `/categories/all` is separate

2. **Testing**:
   - Regression test suite for expense forms
   - Integration tests verifying category dropdown population
   - Manual testing of create/edit expense workflows
   - Verify expense reports still render categories correctly

3. **Monitoring**:
   - Monitor error rates on expense endpoints post-deployment
   - User feedback specifically for expense-related issues

**Contingency Plan**: If expense forms break, immediately deploy rollback, investigate regression, add missing test coverage, redeploy with fix.

---

#### R5: Translation Key Management Complexity
**Description**: As associations create custom categories, translation keys become inconsistent or unmaintainable.

**Impact**: Low - Aesthetic issue, doesn't break functionality
**Probability**: High - Guaranteed to happen as users create custom categories
**Mitigation**:
1. **Design Decision**:
   - Store raw values in database (e.g., "Personnel Expenses")
   - Auto-generate translation keys: `personnel_expenses`
   - Fallback to raw value if translation missing
   - Document naming conventions clearly

2. **Translation Management UI** (future enhancement):
   - Show missing translations in admin panel
   - Export/import functionality for translation files
   - Bulk translation update capability

3. **Default Behavior**:
   - If no translation exists, display the raw database value
   - Maintain English and Romanian defaults for common categories

**Contingency Plan**: If translation issues become widespread, implement a simple key→value override system in the database.

---

#### R6: Scope Creep
**Description**: Stakeholders request additional features mid-project (e.g., category merging, budget limits, approval workflows).

**Impact**: Medium - Delays timeline, increases budget, may compromise quality
**Probability**: High - Feature requests are common once users see working prototype
**Mitigation**:
1. **Process**:
   - Refer to PRD "Out of Scope" section (15 deferred features)
   - Require formal change request for new features
   - Impact analysis: estimate effort, timeline delay, risk
   - Product Manager approval required
   - Defer to Phase 2 roadmap

2. **Communication**:
   - Weekly stakeholder updates on progress
   - Demo sessions to manage expectations
   - Roadmap transparency (show Phase 2 features)
   - Celebrate MVP delivery milestone

**Contingency Plan**: If critical feature requested mid-project, pause development, reassess priorities, potentially extend timeline or reduce scope elsewhere.

---

#### R7: Cross-Browser Compatibility Issues
**Description**: Tree view or advanced UI components don't render correctly in Safari, Firefox, or Edge.

**Impact**: Low - Affects subset of users, workaround available
**Probability**: Low - Naive UI is well-tested across browsers
**Mitigation**:
1. **Technology Choice**:
   - Use Naive UI components (NTree) which are battle-tested
   - Avoid custom CSS hacks or browser-specific features
   - Progressive enhancement approach

2. **Testing**:
   - Cross-browser testing checklist (Chrome, Firefox, Safari, Edge)
   - Automated browser testing in CI/CD pipeline
   - Test on both Windows and macOS

3. **Fallback**:
   - If tree view fails, fall back to flat table view
   - Provide browser compatibility matrix in docs

**Contingency Plan**: If major compatibility issue found post-launch, provide quick patch or temporarily disable problematic feature for affected browsers.

---

#### R8: Accidental Bulk Deletion
**Description**: User accidentally deactivates all categories or critical categories, disrupting expense operations.

**Impact**: High - Blocks expense entry until categories reactivated
**Probability**: Low - Requires user to select many items and confirm
**Mitigation**:
1. **Prevention**:
   - Confirmation dialog showing exact count: "Deactivate 45 categories?"
   - Preview list of categories to be deactivated
   - Prevent deactivation of last active category (validation rule)
   - Warning if attempting to deactivate >50% of active categories

2. **Recovery**:
   - Easy reactivation process (one-click restore)
   - Audit trail in `updated_at` timestamps
   - Support team can quickly reactivate via database

3. **Education**:
   - Best practice guide: deactivate one at a time for testing
   - Warning tooltip on bulk deactivate button

**Contingency Plan**: If bulk deletion occurs, support team restores categories via SQL UPDATE, document incident, improve confirmation dialog messaging.

---

## Dependencies and Blockers

### Internal Dependencies

```
Dependency Graph:

Phase 1 (Backend)
   │
   ├─→ sqlc installation and configuration
   ├─→ Database migration framework (goose)
   ├─→ Access to production database copy for testing
   │
   └─→ Phase 2 (Frontend) ← blocks until backend complete
          │
          ├─→ API documentation from Phase 1
          ├─→ Vue 3 and Naive UI setup
          ├─→ TypeScript configuration
          │
          └─→ Phase 3 (Advanced Features) ← blocks until core UI complete
                 │
                 ├─→ Search debounce implementation
                 ├─→ Bulk operation backend support
                 │
                 └─→ Phase 4 (Testing) ← blocks until all features complete
                        │
                        ├─→ Test frameworks (Vitest, Playwright)
                        ├─→ Staging environment access
                        │
                        └─→ Phase 5 (Deployment) ← blocks until UAT passed
                               │
                               ├─→ Production database access
                               ├─→ Deployment pipeline configured
                               └─→ Maintenance window scheduled
```

### External Dependencies

| Dependency | Type | Owner | Risk | Mitigation |
|------------|------|-------|------|------------|
| Production DB Copy | Infrastructure | DevOps | Medium | Request in advance, create snapshot script |
| Staging Environment | Infrastructure | DevOps | Low | Already available |
| Maintenance Window | Business | Product Manager | Medium | Schedule 2 weeks in advance, communicate to users |
| UAT Participants | Business | Association Admins | Medium | Recruit early, provide incentives |
| CI/CD Pipeline Access | Infrastructure | DevOps | Low | Verify access in Week 1 |
| Design Assets (Optional) | Design | UX Designer | Low | Use default Naive UI styling if unavailable |

### Potential Blockers

| Blocker | Impact | Resolution Strategy | Escalation Path |
|---------|--------|---------------------|-----------------|
| Backend Developer Unavailability | Critical | Cross-train Frontend Dev on basic Go, have Tech Lead step in | Delay Phase 1 by X days, compress later phases |
| Frontend Developer Unavailability | Critical | Cross-train Backend Dev on Vue basics, simplify UI | Reduce Phase 3 scope, delay advanced features |
| Database Migration Failure | Critical | Execute rollback plan, fix issues, retry | Extended maintenance window, phased rollout |
| UAT Rejection | High | Address feedback, schedule re-test | Delay production release, prioritize critical fixes |
| Production Outage During Deploy | Critical | Immediate rollback, investigate | Emergency hotfix team, post-mortem |
| Scope Creep (Major Feature Request) | Medium | Change request process, defer to Phase 2 | Product Manager decision, stakeholder alignment |

### Dependency Management Actions

**Week 1 (Phase 1 Start)**:
- ✅ Confirm sqlc and goose are installed
- ✅ Request production database snapshot from DevOps
- ✅ Verify backend developer has database access
- ✅ Set up development database with test data

**Week 2 (Phase 2 Start)**:
- ✅ Confirm Phase 1 API endpoints are functional
- ✅ Review API documentation with Frontend Dev
- ✅ Verify Naive UI library version compatibility
- ✅ Set up Vue i18n configuration

**Week 3 (Phase 3 Start)**:
- ✅ Confirm core UI (Phase 2) is merged to main branch
- ✅ Deploy Phase 2 to staging for QA testing
- ✅ Begin recruiting UAT participants

**Week 4 (Phase 4 Start)**:
- ✅ All features code-complete and merged
- ✅ Test data seeded in staging environment
- ✅ Playwright test suite configured
- ✅ UAT test plan approved

**Week 5 (Phase 5 Start)**:
- ✅ UAT sign-off received
- ✅ Maintenance window scheduled and communicated
- ✅ Rollback plan tested
- ✅ Production deployment checklist completed

---

## Testing Strategy

### Testing Pyramid

```
                  /\
                 /  \
                /E2E \         10% - End-to-End Tests
               /______\
              /        \
             /Integration\    30% - Integration Tests
            /____________\
           /              \
          /  Unit Tests    \  60% - Unit Tests
         /__________________\
```

### Test Coverage Targets

| Layer | Target | Measurement |
|-------|--------|-------------|
| Unit Tests (Backend) | 80% | Go test coverage |
| Unit Tests (Frontend) | 70% | Vitest coverage |
| Integration Tests | All critical paths | API endpoint tests |
| E2E Tests | 5 core workflows | Playwright scenarios |

---

### Unit Testing

#### Backend (Go)
**File**: `api/internal/handlers/categories_test.go`

**Test Cases**:
1. **HandleGetAllCategories**
   - Returns active categories only when include_inactive=false
   - Returns all categories when include_inactive=true
   - Filters by association_id correctly
   - Returns 403 for unauthorized association access

2. **HandleCreateCategory**
   - Creates category successfully with valid input
   - Returns 400 for missing required fields
   - Returns 409 for duplicate category (type+family+name)
   - Returns 403 for non-member association
   - Validates max length constraints

3. **HandleUpdateCategory**
   - Updates category successfully
   - Returns 409 if update creates duplicate
   - Returns 404 for non-existent category
   - Preserves association_id (cannot change ownership)

4. **HandleDeactivateCategory**
   - Soft deletes category (sets is_deleted=true)
   - Returns 400 if attempting to deactivate last active category
   - Works even if category is in use by expenses

5. **HandleReactivateCategory**
   - Reactivates soft-deleted category
   - Returns 409 if reactivation creates duplicate
   - Returns 404 for non-existent category

6. **HandleGetCategoryUsage**
   - Returns accurate usage count
   - Returns 0 for unused categories
   - Returns recent expense details

7. **HandleBulkDeactivateCategories**
   - Deactivates multiple categories in single transaction
   - Validates all IDs belong to the association
   - Returns partial success if some IDs invalid

**Tools**: Go standard `testing` package, `testify/assert` for assertions

**Execution**:
```bash
cd api
go test ./internal/handlers -v -cover
go test ./internal/handlers -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

#### Frontend (TypeScript/Vue)
**Files**: `*.spec.ts` alongside component files

**Test Cases**:
1. **CategoryForm.vue**
   - Renders in create mode with empty form
   - Renders in edit mode with pre-filled data
   - Validates required fields
   - Validates max length (100 chars)
   - Emits 'saved' event on successful save
   - Emits 'cancelled' event on cancel click
   - Displays error messages from API

2. **CategoriesList.vue**
   - Displays categories in table format
   - Shows empty state when no categories
   - Filters categories by status (active/inactive)
   - Searches categories by text
   - Emits 'edit' event with category ID
   - Emits 'deactivate' event with category ID
   - Displays usage count badges
   - Paginates results (100 per page)

3. **CategoryTree.vue** (if implemented)
   - Renders tree hierarchy (Type → Family → Name)
   - Expands/collapses nodes on click
   - Shows active/inactive status
   - Emits action events correctly

4. **categoryApi (services/api.ts)**
   - Calls correct endpoints with proper parameters
   - Handles successful responses
   - Handles error responses
   - Includes authentication headers

**Tools**: Vitest, Vue Test Utils, Mock Service Worker (MSW) for API mocking

**Execution**:
```bash
cd ui
npm run test:unit
npm run test:coverage
```

---

### Integration Testing

**Scope**: Test interaction between frontend, backend, and database

**Test Scenarios**:

1. **Create Category Flow**
   - POST /v1/api/associations/1/categories
   - Verify category inserted into database
   - Verify category appears in GET /v1/api/associations/1/categories
   - Verify category available in expense form dropdown

2. **Edit Category Flow**
   - Create category
   - PUT /v1/api/associations/1/categories/X with new values
   - Verify database updated
   - Verify existing expense references unchanged

3. **Deactivate Category Flow**
   - Create category and link to expense
   - PUT /v1/api/associations/1/categories/X/deactivate
   - Verify is_deleted=true in database
   - Verify category hidden from GET /categories (active only)
   - Verify category still visible in expense GET /expenses/X

4. **Reactivate Category Flow**
   - Deactivate category
   - PUT /v1/api/associations/1/categories/X/reactivate
   - Verify is_deleted=false in database
   - Verify category appears in active list

5. **Bulk Operations Flow**
   - Create 10 categories
   - POST /bulk-deactivate with [id1, id2, id3]
   - Verify 3 categories deactivated
   - POST /bulk-reactivate with [id1]
   - Verify 1 category reactivated

6. **Association Isolation**
   - Create category for Association A
   - Attempt to access via Association B's endpoint
   - Verify 403 Forbidden response

**Tools**: Go integration tests using test database, Postman/Newman for API testing

**Execution**:
```bash
cd api
go test ./tests/integration -v
```

---

### End-to-End Testing

**Scope**: Simulate real user workflows in browser

**Tool**: Playwright (TypeScript)

**Test Scenarios**:

1. **Happy Path: Create and Use Category**
   - Login as association admin
   - Navigate to /categories
   - Click "Create New Category"
   - Fill form: Type="Budgeted", Family="Utilities", Name="Electricity"
   - Click Save
   - Verify success message
   - Navigate to /expenses
   - Click "Create Expense"
   - Verify "Electricity" appears in category dropdown
   - Select category and complete expense creation
   - Verify expense saved with correct category

2. **Edit Existing Category**
   - Login as association admin
   - Navigate to /categories
   - Search for "Electricity"
   - Click Edit button
   - Change Name to "Electricity Bill"
   - Click Save
   - Verify category name updated in list
   - Navigate to expense using this category
   - Verify updated name displayed

3. **Deactivate and Reactivate**
   - Navigate to /categories
   - Find category with 0 usage
   - Click Deactivate
   - Confirm deactivation dialog
   - Verify category marked as Inactive
   - Filter to show "Inactive" categories
   - Click Reactivate
   - Verify category marked as Active

4. **Bulk Deactivate**
   - Navigate to /categories
   - Select 3 categories via checkboxes
   - Click "Bulk Deactivate"
   - Confirm dialog showing count
   - Verify all 3 categories deactivated

5. **Search and Filter**
   - Navigate to /categories
   - Enter search term "Personnel"
   - Verify only matching categories displayed
   - Change status filter to "Inactive"
   - Verify only inactive categories displayed

**Execution**:
```bash
cd ui
npx playwright test
npx playwright test --headed  # with browser UI
npx playwright show-report
```

**CI/CD Integration**: E2E tests run on every PR to main branch, blocking merge if failures.

---

### Performance Testing

**Objectives**:
- Verify page load time <2 seconds
- Verify search response time <500ms
- Verify bulk operations complete <3 seconds

**Test Dataset**:
- 1000 categories across 3 types, 20 families
- 10,000 expenses referencing categories
- 50 concurrent users

**Tools**: Apache JMeter or k6 for load testing

**Test Scenarios**:
1. GET /categories/all - measure response time with 1000 results
2. GET /categories/all with search query - measure filter performance
3. POST /bulk-deactivate with 100 IDs - measure transaction time
4. Concurrent user simulation: 50 users browsing categories simultaneously

**Acceptance Criteria**:
- 95th percentile response time <2 seconds for page load
- 95th percentile response time <500ms for search
- No database locks or timeouts under load

---

### Security Testing

**Objectives**:
- Verify authentication and authorization enforcement
- Prevent SQL injection and XSS attacks
- Validate input sanitization

**Test Cases**:
1. **Authentication Bypass**
   - Attempt to access /categories without JWT token
   - Verify 401 Unauthorized response

2. **Association Authorization**
   - Login as user for Association A
   - Attempt to access Association B's categories
   - Verify 403 Forbidden response

3. **SQL Injection**
   - Submit category name: `'; DROP TABLE categories; --`
   - Verify query parameterization prevents injection
   - Verify category created with literal string

4. **XSS Attack**
   - Create category with name: `<script>alert('XSS')</script>`
   - Verify Vue escapes HTML on display
   - Verify no script execution in browser

5. **CSRF Protection**
   - Verify CSRF tokens on state-changing operations (if implemented)

**Tools**: OWASP ZAP for automated scanning, manual penetration testing

---

### User Acceptance Testing (UAT)

**Participants**: 3-5 association administrators from different associations

**Duration**: 3 days (Week 4)

**Test Plan**:
1. Provide UAT guide with step-by-step scenarios
2. Seed test data in staging environment
3. Conduct 1-hour training session
4. Users perform scripted tests independently
5. Collect feedback via form and interviews
6. Triage issues (Critical/High/Medium/Low)

**Scenarios**:
- Create 5 custom categories for your association
- Edit a category name to fix a typo
- Deactivate an unused category
- Reactivate a previously deactivated category
- Use new category in expense creation
- Search for a specific category
- Verify inactive categories don't appear in expense dropdown

**Acceptance Criteria**:
- 80% of participants complete all scenarios successfully
- No Critical or High severity bugs reported
- Average satisfaction score >4/5

---

### Test Automation Strategy

**CI/CD Pipeline**:
```
git push → GitHub Actions
   ├─→ Backend: go test (unit + integration)
   ├─→ Frontend: npm run test (unit)
   ├─→ Linting: golangci-lint, eslint
   ├─→ Build: go build, npm run build
   └─→ E2E: Playwright tests (on staging deploy)
```

**Coverage Gates**:
- Pull requests must maintain >80% backend coverage
- Pull requests must maintain >70% frontend coverage
- E2E tests must pass before merging to main

**Test Data Management**:
- Seed script for consistent test data
- Database reset between test runs
- Fixtures for common test scenarios

---

## Detailed Implementation Checklist

### Phase 1: Database & Backend Foundation

#### Backend Developer Tasks

**Day 1: Database Setup**
- [ ] Create migration file: `api/sql/schema/00021_add_category_indexes.sql`
  - [ ] Add composite index on (association_id, is_deleted)
  - [ ] Add partial index on (association_id, type, family, name) WHERE is_deleted=FALSE
  - [ ] Test migration on local database
  - [ ] Write rollback (down migration)
- [ ] Run migration locally: `goose up`
- [ ] Verify indexes created: `PRAGMA index_list('categories');`
- [ ] Create test data: seed 100 categories for testing

**Day 2-3: SQL Queries & Code Generation**
- [ ] Open `api/sql/queries/categories.sql`
- [ ] Add query: `GetAllCategories` (with include_inactive param)
- [ ] Add query: `UpdateCategory` (return updated record)
- [ ] Add query: `ReactivateCategory`
- [ ] Add query: `GetCategoryUsageCount`
- [ ] Add query: `GetCategoryUsageDetails` (last 10 expenses)
- [ ] Add query: `BulkDeactivateCategories` (use sqlc.slice)
- [ ] Add query: `BulkReactivateCategories` (use sqlc.slice)
- [ ] Add query: `CheckCategoryUniqueness` (for validation)
- [ ] Run `sqlc generate`
- [ ] Verify generated code in `api/internal/database/categories.sql.go`
- [ ] Fix any sqlc errors or warnings

**Day 3-4: HTTP Handlers**
- [ ] Open `api/internal/handlers/categories.go`
- [ ] Implement `HandleGetAllCategories`
  - [ ] Parse `include_inactive` query parameter (default: false)
  - [ ] Call `db.GetAllCategories` with association_id and filter
  - [ ] Map database models to response DTOs
  - [ ] Return JSON with status 200
  - [ ] Handle errors (404, 500)
- [ ] Implement `HandleUpdateCategory`
  - [ ] Parse JSON request body
  - [ ] Validate required fields (type, family, name)
  - [ ] Call `db.CheckCategoryUniqueness` to prevent duplicates
  - [ ] Call `db.UpdateCategory`
  - [ ] Return updated category JSON with status 200
  - [ ] Handle errors (400, 404, 409, 500)
- [ ] Implement `HandleReactivateCategory`
  - [ ] Get category ID from path parameter
  - [ ] Call `db.ReactivateCategory`
  - [ ] Check for uniqueness conflicts
  - [ ] Return success message with status 200
  - [ ] Handle errors (404, 409, 500)
- [ ] Implement `HandleGetCategoryUsage`
  - [ ] Call `db.GetCategoryUsageCount`
  - [ ] Call `db.GetCategoryUsageDetails`
  - [ ] Build response with count and expense list
  - [ ] Return JSON with status 200
- [ ] Implement `HandleBulkDeactivateCategories`
  - [ ] Parse JSON request body (array of IDs)
  - [ ] Validate all IDs belong to association
  - [ ] Call `db.BulkDeactivateCategories`
  - [ ] Return success message with count
  - [ ] Handle errors (400, 500)
- [ ] Implement `HandleBulkReactivateCategories`
  - [ ] Similar to bulk deactivate
  - [ ] Add uniqueness validation for reactivated categories

**Day 4: Router Configuration**
- [ ] Open `api/main.go`
- [ ] Add route: `GET /v1/api/associations/{associationId}/categories/all`
  - [ ] Handler: HandleGetAllCategories
  - [ ] Middleware: MiddlewareAssociationResource
- [ ] Add route: `PUT /v1/api/associations/{associationId}/categories/{categoryId}`
  - [ ] Handler: HandleUpdateCategory
  - [ ] Middleware: MiddlewareAssociationResource
- [ ] Add route: `PUT /v1/api/associations/{associationId}/categories/{categoryId}/reactivate`
  - [ ] Handler: HandleReactivateCategory
  - [ ] Middleware: MiddlewareAssociationResource
- [ ] Add route: `GET /v1/api/associations/{associationId}/categories/{categoryId}/usage`
  - [ ] Handler: HandleGetCategoryUsage
  - [ ] Middleware: MiddlewareAssociationResource
- [ ] Add route: `POST /v1/api/associations/{associationId}/categories/bulk-deactivate`
  - [ ] Handler: HandleBulkDeactivateCategories
  - [ ] Middleware: MiddlewareAssociationResource
- [ ] Add route: `POST /v1/api/associations/{associationId}/categories/bulk-reactivate`
  - [ ] Handler: HandleBulkReactivateCategories
  - [ ] Middleware: MiddlewareAssociationResource
- [ ] Run `go build` to verify compilation
- [ ] Start server and test routes via curl/Postman

**Day 5: Unit Tests**
- [ ] Create file: `api/internal/handlers/categories_test.go`
- [ ] Set up test database connection
- [ ] Write test: `TestHandleGetAllCategories_ActiveOnly`
- [ ] Write test: `TestHandleGetAllCategories_IncludeInactive`
- [ ] Write test: `TestHandleCreateCategory_Success`
- [ ] Write test: `TestHandleCreateCategory_DuplicateError`
- [ ] Write test: `TestHandleUpdateCategory_Success`
- [ ] Write test: `TestHandleUpdateCategory_UniquenessConflict`
- [ ] Write test: `TestHandleDeactivateCategory_Success`
- [ ] Write test: `TestHandleReactivateCategory_Success`
- [ ] Write test: `TestHandleGetCategoryUsage`
- [ ] Write test: `TestHandleBulkDeactivate`
- [ ] Write test: `TestAssociationAuthorization`
- [ ] Run `go test -v -cover`
- [ ] Verify coverage >80%
- [ ] Fix failing tests

**Phase 1 Completion Checklist**
- [ ] All handlers implemented and functional
- [ ] All unit tests passing with >80% coverage
- [ ] API endpoints tested manually via Postman
- [ ] No breaking changes to existing expense endpoints
- [ ] Code reviewed by Tech Lead
- [ ] Merged to main branch

---

### Phase 2: Core Frontend UI

#### Frontend Developer Tasks

**Day 1: Project Setup & API Service**
- [ ] Open `ui/src/services/api.ts`
- [ ] Add `categoryApi` object (if not exists)
- [ ] Implement method: `getAllCategories(associationId, includeInactive)`
- [ ] Implement method: `getCategory(associationId, categoryId)`
- [ ] Implement method: `createCategory(associationId, data)`
- [ ] Implement method: `updateCategory(associationId, categoryId, data)`
- [ ] Implement method: `deactivateCategory(associationId, categoryId)`
- [ ] Implement method: `reactivateCategory(associationId, categoryId)`
- [ ] Implement method: `getCategoryUsage(associationId, categoryId)`
- [ ] Implement method: `bulkDeactivate(associationId, categoryIds)`
- [ ] Implement method: `bulkReactivate(associationId, categoryIds)`
- [ ] Open `ui/src/types/api.ts`
- [ ] Define interface: `Category`
- [ ] Define interface: `CategoryCreateRequest`
- [ ] Define interface: `CategoryUpdateRequest`
- [ ] Define interface: `CategoryUsageResponse`
- [ ] Run TypeScript compiler: `npm run type-check`

**Day 2: Categories Page**
- [ ] Create file: `ui/src/pages/categories/index.vue`
- [ ] Import components: NPageHeader, NCard, NModal
- [ ] Import: AssociationSelector component
- [ ] Set up Composition API: `ref`, `computed`, `onMounted`
- [ ] Add state: `associationId` (ref)
- [ ] Add state: `showCategoryModal` (ref)
- [ ] Add state: `editingCategoryId` (ref)
- [ ] Add computed: `modalTitle` (Create vs Edit)
- [ ] Add computed: `canShowCategories` (association selected)
- [ ] Add method: `handleEditCategory(categoryId)`
- [ ] Add method: `handleCreateCategory()`
- [ ] Add method: `handleCategorySaved(category)`
- [ ] Add method: `handleCategoryFormCancelled()`
- [ ] Add method: `handleAssociationChanged(newId)`
- [ ] Add template: Page header with title
- [ ] Add template: Association selector
- [ ] Add template: "Create Category" button
- [ ] Add template: CategoriesList component (placeholder)
- [ ] Add template: CategoryForm modal
- [ ] Add i18n: `t('categories.title')`
- [ ] Test: Page renders without errors

**Day 3: Categories List Component**
- [ ] Create file: `ui/src/components/CategoriesList.vue`
- [ ] Import: NDataTable, NButton, NTag, NSpin, NEmpty
- [ ] Add props: `associationId` (number, required)
- [ ] Add state: `categories` (ref<Category[]>)
- [ ] Add state: `loading` (ref<boolean>)
- [ ] Add state: `error` (ref<string | null>)
- [ ] Add state: `includeInactive` (ref<boolean> = false)
- [ ] Add method: `async fetchCategories()`
  - [ ] Call `categoryApi.getAllCategories`
  - [ ] Set loading state
  - [ ] Handle success and error
  - [ ] Update `categories` ref
- [ ] Add method: `handleDeactivate(categoryId)`
  - [ ] Show confirmation dialog
  - [ ] Call `categoryApi.deactivateCategory`
  - [ ] Refresh list
  - [ ] Show success message
- [ ] Add method: `handleReactivate(categoryId)`
  - [ ] Call `categoryApi.reactivateCategory`
  - [ ] Refresh list
- [ ] Add method: `updateCategory(updatedCategory)` (for parent)
- [ ] Add method: `addCategory(newCategory)` (for parent)
- [ ] Define table columns:
  - [ ] Type column
  - [ ] Family column
  - [ ] Name column
  - [ ] Status column (NTag: Active/Inactive)
  - [ ] Usage Count column
  - [ ] Actions column (Edit, Deactivate/Reactivate buttons)
- [ ] Add template: NSpin wrapper for loading
- [ ] Add template: NDataTable with columns and data
- [ ] Add template: NEmpty for no categories
- [ ] Add template: Filter toggle for "Include Inactive"
- [ ] Add watcher: `watch(associationId)` → fetch categories
- [ ] Add watcher: `watch(includeInactive)` → fetch categories
- [ ] Test: List displays categories correctly

**Day 4: Category Form Component**
- [ ] Create file: `ui/src/components/CategoryForm.vue`
- [ ] Import: NForm, NFormItem, NInput, NButton, NSpace, NSpin
- [ ] Add props: `associationId` (number, required)
- [ ] Add props: `categoryId` (number, optional)
- [ ] Add state: `formData` (reactive)
  - [ ] type: string
  - [ ] family: string
  - [ ] name: string
- [ ] Add state: `loading` (ref<boolean>)
- [ ] Add state: `submitting` (ref<boolean>)
- [ ] Add state: `error` (ref<string | null>)
- [ ] Add formRef: `ref<FormInst | null>`
- [ ] Add computed: `isEditMode` (based on categoryId)
- [ ] Define validation rules:
  - [ ] type: required, max 100 chars
  - [ ] family: required, max 100 chars
  - [ ] name: required, max 100 chars
- [ ] Add method: `async fetchCategoryDetails()`
  - [ ] If categoryId, call `categoryApi.getCategory`
  - [ ] Populate formData
- [ ] Add method: `async handleSubmit()`
  - [ ] Validate form
  - [ ] If edit mode, call `updateCategory`
  - [ ] If create mode, call `createCategory`
  - [ ] Emit 'saved' event with response data
  - [ ] Handle errors
- [ ] Add method: `handleCancel()`
  - [ ] Emit 'cancelled' event
- [ ] Add method: `resetForm()`
- [ ] Add template: NForm with formRef
- [ ] Add template: NFormItem for Type input
- [ ] Add template: NFormItem for Family input
- [ ] Add template: NFormItem for Name input
- [ ] Add template: Error alert (NAlert)
- [ ] Add template: Action buttons (Save, Cancel)
- [ ] Add onMounted: fetch category details if editing
- [ ] Test: Form creates category successfully
- [ ] Test: Form updates category successfully
- [ ] Test: Validation works correctly

**Day 5: i18n & Router Integration**
- [ ] Open `ui/src/i18n/locales/en.json`
- [ ] Add all category-related translation keys (see Architecture section)
- [ ] Open `ui/src/i18n/locales/ro.json`
- [ ] Add Romanian translations
- [ ] Open router file (e.g., `ui/src/router/index.ts`)
- [ ] Add route: `{ path: '/categories', component: () => import('@/pages/categories/index.vue'), meta: { requiresAuth: true } }`
- [ ] (Optional) Add navigation menu link to categories page
- [ ] Test: Navigate to /categories
- [ ] Test: Page requires authentication
- [ ] Test: All translations display correctly
- [ ] Test: Language switching works (EN ↔ RO)

**Phase 2 Completion Checklist**
- [ ] Categories page fully functional
- [ ] Can create new category via UI
- [ ] Can edit existing category via UI
- [ ] Can deactivate category via UI
- [ ] Active/inactive categories visually distinct
- [ ] Association selector filters categories correctly
- [ ] No console errors or warnings
- [ ] i18n translations complete (EN + RO)
- [ ] Code reviewed by Tech Lead
- [ ] Merged to main branch

---

### Phase 3: Advanced Features & Integration

#### Frontend Developer Tasks

**Day 1: Search and Filter**
- [ ] Open `ui/src/components/CategoriesList.vue`
- [ ] Add state: `searchQuery` (ref<string> = '')
- [ ] Add state: `statusFilter` (ref<'all' | 'active' | 'inactive'> = 'all')
- [ ] Add state: `typeFilter` (ref<string> = '')
- [ ] Add computed: `filteredCategories`
  - [ ] Filter by searchQuery (type, family, name)
  - [ ] Filter by statusFilter (is_deleted)
  - [ ] Filter by typeFilter (type field)
- [ ] Add template: Search input (NInput) with magnifying glass icon
- [ ] Add template: Status filter dropdown (NSelect)
- [ ] Add template: Type filter dropdown (NSelect)
- [ ] Add debounce to search input (300ms)
- [ ] Test: Search filters results in real-time
- [ ] Test: Multiple filters work together

**Day 2: Category Tree View (Optional)**
- [ ] Create file: `ui/src/components/CategoryTree.vue`
- [ ] Import: NTree from Naive UI
- [ ] Add props: `categories` (Category[])
- [ ] Add method: `buildTreeData(categories)` → CategoryTreeNode[]
  - [ ] Group by Type → Family → Name
  - [ ] Create hierarchical structure
- [ ] Add state: `expandedKeys` (ref<string[]>)
- [ ] Add template: NTree component
- [ ] Add template: Custom node rendering (status badge, actions)
- [ ] Add method: `handleNodeExpand(keys)`
- [ ] Add method: `handleNodeActionClick(action, node)`
- [ ] Test: Tree expands/collapses correctly
- [ ] Test: Actions trigger events
- [ ] (Optional) Allow parent component to toggle between tree/table view

**Day 3: Bulk Operations**
- [ ] Open `ui/src/components/CategoriesList.vue`
- [ ] Add state: `selectedCategoryIds` (ref<number[]> = [])
- [ ] Add computed: `selectedCount`
- [ ] Add method: `handleSelectionChange(rowKeys)`
- [ ] Add method: `async handleBulkDeactivate()`
  - [ ] Show confirmation dialog with count
  - [ ] Call `categoryApi.bulkDeactivate(associationId, selectedCategoryIds)`
  - [ ] Refresh list
  - [ ] Clear selection
  - [ ] Show success message
- [ ] Add method: `async handleBulkReactivate()`
  - [ ] Similar to bulk deactivate
- [ ] Add table column: Checkbox for selection
- [ ] Add template: "Bulk Deactivate" button (visible when items selected)
- [ ] Add template: "Bulk Reactivate" button
- [ ] Add template: Selected count display
- [ ] Test: Selecting multiple items works
- [ ] Test: Bulk deactivate confirms and executes
- [ ] Test: Bulk reactivate works

**Day 4: Usage Statistics & UX Enhancements**
- [ ] Add method: `async fetchCategoryUsage(categoryId)`
  - [ ] Call `categoryApi.getCategoryUsage`
  - [ ] Display in modal or tooltip
- [ ] Enhance `handleDeactivate` method:
  - [ ] Fetch usage count before confirmation
  - [ ] Show warning if category in use: "Used by 24 expenses"
  - [ ] Include usage count in confirmation dialog
- [ ] Add loading indicators:
  - [ ] NSpin for initial page load
  - [ ] Skeleton screen for table
  - [ ] Button loading state during save
- [ ] Add empty states:
  - [ ] "No categories found" with illustration
  - [ ] "No search results" with suggestion to clear filters
  - [ ] "Create your first category" call-to-action
- [ ] Add toast notifications (NMessage):
  - [ ] Success: "Category created successfully"
  - [ ] Error: "Failed to create category"
  - [ ] Warning: "Category is used by X expenses"
- [ ] Add confirmation dialogs (NModal + NPopconfirm):
  - [ ] Deactivate confirmation
  - [ ] Bulk deactivate confirmation
- [ ] Test: All UX enhancements functional

**Day 5: Integration Testing & Bug Fixes**
- [ ] Write Playwright test: Create and use category in expense
- [ ] Write Playwright test: Edit category and verify update
- [ ] Write Playwright test: Deactivate and reactivate flow
- [ ] Write Playwright test: Bulk operations
- [ ] Write Playwright test: Search and filter
- [ ] Run all E2E tests: `npx playwright test`
- [ ] Fix any failing tests
- [ ] Test integration with existing expense form:
  - [ ] Create category in /categories
  - [ ] Navigate to /expenses
  - [ ] Verify new category in dropdown
  - [ ] Create expense with new category
  - [ ] Verify expense saved correctly
- [ ] Test soft delete preservation:
  - [ ] Create expense with category A
  - [ ] Deactivate category A
  - [ ] Navigate to expense detail
  - [ ] Verify category A still displayed
- [ ] Regression testing: Ensure no existing features broken
- [ ] Performance testing: Load page with 500 categories

**Phase 3 Completion Checklist**
- [ ] Search and filter fully functional
- [ ] Bulk operations working correctly
- [ ] Usage statistics displayed
- [ ] Reactivation flow complete
- [ ] All UX enhancements implemented
- [ ] Integration tests passing
- [ ] No regressions in expense management
- [ ] Performance targets met
- [ ] Code reviewed by Tech Lead
- [ ] Merged to main branch

---

### Phase 4: Testing, Polish & Documentation

#### QA Engineer Tasks

**Day 1-2: Test Execution**
- [ ] Execute unit test suite (backend): `go test -v -cover`
- [ ] Verify backend coverage >80%
- [ ] Execute unit test suite (frontend): `npm run test`
- [ ] Verify frontend coverage >70%
- [ ] Execute integration tests
- [ ] Execute E2E test suite: `npx playwright test`
- [ ] Manual exploratory testing:
  - [ ] Test all happy paths
  - [ ] Test edge cases (last active category, duplicate names, etc.)
  - [ ] Test error scenarios (network failure, validation errors)
- [ ] Browser compatibility testing:
  - [ ] Chrome (latest)
  - [ ] Firefox (latest)
  - [ ] Safari (latest)
  - [ ] Edge (latest)
- [ ] Responsive design testing:
  - [ ] Desktop (1920x1080, 1366x768)
  - [ ] Tablet (1024x768, 768x1024)
  - [ ] Mobile (375x667 - view only)
- [ ] Accessibility testing:
  - [ ] Keyboard navigation (Tab, Enter, Escape)
  - [ ] Screen reader compatibility (NVDA/JAWS)
  - [ ] Color contrast verification (WCAG AA)
  - [ ] Focus indicators visible
- [ ] Performance testing:
  - [ ] Load page with 1000 categories
  - [ ] Measure page load time (target: <2s)
  - [ ] Measure search response time (target: <500ms)
  - [ ] Test bulk operations with 100 items

**Day 2-3: Bug Reporting & Triage**
- [ ] Log all bugs in issue tracker (GitHub Issues / Jira)
- [ ] Categorize by severity: Critical, High, Medium, Low
- [ ] Assign to developers
- [ ] Track bug fix progress
- [ ] Retest fixed bugs
- [ ] Critical bugs:
  - [ ] Data loss issues
  - [ ] Authentication bypass
  - [ ] Application crashes
- [ ] High bugs:
  - [ ] Functional failures
  - [ ] Incorrect data display
  - [ ] Performance issues
- [ ] Medium/Low bugs:
  - [ ] UI inconsistencies
  - [ ] Minor validation issues

**Day 3-4: Security & Performance Audits**
- [ ] Run OWASP ZAP security scan
- [ ] Manual security testing:
  - [ ] SQL injection attempts
  - [ ] XSS payload testing
  - [ ] Authentication bypass attempts
  - [ ] CSRF testing (if applicable)
- [ ] Review security findings
- [ ] Verify all High/Critical vulnerabilities fixed
- [ ] Performance profiling:
  - [ ] Backend: Profile slow queries with EXPLAIN
  - [ ] Frontend: Chrome DevTools performance audit
  - [ ] Identify and report bottlenecks
- [ ] Load testing:
  - [ ] Simulate 50 concurrent users
  - [ ] Monitor response times and error rates
  - [ ] Verify no database locks or deadlocks

**Day 4-5: UAT Preparation**
- [ ] Create UAT test plan document
- [ ] Define test scenarios for association admins
- [ ] Prepare test data in staging environment:
  - [ ] Seed 3 associations with different category sets
  - [ ] Create sample expenses linked to categories
  - [ ] Include deactivated categories
- [ ] Prepare UAT guide (step-by-step instructions)
- [ ] Schedule UAT sessions with 3-5 participants
- [ ] Conduct training session for UAT participants
- [ ] Provide UAT feedback form (Google Forms / SurveyMonkey)
- [ ] Monitor UAT sessions and collect feedback
- [ ] Triage UAT findings
- [ ] Schedule UAT sign-off meeting

#### Frontend/Backend Developer Tasks

**Day 1-3: Bug Fixes**
- [ ] Address all Critical bugs immediately
- [ ] Fix High severity bugs
- [ ] Prioritize Medium bugs based on impact
- [ ] Defer Low severity bugs to post-launch backlog
- [ ] Run regression tests after each fix
- [ ] Request QA retesting for fixed bugs

**Day 3-4: Code Review & Refactoring**
- [ ] Tech Lead: Review all code changes
- [ ] Address code review feedback
- [ ] Refactor duplicated code
- [ ] Remove console.log statements
- [ ] Remove commented-out code
- [ ] Format code consistently (gofmt, prettier)
- [ ] Run linters: golangci-lint, eslint
- [ ] Fix all linting warnings

**Day 4-5: Documentation**
- [ ] Backend Developer: Write API documentation
  - [ ] Endpoint specifications (request/response schemas)
  - [ ] Authentication requirements
  - [ ] Example requests (curl commands)
  - [ ] Error code reference
- [ ] Frontend Developer: Update README
  - [ ] Category management feature overview
  - [ ] Component documentation
  - [ ] State management patterns
  - [ ] Troubleshooting guide
- [ ] Write user guide for association admins:
  - [ ] How to create categories
  - [ ] How to edit/deactivate/reactivate categories
  - [ ] Understanding soft delete
  - [ ] Best practices
  - [ ] Screenshots and GIFs
- [ ] Write deployment runbook:
  - [ ] Pre-deployment checklist
  - [ ] Migration execution steps
  - [ ] Post-deployment validation
  - [ ] Rollback procedure

**Phase 4 Completion Checklist**
- [ ] All Critical and High bugs fixed
- [ ] Test coverage targets met (80% backend, 70% frontend)
- [ ] Security audit passed (no High/Critical vulnerabilities)
- [ ] Performance metrics achieved (page load <2s, search <500ms)
- [ ] UAT completed with >80% success rate
- [ ] All documentation complete and reviewed
- [ ] Code review approved by Tech Lead
- [ ] Deployment runbook finalized
- [ ] Sign-off from Product Manager for production deployment

---

### Phase 5: Migration, Deployment & Rollout

#### Backend Developer Tasks

**Day 1: Data Migration Script**
- [ ] Create migration analysis script:
  - [ ] Query all unique (type, family, name) combinations from expenses
  - [ ] Identify associations with existing expenses
  - [ ] Count categories per association
- [ ] Create migration script: `api/scripts/migrate_categories.go`
  - [ ] Backfill categories for all associations
  - [ ] Validate expense references
  - [ ] Dry-run mode (flag: --dry-run)
  - [ ] Verbose logging
- [ ] Create rollback script: `api/scripts/rollback_migration.sql`
- [ ] Test migration on production database copy:
  - [ ] Run dry-run mode
  - [ ] Review logs
  - [ ] Execute migration
  - [ ] Validate with queries (see PRD Appendix E)
  - [ ] Test rollback
  - [ ] Repeat 3 times to ensure consistency
- [ ] Create default category seed file: `api/sql/samples/default_categories.sql`
  - [ ] Romanian categories (Budgeted, Maintenance, Improvement)
  - [ ] Association-agnostic template
- [ ] Update association creation logic to seed default categories

**Day 2: Deployment Preparation**
- [ ] Create deployment checklist
- [ ] Schedule maintenance window (communicate to users 1 week in advance)
- [ ] Prepare staging deployment:
  - [ ] Deploy backend to staging
  - [ ] Run database migration on staging
  - [ ] Deploy frontend to staging
  - [ ] Smoke test all critical paths
- [ ] Prepare rollback artifacts:
  - [ ] Previous backend binary
  - [ ] Previous frontend build
  - [ ] Rollback SQL script
- [ ] Set up monitoring dashboards:
  - [ ] Error rate alerts
  - [ ] Response time alerts
  - [ ] Database connection pool metrics

**Day 3: Production Deployment**
- [ ] Communicate maintenance window start to users
- [ ] Backup production database (full snapshot)
- [ ] Deploy backend:
  - [ ] Stop API server
  - [ ] Deploy new binary
  - [ ] Run database migration: `goose up`
  - [ ] Start API server
  - [ ] Verify health check endpoint
- [ ] Run migration script:
  - [ ] Execute `migrate_categories.go`
  - [ ] Monitor logs for errors
  - [ ] Validate with integrity queries
- [ ] Deploy frontend:
  - [ ] Build production bundle: `npm run build`
  - [ ] Deploy static assets to CDN/server
  - [ ] Clear CDN cache
- [ ] Post-deployment validation:
  - [ ] Test login and authentication
  - [ ] Navigate to /categories page
  - [ ] Create test category
  - [ ] Create test expense with new category
  - [ ] Verify existing expenses still have category references
  - [ ] Test search and filter
  - [ ] Test deactivate/reactivate
- [ ] Communicate maintenance window end to users

**Day 4: Monitoring & Support**
- [ ] Monitor application logs for errors
- [ ] Monitor error rate (target: <1%)
- [ ] Monitor performance metrics (response times)
- [ ] Monitor user feedback channels (support tickets, emails)
- [ ] Be on-call for critical issues (24-hour window)
- [ ] Triage any bugs reported:
  - [ ] Critical: Hotfix immediately
  - [ ] High: Schedule fix within 48 hours
  - [ ] Medium/Low: Add to backlog

**Day 5: Post-Launch Review & Handoff**
- [ ] Run success metrics queries:
  - [ ] Verify 100% expense-category integrity
  - [ ] Check category creation rate
  - [ ] Measure page load times
  - [ ] Count support tickets
- [ ] Conduct post-launch retrospective meeting:
  - [ ] What went well?
  - [ ] What could be improved?
  - [ ] Lessons learned
  - [ ] Action items for future projects
- [ ] Create post-launch report:
  - [ ] Deployment summary
  - [ ] Issues encountered and resolutions
  - [ ] Performance metrics
  - [ ] User feedback summary
  - [ ] Next steps and Phase 2 roadmap
- [ ] Hand off to support team:
  - [ ] Knowledge transfer session
  - [ ] Provide troubleshooting guide
  - [ ] Escalation path for critical issues

**Phase 5 Completion Checklist**
- [ ] Zero downtime deployment achieved
- [ ] Database migration completed successfully
- [ ] 100% of existing expenses retain valid category references
- [ ] All associations have default categories
- [ ] No Critical bugs reported in first 48 hours
- [ ] Monitoring and alerts configured
- [ ] Post-launch report completed
- [ ] Knowledge transfer to support team
- [ ] Project officially closed

---

## Migration and Deployment

### Migration Strategy

#### Pre-Migration Analysis

**Identify Existing Categories**:
```sql
-- Extract all unique category combinations currently in use
SELECT DISTINCT
  c.type,
  c.family,
  c.name,
  e.association_id,
  COUNT(DISTINCT e.id) as expense_count
FROM expenses e
JOIN categories c ON e.category_id = c.id
GROUP BY c.type, c.family, c.name, e.association_id
ORDER BY e.association_id, c.type, c.family, c.name;
```

**Expected Results**:
- 10-50 unique categories per association
- Categories already exist in the `categories` table (soft delete feature)
- No migration of category data needed, only validation

#### Migration Steps

**Step 1: Validate Current State**
```sql
-- Check for orphaned expenses (should return 0)
SELECT COUNT(*) as orphaned_expenses
FROM expenses e
LEFT JOIN categories c ON e.category_id = c.id
WHERE c.id IS NULL;

-- Check for duplicate active categories (should return 0)
SELECT association_id, type, family, name, COUNT(*) as count
FROM categories
WHERE is_deleted = FALSE
GROUP BY association_id, type, family, name
HAVING COUNT(*) > 1;
```

**Step 2: Backfill Association Categories** (if needed)
```go
// api/scripts/migrate_categories.go
package main

import (
    "database/sql"
    "flag"
    "fmt"
    "log"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    dryRun := flag.Bool("dry-run", false, "Validate only, do not make changes")
    dbPath := flag.String("db", "", "Path to database file")
    flag.Parse()

    db, err := sql.Open("sqlite3", *dbPath)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Validation queries
    orphanedCount := validateExpenseReferences(db)
    if orphanedCount > 0 {
        log.Fatalf("Found %d orphaned expenses - aborting", orphanedCount)
    }

    // Check for associations without any categories
    associations := findAssociationsWithoutCategories(db)
    if len(associations) == 0 {
        log.Println("All associations already have categories")
        return
    }

    log.Printf("Found %d associations needing default categories\n", len(associations))

    if *dryRun {
        log.Println("DRY RUN - No changes made")
        return
    }

    // Seed default categories for new associations
    for _, assocID := range associations {
        err := seedDefaultCategories(db, assocID)
        if err != nil {
            log.Printf("Error seeding association %d: %v\n", assocID, err)
        } else {
            log.Printf("Seeded categories for association %d\n", assocID)
        }
    }

    log.Println("Migration complete")
}

func seedDefaultCategories(db *sql.DB, associationID int64) error {
    defaultCategories := []struct {
        Type   string
        Family string
        Name   string
    }{
        {"Budgeted", "Personnel", "Salary Janitor"},
        {"Budgeted", "Personnel", "Salary Administrator"},
        {"Budgeted", "Utilities", "Internet"},
        {"Budgeted", "Utilities", "Electricity"},
        {"Maintenance", "Repairs", "Building Maintenance"},
        {"Improvement", "Community", "Events"},
    }

    tx, err := db.Begin()
    if err != nil {
        return err
    }

    stmt, err := tx.Prepare(`
        INSERT INTO categories (type, family, name, association_id)
        VALUES (?, ?, ?, ?)
    `)
    if err != nil {
        tx.Rollback()
        return err
    }
    defer stmt.Close()

    for _, cat := range defaultCategories {
        _, err := stmt.Exec(cat.Type, cat.Family, cat.Name, associationID)
        if err != nil {
            tx.Rollback()
            return err
        }
    }

    return tx.Commit()
}
```

**Step 3: Run Migration**
```bash
# Dry run first
./migrate_categories --db=/path/to/prod.db --dry-run

# Execute migration
./migrate_categories --db=/path/to/prod.db
```

**Step 4: Validate Post-Migration**
```sql
-- Verify all expenses still have valid category references
SELECT COUNT(*) as valid_expenses
FROM expenses e
JOIN categories c ON e.category_id = c.id;

-- Verify no duplicate active categories
SELECT association_id, type, family, name, COUNT(*) as count
FROM categories
WHERE is_deleted = FALSE
GROUP BY association_id, type, family, name
HAVING COUNT(*) > 1;

-- Check category distribution
SELECT association_id, COUNT(*) as category_count
FROM categories
WHERE is_deleted = FALSE
GROUP BY association_id;
```

#### Rollback Plan

**Rollback SQL** (`api/scripts/rollback_migration.sql`):
```sql
-- This migration adds indexes only, so rollback is dropping them
BEGIN TRANSACTION;

DROP INDEX IF EXISTS idx_categories_association_deleted;
DROP INDEX IF EXISTS idx_categories_lookup;

-- If categories were created by migration, remove them
-- (Only if migration created new categories for associations)
DELETE FROM categories
WHERE id IN (
  SELECT id FROM categories
  WHERE created_at > '2025-12-XX 00:00:00'  -- Replace with migration date
  AND association_id NOT IN (SELECT DISTINCT association_id FROM expenses)
);

COMMIT;
```

**Application Rollback**:
1. Stop API server
2. Deploy previous binary version
3. Execute rollback SQL script
4. Start API server
5. Verify existing functionality works

### Deployment Runbook

#### Pre-Deployment Checklist

- [ ] All UAT tests passed and signed off
- [ ] All Critical and High bugs fixed
- [ ] Deployment runbook reviewed by team
- [ ] Maintenance window scheduled and communicated (1 week notice)
- [ ] Backup of production database created
- [ ] Rollback artifacts prepared (previous binaries, SQL scripts)
- [ ] Monitoring dashboards configured
- [ ] On-call rotation scheduled for 48-hour post-deploy window

#### Deployment Steps

**1. Staging Deployment (Day -1)**
```bash
# Backend
cd api
git pull origin main
go build -o apc-api-staging
scp apc-api-staging staging-server:/opt/apc/
ssh staging-server "systemctl restart apc-api"

# Frontend
cd ui
npm run build
scp -r dist/* staging-server:/var/www/apc/

# Database migration
ssh staging-server
cd /opt/apc
goose -dir sql/schema sqlite3 apc-staging.db up

# Smoke tests
curl https://staging.apc.example.com/health
# Manual testing of categories page
```

**2. Production Deployment (Maintenance Window)**

**Time: 02:00 AM (low traffic)**

```bash
# Step 1: Announce maintenance (T-0 minutes)
# Send notification email to all users

# Step 2: Backup database (T+0)
ssh prod-server
cd /opt/apc
cp apc-prod.db apc-prod.db.backup-$(date +%Y%m%d-%H%M%S)

# Step 3: Deploy backend (T+5)
cd api
git pull origin main
go build -o apc-api
sudo systemctl stop apc-api
cp apc-api /opt/apc/apc-api
cd /opt/apc
goose -dir sql/schema sqlite3 apc-prod.db up
sudo systemctl start apc-api
sudo systemctl status apc-api

# Step 4: Verify backend (T+10)
curl http://localhost:8080/health
# Check logs: journalctl -u apc-api -f

# Step 5: Run migration script (T+15)
./migrate_categories --db=apc-prod.db --dry-run
./migrate_categories --db=apc-prod.db
# Verify migration logs

# Step 6: Deploy frontend (T+20)
cd ui
npm run build
sudo rsync -av dist/ /var/www/apc/
sudo systemctl reload nginx

# Step 7: Post-deployment validation (T+25)
# Test scenarios:
# - Login as association admin
# - Navigate to /categories
# - Create new category
# - Edit existing category
# - Create expense with new category
# - Verify existing expenses still display categories
# - Test search and filter
# - Test deactivate/reactivate

# Step 8: Announce maintenance complete (T+40)
# Send notification email to all users
```

#### Post-Deployment Monitoring

**First 2 Hours**:
- [ ] Monitor error logs: `journalctl -u apc-api -f`
- [ ] Monitor application metrics dashboard
- [ ] Check error rate: Should remain <1%
- [ ] Monitor support channels (email, chat)
- [ ] Verify no Critical bugs reported

**First 24 Hours**:
- [ ] Run validation queries every 4 hours
- [ ] Monitor user adoption: category creation rate
- [ ] Review user feedback and support tickets
- [ ] Address any High severity issues immediately

**First Week**:
- [ ] Daily check of success metrics
- [ ] Weekly report to stakeholders
- [ ] Triage and prioritize post-launch issues
- [ ] Plan hotfix releases if needed

#### Emergency Rollback Procedure

**Trigger Conditions**:
- Critical data loss detected
- Authentication/authorization bypass
- Application crashes preventing expense entry
- >10% error rate sustained for 15 minutes

**Rollback Steps** (15-minute SLA):
```bash
# 1. Stop current deployment
sudo systemctl stop apc-api

# 2. Restore previous backend
cp /opt/apc/apc-api.previous /opt/apc/apc-api
sudo systemctl start apc-api

# 3. Restore database (if migration caused issue)
cd /opt/apc
cp apc-prod.db apc-prod.db.failed
mv apc-prod.db.backup-YYYYMMDD-HHMMSS apc-prod.db

# 4. Restore frontend
sudo rsync -av /var/www/apc.previous/ /var/www/apc/
sudo systemctl reload nginx

# 5. Verify rollback successful
curl http://localhost:8080/health
# Test critical paths manually

# 6. Announce incident and rollback
# Send notification to users and stakeholders

# 7. Incident post-mortem
# Schedule team meeting within 24 hours
```

---

## Success Metrics

### Development Velocity Metrics

| Metric | Target | Measurement Method | Frequency |
|--------|--------|-------------------|-----------|
| Sprint Velocity | 30-40 story points/week | JIRA/GitHub Issues | Weekly |
| Code Review Turnaround | <1 business day | GitHub PR timestamps | Daily |
| Bug Fix Time (Critical) | <4 hours | Issue tracker | Per incident |
| Bug Fix Time (High) | <24 hours | Issue tracker | Per incident |
| Build Success Rate | >95% | CI/CD pipeline | Per commit |

### Quality Metrics

| Metric | Target | Measurement Method | Tracking |
|--------|--------|-------------------|----------|
| Unit Test Coverage (Backend) | >80% | `go test -cover` | Per PR |
| Unit Test Coverage (Frontend) | >70% | Vitest coverage report | Per PR |
| Integration Test Pass Rate | 100% | Test suite results | Per deployment |
| E2E Test Pass Rate | 100% | Playwright results | Per deployment |
| Critical Bugs in Production | 0 | Issue tracker | Weekly |
| High Bugs in Production | <3 | Issue tracker | Weekly |
| Code Quality Score | A (SonarQube) | Static analysis | Per PR |

### Performance Metrics

| Metric | Target | Measurement Method | Frequency |
|--------|--------|-------------------|-----------|
| Category Page Load Time | <2 seconds | Browser Performance API | Continuous |
| Search Response Time | <500ms | Backend logging | Continuous |
| Category Creation Time | <1 second | Backend logging | Continuous |
| Bulk Operation Time (100 items) | <3 seconds | Backend logging | On-demand |
| API Error Rate | <1% | Error monitoring | Real-time |
| Database Query Time (95th percentile) | <100ms | Query profiling | Daily |

### User Adoption Metrics

| Metric | Target | Measurement Method | Frequency |
|--------|--------|-------------------|-----------|
| % Associations with Custom Categories | >70% within 30 days | Database query | Weekly |
| Average Categories per Association | 15-30 | Database query | Weekly |
| Category Creation Rate | >10 new categories/week | Database query | Weekly |
| Category Deactivation Rate | <5% of total/month | Database query | Monthly |
| Support Tickets (Category-related) | <5/week | Support system | Weekly |
| User Satisfaction Score | >4/5 | UAT surveys | Post-launch |

### Data Integrity Metrics

| Metric | Target | Measurement Method | Frequency |
|--------|--------|-------------------|-----------|
| Expense-Category Link Validity | 100% | Validation query (see below) | Daily |
| Duplicate Active Categories | 0 | Validation query | Daily |
| Orphaned Expenses | 0 | Validation query | Daily |

**Validation Queries**:

```sql
-- Data Integrity Checks (run daily)

-- 1. Check for orphaned expenses (should always return 0)
SELECT COUNT(*) as orphaned_expenses
FROM expenses e
LEFT JOIN categories c ON e.category_id = c.id
WHERE c.id IS NULL;

-- 2. Check for duplicate active categories (should always return 0)
SELECT COUNT(*) as duplicate_categories
FROM (
  SELECT association_id, type, family, name, COUNT(*) as count
  FROM categories
  WHERE is_deleted = FALSE
  GROUP BY association_id, type, family, name
  HAVING COUNT(*) > 1
);

-- 3. Measure category usage distribution
SELECT
  CASE
    WHEN usage_count = 0 THEN 'Unused'
    WHEN usage_count < 10 THEN 'Low'
    WHEN usage_count < 50 THEN 'Medium'
    ELSE 'High'
  END as usage_tier,
  COUNT(*) as category_count
FROM (
  SELECT c.id, COUNT(e.id) as usage_count
  FROM categories c
  LEFT JOIN expenses e ON c.id = e.category_id
  WHERE c.is_deleted = FALSE
  GROUP BY c.id
) usage
GROUP BY usage_tier;

-- 4. Track category creation over time
SELECT
  DATE(created_at) as date,
  COUNT(*) as categories_created
FROM categories
WHERE created_at > DATE('now', '-30 days')
GROUP BY DATE(created_at)
ORDER BY date;
```

### Business Impact Metrics

| Metric | Target | Measurement Method | Frequency |
|--------|--------|-------------------|-----------|
| Support Ticket Reduction | 90% reduction in category-related tickets | Support system comparison | Monthly |
| Category Customization Rate | 80% of associations customize at least 1 category | Database query | Monthly |
| Time to Create Category | <2 minutes average | User session analytics | Monthly |
| Feature Adoption Rate | 90% of active associations use feature | Database query | Monthly |

---

## Conclusion

This implementation plan provides a comprehensive, developer-ready roadmap for building the Expense Taxonomy Management feature. The plan is structured to:

1. **Minimize Risk**: Phased approach with extensive testing and validation
2. **Maximize Efficiency**: Clear task breakdown with parallel work opportunities
3. **Ensure Quality**: 80%+ test coverage, performance benchmarks, security audits
4. **Preserve Data Integrity**: Soft delete only, comprehensive validation queries
5. **Enable Success**: Detailed checklists, success metrics, monitoring strategy

**Key Success Factors**:
- Follow existing architectural patterns (accounts, expenses pages)
- Maintain backward compatibility with existing components
- Comprehensive testing at every phase
- Clear communication with stakeholders
- Robust rollback procedures

**Next Steps**:
1. Review this plan with the development team
2. Confirm resource allocation and timeline
3. Schedule kickoff meeting
4. Begin Phase 1: Database & Backend Foundation

**Questions or Concerns**: Escalate to Tech Lead or Product Manager

---

**Document Version**: 1.0
**Last Updated**: 2025-12-07
**Maintained By**: Tech Lead

---
