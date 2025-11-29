# File Upload Service

A simple file upload service built with Go following best practices for project structure.

## Project Structure

```
uploads/
├── cmd/                 # Application entry points
│   └── server/          # Main application
│       └── main.go      # Application entry point
├── internal/            # Internal application code
│   ├── config/          # Configuration management
│   │   └── config.go
│   ├── handlers/        # HTTP request handlers
│   │   └── handlers.go
│   └── middleware/      # HTTP middleware
│       └── middleware.go
├── configs/             # Configuration files
│   └── config.json
├── assets/              # Static assets (HTML, CSS, JS)
│   ├── index.html
│   ├── style.css
│   └── script.js
├── data/                # Uploaded files storage
├── go.mod
├── go.sum
└── README.md
```

## Features

- File upload with API key authentication
- File listing
- File download
- File deletion
- UI login with username/password
- CORS support

## Configuration

The application is configured via `configs/config.json`:

```json
{
  "api_key": "your_api_key_here",
  "ui_username": "admin",
  "ui_password_hash": "md5_hash_of_password",
  "data_dir": "data",
  "port": 80
}
```

## Running the Application

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`.

## API Endpoints

- `POST /api/upload` - Upload a file (requires X-API-Key header)
- `GET /api/list` - List all files (requires X-API-Key header)
- `GET /api/download?name=filename` - Download a file (requires X-API-Key header)
- `DELETE /api/delete` - Delete a file (requires X-API-Key header and JSON body with filename)
- `POST /api/login` - Login to get API key for UI

## Security

- API endpoints are protected with API key authentication
- Path traversal attacks are prevented
- Passwords are stored as MD5 hashes
