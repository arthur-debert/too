# Bug: Multi-line todos are not properly indented in output

## Description

When a todo item contains multiple lines (newlines in the text), the output doesn't properly indent the continuation lines. This makes it difficult to visually distinguish where one todo ends and another begins, especially in lists with many items.

## Current Behavior

Multi-line todos display with subsequent lines starting at the beginning of the line:

```
1 | ✕ Hello, item one
2 | ✕ Hello, Item two, multiline
continues here
3 | ✕ Hello, item three
```

This breaks the visual hierarchy and makes it unclear that "continues here" is part of item 2.

## Expected Behavior

Subsequent lines of multi-line todos should be indented to align with the todo text:

```
1 | ✕ Hello, item one
2 | ✕ Hello, Item two, multiline
      continues here
3 | ✕ Hello, item three
```

## Affected Formatters

- Terminal output formatter (default)
- Markdown output formatter

## Steps to Reproduce

1. Create a multi-line todo:
   ```bash
   tdh add "First line
   second line
   third line"
   ```

2. List todos:
   ```bash
   tdh list
   ```

3. Observe that the second and third lines are not indented

## Impact

This affects the readability of todo lists, especially for users who use multi-line descriptions to provide more context or break down complex tasks. The lack of proper indentation makes it harder to scan through todo lists quickly.

## Proposed Solution

1. Update the terminal formatter to detect newlines in todo text and indent subsequent lines
2. Update the markdown formatter to handle multi-line todos appropriately
3. Calculate the proper indentation based on the position number width and symbols

## Testing

- Unit tests for formatters with multi-line todo text
- Integration tests verifying output formatting
- Test with nested todos to ensure proper indentation at all levels