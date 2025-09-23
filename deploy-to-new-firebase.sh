#!/bin/bash

# Deploy Corridor OS Professional to New Firebase Project
echo "🚀 Deploying Corridor OS Professional to Firebase..."
echo "Project: corridor-os-80410062"
echo "URL: https://studio.firebase.google.com/corridor-os-80410062"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if Firebase CLI is installed
if ! command -v firebase &> /dev/null; then
    echo -e "${RED}❌ Firebase CLI not found.${NC}"
    echo "Please install it first:"
    echo "   npm install -g firebase-tools"
    exit 1
fi

# Check if logged in to Firebase
echo -e "${BLUE}🔐 Checking Firebase authentication...${NC}"
if ! firebase projects:list &> /dev/null; then
    echo -e "${YELLOW}⚠️  Please log in to Firebase:${NC}"
    firebase login
fi

# Verify project exists
echo -e "${BLUE}📋 Verifying Firebase project...${NC}"
if firebase use corridor-os-80410062 &> /dev/null; then
    echo -e "${GREEN}✅ Successfully connected to project: corridor-os-80410062${NC}"
else
    echo -e "${YELLOW}⚠️  Setting up project connection...${NC}"
    firebase use --add corridor-os-80410062
fi

# Test the build locally first
echo -e "${BLUE}🧪 Testing Professional Corridor OS locally...${NC}"
echo "   Starting local server on port 5000..."

# Kill any existing servers
pkill -f "firebase serve" 2>/dev/null || true
pkill -f "python3 -m http.server" 2>/dev/null || true

# Start Firebase local server
firebase serve --port 5000 &
SERVER_PID=$!

# Wait for server to start
sleep 4

# Check if server is running
if curl -s http://localhost:5000 > /dev/null; then
    echo -e "${GREEN}✅ Local server running successfully${NC}"
    echo "   🌐 Local URL: http://localhost:5000"
    echo "   📱 Testing Professional Corridor OS..."
    
    # Test the professional version
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:5000/corridor-os-professional.html)
    if [ "$RESPONSE" = "200" ]; then
        echo -e "${GREEN}✅ Professional Corridor OS loading correctly${NC}"
    else
        echo -e "${YELLOW}⚠️  Professional version response: $RESPONSE${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}🎯 Ready for deployment!${NC}"
    read -p "Press Enter to continue with Firebase deployment..."
    
    # Stop local server
    kill $SERVER_PID 2>/dev/null
    sleep 2
else
    echo -e "${RED}❌ Local server failed to start${NC}"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# Deploy to Firebase
echo -e "${BLUE}☁️  Deploying to Firebase Hosting...${NC}"
echo "   Project: corridor-os-80410062"
echo "   Target: https://corridor-os-80410062.web.app"
echo ""

firebase deploy --only hosting

# Check deployment status
if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ Deployment successful!${NC}"
    echo ""
    echo -e "${BLUE}🌐 Your Professional Corridor OS is now live at:${NC}"
    echo "   • Primary URL: https://corridor-os-80410062.web.app"
    echo "   • Firebase URL: https://corridor-os-80410062.firebaseapp.com"
    echo ""
    echo -e "${BLUE}📱 Features deployed:${NC}"
    echo "   ✅ Professional theme with business colors"
    echo "   ✅ Direct loading (no demo/mode selection)"
    echo "   ✅ Built-in contact section"
    echo "   ✅ Quantum-Photonic computing interface"
    echo "   ✅ Ubuntu-inspired desktop environment"
    echo "   ✅ Advanced window management"
    echo "   ✅ Settings accessible within OS"
    echo ""
    echo -e "${GREEN}🎉 Corridor OS Professional deployment complete!${NC}"
    echo ""
    echo -e "${BLUE}📋 Next steps:${NC}"
    echo "   1. Visit your live site to test"
    echo "   2. Add your company information tomorrow"
    echo "   3. Configure custom domain if needed"
    echo ""
    echo -e "${BLUE}🔧 Firebase Console:${NC}"
    echo "   https://studio.firebase.google.com/corridor-os-80410062"
else
    echo -e "${RED}❌ Deployment failed!${NC}"
    echo "Check the error messages above and try again."
    exit 1
fi

