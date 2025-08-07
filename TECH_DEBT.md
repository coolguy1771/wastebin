# Technical Debt Analysis

This document outlines areas of technical debt identified in the codebase. Addressing these issues will improve maintainability, performance, and developer experience.

## Frontend

### 1. `useForm` Hook Complexity
The `useForm` hook in `web/src/hooks/useForm.ts` is a powerful but complex piece of logic. While it has been made more type-safe, it could be simplified and improved.

*   **Recommendation:**
    *   Consider breaking down the hook into smaller, more focused hooks (e.g., `useFormValues`, `useFormValidation`).
    *   Provide more targeted helper functions for different input types (e.g., `useCheckbox`, `useSelect`) instead of the generic `setValue`.
    *   Improve the documentation for the hook to make it easier to use.

### 2. Styling with `sx` Prop
Many components use the `sx` prop for styling. While convenient for small-scale styling, it can lead to inconsistent and hard-to-maintain styles in a larger application.

*   **Recommendation:**
    *   Adopt a more systematic approach to styling, such as CSS-in-JS (e.g., Emotion's `styled` components) or CSS Modules.
    *   Create a theme file with a more comprehensive set of design tokens (colors, spacing, typography) to ensure consistency.

### 3. Magic Strings
There are several instances of "magic strings" in the codebase, such as local storage keys (`wastebin-theme-mode`) and API routes.

*   **Recommendation:**
    *   Centralize all constants in a dedicated `constants.ts` file or a set of files. This will make them easier to manage and prevent typos.

## Backend

### 1. Overly Strict Linting
The `.golangci.yml` configuration enables all linters (`default: all`). This is a good starting point, but it can create a lot of noise and make it difficult to contribute code without fighting the linter.

*   **Recommendation:**
    *   Curate a specific set of linters that are most valuable for the project. This will reduce noise and focus on the most important issues.
    *   Consider introducing a tool like `revive` with a more granular configuration.

### 2. Basic Error Handling
The error handling in the backend is basic. Errors are often returned as simple strings, which makes debugging difficult.

*   **Recommendation:**
    *   Implement structured logging (e.g., using a library like `zap` or `zerolog`). This will allow for more detailed and searchable logs in a production environment.
    *   Define a set of custom error types to provide more context about what went wrong.

### 3. Data Access Layer
The database access is sometimes done directly in the HTTP handlers. This couples the handlers to the database implementation and makes the code harder to test and maintain.

*   **Recommendation:**
    *   Introduce a dedicated data access layer (DAL) or repository pattern. This would abstract the database logic from the business logic in the handlers, making the code more modular and easier to test.
