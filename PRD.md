# Libra Link Product Requirements Document (PRD)

## 1. Document Information

- **Document Title**: Libra Link PRD (TUI-First MVP)
- **Version**: 1.2
- **Date**: February 13, 2026
- **Author**: Jay with AI support
- **Stakeholders**: Jay as primary developer and user
- **Purpose of Document**: This PRD defines the TUI-first product direction for Libra Link, including MVP scope, backend requirements, user flows, and delivery milestones. Mobile and web are documented as post-MVP expansions.

## 2. Executive Summary

Libra Link is an ebook reader and community sharing platform built around a terminal-first user experience. The MVP delivers a focused **TUI client + backend service**. The system supports personal reading, community sharing, borrowing, and metadata enrichment from Google Books API, while keeping offline-first behavior and sync as core product values.

Key goals:

- Deliver a fast, keyboard-first reading workflow for terminal users.
- Support a community library with sharing, borrowing, ratings/reviews, and reporting.
- Keep reading functional offline with background sync on reconnect.
- Provide consistent backend contracts so future mobile/web clients can reach feature parity.

MVP timeline target: **4-6 months** for solo development.

## 3. Problem Statement and Objectives

### 3.1 Problem Statement

- Existing ebook tools are often platform-locked, community-light, or weak for terminal-first workflows.
- Terminal users want a local-first reading experience with reliable sync and clear keyboard interactions.
- Community discovery and controlled borrowing are usually split across different products.
- Product quality depends on balancing format support, legal constraints around sharing, and resilient sync behavior.

### 3.2 Objectives

- **Functional**: Enable reading, metadata enrichment, sharing/browsing, borrowing with time limits, focused reading modes, and reader customization (themes + typography profiles).
- **User-Centric**: Support offline reading, configurable UI (including theme and typography preferences), and low-friction keyboard interaction.
- **Technical**: Reuse current monorepo architecture (`apps/api` + shared contracts) and add TUI app as first client.
- **Community**: Encourage ethical sharing with reporting and moderation-ready flows.
- **Success Metrics**: Reading session retention, borrow/share activity, sync success rate, and low error rates in reading mode.

## 4. Target Audience and User Personas

### 4.1 Target Audience

- Book lovers, developers, students, and researchers comfortable with terminal tools.
- Users who value keyboard-driven interfaces and offline reliability.
- Community-oriented readers interested in sharing and discovering books.

### 4.2 User Personas (MVP-Relevant)

1. **CLI Enthusiast (Developer Dave)**: Uses terminal daily and wants rapid navigation, offline reading, and resume-on-open.
2. **Community Moderator (Librarian Lisa)**: Shares books, monitors reports, and expects borrowing and review flows to be predictable.

Post-MVP personas:

- Mobile reader and web browser personas remain in roadmap scope for later phases.

## 5. Scope

### 5.1 In Scope (MVP)

- TUI client for reading, library management, and community features.
- Backend APIs to power auth, library, metadata, sharing, borrowing, reviews, reports, and sync.
- Ebook format support: **EPUB + PDF + TXT**.
- Offline-first behavior with background sync and conflict handling.
- Reading mode toggle in TUI: **Normal / Zen** (hard immersive behavior).
- Reader customization in TUI: preset themes, custom color overrides, and typography profile selection persisted through resume and sync.

### 5.2 Out of Scope (Post-MVP)

- Mobile app implementation.
- Web app implementation.
- Zen Mode implementation on mobile/web (requirement defined now, delivery later).
- Advanced recommendation engine.
- Payment integration.
- Full admin moderation dashboard UI.

## 6. Features and Requirements

Features are organized by MVP core, MVP TUI-specific UX, and post-MVP cross-platform parity.

### 6.1 MVP Core Features

1. **Ebook Reading**
   - User Story: As a user, I can open and read my ebooks in terminal with stable navigation and saved progress.
   - Details: EPUB/PDF/TXT rendering, progress persistence, bookmarks/last-position resume.
   - Acceptance Criteria: Accurate rendering per supported format; reopen restores latest position.
   - Priority: High.

2. **Reader Customization (Themes + Typography Profiles)**
   - User Story: As a user, I can personalize reading appearance for comfort and readability.
   - Details: Preset themes (`light|dark|sepia|high_contrast`), custom color overrides for allowlisted tokens, and typography profile selection (`compact|comfortable|large`) applied globally.
   - Behavior: Changes apply without restarting the reader, persist across app restart, and sync through offline reconnect flows.
   - Acceptance Criteria: Theme/profile updates are immediate; customization persists across resume and sync; invalid override keys/values are rejected with clear validation errors.
   - Priority: High.

