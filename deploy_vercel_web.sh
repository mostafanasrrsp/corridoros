#!/bin/bash

echo "🚀 Deploying CorridorOS to Vercel (Web Interface)"
echo "================================================"

# Navigate to project directory
cd /Users/mnasr/Desktop/COS

echo "📁 Current directory: $(pwd)"
echo "📦 Preparing for Vercel deployment..."

# Create a simple deployment package
echo "📦 Creating deployment package..."
tar -czf corridoros-deploy.tar.gz \
    --exclude='.git' \
    --exclude='node_modules' \
    --exclude='*.log' \
    --exclude='.DS_Store' \
    .

echo ""
echo "✅ Deployment package created: corridoros-deploy.tar.gz"
echo ""
echo "🌐 Deploy to Vercel using web interface:"
echo "   1. Go to: https://vercel.com/new"
echo "   2. Click 'Browse all templates'"
echo "   3. Select 'Other' or 'Static Site'"
echo "   4. Upload the corridoros-deploy.tar.gz file"
echo "   5. Or connect your GitHub repository: mostafanasrrsp/corridoros"
echo ""
echo "🔗 Your site will be available at a Vercel URL like:"
echo "   https://corridoros-[random].vercel.app"
echo ""
echo "📋 Alternative: Use Vercel CLI after authentication:"
echo "   npx vercel --prod"
