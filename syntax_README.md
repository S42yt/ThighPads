# ThighPads Syntax Highlighting

ThighPads now supports syntax highlighting for code entries. The syntax highlighting system uses tags to detect what language an entry contains and applies the appropriate highlighting.

## How Syntax Highlighting Works

1. Each entry in ThighPads has tags (e.g., "typescript", "python", "markdown")
2. Syntax highlighting definitions specify which tags they can handle
3. When viewing an entry, if its tags match a loaded syntax highlighter, the content will be highlighted

## Syntax Highlighting JSON Schema

Syntax highlighters are defined in JSON files with the following structure:

```json
{
  "name": "TypeScript",
  "author": "ThighPads",
  "version": "1.0",
  "language": "typescript",
  "tags": ["typescript", "ts", "javascript", "js"],
  "tokenColors": {
    "keyword": "#569CD6",
    "string": "#CE9178",
    "number": "#B5CEA8",
    "comment": "#6A9955",
    "function": "#DCDCAA",
    "type": "#4EC9B0",
    "variable": "#9CDCFE",
    "operator": "#D4D4D4"
  },
  "rules": [
    {
      "pattern": "\\b(const|let|var|function|return|if|else|for|while|class|extends|import|export)\\b",
      "token": "keyword"
    },
    {
      "pattern": "\"(?:\\\\\"|[^\"])*?\"|'(?:\\\\'|[^'])*?'|`(?:\\\\`|[^`])*?`",
      "token": "string"
    },
    {
      "pattern": "\\b\\d+(\\.\\d+)?\\b",
      "token": "number"
    },
    {
      "pattern": "\\/\\/.*$|\\/\\*[\\s\\S]*?\\*\\/",
      "token": "comment"
    },
    {
      "pattern": "\\b[A-Za-z_][A-Za-z0-9_]*(?=\\s*\\()",
      "token": "function"
    },
    {
      "pattern": "\\b(interface|type|class|enum)\\b|\\b[A-Z][A-Za-z0-9_]*\\b",
      "token": "type"
    }
  ]
}
```

### Schema Attributes

- `name`: The name of the syntax highlighter (displayed in settings)
- `author`: The creator of the syntax highlighter
- `version`: The syntax highlighter version
- `language`: The primary language this highlighter supports
- `tags`: Array of tags that will trigger this syntax highlighter
- `tokenColors`: Object mapping token types to colors (hex values)
- `rules`: Array of pattern rules that match text and assign tokens

### Rule Properties

- `pattern`: A regular expression pattern that matches specific syntax elements
- `token`: The token type to assign to matched text (corresponds to `tokenColors`)

## Available Token Types

- `keyword`: Language keywords (if, for, function, etc.)
- `string`: String literals
- `number`: Numeric literals
- `comment`: Comments
- `function`: Function names
- `type`: Type names, classes, interfaces
- `variable`: Variable names
- `operator`: Operators (+, -, =, etc.)

## Using Syntax Highlighting

1. Create a JSON file with your syntax highlighter, following the schema above
2. In ThighPads, go to Settings
3. Select "Import syntax highlighting"
4. Enter the path to your JSON file and press Enter
5. Your syntax highlighter will be imported
6. Go to "Manage syntax highlighting" and enable your highlighter
7. Create or view entries with tags that match your syntax highlighter

## Example Rules

Here are some example regex patterns for common language features:

| Feature | Pattern Example |
|---------|-----------------|
| Keywords | `\\b(if|else|for|while|function)\\b` |
| Strings | `"(?:\\\\\"|[^\"])*?\"|'(?:\\\\'|[^'])*?'` |
| Numbers | `\\b\\d+(\\.\\d+)?\\b` |
| Comments | `\\/\\/.*$|\\/\\*[\\s\\S]*?\\*\\/` |
| Functions | `\\b[A-Za-z_][A-Za-z0-9_]*(?=\\s*\\()` |
| Types | `\\b(interface|type|class)\\b|\\b[A-Z][A-Za-z0-9_]*\\b` |

## Multiple Syntax Highlighters

You can have multiple syntax highlighters enabled at once. ThighPads will:

1. Check the entry's tags against all enabled syntax highlighters
2. Apply highlighting from the first syntax highlighter with matching tags
3. If multiple syntax highlighters match different tags, they won't conflict as long as their patterns don't overlap

## Example

See the included `example_files/typescript_syntax.json` for an example syntax highlighter for TypeScript.

## Syntax Highlighter Location

Imported syntax highlighters are stored in the ThighPads config directory:
- Windows: `%USERPROFILE%\.config\thighpads\syntax\`
- macOS/Linux: `~/.config/thighpads/syntax\`

You can also place syntax files directly in this folder.

## Regular Expression Tips

When creating patterns:
- Use `\\b` for word boundaries
- Escape backslashes with another backslash
- Test your regex patterns in a regex tester before using them
- More specific patterns should be placed before general patterns

## Sharing Syntax Highlighters

If you create a useful syntax highlighter, consider sharing it with the ThighPads community!