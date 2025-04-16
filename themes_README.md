# ThighPads Themes

ThighPads supports custom themes to personalize your experience. You can create, import, and manage themes through the settings menu.

## Theme JSON Schema

Themes are defined in JSON files with the following structure:

```json
{
  "name": "Midnight",
  "author": "ThighPads",
  "version": "1.0",
  "accent": "#7D56F4",
  "secondary": "#AE88FF",
  "text": "#E4E4E4",
  "subtle": "#888888",
  "error": "#FF5555",
  "success": "#55FF55",
  "warning": "#FFAA55",
  "background": "#1A1A1A"
}
```

### Color Properties

| Property | Description | Required | Example |
|----------|-------------|----------|---------|
| `name` | Theme name shown in settings | Yes | `"Midnight"` |
| `author` | Creator of the theme | No | `"ThighPads"` |
| `version` | Theme version | No | `"1.0"` |
| `accent` | Primary accent color for titles and borders | Yes | `"#7D56F4"` |
| `secondary` | Secondary accent for subtitles and highlights | Yes | `"#AE88FF"` |
| `text` | Main text color | Yes | `"#FFFFFF"` |
| `subtle` | Less important text color | No | `"#888888"` |
| `error` | Color for error messages | No | `"#FF5555"` |
| `success` | Color for success messages | No | `"#55FF55"` |
| `warning` | Color for warnings | No | `"#FFAA55"` |
| `background` | Application background color | Yes | `"#222222"` |

All colors must be specified as hex values in the format `#RRGGBB` or `#RGB`.

## Theme Validation

ThighPads validates themes against the above schema. The required fields are:
- `name`
- `accent`
- `secondary`
- `text`
- `background`

If your theme is missing any required fields or has invalid color values, it will not be loaded.

## Using Themes

### Importing a Theme

1. Create a JSON file with your theme, following the schema above
2. Go to Settings → Appearance → Themes
3. Select "Import theme"
4. Enter the path to your JSON file
5. Your theme will be imported and appear in the themes list

### Applying a Theme

1. Go to Settings → Appearance → Themes
2. Select a theme from the list
3. The theme will be applied immediately

### Creating a Theme

1. Go to Settings → Appearance → Themes
2. Select "Create new theme"
3. Enter a name for your theme
4. Adjust the color values
5. Save your theme

## Example Themes

### Dracula Theme

```json
{
  "name": "Dracula",
  "author": "ThighPads",
  "version": "1.0",
  "accent": "#BD93F9",
  "secondary": "#FF79C6",
  "text": "#F8F8F2",
  "subtle": "#6272A4",
  "error": "#FF5555",
  "success": "#50FA7B",
  "warning": "#FFB86C",
  "background": "#282A36"
}
```

### Nord Theme

```json
{
  "name": "Nord",
  "author": "ThighPads",
  "version": "1.0",
  "accent": "#88C0D0",
  "secondary": "#81A1C1",
  "text": "#ECEFF4",
  "subtle": "#4C566A",
  "error": "#BF616A",
  "success": "#A3BE8C",
  "warning": "#EBCB8B",
  "background": "#2E3440"
}
```

See the included `example_files/themes/` directory for additional theme examples.

## Theme Location

Imported themes are stored in the ThighPads config directory:
- Windows: `%USERPROFILE%\.config\thighpads\themes\`
- macOS/Linux: `~/.config/thighpads/themes\`

You can also place theme files directly in this folder to make them available in the theme selector.

## Color Harmony Tips

When creating themes, consider these tips for harmonious color schemes:

1. Use a color wheel to find complementary or analogous colors
2. Maintain sufficient contrast between text and background (WCAG recommends at least 4.5:1)
3. Use accent colors sparingly for visual hierarchy
4. Test your theme in both light and dark environments

## Sharing Themes

If you create a theme you're proud of, consider sharing it with the ThighPads community!