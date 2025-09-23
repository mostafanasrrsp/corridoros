# Firebase Deployment Instructions

## Step 1: Login to Firebase
Run this command in your terminal:
```bash
firebase login
```
- Choose "n" when asked about Gemini features
- Follow the browser authentication process

## Step 2: Initialize Firebase Project
```bash
firebase init hosting
```
- Choose "Use an existing project" or "Create a new project"
- Set public directory to "." (current directory)
- Configure as single-page app: Yes
- Set up automatic builds: No (for now)

## Step 3: Update Project ID
Edit `.firebaserc` and replace `your-project-id` with your actual Firebase project ID.

## Step 4: Deploy
```bash
firebase deploy
```

## Step 5: Set Custom Domain (Optional)
1. Go to Firebase Console → Hosting
2. Click "Add custom domain"
3. Enter your domain name
4. Follow the DNS configuration instructions

## Files Included in Deployment:
- ✅ index.html
- ✅ styles.css
- ✅ All JavaScript files (quantum.js, photon.js, etc.)
- ✅ README.md

## Files Excluded:
- ❌ PDF files (research paper)
- ❌ Video files (MP4)
- ❌ Design files (DXF, SVG)
- ❌ CSV files (BOM)

## After Deployment:
Your Corridor Computer emulator will be available at:
`https://your-project-id.web.app`

Or with custom domain:
`https://yourdomain.com`
