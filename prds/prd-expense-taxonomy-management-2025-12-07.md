# Product Requirements Document: Expense Taxonomy Management

**Version:** 1.0
**Date:** 2025-12-07
**Status:** Draft
**Author:** Product Management Team
**Product:** APC (Apartment Owners Association) Management Application

---

## Executive Summary

### Overview
This PRD defines the requirements for implementing a comprehensive Expense Taxonomy Management system within the APC application. Currently, expense categories are hardcoded values with a three-level hierarchy (type, family, name). This feature will enable association administrators to manage their own expense taxonomy through an intuitive admin interface while preserving data integrity for existing expenses.

### Business Value
- **Flexibility**: Each association can customize their expense categories to match their specific needs and accounting practices
- **Scalability**: New associations can start with sensible defaults and adjust as needed
- **Data Integrity**: Soft deletion ensures historical expense data remains intact and reportable
- **User Autonomy**: Reduces dependency on technical team for category management
- **Compliance**: Better alignment with local accounting standards and reporting requirements

### Success Metrics
- 100% of existing expenses maintain valid category references post-launch
- Association admins can create/modify categories without technical assistance
- Zero data loss from category deletions (soft delete only)
- Page load time under 2 seconds for category management interface
- 90% reduction in support requests for category-related changes

---

## Product Vision

Enable association administrators to independently manage their expense categorization system through an intuitive, hierarchical interface that respects the existing three-level taxonomy (type → family → name) while ensuring complete data preservation and multi-language support.

---

## Target Users & User Stories

### Primary Users
1. **Association Administrators**: Users with administrative privileges for a specific association who manage day-to-day operations
2. **System Administrators**: Platform-level administrators (is_admin = true) who may need to manage categories across associations

### User Stories

#### US-1: View All Categories (Priority: Must Have)
**As an** association administrator
**I want to** view all active and inactive categories in a structured, hierarchical format
**So that** I can understand my current expense categorization system

**Acceptance Criteria:**
- Categories are displayed in a tree structure: Type → Family → Name
- Active categories are clearly distinguished from inactive (soft-deleted) categories
- Categories are sorted alphabetically within each level
- Both English and Romanian translations are visible or can be toggled
- The interface shows category metadata (created date, last modified, usage count)

#### US-2: Create New Categories (Priority: Must Have)
**As an** association administrator
**I want to** create new expense categories at any level of the hierarchy
**So that** I can categorize expenses according to my association's specific needs

**Acceptance Criteria:**
- Can create new Type values (e.g., "Emergency", "Special Projects")
- Can create new Family values under existing or new Types
- Can create new Name values under existing or new Families
- Form validates that all three fields (type, family, name) are provided
- System prevents duplicate combinations of type-family-name within the association
- Successfully created categories are immediately available in expense forms
- Creation timestamp is recorded
- Both English and Romanian labels can be provided

#### US-3: Deactivate Categories (Priority: Must Have)
**As an** association administrator
**I want to** deactivate categories I no longer use
**So that** they don't clutter dropdown menus while preserving historical data

**Acceptance Criteria:**
- Deactivation is a soft delete (sets is_deleted = TRUE)
- System warns if the category is currently in use by expenses
- Deactivated categories are hidden from expense creation/editing forms
- Deactivated categories remain visible in historical expense views
- Deactivated categories can be filtered/viewed in the management interface
- Usage count is displayed before deactivation confirmation

#### US-4: Reactivate Categories (Priority: Should Have)
**As an** association administrator
**I want to** reactivate previously deactivated categories
**So that** I can resume using categories without creating duplicates

**Acceptance Criteria:**
- Can view list of deactivated categories
- Can reactivate with a single action
- Reactivated categories immediately appear in expense forms
- Updated timestamp reflects reactivation

#### US-5: Edit Category Labels (Priority: Should Have)
**As an** association administrator
**I want to** modify the display names and translations of categories
**So that** I can correct errors or improve clarity without breaking references

**Acceptance Criteria:**
- Can edit type, family, and name fields
- System validates uniqueness after edit
- All existing expenses maintain their category_id references
- Changes are reflected immediately in the UI
- Both English and Romanian translations can be updated
- Audit trail shows modification history

#### US-6: Bulk Operations (Priority: Could Have)
**As an** association administrator
**I want to** perform bulk actions on multiple categories
**So that** I can efficiently manage large taxonomy changes

