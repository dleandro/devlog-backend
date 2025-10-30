#!/bin/bash

# Generate Admin API Key
# This script generates a secure random API key for admin operations

echo "🔐 Generating Admin API Key..."
echo ""

# Generate a 32-character hex key
API_KEY=$(openssl rand -hex 32)

echo "✅ Generated API Key:"
echo "ADMIN_API_KEYS=$API_KEY"
echo ""
echo "📝 Next steps:"
echo "1. Add this to your .env file (you can add multiple keys separated by commas)"
echo "2. Restart your server"
echo "3. Use this key in Authorization header: Bearer $API_KEY"
echo ""
echo "⚠️  Keep this key secure and don't commit it to version control!"