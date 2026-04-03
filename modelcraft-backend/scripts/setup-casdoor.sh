#!/bin/bash
# Quick setup script for Casdoor integration with ModelCraft

set -e

echo "=========================================="
echo "Casdoor Setup for ModelCraft"
echo "=========================================="
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    echo "❌ .env file not found!"
    echo "Please copy .env.docker.example to .env first:"
    echo "  cp .env.docker.example .env"
    echo ""
    exit 1
fi

echo "Step 1: Starting Casdoor and its database..."
docker compose up -d casdoor-db casdoor

echo ""
echo "⏳ Waiting for Casdoor to be ready (this may take 30-60 seconds)..."
echo ""

# Wait for Casdoor to be healthy
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if docker compose ps casdoor | grep -q "healthy"; then
        echo "✅ Casdoor is ready!"
        break
    fi

    attempt=$((attempt + 1))
    if [ $attempt -eq $max_attempts ]; then
        echo "❌ Casdoor did not become healthy in time"
        echo "Check logs with: docker compose logs casdoor"
        exit 1
    fi

    echo -n "."
    sleep 2
done

echo ""
echo "=========================================="
echo "✅ Casdoor is now running!"
echo "=========================================="
echo ""
echo "Next steps:"
echo ""
echo "1. Open Casdoor Admin Console:"
echo "   🌐 http://localhost:8000"
echo ""
echo "2. Login with default credentials:"
echo "   👤 Username: admin"
echo "   🔑 Password: 123"
echo ""
echo "   ⚠️  IMPORTANT: Change this password immediately!"
echo ""
echo "3. Create an Organization:"
echo "   - Go to: Organizations → Add"
echo "   - Name: modelcraft-default"
echo "   - Display Name: ModelCraft Default Organization"
echo "   - Website URL: http://localhost:8080"
echo ""
echo "4. Create an Application:"
echo "   - Go to: Applications → Add"
echo "   - Name: modelcraft"
echo "   - Display Name: ModelCraft Runtime"
echo "   - Organization: modelcraft-default"
echo "   - Redirect URLs: http://localhost:8080/api/auth/callback"
echo "   - Grant Types: Authorization Code, Refresh Token"
echo "   - Token Format: JWT"
echo ""
echo "5. Get Configuration Values:"
echo "   - Copy Client ID and Client Secret from the Application"
echo "   - Go to: Certs → cert-built-in → Copy Certificate"
echo ""
echo "6. Update .env file with:"
echo "   - CASDOOR_CLIENT_ID"
echo "   - CASDOOR_CLIENT_SECRET"
echo "   - CASDOOR_ORGANIZATION (e.g., modelcraft-default)"
echo "   - CASDOOR_CERTIFICATE"
echo ""
echo "7. Restart ModelCraft:"
echo "   docker compose up -d modelcraft"
echo ""
echo "=========================================="
echo "📖 For detailed instructions, see: casdoor/README.md"
echo "=========================================="