**Acceptance Criteria:**
- Can select multiple categories via checkboxes
- Can bulk deactivate selected categories
- Can bulk reactivate selected categories
- System shows warning for categories in use before bulk deactivation
- Confirmation dialog shows count of affected categories

#### US-7: Search and Filter (Priority: Should Have)
**As an** association administrator
**I want to** search and filter categories
**So that** I can quickly find specific categories in a large list

**Acceptance Criteria:**
- Can search by type, family, or name (free text)
- Can filter by status (active/inactive)
- Can filter by type value
- Search works across both English and Romanian labels
- Results update in real-time as filters change

#### US-8: View Category Usage (Priority: Should Have)
**As an** association administrator
**I want to** see which categories are actively used in expenses
**So that** I can make informed decisions about deactivation

**Acceptance Criteria:**
- Usage count displays number of expenses using each category
- Can drill down to view list of expenses for a category
- Shows date of last use
- Clearly indicates categories with zero usage

#### US-9: Import Default Categories (Priority: Must Have)
**As a** system administrator
**I want to** provide default category templates for new associations
**So that** new associations can start with sensible defaults

**Acceptance Criteria:**
- New associations receive default Romanian categories on creation
- Default categories are scoped to the specific association
- Defaults can be customized after association creation
- Migration script seeds existing associations with current hardcoded categories

---

## Functional Requirements

### FR-1: Category Data Model
**Priority:** Must Have

The system SHALL maintain the existing three-level category hierarchy:

**Database Schema** (existing):
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

**Additional Requirements:**
- Unique constraint on (type, family, name, association_id) for active categories
- Index on (association_id, is_deleted) for performance
- All fields are required; no NULL values permitted
- Cascade behavior: categories cannot be hard-deleted if referenced by expenses

### FR-2: Category Management API Endpoints
**Priority:** Must Have

The backend SHALL provide RESTful API endpoints:

#### Existing Endpoints (to maintain):
- `GET /v1/api/associations/{associationId}/categories` - Get all active categories
- `GET /v1/api/associations/{associationId}/categories/{categoryId}` - Get single category
- `POST /v1/api/associations/{associationId}/categories` - Create new category
- `PUT /v1/api/associations/{associationId}/categories/{categoryId}/deactivate` - Soft delete category

#### New Endpoints Required:
- `GET /v1/api/associations/{associationId}/categories/all?include_inactive=true` - Get all categories including inactive
- `PUT /v1/api/associations/{associationId}/categories/{categoryId}` - Update category details
- `PUT /v1/api/associations/{associationId}/categories/{categoryId}/reactivate` - Reactivate category
- `GET /v1/api/associations/{associationId}/categories/{categoryId}/usage` - Get usage statistics
- `POST /v1/api/associations/{associationId}/categories/bulk-deactivate` - Bulk deactivate
- `POST /v1/api/associations/{associationId}/categories/bulk-reactivate` - Bulk reactivate

**Validation Rules:**
- Association ID must match authenticated user's association scope
- Category uniqueness validation (type + family + name within association)
- Prevent deactivation if category is the only active one in the association
- Return 409 Conflict if attempting to create duplicate category
- Return 403 Forbidden if user lacks administrative privileges

### FR-3: Category Management UI Page
**Priority:** Must Have

The frontend SHALL provide a dedicated category management page at `/categories` following the existing patterns from `/accounts` and `/expenses`.

**Layout Structure:**
```
┌─────────────────────────────────────────────────────────────┐
│ Page Header: "Expense Categories"                          │
│ Association Selector                                        │
│ [+ Create Category Button]                                 │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│ Filters & Search                                            │
│ [Search] [Status: All/Active/Inactive] [Type Filter]       │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│ Categories Tree/Table                                       │
│ ├─ Budgeted (Type)                                         │
│ │  ├─ Personnel Expenses (Family)                          │
│ │  │  ├─ Salary Janitor (Name) [Edit] [Deactivate] [12]   │
│ │  │  └─ Salary Administrator (Name) [Edit] [Deactivate]   │
│ │  └─ Utilities (Family)                                   │
│ │     └─ Internet (Name) [Edit] [Deactivate] [3]          │
│ ├─ Maintenance (Type)                                      │
│ └─ Improvement (Type)                                      │
└─────────────────────────────────────────────────────────────┘
```

