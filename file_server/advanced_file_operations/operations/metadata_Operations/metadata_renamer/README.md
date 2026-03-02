# Metadata Renamer

Intelligent file renaming based on EXIF metadata extracted from images.

## Features

- Rename files using EXIF date/time
- Rename files using camera make/model
- Rename files using GPS coordinates
- Combined metadata patterns

## Patterns

- `exif:datetime` - Uses DateTimeOriginal, CreateDate, or ModifyDate
- `exif:camera` - Uses Make and Model fields
- `exif:location` - Uses GPSLatitude and GPSLongitude
- `exif:combined` - Combines camera, date, and ISO

## Usage

This feature requires file content (uses `ProcessFileWithContent`) and integrates with the ExifTool service for metadata extraction.

