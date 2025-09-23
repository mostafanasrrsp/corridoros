#!/bin/bash

echo "🚀 Deploying CorridorOS to Vercel..."
echo "=================================="

# Check if Vercel CLI is installed
if ! command -v vercel &> /dev/null; then
    echo "❌ Vercel CLI not found. Installing..."
    npm install -g vercel
fi

# Check if we're logged in
if ! vercel whoami &> /dev/null; then
    echo "🔐 Please log in to Vercel first:"
    vercel login
fi

# Deploy the site
echo "📦 Deploying to Vercel..."
vercel --prod --yes

echo "✅ Deployment complete!"
echo "🌐 Your CorridorOS is now live with HTTPS!"
echo "🔗 Check the URL above for your live site"