**UI Components Required:**
- **CategoriesList.vue**: Main list/tree component displaying categories
- **CategoryForm.vue**: Modal form for create/edit operations
- **/pages/categories/index.vue**: Page wrapper following accounts pattern
- **CategoryTree.vue**: Tree visualization component using Naive UI NTree
- Status badges for active/inactive categories
- Usage count indicators
- Confirmation dialogs for destructive actions

**Interaction Patterns:**
- Click "Create Category" opens modal form
- Click "Edit" on category row opens pre-populated modal form
- Click "Deactivate" shows confirmation with usage warning
- Inactive categories shown in gray with "Reactivate" option
- Tree nodes are collapsible/expandable
- Inline search filters results without full page reload

### FR-4: Internationalization Support
**Priority:** Must Have

The system SHALL support bilingual (English/Romanian) category display:

**Implementation:**
- Category database stores raw values (e.g., "Budgeted", "personnel", "salary_janitor")
- Translation keys follow pattern: `categories.types.{type}`, `categories.families.{family}`, `categories.names.{name}`
- Management UI allows editing both English and Romanian translations
- Existing LocalizedCategoryDisplay component continues to work
- Category form provides fields for both language translations

**Translation File Structure** (extend existing):
```json
{
  "categories": {
    "title": "Expense Categories",
    "createNew": "Create New Category",
    "editCategory": "Edit Category",
    "types": { "Budgeted": "Budgeted", ... },
    "families": { "personnel": "Personnel Expenses", ... },
    "names": { "salary_janitor": "Salary Janitor", ... }
  }
}
```

### FR-5: Data Migration & Seeding
**Priority:** Must Have

The system SHALL provide migration path for existing data:

**Migration Script Requirements:**
1. Extract all unique type-family-name combinations from current hardcoded values
2. Create category records for each association with existing expenses
3. Backfill association_id for all existing categories
4. Create seed data SQL file with default Romanian categories
5. Create seed template for new associations

**Default Categories Template:**
- Budgeted: Personnel, Security, Administrative, Utilities, Reserve Fund, Miscellaneous
- Maintenance: Building Repairs, Cleaning, Materials, Equipment Repairs
- Improvement: Community Activities, Building Improvements, Accessibility

### FR-6: Validation & Business Rules
**Priority:** Must Have

**Category Creation:**
- All three fields (type, family, name) are mandatory
- No special characters except underscore and hyphen
- Maximum length: 100 characters per field
- Uniqueness check within association scope
- Auto-trim whitespace

**Category Deactivation:**
- Cannot deactivate if it's the last active category in association
- Warning dialog must show usage count
- Can deactivate categories currently in use (soft delete preserves history)
- Deactivated categories hidden from dropdowns but visible in reports

**Category Reactivation:**
- Only deactivated categories can be reactivated
- Reactivation checks for uniqueness conflicts
- If conflict exists, show error and suggest merging or renaming

**Category Editing:**
- Cannot change category if it creates a duplicate
- Changes apply immediately to all UI components
- Existing expense references remain intact (linked by ID)

### FR-7: Permission & Access Control
**Priority:** Must Have

**Authorization Rules:**
- Only users with association membership can view categories for that association
- Only association administrators can create/edit/deactivate categories
- System administrators (is_admin = true) can manage all associations
- Categories are fully isolated per association (no cross-association visibility)

**Frontend Route Protection:**
- `/categories` route requires authentication
- Association selector limits options to user's authorized associations
- Create/Edit/Delete buttons hidden for non-admin users

---

## Non-Functional Requirements

### NFR-1: Performance
**Priority:** Must Have

- Category list page SHALL load within 2 seconds for datasets up to 500 categories
- Search/filter operations SHALL respond within 500ms
- Category creation SHALL complete within 1 second
- API endpoints SHALL handle 100 concurrent requests without degradation
- Database queries SHALL use indexes for association_id and is_deleted

### NFR-2: Security
**Priority:** Must Have

- All API endpoints SHALL validate association membership via JWT token
- Input validation SHALL prevent SQL injection (using sqlc prepared statements)
- XSS protection through Vue's automatic escaping
- Rate limiting: 100 requests per minute per user
- Audit logging for all category modifications (create, update, deactivate, reactivate)

### NFR-3: Accessibility
**Priority:** Should Have

