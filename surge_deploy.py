#!/usr/bin/env python3

import subprocess
import sys
import os

def deploy_to_surge():
    print("🚀 CorridorOS Surge Deployment")
    print("================================")
    
    # Get user input
    email = "mostafanasr@aucegypt.edu"
    password = input("Enter your password for Surge: ")
    
    print(f"\n📧 Using email: {email}")
    print("🔐 Deploying with your password...")
    
    try:
        # Create surge deployment
        import random
        import string
        random_domain = ''.join(random.choices(string.ascii_lowercase, k=8)) + '.surge.sh'
        cmd = ["surge", "--project", ".", "--domain", random_domain]
        
        # Start the process
        process = subprocess.Popen(
            cmd,
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            cwd="/Users/mnasr/Desktop/COS"
        )
        
        # Send email and password
        output, _ = process.communicate(input=f"{email}\n{password}\n")
        
        print("\n📋 Surge Output:")
        print(output)
        
        if "Success!" in output:
            print("✅ Deployment successful!")
            print(f"🌐 Your site is live at: https://{random_domain}")
        else:
            print("❌ Deployment may have failed. Check output above.")
            
    except Exception as e:
        print(f"❌ Error: {e}")
        print("\n🔧 Alternative: Try manual deployment")
        print("   1. Run: surge")
        print("   2. Enter email: mostafanasr@aucegypt.edu")
        print("   3. Enter your password")
        print("   4. Accept default domain or enter: corridoros.surge.sh")

if __name__ == "__main__":
    deploy_to_surge()
