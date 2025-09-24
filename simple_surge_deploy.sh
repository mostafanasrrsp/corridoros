#!/bin/bash

echo "🚀 Simple Surge Deployment for CorridorOS"
echo "=========================================="

# Navigate to project directory
cd /Users/mnasr/Desktop/COS

echo "📁 Current directory: $(pwd)"
echo "📦 Project contents:"
ls -la | head -10

echo ""
echo "🔐 Starting Surge deployment..."
echo "Please enter your surge credentials when prompted:"
echo "Email: mostafanasr@aucegypt.edu"
echo "Password: [enter your password]"
echo ""
echo "For domain, just press ENTER to auto-generate a random domain"
echo ""

# Run surge with manual input
surge

echo ""
echo "✅ Surge deployment process completed!"
echo "🌐 Check the output above for your live URL"