- Interface SHALL meet WCAG 2.1 Level AA standards
- Keyboard navigation for all actions (Tab, Enter, Escape)
- Screen reader support with proper ARIA labels
- Sufficient color contrast (4.5:1 minimum)
- Focus indicators visible on all interactive elements
- Error messages announced to screen readers

### NFR-4: Usability
**Priority:** Must Have

- Interface SHALL follow existing Naive UI design patterns
- Consistent with other admin pages (accounts, expenses, units)
- Informative error messages in user's selected language
- Confirmation dialogs for destructive actions
- Undo option for accidental deactivations (via reactivate)
- Loading states displayed during async operations
- Empty states with helpful guidance

### NFR-5: Compatibility
**Priority:** Must Have

- Support modern browsers: Chrome 90+, Firefox 88+, Safari 14+, Edge 90+
- Responsive design for desktop (1024px+) and tablet (768px+)
- Mobile support (phone <768px) for view-only mode
- Backend compatible with SQLite 3.35+
- Go 1.21+ and Vue 3.3+ compatibility

### NFR-6: Maintainability
**Priority:** Should Have

- Code follows existing project structure and naming conventions
- API handlers use sqlc for type-safe database queries
- Frontend components are modular and reusable
- Comprehensive inline documentation
- Unit tests for business logic (80% coverage target)
- Integration tests for critical paths
- Database migrations use goose framework

### NFR-7: Data Integrity
**Priority:** Must Have

- Soft delete ensures zero data loss
- Foreign key constraints prevent orphaned expense records
- Transaction support for bulk operations
- Database backups include category history
- Audit trail for all category changes (via created_at/updated_at)

---

## Technical Considerations

### Database Changes

**New Indexes Required:**
```sql
CREATE INDEX idx_categories_association_deleted
ON categories(association_id, is_deleted);

CREATE INDEX idx_categories_lookup
ON categories(association_id, type, family, name)
WHERE is_deleted = FALSE;
```

**New Queries** (sqlc format):
```sql
-- name: GetAllCategories :many
SELECT * FROM categories
WHERE association_id = ?
ORDER BY is_deleted, type, family, name;

-- name: UpdateCategory :one
UPDATE categories
SET type = ?, family = ?, name = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND association_id = ?
RETURNING *;

-- name: ReactivateCategory :exec
UPDATE categories
SET is_deleted = FALSE, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: GetCategoryUsageCount :one
SELECT COUNT(*) as usage_count
FROM expenses
WHERE category_id = ?;

-- name: BulkDeactivateCategories :exec
UPDATE categories
SET is_deleted = TRUE, updated_at = CURRENT_TIMESTAMP
WHERE id IN (sqlc.slice('category_ids')) AND association_id = ?;
```

### Backend Architecture

**New/Modified Files:**
```
api/
├── sql/
│   ├── queries/
│   │   └── categories.sql (modify - add new queries)
│   └── schema/
│       └── 00021_add_category_indexes.sql (new)
├── internal/
│   ├── database/
│   │   └── categories.sql.go (regenerate with sqlc)
│   └── handlers/
│       ├── categories.go (modify - add new handlers)
│       └── categories_test.go (new)
└── main.go (modify - add new routes)
```

**New HTTP Handlers:**
- HandleGetAllCategories (with query param for include_inactive)
- HandleUpdateCategory
- HandleReactivateCategory
- HandleGetCategoryUsage
- HandleBulkDeactivateCategories
- HandleBulkReactivateCategories

**Error Handling:**
- 400 Bad Request: Invalid input, validation failures
- 403 Forbidden: Insufficient permissions
- 404 Not Found: Category doesn't exist
- 409 Conflict: Duplicate category, constraint violation
- 500 Internal Server Error: Database errors

### Frontend Architecture

**New Components:**
```
ui/src/
├── pages/
│   └── categories/
│       └── index.vue (new)
├── components/
│   ├── CategoriesList.vue (new)
│   ├── CategoryForm.vue (new)
│   ├── CategoryTree.vue (new)
│   ├── CategorySelector.vue (existing - no changes needed)
│   └── LocalizedCategoryDisplay.vue (existing - no changes needed)
├── services/
│   └── api.ts (modify - add category management endpoints)
└── i18n/
    └── locales/
        ├── en.json (modify - add category management keys)
        └── ro.json (modify - add category management keys)
```

