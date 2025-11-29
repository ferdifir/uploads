# Deployment Guide

This project includes a deployment script to make it easy to deploy the application to a server.

## Prerequisites

- Git
- Go (1.25 or higher)
- SQLite3 development headers (for CGO)

## Deployment Script Usage

The `deploy.sh` script handles the entire deployment process including:
- Cloning/updating the repository
- Building the application
- Managing the server process

### Basic Usage

```bash
./deploy.sh
```

This will:
1. Stop any existing running instance
2. Clone the repository (or update if it already exists)
3. Build the application
4. Create a default configuration file
5. Start the server

### Command Line Options

```bash
./deploy.sh [options]
```

Available options:
- `-b, --branch BRANCH`: Specify the git branch to deploy (default: main)
- `-r, --restart`: Restart the server after update (for update action)
- `-u, --update`: Update the existing deployment
- `-s, --start`: Start the server without updating
- `-t, --stop`: Stop the running server
- `--port PORT`: Specify the port for the server (default: 8081)

### Examples

1. **Deploy for the first time:**
   ```bash
   ./deploy.sh
   ```

2. **Update existing deployment:**
   ```bash
   ./deploy.sh -u
   ```

3. **Update and restart:**
   ```bash
   ./deploy.sh -u -r
   ```

4. **Deploy from a specific branch:**
   ```bash
   ./deploy.sh -b develop
   ```

5. **Stop the server:**
   ```bash
   ./deploy.sh -t
   ```

6. **Start the server:**
   ```bash
   ./deploy.sh -s
   ```

## Configuration

The deployment script will create a default configuration file at `configs/config.json` if it doesn't exist. You can customize this file after deployment.

The default configuration includes:
- API key: "RahasiaAPIKey123"
- UI username: "admin"
- UI password hash: "bb77d0d3b3f239fa5db73bdf27b8d29a" (for password "securepass")
- Data directory: "data"
- Database path: "uploads.db"
- Port: 8081 (or as specified with --port)

## Server Management

The script supports both systemd and nohup-based server management:

- If systemd is available, it will create a systemd service
- If systemd is not available, it will use nohup to run the server in the background

## Notes

1. Before using the deployment script, make sure to update the `REPO_URL` variable in the script with your actual repository URL
2. The server runs on port 8081 by default, but this can be changed with the `--port` option
3. The application data will be stored in the `data` directory
4. The SQLite database will be created as `uploads.db`
