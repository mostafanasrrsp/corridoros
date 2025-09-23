#!/bin/bash

# Corridor OS Deployment Script
echo "🚀 Deploying Corridor OS to Firebase..."

# Check if Firebase CLI is installed
if ! command -v firebase &> /dev/null; then
    echo "❌ Firebase CLI not found. Please install it first:"
    echo "   npm install -g firebase-tools"
    exit 1
fi

# Check if logged in to Firebase
if ! firebase projects:list &> /dev/null; then
    echo "🔐 Please log in to Firebase:"
    firebase login
fi

# Build the project
echo "🏗️  Building Corridor OS..."
python3 create-bootable-iso.py

# Test the build locally
echo "🧪 Testing build locally..."
echo "   Starting local server on port 5000..."
firebase serve --port 5000 &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Check if server is running
if curl -s http://localhost:5000 > /dev/null; then
    echo "✅ Local server running successfully"
    echo "   Visit: http://localhost:5000"
    echo "   Press Ctrl+C to stop and continue with deployment"
    
    # Wait for user input
    read -p "   Press Enter to continue with deployment..."
    
    # Stop local server
    kill $SERVER_PID 2>/dev/null
else
    echo "❌ Local server failed to start"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# Deploy to Firebase
echo "☁️  Deploying to Firebase Hosting..."
firebase deploy

# Check deployment status
if [ $? -eq 0 ]; then
    echo "✅ Deployment successful!"
    echo ""
    echo "🌐 Your Corridor OS is now live at:"
    echo "   https://redseaportal.com"
    echo ""
    echo "📱 Access points:"
    echo "   • Main OS: https://redseaportal.com/"
    echo "   • Computer Demo: https://redseaportal.com/computer"
    echo "   • Direct OS: https://redseaportal.com/os"
    echo ""
    echo "🎉 Corridor OS deployment complete!"
else
    echo "❌ Deployment failed!"
    exit 1
fi