**Component Hierarchy:**
```
/pages/categories/index.vue
├── AssociationSelector (existing)
├── CategoriesList.vue
│   ├── CategoryTree.vue (tree visualization)
│   │   └── LocalizedCategoryDisplay (existing)
│   └── CategoryForm.vue (modal)
│       ├── NForm (Naive UI)
│       ├── NFormItem (type, family, name inputs)
│       └── Action buttons
└── Confirmation Modals (NModal)
```

**State Management:**
- Use Vue 3 Composition API with `ref` and `computed`
- No Vuex/Pinia needed (follow existing pattern)
- Local component state for UI interactions
- API calls via composables/services layer

**Naive UI Components Used:**
- NPageHeader, NCard, NButton, NModal, NForm, NFormItem, NInput
- NTree (for hierarchical display)
- NDataTable (alternative to tree view)
- NSpin (loading states)
- NTag (status badges)
- NPopconfirm (delete confirmations)
- NSpace, NDivider (layout)

### API Service Extensions

**New API Methods** (/ui/src/services/api.ts):
```typescript
export const categoryApi = {
  getCategories: (associationId: number, includeInactive = false) =>
    axios.get(`/v1/api/associations/${associationId}/categories/all?include_inactive=${includeInactive}`),

  getCategory: (associationId: number, categoryId: number) =>
    axios.get(`/v1/api/associations/${associationId}/categories/${categoryId}`),

  createCategory: (associationId: number, data: CategoryCreateDTO) =>
    axios.post(`/v1/api/associations/${associationId}/categories`, data),

  updateCategory: (associationId: number, categoryId: number, data: CategoryUpdateDTO) =>
    axios.put(`/v1/api/associations/${associationId}/categories/${categoryId}`, data),

  deactivateCategory: (associationId: number, categoryId: number) =>
    axios.put(`/v1/api/associations/${associationId}/categories/${categoryId}/deactivate`),

  reactivateCategory: (associationId: number, categoryId: number) =>
    axios.put(`/v1/api/associations/${associationId}/categories/${categoryId}/reactivate`),

  getCategoryUsage: (associationId: number, categoryId: number) =>
    axios.get(`/v1/api/associations/${associationId}/categories/${categoryId}/usage`),

  bulkDeactivate: (associationId: number, categoryIds: number[]) =>
    axios.post(`/v1/api/associations/${associationId}/categories/bulk-deactivate`, { ids: categoryIds }),

  bulkReactivate: (associationId: number, categoryIds: number[]) =>
    axios.post(`/v1/api/associations/${associationId}/categories/bulk-reactivate`, { ids: categoryIds })
}
```

### TypeScript Type Definitions

**Extend** /ui/src/types/api.ts:
```typescript
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

export interface CategoryCreateDTO {
  type: string
  family: string
  name: string
}

export interface CategoryUpdateDTO {
  type: string
  family: string
  name: string
}

export interface CategoryUsage {
  category_id: number
  usage_count: number
  last_used_at: string | null
  expenses: Array<{
    id: number
    description: string
    amount: number
    created_at: string
  }>
}
```

---

## Timeline and Milestones

### Phase 1: Foundation (Week 1)
**Deliverables:**
- Database migration script for indexes
- New SQL queries added to categories.sql
- sqlc regeneration
- Backend handlers implemented
- Unit tests for backend logic
- API documentation updated

**Success Criteria:**
- All API endpoints functional
- 100% of existing expenses maintain category references
- Unit test coverage >80%

### Phase 2: Core UI (Week 2)
**Deliverables:**
- Categories page scaffolding
- CategoriesList component
- CategoryForm component
- Basic CRUD operations working
- i18n translations for EN/RO

**Success Criteria:**
- Can create, view, edit, deactivate categories via UI
- Categories immediately available in expense forms
- No console errors or warnings

### Phase 3: Enhanced Features (Week 3)
**Deliverables:**
- CategoryTree hierarchical view
- Search and filtering
- Bulk operations
- Usage statistics display
- Reactivate functionality
- Empty states and loading indicators

**Success Criteria:**
- All user stories marked "Must Have" and "Should Have" completed
- Responsive design working on desktop and tablet
- Accessibility checklist verified

### Phase 4: Polish & Testing (Week 4)
**Deliverables:**
- Integration tests
- End-to-end testing
- Performance optimization
- Security audit
- User acceptance testing
- Documentation (user guide)

