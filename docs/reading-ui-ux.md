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
- Clicking the image zooms in.
- Clicking again zooms out.
- Moving the cursor while zoomed shifts the zoom origin.
- Leaving the page frame resets the zoom.

The reset boundary is the page frame, not the image pixels. This gives the cursor some room to move over the black side padding without dropping the zoom immediately.

Mobile keeps the touch zoom and pan behavior. The desktop click zoom does not replace mobile gestures.
