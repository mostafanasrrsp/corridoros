#!/bin/bash

echo "🚀 Deploying CorridorOS to Netlify..."
echo "=================================="

# Check if Netlify CLI is installed
if ! command -v netlify &> /dev/null; then
    echo "❌ Netlify CLI not found. Installing..."
    npm install -g netlify-cli
fi

# Check if we're logged in
if ! netlify status &> /dev/null; then
    echo "🔐 Please log in to Netlify first:"
    netlify login
fi

# Deploy the site
echo "📦 Deploying to Netlify..."
netlify deploy --prod --dir .

echo "✅ Deployment complete!"
echo "🌐 Your CorridorOS is now live with HTTPS!"
echo "🔗 Check your Netlify dashboard for the URL"