**Success Criteria:**
- All success metrics achieved
- Zero critical bugs
- Performance targets met
- User documentation complete

### Phase 5: Migration & Deployment (Week 5)
**Deliverables:**
- Migration script for existing associations
- Default category seeding
- Production deployment
- Monitoring setup
- Rollback plan tested

**Success Criteria:**
- Zero downtime deployment
- All existing data migrated successfully
- Monitoring alerts configured
- Rollback plan documented

---

## Risks and Mitigation

### Risk 1: Data Integrity During Migration
**Severity:** High
**Probability:** Medium

**Description:** Migration script could fail to preserve existing expense-category relationships, causing data loss or broken reports.

**Mitigation:**
- Comprehensive testing on production database copy
- Dry-run mode for migration script
- Database backup before migration
- Validation queries to verify relationship integrity
- Rollback script prepared
- Gradual rollout (one association at a time for early adopters)

### Risk 2: Performance Degradation with Large Datasets
**Severity:** Medium
**Probability:** Medium

**Description:** Associations with thousands of categories might experience slow page loads or sluggish UI interactions.

**Mitigation:**
- Database indexes on critical columns
- Pagination for category lists (100 items per page)
- Virtual scrolling for tree views
- Lazy loading of usage statistics
- Performance testing with 10,000+ category dataset
- Caching strategy for frequently accessed categories

### Risk 3: User Confusion About Soft Delete
**Severity:** Medium
**Probability:** High

**Description:** Users might not understand why deactivated categories still appear in historical reports, leading to support requests.

**Mitigation:**
- Clear UI labeling: "Deactivate" instead of "Delete"
- Informative tooltips explaining soft delete
- Visual distinction (gray text, strikethrough) for inactive categories
- Help documentation with screenshots
- In-app guidance on first use
- "Why can't I delete this?" FAQ section

### Risk 4: Translation Key Management Complexity
**Severity:** Low
**Probability:** High

**Description:** As associations create custom categories, translation keys might become inconsistent or unmaintainable.

**Mitigation:**
- Auto-generate translation keys from category values (e.g., "Personnel Expenses" → "personnel_expenses")
- Fallback to raw value if translation missing
- Translation management UI showing missing translations
- Export/import functionality for translation files
- Clear naming conventions documented

### Risk 5: Accidental Bulk Deactivation
**Severity:** High
**Probability:** Low

**Description:** User accidentally deactivates all categories, disrupting expense entry operations.

**Mitigation:**
- Confirmation dialog showing affected count
- Preview of categories to be deactivated
- Prevent deactivation of last active category
- Easy reactivation process (one-click restore)
- Audit trail for recovery
- Warning when attempting to deactivate >50% of active categories

### Risk 6: Cross-Browser Compatibility Issues
**Severity:** Low
**Probability:** Low

**Description:** Tree view or other advanced UI components might not work correctly in all supported browsers.

**Mitigation:**
- Use well-tested Naive UI components (NTree)
- Cross-browser testing in CI/CD pipeline
- Fallback to table view if tree rendering fails
- Progressive enhancement approach
- Browser compatibility matrix in documentation

### Risk 7: Scope Creep
**Severity:** Medium
**Probability:** High

**Description:** Stakeholders might request additional features (e.g., category hierarchies beyond 3 levels, category merging, duplicate detection AI).