3. **Library Management**
   - User Story: As a user, I can import, organize, and search my books.
   - Details: Local import, metadata extraction/editing, filtering/searching by title/author/tags/format, and soft-delete/restore lifecycle for user-owned records.
   - Acceptance Criteria: Import and query workflows are keyboard-driven and performant.
   - Priority: High.

4. **Google Books API Metadata Enrichment**
   - User Story: As a user, I can optionally attach Google Books metadata to improve book details.
   - Details: Search by title/ISBN; attach/detach metadata; failure must not block reading. MVP stores one current metadata attachment per ebook (reattach updates/undeletes current record rather than creating history versions).
   - Acceptance Criteria: Graceful handling of API failure/rate limits.
   - Priority: High.

5. **Community Sharing, Borrowing, and Discovery**
   - User Story: As a user, I can share books, browse community content, borrow with expiry, and leave feedback.
   - Details: Share listing, borrow timers, ratings/reviews, report submission for problematic content, and soft-delete support for user-generated community content where recovery is needed.
   - Acceptance Criteria: Borrow expiry enforcement works consistently; one user one rating per share; reports are persisted.
   - Priority: High.

6. **Authentication**
   - User Story: As a user, I can sign up and log in securely.
   - Details: Email/password and Google OAuth, token refresh, logout/logout-all.
   - Acceptance Criteria: Existing auth routes remain source of truth.
   - Priority: High.

7. **Offline-First Sync**
   - User Story: As a user, I can read offline and sync progress/annotations when reconnected.
   - Details: Local cache + sync queue; deterministic hybrid conflict policy. Use LWW for low-risk reader state/preferences (including customization), and per-entity version checks for progress/bookmarks/annotations with explicit conflict responses on version mismatch.
   - Acceptance Criteria: Offline reads are uninterrupted; queued sync retries automatically.
   - Priority: High.

### 6.2 MVP TUI-Specific UX Features

1. **Reading Modes: Normal and Zen**
   - User Story: As a reader, I can toggle between normal and distraction-free reading modes.
   - Details:
     - Toggle key in reading mode: `z`.
     - **Normal Mode**: Standard reading chrome (progress, metadata/context panels as configured).
     - **Zen Mode (Hard Immersive)**:
       - Hide navigation chrome, menus, metadata panels, community widgets, and sync/status clutter.
       - Show reading content only by default.
       - Reveal controls only via explicit user action (toggle/escape command).
     - Theme and typography choices remain active in both modes; Zen only controls chrome visibility and interaction surfaces.
   - State Persistence:
     - Default `reading_mode`: `normal`.
     - Restore mode on reopen when `zen_restore_on_open = true` (default true).
   - Acceptance Criteria:
     - User can switch modes without leaving reading session.
     - Zen mode does not accidentally reveal non-reading UI.
     - Mode persists across resume and sync.
   - Priority: High.

2. **Keyboard-First Navigation**
   - User Story: As a TUI user, I can complete all primary workflows without mouse input.
   - Details: Consistent shortcuts for library, reading, sharing, borrowing, and reporting.
   - Acceptance Criteria: No core path requires pointer interaction.
   - Priority: High.

3. **Customization Panel (Keyboard-First)**
   - User Story: As a user, I can configure theme/colors/typography using only keyboard controls while reading.
   - Details: Open customization panel from reading UI, select preset theme, edit allowlisted color tokens, switch typography profile, preview/apply, and reset to defaults.
   - Acceptance Criteria: Customization workflow is keyboard-only, changes are visible immediately, and reset restores defaults.
   - Priority: High.

4. **Session Resume**
   - User Story: As a user, I can restart the app and continue from my previous context.
   - Details: Restore open book, position, active reading mode, and active customization preferences.
   - Acceptance Criteria: Resume state is deterministic and reliable.
   - Priority: Medium.

### 6.3 Cross-Platform Parity Requirement (Post-MVP)

- Zen Mode semantics must be consistent on mobile and web:
  - Same `normal|zen` mode model.
  - Same hard immersive intent.
  - Platform-specific explicit exit action allowed, but behavior parity required.
- Mobile/web implementation remains out of MVP delivery scope.

## 7. User Flows and Journeys

### 7.1 Key MVP User Flows (TUI)

