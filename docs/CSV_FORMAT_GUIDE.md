# CSV Import/Export Documentation

## Overview

The Pivot CLI supports importing and exporting GitHub issues via CSV files, enabling seamless integration with external project management tools and data migration workflows.

## CSV File Format Guidelines

### Required Fields

- **title**: The issue title (required, must not be empty)

### Supported Fields

All fields are optional except for `title`:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `id` | Integer | Issue ID | `123` |
| `title` | String | Issue title (required) | `"Fix authentication bug"` |
| `state` | String | Issue state | `"open"`, `"closed"` |
| `priority` | String | Issue priority | `"high"`, `"medium"`, `"low"` |
| `labels` | String List | Comma-separated labels | `"bug,urgent,security"` |
| `assignee` | String | Assigned user | `"john.doe"` |
| `milestone` | String | Milestone name | `"v1.0.0"` |
| `body` | String | Issue description | `"Detailed description..."` |
| `estimated_hours` | Integer | Estimated work hours | `8` |
| `story_points` | Integer | Story points | `5` |
| `epic` | String | Epic name | `"User Authentication"` |
| `dependencies` | Integer List | Comma-separated issue IDs | `"45,67,89"` |
| `acceptance_criteria` | String | Acceptance criteria | `"User can login successfully"` |
| `created_at` | DateTime | Creation timestamp | `"2024-01-15T10:00:00Z"` |
| `updated_at` | DateTime | Last update timestamp | `"2024-01-15T10:30:00Z"` |

### CSV Formatting Rules

#### 1. Header Row
The first row must contain column headers matching the field names above (case-insensitive).

#### 2. String Values
- Enclose string values in double quotes if they contain commas, line breaks, or quotes
- Escape internal quotes by doubling them: `"Issue with ""special"" handling"`

#### 3. Multi-value Fields
- **Labels**: Separate multiple labels with commas: `"bug,urgent,security"`
- **Dependencies**: Separate multiple issue IDs with commas: `"123,456,789"`

#### 4. Date/Time Format
- Use RFC3339 format: `"2024-01-15T10:00:00Z"`
- UTC timezone recommended

#### 5. Empty Values
- Leave empty for optional fields: `""`
- Do not use `null` or `N/A`

### Example CSV Files

#### Minimal CSV
```csv
title,state
"Fix login bug","open"
"Add user dashboard","closed"
```

#### Complete CSV
```csv
id,title,state,priority,labels,assignee,milestone,body,estimated_hours,story_points,epic,dependencies,acceptance_criteria,created_at,updated_at
1,"Implement user authentication","open","high","backend,security","john.doe","v1.0","Add secure user login and registration system",16,8,"User Management","","User can register and login securely","2024-01-15T10:00:00Z","2024-01-15T10:30:00Z"
2,"Design login UI","open","medium","frontend,ui","jane.smith","v1.0","Create user-friendly login interface",8,5,"User Management","1","Login form is intuitive and accessible","2024-01-16T09:00:00Z","2024-01-16T09:15:00Z"
```

## Commands

### Import CSV

```bash
# Preview import (recommended first step)
pivot import csv --preview issues.csv

# Dry-run import (validates without creating issues)
pivot import csv --dry-run issues.csv

# Import to specific repository
pivot import csv --repository owner/repo issues.csv

# Skip duplicate issues during import
pivot import csv --skip-duplicates issues.csv
```

### Export CSV

```bash
# Export all issues to default file (issues.csv)
pivot export csv

# Export to custom file
pivot export csv --output my-issues.csv

# Export specific fields only
pivot export csv --fields title,state,priority,labels

# Export from specific repository
pivot export csv --repository owner/repo
```

## Common Use Cases

### 1. Migration from Other Tools

Many project management tools can export to CSV. To migrate:

1. Export from your current tool
2. Map the columns to Pivot's format (see field mapping below)
3. Preview the import: `pivot import csv --preview exported-issues.csv`
4. Import: `pivot import csv exported-issues.csv`

### 2. Bulk Issue Creation

Create a CSV file with your issues and import:

```bash
# Create issues.csv with your data
pivot import csv --preview issues.csv  # Verify first
pivot import csv issues.csv            # Import
```

### 3. Data Analysis and Reporting

Export issues for analysis in spreadsheet applications:

```bash
pivot export csv --output analysis-data.csv
# Open in Excel, Google Sheets, etc.
```

### 4. Backup and Restore

```bash
# Backup
pivot export csv --output backup-$(date +%Y%m%d).csv

# Restore (use with caution)
pivot import csv --preview backup-20240115.csv
pivot import csv backup-20240115.csv
```

## Field Mapping Guide

When migrating from other tools, map their fields to Pivot's format:

| Other Tool Field | Pivot Field | Notes |
|------------------|-------------|-------|
| Summary | title | Required field |
| Description | body | Can contain markdown |
| Status | state | Map to "open" or "closed" |
| Severity/Importance | priority | Map to "high", "medium", "low" |
| Tags/Categories | labels | Comma-separated list |
| Assignee | assignee | Username only |
| Version/Release | milestone | Version string |
| Effort/Points | story_points | Integer value |
| Time Estimate | estimated_hours | Integer hours |
| Blocked By | dependencies | Comma-separated issue IDs |

## Troubleshooting

### Common Issues

#### 1. "CSV validation failed: EOF"
- **Cause**: Empty CSV file or file with only headers
- **Solution**: Ensure file has at least one data row

#### 2. "Required column 'title' not found"
- **Cause**: Missing title column in header
- **Solution**: Add `title` column to your CSV

#### 3. "Column count mismatch"
- **Cause**: Some rows have different number of columns than header
- **Solution**: Ensure all rows have the same number of fields, use empty strings for missing values

#### 4. "Failed to parse date"
- **Cause**: Invalid date format
- **Solution**: Use RFC3339 format: `"2024-01-15T10:00:00Z"`

#### 5. Import appears to hang
- **Cause**: Large file or network issues
- **Solution**: Try smaller batches, check network connectivity

### Best Practices

1. **Always preview first**: Use `--preview` flag before importing
2. **Use dry-run**: Test with `--dry-run` for large imports
3. **Backup before import**: Export current issues before importing new ones
4. **Validate data**: Check for required fields and correct formats
5. **Test with small batches**: Import a few issues first to verify format
6. **Handle encoding**: Ensure CSV files are UTF-8 encoded
7. **Quote special characters**: Use double quotes for values containing commas or quotes

### File Encoding

- **Recommended**: UTF-8 without BOM
- **Supported**: UTF-8 with BOM (automatically detected)
- **Line endings**: Both Unix (LF) and Windows (CRLF) are supported

### Performance Considerations

- **File size**: Large files (>10MB) may take longer to process
- **API limits**: GitHub API has rate limits; large imports may be throttled
- **Memory usage**: Very large files may require significant memory

## Examples Repository

For complete examples and sample CSV files, see the `examples/csv/` directory in the Pivot repository.

## Support

If you encounter issues with CSV import/export:

1. Check this documentation for common solutions
2. Use `--preview` and `--dry-run` flags for debugging
3. Validate your CSV format against the examples
4. Create an issue in the Pivot repository with sample data (remove sensitive information)
