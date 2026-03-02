# Metadata Extraction Feature

This feature provides comprehensive metadata extraction capabilities using ExifTool, supporting hundreds of file formats. It includes both backend processing and an enhanced frontend component for displaying structured metadata.

**Note**: The frontend `MetadataViewer` component was enhanced to use this backend service instead of creating a separate component, providing a unified metadata viewing experience.

## Supported File Types

ExifTool supports **hundreds of file formats** including (but not limited to):

### Images
- **Common formats**: JPEG, PNG, GIF, WebP, TIFF, BMP, HEIC, HEIF
- **Raw formats**: CR2, NEF, ARW, DNG, and many more
- **EXIF data**: Camera info, location, timestamps, dimensions, color space

### Videos
- MP4, AVI, MOV, MKV, WebM, FLV, WMV, M4V
- Duration, codec info, creation dates, GPS data

### Audio
- MP3, WAV, FLAC, M4A, AAC, OGG, WMA
- Duration, bitrate, codec, album/artist info

### Documents
- **PDF**: Title, author, subject, keywords, page count, creation dates
- **Office**: Word (DOC/DOCX), Excel (XLS/XLSX), PowerPoint (PPT/PPTX)
- **Text**: Plain text, JSON, XML

### Archives
- ZIP, RAR, 7Z, TAR, GZ
- File counts, compression info, creation dates

### And many more formats supported by ExifTool!

## API Endpoints

### POST /api/metadata/extract

Extracts metadata from an uploaded file.

**Request**: Multipart form data with a file upload
**Response**: JSON object containing extracted metadata

```json
{
  "success": true,
  "filename": "photo.jpg",
  "metadata": {
    "filename": "photo.jpg",
    "size": 2048576,
    "contentType": "image/jpeg",
    "hasMetadata": true,
    "image": {
      "width": 1920,
      "height": 1080,
      "hasExif": true,
      "dateTime": "2024-01-15 10:30:45",
      "cameraMake": "Canon",
      "cameraModel": "EOS R5",
      "colorSpace": "sRGB",
      "location": {
        "latitude": 40.7128,
        "longitude": -74.0060,
        "altitude": 10.5
      },
      "exifData": {
        "Make": "Canon",
        "Model": "EOS R5",
        "DateTimeOriginal": "2024:01:15 10:30:45",
        "ImageWidth": 1920,
        "ImageHeight": 1080,
        "GPSLatitude": 40.7128,
        "GPSLongitude": -74.0060,
        "GPSAltitude": 10.5,
        // ... many more EXIF fields
      }
    },
    "raw": {
      "FileType": "JPEG",
      "MIMEType": "image/jpeg",
      "ImageWidth": 1920,
      "ImageHeight": 1080,
      "EncodingProcess": "Baseline DCT, Huffman coding",
      "BitsPerSample": 8,
      "ColorComponents": 3,
      // ... hundreds of additional metadata fields
    }
  }
}
```

## Architecture

The metadata extraction feature consists of:

### Backend Components
- **Processor** (`processor.go`): Core extraction logic
- **Handler** (`handler.go`): HTTP request handling
- **Types**: Structured metadata representations

### Frontend Components
- **MetadataViewer** (`frontend/src/components/ui/MetadataViewer.jsx`): Enhanced UI component that uses the metadata service
- **Metadata Service** (`frontend/src/features/file-processing/services/metadataService.js`): Service layer with backend priority and local fallback

## Dependencies

- `github.com/barasher/go-exiftool`: Go wrapper for ExifTool (metadata extraction and removal)
- Standard Go libraries for file handling, temporary files, and command execution

**Note**: ExifTool now handles all metadata operations comprehensively, including both extraction and removal for all supported file types.

## System Requirements

**ExifTool Installation Required**: This feature requires ExifTool to be installed on the system. ExifTool is a command-line tool that must be available in the system PATH.

### Installation Instructions:

**Ubuntu/Debian:**
```bash
sudo apt-get install libimage-exiftool-perl
```

**macOS (with Homebrew):**
```bash
brew install exiftool
```

**Windows:**
Download from: https://exiftool.org/

**Docker:**
```dockerfile
RUN apt-get update && apt-get install -y libimage-exiftool-perl
```

## Frontend Usage

### MetadataViewer Component

The enhanced `MetadataViewer` component provides a structured display of metadata:

```jsx
import { MetadataViewer } from './components/ui';

// Basic usage - automatically extracts and displays metadata
<MetadataViewer file={selectedFile} />
```

**Features:**
- **Universal Format Support**: Extracts metadata from hundreds of file types (images, videos, audio, documents, archives)
- **Structured Display**: Organized sections for file info, image metadata, PDF metadata, and raw data
- **Backend Priority**: Uses ExifTool backend processing with automatic fallback to local extraction
- **Comprehensive Metadata**: EXIF, GPS, camera info, document properties, codec info, timestamps, and much more
- **Raw Data Access**: Full access to all extracted metadata fields for advanced users
- **Loading States**: Skeleton animations during processing
- **Error Handling**: User-friendly error messages with graceful degradation

### Metadata Service

Direct access to the metadata extraction service:

```javascript
import { getFileMetadata } from './features/file-processing/services/metadataService';

const metadata = await getFileMetadata(file);
// Returns structured metadata with backend/local mode indicator
```

## Backend Usage in Other Services

The metadata extraction processor can be integrated into other services:

```go
processor := metadata_extraction.NewProcessor()
metadata, err := processor.ProcessFileWithContent(ctx, filename, content, contentType)
```
