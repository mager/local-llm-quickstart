# DESIGN.md

## Design Intent

`local-llm-quickstart` should feel like a focused terminal instrument.

It lives in a developer's terminal beside `llama-server`, logs, editors, and project shells. It should be quiet enough for real work, but polished enough that long model answers feel readable instead of dumped.

The UI should say: local, capable, lightweight, precise.

## Scene

A developer is using this in a dark terminal on a laptop or external monitor, likely with multiple panes open. They are testing prompts, watching local model logs, and deciding whether a local runtime is useful enough for a real project.

This scene favors a dark terminal aesthetic, restrained contrast, strong text hierarchy, and low visual noise.

## Design Register

Product UI.

This is a repeated-use tool, not a marketing surface. Density is welcome. Decoration is suspect. Every flourish must improve scanning, orientation, or confidence.

## Visual Strategy

Restrained terminal color with two accents:

- Cool cyan for status and structure.
- Lime for assistant identity.

The palette should feel crisp on a dark terminal without turning into neon overload.

## Current Palette

These colors are implemented in Go via Lip Gloss and Glamour.

```text
Header text        #c7d2fe
Status pill bg     #7dd3fc
Status pill fg     #111827
Meta text          #8a93a6
Help/footer text   #9aa6b2
Input hint         #cbd5e1
User label         #93c5fd
Assistant label    #bef264
Error text         #fca5a5
Markdown body      #cbd5e1
Markdown heading   #38bdf8
Markdown strong    #f8fafc
Markdown emphasis  #e9d5ff
Markdown rule      #475569
Inline code fg     #bef264
Inline code bg     #1f2937
Code block fg      #dbeafe
Code block bg      #111827
Blockquote         #a7f3d0
```

## Layout

The app has four vertical regions:

1. Header
2. Conversation viewport
3. Input textarea
4. Footer

All regions share a two-character horizontal gutter.

This gutter is not decoration. It keeps the left edge calm and gives Glamour-rendered Markdown room to breathe.

## Header

The header should answer three questions at a glance:

- What app am I in?
- What token mode is active?
- What endpoint am I talking to?

Current structure:

```text
local-llm  [tokens=auto last=4096]  http://127.0.0.1:8080  temp=0.70
```

The token mode is styled as a pill because it changes user expectations. If the model cuts off, the user needs to know whether auto mode or a fixed limit caused it.

## Footer

The footer is command memory, not documentation.

Keep it one line when possible:

```text
pgup/pgdn scroll · ctrl+u/d half · ctrl+g/b top/bottom · /continue · /tokens auto|4096 · esc quit
```

When the model is generating, prepend a status pill:

```text
[thinking · 4096 tokens] pgup/pgdn scroll · ...
```

Avoid stuffing every command into the footer. `/help` can carry the full list.

## Input

The textarea should feel light.

Use:

```text
› Ask a coding question...
```

Avoid heavy rails, gray slabs, or cursor-line backgrounds. The input should be visible but not louder than the answer.

Current rules:

- Prompt: `› `
- Height: 3 lines
- No line numbers
- No cursor-line background
- High-contrast placeholder

## Conversation Labels

Labels are intentionally tiny:

```text
you
llm
```

They create orientation without pretending to be chat bubbles.

Rules:

- `you` is blue.
- `llm` is lime.
- Labels are lowercase.
- Do not wrap messages in boxes.
- Do not use role avatars.

## Markdown Rendering

Assistant output should be rendered with Glamour.

The model naturally returns Markdown. Raw Markdown makes the TUI feel unfinished. Rendered Markdown turns generated output into something scan-friendly.

Current approach:

- User messages use plain wrapping.
- Assistant messages use Glamour.
- Glamour word wrap is tied to viewport width.
- A custom Glamour style is used instead of `WithAutoStyle`.

## Markdown Style Rules

### Headings

Use compact symbolic prefixes:

```text
◆ H1/H2
▸ H3
• H4
```

Do not preserve raw `###` prefixes. They read like source text, not rendered UI.

