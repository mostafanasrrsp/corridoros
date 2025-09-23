#!/bin/bash

echo "🚀 Deploying Corridor Computer Emulator to Firebase..."

# Check if Firebase CLI is installed
if ! command -v firebase &> /dev/null; then
    echo "❌ Firebase CLI not found. Please install it first."
    exit 1
fi

# Check if user is logged in
if ! firebase projects:list &> /dev/null; then
    echo "🔐 Please login to Firebase first:"
    echo "   firebase login"
    exit 1
fi

# Deploy to Firebase
echo "📦 Deploying to Firebase Hosting..."
firebase deploy

echo "✅ Deployment complete!"
echo "🌐 Your emulator should be available at your Firebase hosting URL"
echo "📋 Check the deploy.md file for custom domain setup instructions"
