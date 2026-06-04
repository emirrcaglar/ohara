# Reading UI/UX

## Reader modes

Ohara now supports two manga reading modes:

| Mode | Behavior |
| :--- | :--- |
| Swipe | One page at a time. Mobile uses horizontal swipes. Desktop uses page controls. |
| Scroll | Pages are stacked vertically. Progress follows the page currently in view. |

The mode is saved as a user preference, the same way swipe direction is saved.

## Direction

Swipe direction remains separate from reading mode.

- Left-to-right keeps page order in the natural index order.
- Right-to-left reverses the visual order and button meaning.
- Scroll mode uses the same visual order, so RTL manga still reads in the expected sequence.

## Desktop zoom

Desktop reader images are zoomable.

- Hovering an image shows a magnifier cursor.
- Clicking the image zooms in around the cursor.
- Clicking again zooms out.
- Moving the cursor while zoomed shifts the zoom origin, so the page drifts under the pointer.
- Leaving the page frame resets the zoom in single-page desktop mode.

In scroll mode, zoom behaves like one continuous surface across page boundaries:

- The active zoomed page is always stacked above neighboring pages.
- Neighboring non-zoomed pages do not clip or block the expanded zoomed page.
- When the cursor crosses into the next or previous page, both pages briefly stay zoomed during the handoff.
- During that handoff, both pages use one shared viewport cursor point, so the new page feels appended to the old zoomed page rather than acting as a separate zoom.
- The outgoing page is removed after a short cursor-distance hysteresis, avoiding a snap at the exact boundary.
- If desktop auto-scroll moves a new page under a stationary cursor, zoom focus advances from page layout geometry instead of relying only on `mouseenter`. This prevents pages from being skipped when the enlarged zoomed page visually overlaps the next page.

The reset boundary is the page frame, not the image pixels. This gives the cursor some room to move over the black side padding without dropping the zoom immediately.

Mobile keeps the touch zoom and pan behavior. The desktop click zoom does not replace mobile gestures.
