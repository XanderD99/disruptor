#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: $0 migration_name [-o output_dir]"
    exit 1
fi

NAME="$1"
shift

# Default migration directory
MIGRATION_DIR="cmd/migrate/migrations"

# Parse optional -o flag for output directory
while [[ $# -gt 0 ]]; do
    case "$1" in
        -o|--output)
            MIGRATION_DIR="$2"
            shift 2
            ;;
        *)
            shift
            ;;
    esac
done

TIMESTAMP=$(date +"%Y%m%d%H%M%S")
NAME=$(echo "$NAME" | tr ' ' '_' | tr '[:upper:]' '[:lower:]')
FILENAME="${MIGRATION_DIR}/${TIMESTAMP}_${NAME}.go"

mkdir -p "$MIGRATION_DIR"

cat > "$FILENAME" <<EOF
package migrations

import (
    "context"

    "github.com/uptrace/bun"
)

func init() {
    Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
        // TODO: Add migration code here
        return nil
    }, func(ctx context.Context, db *bun.DB) error {
        // TODO: Add rollback code here
        return nil
    })
}
EOF

echo "Created migration: $FILENAME"