1. **Onboarding and First Read**
   - Sign up/login -> import ebook -> optional Google metadata attach -> open reader -> start in normal mode.

2. **Reading Session with Mode Toggle**
   - Open book -> read in normal mode -> press `z` for Zen -> read in immersive mode -> explicit action to reveal/exit -> close app -> reopen and restore mode/position.

3. **Reading Session with Customization**
   - Open book -> open customization panel -> select preset theme or custom overrides -> choose typography profile -> continue reading -> close app -> reopen with preferences restored.

4. **Sharing and Borrowing**
   - Share ebook -> set borrow rules -> publish -> another user browses -> borrows -> borrow expires -> access removed.

5. **Community Feedback and Safety**
   - Browse share -> submit rating/review -> optionally report abusive/infringing content.

6. **Offline and Reconnect Sync**
   - Read offline -> queue progress/annotation/mode/customization updates -> reconnect -> background sync applies updates/conflict policy.

## 8. Technical Architecture

### 8.1 High-Level Architecture

- **Current Repo Reality**:
  - `apps/api` (Go/Fiber backend) as source of backend functionality.
  - `packages/zod` and `packages/openapi` for API contract and OpenAPI generation.
- **Planned MVP Addition**:
  - `apps/tui` (Go, Bubble Tea) as first and only client shipped in MVP.
- **Storage and Services**:
  - PostgreSQL for relational data.
  - Redis for caching/jobs/session-related components already present in API architecture.

### 8.2 Data Model Additions

- Existing entities (User, Ebook, Share, Annotation, Borrow) remain.
- User preferences must include:
  - `reading_mode`: `"normal" | "zen"`
  - `zen_restore_on_open`: `boolean` (default `true`)
  - `theme_mode`: `"light" | "dark" | "sepia" | "high_contrast"` (default `dark`)
  - `theme_overrides`: object map for allowlisted tokens (default empty object)
  - `typography_profile`: `"compact" | "comfortable" | "large"` (default `comfortable`)
- Reading session/progress payloads include current `reading_mode` for restore and sync consistency.
- Preference updates must support partial patch semantics and validate allowlisted override tokens and color value formats.
- Sync conflict model is hybrid for MVP:
  - LWW (server timestamp) for `user_preferences` and `user_reader_state`.
  - Versioned updates for `reading_progress`, `bookmarks`, and `annotations` using row versions and conditional writes.
  - Conflict responses must return server entity + version so clients can resolve deterministically.
- Soft-deletion model is required for user-generated and recoverable entities using `deleted_at` timestamps.
  - Soft-delete by default for: users, ebooks, authors, tags, ebook metadata attachments, reading progress, bookmarks, annotations, shares, and share reviews.
  - Queries for active records must filter `deleted_at IS NULL`.
  - Uniqueness constraints for soft-deleted entities should use active-record semantics (partial unique indexes where needed).

### 8.3 Backend API Surface (MVP Targets)

- Auth: existing `/api/v1/auth/*` endpoints.
- Library: create/list/get/update/delete books.
- Metadata: attach/detach Google Books metadata.
- Community: share create/list/get, borrow/return, reviews/ratings, reports.
- Sync: progress + annotations + preferences/session bootstrap.
- Preferences: if no existing preferences endpoint, add `PATCH /api/v1/users/preferences`.
  - Include `reading_mode`, `zen_restore_on_open`, `theme_mode`, `theme_overrides`, and `typography_profile`.
  - Enforce enum validation for mode/profile and allowlist + format validation for `theme_overrides`.

### 8.4 TUI Technology Stack Decision (MVP Locked)

- Primary stack for `apps/tui`:
  - `bubbletea` for app loop and state transitions.
  - `bubbles` for reusable keyboard-first UI components.
  - `lipgloss` for theme tokens and runtime style switching.
- Local-first persistence and sync queue:
  - SQLite as the on-device store for library cache, session resume state, preferences, and offline sync events.
  - `sqlc` as the required data access layer for typed SQL queries and deterministic behavior.
  - Local schema must model event queue replay for hybrid sync conflict handling (LWW + versioned entities).
- API contract integration:
  - TUI client should consume generated Go API types/client from OpenAPI (`apps/api/static/openapi.json`) to preserve backend contract parity.
- Why this stack is selected:
  - Matches keyboard-first UX and deterministic state requirements for Normal/Zen mode.
  - Supports near-instant theme and typography updates in active reading sessions.
  - Reduces solo-dev MVP risk by using established Go TUI patterns while keeping offline sync explicit and testable.

