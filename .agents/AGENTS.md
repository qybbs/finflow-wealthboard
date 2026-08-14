# Rules & Guidelines for FinFlow Wealthboard Agents

This document contains rules and guidelines that all AI agents working on this repository must adhere to.

---

## 1. Frontend Development & Design System Compliance

All frontend features, enhancements, and UI adjustments must follow the project's official Design System.

- **Design System Reference:** All agents MUST read [docs/DESIGN_SYSTEM.md](../docs/DESIGN_SYSTEM.md) before writing any CSS or modifying HTML styles.
- **Design Tokens:** Always use CSS Custom Properties (e.g., `var(--bg-primary)`, `var(--brand-primary)`, `var(--text-primary)`) instead of hardcoding hexadecimal, RGB, or named colors.
- **Theme Reactivity:** Ensure all components support both light and dark modes natively via the `:root[data-theme="dark"]` attribute selectors and CSS variables. Do not create separate layouts for light and dark modes.
- **Transitions:** Every transition and interactive element state change must be smooth, utilizing the transition speed tokens `var(--transition-fast)` or `var(--transition-normal)`.
- **CSS Architecture:**
  - Place common rules, animations, theme tokens, scrollbars, and generic layouts inside `/frontend/src/design-system.css`.
  - Component-specific scoped styling is allowed in Svelte `<style>` blocks, but they must utilize global design tokens.

---

## 2. Code Quality & Codebase Integrity

- **Database Safety:** Ensure Go backend handlers perform validation (like positive numeric values, validation of identifiers) before executing database writes.
- **Error Handling:** Always wrap Svelte `fetch()` requests in try/catch blocks and check `response.ok` before attempting to parse JSON.
