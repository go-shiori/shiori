# Shiori E2E User Stories

Functional user stories for end-to-end testing of the Shiori bookmark manager.
Each story represents a discrete action a single user can perform through the UI or API.

There are two test layers:
- **Server** (`e2e/server/`): API-level tests via HTTP calls
- **Playwright** (`e2e/playwright/`): Browser UI tests via Playwright

Legend:
- ✅ Covered by existing e2e tests
- ❌ Not yet covered
- n/a Not applicable for this test layer

---

## 1. Authentication

| # | User Story | Server | Playwright |
|---|-----------|--------|------------|
| 1.1 | As a user, I can log in with valid credentials and be redirected to the library | ✅ | ✅ |
| 1.2 | As a user, I see an error when logging in with a wrong username | ✅ | ✅ |
| 1.3 | As a user, I see an error when logging in with a wrong password | ✅ | ✅ |
| 1.4 | As a user, I see a validation error when submitting an empty username | ✅ | ✅ |
| 1.5 | As a user, I can log out and be returned to the login page | ✅ | ✅ |
| 1.6 | As a user, I can check "Remember me" so my session lasts longer | ✅ | ✅ |
| 1.7 | As a user, I am redirected to the login page when my session expires | ✅ | ✅ |
| 1.8 | As a user, I am redirected to my original destination after logging in (deep-link redirect) | n/a | ✅ |

## 2. Account Management (Admin)

| # | User Story | Server | Playwright |
|---|-----------|--------|------------|
| 2.1 | As an admin, I can create a new admin account | ✅ | ✅ |
| 2.2 | As an admin, I can create a new regular user account | ✅ | ✅ |
| 2.3 | As an admin, I can delete a user account | ✅ | ✅ |
| 2.4 | As an admin, I can change another user's password | ✅ | ✅ |
| 2.5 | As an admin, I can see the list of all accounts in settings | ✅ | ✅ |
| 2.6 | As an admin, I can see the system information panel | ✅ | ✅ |
| 2.7 | As an admin, I can change a user's role (promote/demote admin) | ✅ | n/a |
| 2.8 | As an admin, I cannot delete my own account | ✅ | ✅ |

## 3. Account Management (Self-service)

| # | User Story | Server | Playwright |
|---|-----------|--------|------------|
| 3.1 | As a user, I can change my own password (providing the current one) | ✅ | ✅ |
| 3.2 | As a user, I cannot see the system information panel | ✅ | ✅ |
| 3.3 | As a user, I can see my account settings | ✅ | ✅ |
| 3.4 | As a user, I cannot access the admin account management section | ✅ | ✅ |
| 3.5 | As a user, I see an error when providing the wrong current password | ✅ | ✅ |

## 4. Bookmark Creation

| # | User Story | Server | Playwright |
|---|-----------|--------|------------|
| 4.1 | As a user, I can add a bookmark by providing a URL | ✅ | ✅ |
| 4.2 | As a user, I can add a bookmark with a custom title and excerpt | ✅ | ✅ |
| 4.3 | As a user, I can add a bookmark and request archive creation | ✅ | ✅ |
| 4.4 | As a user, I can add a bookmark and request ebook/EPUB generation | ✅ | ✅ |
| 4.5 | As a user, I can add a bookmark and assign tags during creation | ✅ | ✅ |
| 4.6 | As a user, I can add a bookmark marked as public | ✅ | ✅ |
| 4.7 | As a user, I see the new bookmark appear in my library after creation | ✅ | ✅ |
| 4.8 | As a user, I see an error when adding a bookmark with an invalid URL | ✅ | ✅ |

## 5. Bookmark Viewing & Reading

| # | User Story | Server | Playwright |
|---|-----------|--------|------------|
| 5.1 | As a user, I can see my bookmarks listed in the library | ❌ | ❌ |
| 5.2 | As a user, I can switch between list view and card view | ❌ | ❌ |
| 5.3 | As a user, I can view the readable/extracted content of a bookmark | ❌ | ❌ |
| 5.4 | As a user, I can view the archived version of a bookmark | ❌ | ❌ |
| 5.5 | As a user, I can download the EPUB version of a bookmark | ❌ | ❌ |
| 5.6 | As a user, I can see bookmark thumbnails in the library | ❌ | ❌ |
| 5.7 | As a user, I can see bookmark details (title, excerpt, URL, tags) | ❌ | ❌ |

## 6. Bookmark Editing

| # | User Story | Server | Playwright |
|---|-----------|--------|------------|
| 6.1 | As a user, I can edit a bookmark's title | ❌ | ❌ |
| 6.2 | As a user, I can edit a bookmark's excerpt | ❌ | ❌ |
| 6.3 | As a user, I can change a bookmark's public/private status | ❌ | ❌ |
| 6.4 | As a user, I can update a bookmark's tags | ❌ | ❌ |
| 6.5 | As a user, I can re-fetch a bookmark's content/archive | ❌ | ❌ |

## 7. Bookmark Deletion

| # | User Story | Server | Playwright |
|---|-----------|--------|------------|
| 7.1 | As a user, I can delete a single bookmark | ❌ | ❌ |
| 7.2 | As a user, I am asked to confirm before deleting a bookmark | ❌ | ❌ |
| 7.3 | As a user, I can select and delete multiple bookmarks at once | ❌ | ❌ |
| 7.4 | As a user, the bookmark disappears from the library after deletion | ❌ | ❌ |