## 9. Non-Functional Requirements

### 9.1 Performance

- API response target: < 500ms for common reads.
- Reader startup target: fast enough for terminal workflow (practical target < 2s for normal books).
- Mode toggle target: near-instant in active reading session.
- Theme/profile apply target: near-instant in active reading session.

### 9.2 Security and Compliance

- Enforce HTTPS, auth/session best practices, and input sanitization.
- Preserve legal disclaimer and user acknowledgment for shared content.
- Reporting pipeline must persist report lifecycle data during normal operations; explicit hard purge flows are allowed to delete reports and related moderation history.

### 9.3 Accessibility

- TUI keyboard navigation is mandatory.
- Reading mode must preserve readability in both normal and Zen modes.
- High-contrast preset must be available.
- Custom color overrides must pass contrast/readability validation against minimum thresholds.

### 9.4 Reliability

- Offline reading should not be blocked by network failures.
- Sync queue retries must be deterministic and observable.
- Offline customization updates must queue and sync deterministically on reconnect.
- Soft-deleted records must not appear in normal read paths while remaining recoverable for restore/audit workflows, except when explicit hard purge is invoked.

### 9.5 Internationalization

- English-first MVP; data structures should remain i18n-ready.

## 10. Assumptions, Risks, and Dependencies

### 10.1 Assumptions

- Users are responsible for legal rights of uploaded/shared ebooks.
- Google API keys are available for metadata enrichment.
- MVP remains TUI-only even though mobile/web parity requirements are documented.
- Typography preferences in MVP are app-level rendering profiles, not guaranteed OS-level font family switching.
- Hard deletion is reserved for explicit purge paths; default delete behavior in MVP is soft delete.
- Hard purge is allowed to remove `share_reports` and related moderation history when explicitly invoked.

### 10.2 Risks

- Legal risk from user-shared content.
- Format-specific rendering issues, especially PDF behavior in terminal contexts.
- Sync conflicts causing perceived data inconsistency if conflict policy is unclear.
- Unreadable custom theme overrides if validation is too permissive.

### 10.3 Dependencies

- External: Google Books API, Google OAuth, ebook parsing/rendering libraries.
- Internal: Existing API architecture, shared Zod/OpenAPI contract pipeline.
- TUI stack dependencies: Bubble Tea ecosystem (`bubbletea`, `bubbles`, `lipgloss`), SQLite driver, and `sqlc` code generation.

## 11. Roadmap and Milestones

- **Phase 1 (Month 1-2)**: Backend hardening for auth/library/sync contracts, TUI app scaffold.
- **Phase 2 (Month 2-3)**: Reader engine (EPUB/PDF/TXT), library UX, offline cache, and reader customization (themes + typography profiles).
- **Phase 3 (Month 3-4)**: Community sharing/borrowing, ratings/reviews/reports, Zen mode in TUI, and customization hardening (validation/accessibility).
- **Phase 4 (Month 4-6)**: Stability, test coverage, packaging, and launch readiness.
- **Post-MVP**: Mobile/web clients, including Zen mode parity and broader product expansion.

## 12. Acceptance Tests (PRD-Level)

1. User toggles normal <-> zen in active reading session.
2. Zen mode hides all non-essential UI; controls appear only via explicit action.
3. Reading mode persists across session resume and reconnect sync.
4. Borrowed content expires exactly at configured time.
5. One user can only maintain one active rating per shared book (update allowed).
6. Reporting flow records issue and links it to reporting user/share.
7. Offline reading works for downloaded EPUB/PDF/TXT and syncs once online.
8. Normal mode behavior remains unchanged by Zen mode introduction.
9. Soft-deleted records are hidden from normal list/get APIs and can be restored when business rules allow.
10. Sync conflict handling follows the hybrid model: LWW for preferences/reader state, version conflicts for progress/bookmarks/annotations.
11. Metadata attach/detach keeps one current active Google metadata record per ebook in MVP.
12. User can apply a preset `theme_mode` and it persists across restart/resume.
13. User can apply `theme_overrides`; valid values persist locally and across reconnect sync.
14. Invalid `theme_overrides` keys/values are rejected with validation errors.
15. User can switch `typography_profile`; reading layout updates immediately and persists.
16. Zen mode behavior remains immersive while preserving active theme/profile styling.
17. Offline customization changes queue and sync deterministically once online.
18. Concurrent preference edits resolve by LWW for `user_preferences`.