### Lists

Lists should be clean and compact:

```text
  • item
  • item
```

Nested lists should remain readable, but do not over-indent them. Terminal width is precious.

### Strong Text

Strong text should be brighter, not loud.

Use strong text to support scanability in generated itineraries, plans, and code explanations.

### Tables

Tables are useful but dangerous in terminals.

Rules:

- Keep separators subtle.
- Wrap the viewport before the table hits the right edge.
- Prefer readable columns over decorative grid density.

If model-generated tables keep looking awkward, consider adding a system hint that asks for lists instead of wide tables.

### Code

Inline code should have a subtle dark background and lime text.

Code blocks should use a dark background and syntax highlighting through Glamour/Chroma.

The code block should feel distinct but not like a giant card.

## Spacing

The TUI should use varied vertical rhythm:

- Header: compact
- Conversation: breathable
- Markdown paragraphs: one-line margin
- Footer: compact
- Input: small but comfortable

Avoid making every region equally padded. Equal padding everywhere feels mechanical.

## What To Avoid

- Raw Markdown markers in assistant output.
- Heavy boxes around messages.
- Chat bubble metaphors.
- Gradient terminal text.
- Decorative side rails.
- Huge banners.
- Footer command soup.
- Low-contrast placeholder text.
- Wrapping that touches the right edge.
- Tables wider than the viewport.

## Interaction Rules

### Sending

Enter sends.

Shift+Enter inserts a newline.

### Quit

`Ctrl+C` and `Esc` should quit, even during a request.

### Scroll

Long answers must be navigable:

```text
PageUp / PageDown
Ctrl+U / Ctrl+D
Ctrl+G top
Ctrl+B bottom
Mouse wheel
```

### Continue

When a response hits the token cap, the app should offer `/continue`.

The user should not have to manually craft a continuation prompt.

## Token UX

Token budget is part of the interface.

The app should show the current mode:

```text
tokens=auto next~548
tokens=auto last=4096
tokens=2048
```

Auto mode should be understandable, not mysterious. It can make a rough guess based on prompt length and task words like "itinerary", "draft", "detailed", "code", or "debug".

Manual override should remain obvious:

```text
/tokens 4096
/tokens auto
```

## Error UX

Errors should be direct:

Good:

```text
request local model: connection refused
```

Better if contextual:

```text
Request failed. Is llama-server running?
```

Avoid jokes or cute failure states. Local runtime errors are already confusing enough.

## Accessibility

Terminal accessibility mostly means contrast, restraint, and predictable keyboard behavior.

Rules:

- Do not rely on color alone for state.
- Keep command labels textual.
- Keep contrast high for placeholder and footer text.
- Preserve plain text transcripts.
- Avoid animated or blinking status except the terminal cursor.

## Implementation Notes

Important files:

```text
cmd/local-llm/main.go
internal/tui/model.go
internal/llm/client.go
```

Rendering stack:

```text
Bubble Tea       app loop
Bubbles textarea input
Bubbles viewport output
Lip Gloss        app chrome
Glamour          assistant Markdown
```

The TUI should keep using these Charmbracelet libraries rather than introducing an unrelated UI framework.

## Future Design Work

### Near-term

- Add a `/theme` command for `default`, `plain`, and `dense`.
- Add a compact mode for small terminal panes.
- Add a "response complete" status with elapsed time if usage data becomes available.
- Add a prettier empty state after `/reset`.

### Maybe

- Streaming output.
- Search within transcript.
- Copy last response.
- Markdown table fallback for narrow terminals.

### Probably not

- Message cards.
- Avatars.
- Mouse-first controls.
- Full-screen dashboards.
- Decorative ASCII art.

## Quality Bar

A screenshot of the app should look like a purpose-built terminal product.

It should not look like:

- Raw stdout
- A debug log
- A generic chat clone
- A web app squeezed into a terminal
- A README renderer with an input bolted underneath

The best version feels like a small, sharp local AI workbench.

