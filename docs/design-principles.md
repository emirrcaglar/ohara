# High-End Editorial Design System: The Kinetic Terminal

## 1. Overview & Creative North Star

### The Creative North Star: "The Kinetic Terminal"
This design system is a rejection of the "soft" modern web. It is built for speed, performance, and raw technical authority. Inspired by retro-futurist command centers and low-poly aesthetics, the "Kinetic Terminal" aesthetic prioritizes sharp geometry over organic curves and high-contrast intentionality over subtle shadows. 

We are moving away from the "template" look. Instead of centered, balanced grids, we utilize **intentional asymmetry**. Elements should feel like they were snapped into place on a high-speed mainframe. By leveraging overlapping containers and a brutalist approach to depth, we create a premium experience that feels both nostalgic and cutting-edge, while remaining incredibly lightweight for low-end hardware.

---

## 2. Colors

The color palette is a high-octane mix of absolute depth and radioactive vibrance. We use a "Black-Out" foundation to ensure hardware performance on OLED and low-end displays alike.

### The "No-Line" Rule
**Explicit Instruction:** 1px solid borders for sectioning are strictly prohibited. 
In this system, boundaries are defined exclusively by **background color shifts**. To separate a sidebar from a main content area, do not draw a line; instead, place a `surface-container-low` (#1C1B1B) section against the main `surface` (#131313). This creates a sophisticated, "carved" look rather than a "sketched" one.

### Surface Hierarchy & Nesting
We achieve depth through "Tonal Stacking." The UI should be treated as a series of physical plates:
*   **Base:** `surface` (#131313)
*   **Deep Inset:** `surface_container_lowest` (#0E0E0E) - used for background wells or inactive zones.
*   **Raised Content:** `surface_container_high` (#2A2A2A) - used for active cards or hovered states.
*   **Primary Action:** `primary_container` (#FF8C00) - reserved for the most critical user paths.

### Signature Textures & Glass
While complex gradients are forbidden, use "Digital Grain"—a 2% opacity noise overlay—on large `surface` areas to prevent color banding on low-end screens. For floating overlays (menus/tooltips), use a **hard-edge Glassmorphism**: `surface_container_highest` at 80% opacity with a heavy `backdrop-blur`.

---

## 3. Typography

**Font Family:** Space Grotesk (Monospaced/Tech feel).

The typography is the voice of the system. It should feel like a technical manual curated by a high-fashion editor.

*   **Display (lg/md/sm):** Use for hero moments and media titles. Tracking should be set to `-0.05em` to create a dense, "compressed" high-end feel.
*   **Headline & Title:** Use for section headers. These must always be Uppercase to reinforce the terminal aesthetic.
*   **Body (lg/md/sm):** High-readability monospaced rhythm. Ensure line-height is generous (1.5) to balance the sharp geometry of the UI.
*   **Labels:** Use the Pink accent (`secondary_container`: #FF4A8D) for labels to catch the eye in peripheral vision without distracting from the main content.

---

## 4. Elevation & Depth

### The Layering Principle
Forget Z-index shadows. Hierarchy is achieved by stacking the surface-container tiers. 
*   **Standard:** A `surface_container_low` card sitting on a `surface` background.
*   **Focus:** Transition the card to `surface_container_highest` on hover.

### Ambient Shadows
If a floating element requires a shadow (e.g., a modal), do not use grey. Use a **Tinted Ambient Shadow**:
*   **Color:** `surface_container_lowest` at 40% opacity.
*   **Style:** 0px offset, 20px - 40px blur. This creates a "glow in reverse" effect that feels integrated into the dark theme.

### The "Ghost Border" Fallback
If contrast testing fails for accessibility, use a "Ghost Border": the `outline_variant` token (#564334) at **15% opacity**. It should be felt, not seen.

---

## 5. Components

### Sharp Geometry
**Roundedness Scale:** All `borderRadius` values are `0px`. No exceptions. This reinforces the low-poly, performant vibe.

### Buttons
*   **Primary:** Background: `primary_container` (#FF8C00), Text: `on_primary_container` (#623200). Sharp corners. On hover: Shift to `primary` (#FFB77D).
*   **Secondary:** Background: Transparent, Border: 2px `primary_container`. 
*   **Tertiary (The "Side" Action):** Text: `secondary` (#FFB1C4). No background. Used for "Cancel" or "Secondary Meta" actions.

### Cards & Media Cells
Forbid the use of divider lines. Separate metadata from imagery using the Spacing Scale (specifically `spacing-4` or `spacing-6`). Use `surface_container_low` for the card body to create a subtle "lift" from the background.

### Input Fields
*   **State:** Default uses `surface_container_highest` background.
*   **Focus:** The bottom edge gains a 2px `primary_container` (Orange) bar.
*   **Error:** The label shifts to `error` (#FFB4AB) and the bottom bar becomes `secondary_container` (Pink) for a non-traditional, high-contrast warning.

### Status Indicators (Pink Accents)
Use the Pink `secondary` (#FFB1C4) exclusively for "Live" indicators, "New" badges, or "System Alert" pings. This keeps the orange reserved for "User Intent" and pink for "System Information."

---

## 6. Do’s and Don’ts

### Do:
*   **Use Asymmetry:** Place labels in the top-right of a container while the value is in the bottom-left.
*   **Leverage High Contrast:** Ensure `on_surface` text always sits on a sufficiently dark `surface` tier.
*   **Use Space Grotesk for everything:** Let the monospaced rhythm provide the "grid."

### Don't:
*   **Don't use 1px borders:** Rely on background shifts.
*   **Don't use Rounded Corners:** Even a 2px radius breaks the "Kinetic Terminal" aesthetic.
*   **Don't use traditional "Success Green":** Use the Orange `primary` for success and Pink `secondary` for alerts. This maintains the bespoke brand identity.
*   **Don't use heavy gradients:** If depth is needed, use two solid colors side-by-side (Brutalist style) rather than a smooth transition.