## 8. Tag Management

| # | User Story | Server | Playwright |
|---|-----------|--------|------------|
| 8.1 | As a user, I can view the list of all tags | ❌ | ❌ |
| 8.2 | As a user, I can create a new tag | ❌ | ❌ |
| 8.3 | As a user, I can rename a tag | ❌ | ❌ |
| 8.4 | As an admin, I can delete a tag | ❌ | ❌ |
| 8.5 | As a user, I can see the bookmark count per tag | ❌ | ❌ |
| 8.6 | As a user, I can add a tag to a bookmark | ❌ | ❌ |
| 8.7 | As a user, I can remove a tag from a bookmark | ❌ | ❌ |
| 8.8 | As a user, I can bulk-update tags across multiple bookmarks | ❌ | ❌ |

## 9. Search & Filtering

| # | User Story | Server | Playwright |
|---|-----------|--------|------------|
| 9.1 | As a user, I can search bookmarks by keyword | ❌ | ❌ |
| 9.2 | As a user, I can filter bookmarks by one or more tags | ❌ | ❌ |
| 9.3 | As a user, I can exclude bookmarks with specific tags | ❌ | ❌ |
| 9.4 | As a user, I can search tags by name | ❌ | ❌ |
| 9.5 | As a user, I see results update as I type (debounced search) | ❌ | ❌ |
| 9.6 | As a user, I see a message when no results match my search | ❌ | ❌ |

## 10. Pagination

| # | User Story | Server | Playwright |
|---|-----------|--------|------------|
| 10.1 | As a user, I can navigate between pages of bookmarks | ❌ | ❌ |
| 10.2 | As a user, I can navigate between pages of tags | ❌ | ❌ |
| 10.3 | As a user, pagination resets when I change search/filter criteria | ❌ | ❌ |

## 11. User Preferences / Settings

| # | User Story | Server | Playwright |
|---|-----------|--------|------------|
| 11.1 | As a user, I can change the application theme (light/dark/system) | ❌ | ❌ |
| 11.2 | As a user, I can change the display language | ❌ | ❌ |
| 11.3 | As a user, I can set my default view mode (list/card) | ❌ | ❌ |
| 11.4 | As a user, I can toggle showing bookmark IDs | ❌ | ❌ |
| 11.5 | As a user, I can toggle hiding thumbnails | ❌ | ❌ |
| 11.6 | As a user, I can toggle hiding excerpts | ❌ | ❌ |
| 11.7 | As a user, I can set default bookmark creation options (archive, ebook, public) | ❌ | ❌ |
| 11.8 | As a user, my preferences persist across sessions | ❌ | ❌ |

## 12. Public Bookmarks

| # | User Story | Server | Playwright |
|---|-----------|--------|------------|
| 12.1 | As an unauthenticated visitor, I can view a public bookmark's content | ❌ | ❌ |
| 12.2 | As an unauthenticated visitor, I cannot view a private bookmark | ❌ | ❌ |
| 12.3 | As an unauthenticated visitor, I cannot access the library or settings | ❌ | ❌ |

## 13. System / Health

| # | User Story | Server | Playwright |
|---|-----------|--------|------------|
| 13.1 | As an operator, I can verify the server is alive via the liveness endpoint | ✅ | n/a |
| 13.2 | As an operator, I can ping the API to check availability | ❌ | n/a |
| 13.3 | As an admin, I can view system info (version, database, OS) in settings | ❌ | ❌ |

## 14. API Authentication (v1)

| # | User Story | Server | Playwright |
|---|-----------|--------|------------|
| 14.1 | As an API consumer, I can authenticate and receive a token | ✅ | n/a |
| 14.2 | As an API consumer, I receive an error for invalid credentials | ✅ | n/a |
| 14.3 | As an API consumer, I can refresh my authentication token | ❌ | n/a |
| 14.4 | As an API consumer, I receive 401 when using an expired token | ❌ | n/a |

---

## Coverage Summary

| Area | Server | Playwright | Total Stories |
|------|--------|------------|---------------|
| Authentication | 7/8 | 8/8 | 8 |
| Account Management (Admin) | 7/8 | 7/8 | 8 |
| Account Management (Self-service) | 5/5 | 5/5 | 5 |
| Bookmark Creation | 8/8 | 8/8 | 8 |
| Bookmark Viewing & Reading | 0/7 | 0/7 | 7 |
| Bookmark Editing | 0/5 | 0/5 | 5 |
| Bookmark Deletion | 0/4 | 0/4 | 4 |
| Tag Management | 0/8 | 0/8 | 8 |
| Search & Filtering | 0/6 | 0/6 | 6 |
| Pagination | 0/3 | 0/3 | 3 |
| User Preferences / Settings | 0/8 | 0/8 | 8 |
| Public Bookmarks | 0/3 | 0/3 | 3 |
| System / Health | 1/3 | 0/3 | 3 |
| API Authentication (v1) | 3/4 | 0/4 | 4 |
| **Total** | **31/80** | **28/80** | **80** |
