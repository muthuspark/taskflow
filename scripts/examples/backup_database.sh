#!/bin/bash
# Example backup script for TaskFlow

set -e

echo "Starting database backup..."
BACKUP_DIR="${BACKUP_DIR:-./backups}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/taskflow_$TIMESTAMP.db.gz"

mkdir -p "$BACKUP_DIR"

# Backup the database
if [ -f "taskflow.db" ]; then
    gzip -c taskflow.db > "$BACKUP_FILE"
    echo "Backup completed: $BACKUP_FILE"
    ls -lh "$BACKUP_FILE"
else
    echo "Error: taskflow.db not found"
    exit 1
fi

# Keep only last 7 backups
find "$BACKUP_DIR" -name "taskflow_*.db.gz" -type f -mtime +7 -delete
echo "Old backups cleaned up"
