#!/bin/bash

# Deployment script for File Upload Manager
# This script will clone the repository, build the application, and start the server

set -e  # Exit on any error

# Configuration
REPO_URL="https://github.com/ferdifir/uploads.git"  # Update this with your actual repo URL
PROJECT_NAME="uploads"
DEPLOY_DIR="/var/www/$PROJECT_NAME"
BINARY_NAME="file-manager"
PORT=${PORT:-8081}  # Use PORT environment variable or default to 8081

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
print_status "Checking prerequisites..."

if ! command_exists git; then
    print_error "git is not installed. Please install git first."
    exit 1
fi

if ! command_exists go; then
    print_error "go is not installed. Please install Go first."
    exit 1
fi

print_status "Prerequisites check passed."

# Parse command line arguments
ACTION="deploy"
BRANCH="main"
RESTART=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -b|--branch)
            BRANCH="$2"
            shift 2
            ;;
        -r|--restart)
            RESTART=true
            shift
            ;;
        -u|--update)
            ACTION="update"
            shift
            ;;
        -s|--start)
            ACTION="start"
            shift
            ;;
        -t|--stop)
            ACTION="stop"
            shift
            ;;
        --port)
            PORT="$2"
            shift 2
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Usage: $0 [-b|--branch BRANCH] [-r|--restart] [-u|--update] [-s|--start] [-t|--stop] [--port PORT]"
            exit 1
            ;;
    esac
done

# Function to stop the server if running
stop_server() {
    if pgrep -f "$BINARY_NAME" > /dev/null; then
        print_status "Stopping existing server..."
        pkill -f "$BINARY_NAME" || true
        sleep 2
        if pgrep -f "$BINARY_NAME" > /dev/null; then
            sleep 5  # Give more time for graceful shutdown
            pkill -9 -f "$BINARY_NAME" || true  # Force kill if still running
        fi
        print_status "Server stopped."
    else
        print_status "No running server found."
    fi
}

# Function to start the server
start_server() {
    print_status "Starting server on port $PORT..."
    
    # Create a simple systemd service or run in background
    if command_exists systemctl && [ -f "/etc/systemd/system/$BINARY_NAME.service" ]; then
        sudo systemctl start $BINARY_NAME
        sudo systemctl status $BINARY_NAME
    else
        # Run in background with nohup
        nohup "$DEPLOY_DIR/$BINARY_NAME" &
        echo $! > "$DEPLOY_DIR/$BINARY_NAME.pid"
        print_status "Server started in background. PID: $(cat $DEPLOY_DIR/$BINARY_NAME.pid)"
    fi
}

# Function to deploy/update the application
deploy_application() {
    print_status "Deploying application to $DEPLOY_DIR..."
    
    # Create deployment directory if it doesn't exist
    sudo mkdir -p "$DEPLOY_DIR"
    sudo chown $USER:$USER "$DEPLOY_DIR"
    
    # Clone or update the repository
    if [ -d "$DEPLOY_DIR/.git" ]; then
        print_status "Repository already exists. Updating..."
        cd "$DEPLOY_DIR"
        git fetch origin
        git checkout "$BRANCH"
        git pull origin "$BRANCH"
    else
        print_warning "Repository not found at $DEPLOY_DIR. This script expects the code to be in a git repository."
        print_status "If deploying from local files, copy the project files to $DEPLOY_DIR manually."
        print_status "For initial deployment, you should:"
        print_status "1. Push this project to a git repository"
        print_status "2. Update the REPO_URL variable in this script"
        print_status "3. Run this script again"
        
        # If directory doesn't exist, we'll create a simple copy-based deployment
        if [ ! -d "$DEPLOY_DIR" ]; then
            print_status "Creating deployment directory..."
            mkdir -p "$DEPLOY_DIR"
        fi
        
        # Copy current project files to deployment directory
        print_status "Copying project files to $DEPLOY_DIR..."
        rsync -av --exclude='deploy.sh' --exclude='README-DEPLOY.md' --exclude='.git' . "$DEPLOY_DIR/"
        cd "$DEPLOY_DIR"
    fi
    
    # Build the application
    print_status "Building application..."
    go mod tidy
    CGO_ENABLED=1 go build -o "$BINARY_NAME" cmd/server/main.go
    
    # Copy config file if it doesn't exist
    if [ ! -f "$DEPLOY_DIR/configs/config.json" ]; then
        print_status "Creating default config file..."
        mkdir -p "$DEPLOY_DIR/configs"
        cat > "$DEPLOY_DIR/configs/config.json" << EOF
{
  "api_key": "RahasiaAPIKey123",
  "ui_username": "admin",
  "ui_password_hash": "bb77d0d3b3f239fa5db73bdf27b8d29a",
  "data_dir": "data",
  "db_path": "uploads.db",
  "port": $PORT
}
EOF
    fi
    
    # Create data directory
    mkdir -p "$DEPLOY_DIR/data"
    
    print_status "Build completed successfully."
}

# Function to setup systemd service (optional)
setup_systemd_service() {
    if command_exists systemctl; then
        print_status "Setting up systemd service..."
        
        sudo tee "/etc/systemd/system/$BINARY_NAME.service" > /dev/null << EOF
[Unit]
Description=File Upload Manager
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$DEPLOY_DIR
ExecStart=$DEPLOY_DIR/$BINARY_NAME
Restart=always
RestartSec=10
Environment=PORT=$PORT

[Install]
WantedBy=multi-user.target
EOF
        
        sudo systemctl daemon-reload
        print_status "Systemd service created."
    else
        print_warning "systemctl not available. Will run with nohup instead."
    fi
}

# Main logic based on action
case $ACTION in
    "deploy")
        print_status "Starting deployment process..."
        stop_server
        deploy_application
        setup_systemd_service
        start_server
        print_status "Deployment completed successfully!"
        ;;
    "update")
        print_status "Updating application..."
        stop_server
        deploy_application
        if [ "$RESTART" = true ]; then
            start_server
        fi
        print_status "Update completed successfully!"
        ;;
    "start")
        start_server
        ;;
    "stop")
        stop_server
        ;;
    *)
        print_error "Unknown action: $ACTION"
        exit 1
        ;;
esac

print_status "Deployment script finished."