**Mitigation:**
- Clear PRD with prioritized features (Must/Should/Could/Won't)
- Change request process requiring impact analysis
- Defer "Could Have" features to Phase 2
- Regular stakeholder alignment meetings
- Roadmap visibility showing Phase 2 features

---

## Assumptions

1. **User Permissions**: Users with association membership have appropriate permissions. No new user roles need to be created for this feature.

2. **Database Size**: Maximum 1,000 categories per association. System will not optimize for associations with >10,000 categories.

3. **Translation Workflow**: Associations are responsible for providing both English and Romanian translations. No machine translation will be provided.

4. **Category Hierarchy**: The three-level hierarchy (type → family → name) is sufficient. No support for dynamic hierarchies or additional levels.

5. **Backward Compatibility**: Existing CategorySelector and LocalizedCategoryDisplay components will continue to work without modification.

6. **Migration Timing**: Migration will occur during planned maintenance window with user communication.

7. **Default Categories**: All new associations will receive the same default Romanian category set from samples/categories.sql.

8. **Browser Support**: Users are on modern, evergreen browsers. No IE11 support required.

9. **Mobile Usage**: Category management is primarily a desktop activity. Mobile interface can be view-only or simplified.

10. **Audit Requirements**: created_at and updated_at timestamps are sufficient for audit trails. No detailed change history (field-level diffs) required.

11. **Uniqueness Scope**: Category uniqueness is enforced at the (type, family, name, association_id) level. Same names can exist across different associations.

12. **Hard Delete**: Hard deletion of categories will never be supported to ensure referential integrity with expenses.

---

## Out of Scope (Future Enhancements)

The following features are explicitly out of scope for the initial release but may be considered for future iterations:

1. **Category Merging**: Ability to merge duplicate categories and update all expense references
2. **Category Templates**: Sharing category templates between associations or importing from a marketplace
3. **Advanced Hierarchy**: More than 3 levels or dynamic hierarchy depth
4. **Category Metadata**: Additional fields like description, color coding, icons, budget limits
5. **Expense Recategorization**: Bulk updating of expense categories
6. **Category Analytics**: Spending trends, budget vs actual, forecasting
7. **Category Import/Export**: CSV/Excel import of categories
8. **Category Versioning**: Detailed change history with rollback capability
9. **Duplicate Detection**: AI-powered suggestions for similar categories
10. **Category Workflows**: Approval process for category creation/deletion
11. **Multi-Level Soft Delete**: Deactivating a Type deactivates all child Families and Names
12. **Category Aliases**: Multiple names for the same category (e.g., "Elevator" and "Lift")
13. **Custom Fields**: Association-specific custom fields for categories
14. **Category Notes**: Internal notes or documentation for each category
15. **Category Restrictions**: Limiting which users can use specific categories

---

## Appendix

### A. Existing System Architecture

**Backend Stack:**
- Language: Go 1.21+
- Database: SQLite 3.35+
- Query Builder: sqlc (type-safe SQL)
- HTTP Router: Standard library net/http
- Migration Tool: goose

**Frontend Stack:**
- Framework: Vue 3.3+ (Composition API)
- UI Library: Naive UI
- State Management: Composables (no Vuex/Pinia)
- HTTP Client: Axios
- i18n: vue-i18n
- Build Tool: Vite

**Authentication:**
- JWT tokens
- TOTP two-factor authentication
- User-association membership via users_associations table

### B. Reference Implementations

**Similar Pages to Follow:**
- `/accounts` - Account management page with association selector, list view, modal forms
- `/expenses` - Expense tracking with category selector, filtering, charts
- `/units` - Unit management with CRUD operations
- `/owners` - Owner management with relationship handling

**Components to Reuse:**
- AssociationSelector - Association dropdown (existing)
- CategorySelector - Category dropdown (existing, no changes)
- LocalizedCategoryDisplay - i18n display component (existing, no changes)

### C. Example Category Data Structure

**Type Values:**
- Budgeted (recurring, planned expenses)
- Maintenance (repairs and upkeep)
- Improvement (capital improvements)

**Family Values (sample):**
- Personnel Expenses (Cheltuieli de Personal)
- Security and Safety (Securitate)
- Building Repairs (Reparații Clădire)
- Administrative Services (Servicii Administrative)
- Utilities (Utilități)
- Community Activities (Activități Comunitare)

**Name Values (sample):**
- Salary Janitor (Salariu Maturator)
- Internet Services (Servicii Internet)
- Building Painting (Vopsire Clădire)
- Annual Meeting Expenses (Cheltuieli Adunare Anuală)

### D. UI Mockup (Wireframe)

```
┌───────────────────────────────────────────────────────────────┐
│ ☰ APC Management                                   [EN] User▼ │
├───────────────────────────────────────────────────────────────┤
│                                                               │
│  ← Back   Expense Categories                                 │
│                                                               │
│  Association: [Bloc 123, Str. Example ▼]                     │
│                                                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │ [Search categories...] [Status: All ▼] [+ Create]      │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │ ▼ Budgeted (15 categories)                              │ │
│  │   ▼ Personnel Expenses                                  │ │
│  │     ├─ Salary Janitor              [📝 Edit] [🗑 Deact] │ │
│  │     │  Usage: 24 expenses                               │ │
│  │     ├─ Salary Administrator        [📝 Edit] [🗑 Deact] │ │
│  │     │  Usage: 12 expenses                               │ │
│  │     └─ CNAS Contributions          [📝 Edit] [🗑 Deact] │ │
│  │        Usage: 24 expenses                               │ │
│  │   ▼ Utilities                                           │ │
│  │     └─ Internet Services           [📝 Edit] [🗑 Deact] │ │
│  │        Usage: 3 expenses                                │ │
│  │                                                          │ │
│  │ ▼ Maintenance (8 categories)                            │ │
│  │   ▼ Building Repairs                                    │ │
│  │     └─ Roof Repair                 [📝 Edit] [🗑 Deact] │ │
│  │        Usage: 0 expenses                                │ │
│  │                                                          │ │
│  │ ▶ Improvement (6 categories)                            │ │
│  │                                                          │ │
│  │ ▶ Inactive Categories (2)                               │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                               │
└───────────────────────────────────────────────────────────────┘

Modal Dialog (Create/Edit):
┌─────────────────────────────────────┐
│ Create New Category             [✕] │
├─────────────────────────────────────┤
│ Type:                               │
│ [Budgeted              ▼]           │
│                                     │
│ Family:                             │
│ [Personnel Expenses    ▼]           │
│                                     │
│ Name:                               │
│ [Bonus Payments                  ]  │
│                                     │
│ English Label:                      │
│ [Bonus Payments                  ]  │
│                                     │
│ Romanian Label:                     │
│ [Plăți Bonusuri                  ]  │
│                                     │
│              [Cancel]  [Save]       │
└─────────────────────────────────────┘
```

### E. Success Metrics Tracking

**Metric Dashboard** (to be implemented in analytics):

| Metric | Target | Measurement Method | Tracking Frequency |
|--------|--------|-------------------|-------------------|
| Data Integrity | 100% expense-category links valid | Daily validation query | Daily |
| Admin Self-Service | 90% reduction in support tickets | Support system tags | Weekly |
| Page Load Time | <2 seconds | Browser performance API | Continuous |
| Search Response | <500ms | Backend logging | Continuous |
| User Adoption | 70% of associations create custom category within 30 days | Database query | Monthly |
| Error Rate | <1% of category operations fail | Error logging | Daily |

**Validation Queries:**
```sql
-- Check for orphaned expenses (should always return 0)
SELECT COUNT(*)
FROM expenses e
LEFT JOIN categories c ON e.category_id = c.id
WHERE c.id IS NULL;

-- Check for duplicate active categories
SELECT association_id, type, family, name, COUNT(*)
FROM categories
WHERE is_deleted = FALSE
GROUP BY association_id, type, family, name
HAVING COUNT(*) > 1;

-- Measure category usage distribution
SELECT
  CASE WHEN usage_count = 0 THEN 'Unused'
       WHEN usage_count < 10 THEN 'Low'
       WHEN usage_count < 50 THEN 'Medium'
       ELSE 'High' END as usage_tier,
  COUNT(*) as category_count
FROM (
  SELECT c.id, COUNT(e.id) as usage_count
  FROM categories c
  LEFT JOIN expenses e ON c.id = e.category_id
  WHERE c.is_deleted = FALSE
  GROUP BY c.id
) usage
GROUP BY usage_tier;
```

---

## Glossary

- **APC**: Apartment Owners Association (Asociația Proprietarilor de Condominiu)
- **Category**: A three-level taxonomy for classifying expenses (type → family → name)
- **Soft Delete**: Marking a record as deleted (is_deleted = TRUE) without removing it from the database
- **Hard Delete**: Permanent removal of a record from the database (not permitted for categories)
- **Association**: A legal entity representing an apartment building or complex
- **sqlc**: SQL compiler that generates type-safe Go code from SQL queries
- **goose**: Database migration tool for managing schema changes
- **Naive UI**: Vue 3 component library used throughout the application
- **i18n**: Internationalization (supporting multiple languages)
- **CRUD**: Create, Read, Update, Delete operations
- **JWT**: JSON Web Token used for authentication
- **TOTP**: Time-based One-Time Password (two-factor authentication)
- **WCAG**: Web Content Accessibility Guidelines

---

## Approval & Sign-off

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Product Manager | | | |
| Engineering Lead | | | |
| UX Designer | | | |
| QA Lead | | | |
| Stakeholder Representative | | | |

---

## Document Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-12-07 | Product Team | Initial draft |

---

**Document End**
