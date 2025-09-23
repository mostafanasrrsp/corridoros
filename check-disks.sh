#!/bin/bash

# Comprehensive Disk Detection Script
# Shows all available disks and USB devices

echo "🔍 DISK DETECTION REPORT"
echo "========================"
echo

echo "📋 1. ALL DISKS (diskutil list):"
echo "--------------------------------"
diskutil list
echo

echo "📋 2. EXTERNAL DISKS ONLY:"
echo "-------------------------"
diskutil list | grep -i external
echo

echo "📋 3. MOUNTED VOLUMES:"
echo "---------------------"
ls -la /Volumes/
echo

echo "📋 4. USB DEVICES (system_profiler):"
echo "-----------------------------------"
system_profiler SPUSBDataType | grep -A 20 -B 5 "USB"
echo

echo "📋 5. ALL BLOCK DEVICES:"
echo "-----------------------"
ls -la /dev/disk*
echo

echo "📋 6. DISK UTILITY INFO:"
echo "-----------------------"
diskutil info /dev/disk0 2>/dev/null || echo "No disk0"
diskutil info /dev/disk1 2>/dev/null || echo "No disk1"
diskutil info /dev/disk2 2>/dev/null || echo "No disk2"
diskutil info /dev/disk3 2>/dev/null || echo "No disk3"
diskutil info /dev/disk4 2>/dev/null || echo "No disk4"
diskutil info /dev/disk5 2>/dev/null || echo "No disk5"
echo

echo "📋 7. RECENTLY PLUGGED DEVICES:"
echo "------------------------------"
system_profiler SPUSBDataType | grep -A 10 -B 2 "Product ID\|Vendor ID\|Serial Number"
echo

echo "📋 8. DISK USAGE:"
echo "----------------"
df -h
echo

echo "📋 9. MOUNT POINTS:"
echo "------------------"
mount | grep -E "(disk|usb|external)"
echo

echo "✅ Disk detection complete!"
echo
echo "💡 COMMON USB LOCATIONS:"
echo "  • /Volumes/[USB_NAME]"
echo "  • /dev/disk[number]"
echo "  • /media/[username]/[USB_NAME] (Linux)"
echo
echo "🔧 TO MOUNT USB MANUALLY:"
echo "  • diskutil mount /dev/disk[number]"
echo "  • diskutil unmount /dev/disk[number]"
echo "  • diskutil eject /dev/disk[number]"
