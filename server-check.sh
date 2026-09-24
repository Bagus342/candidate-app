#!/bin/bash

CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | sed "s/.*, *\([0-9.]*\)%* id.*/\1/" | awk '{printf "%.0f", 100 - $1}')
if [ "$CPU_USAGE" -lt 80 ]; then
    echo "CPU : OK"
else
    echo "CPU : WARNING"
fi

MEM_FREE=$(free -m | awk 'NR==2{printf "%.0f", $4*100/$2 }')
if [ "$MEM_FREE" -gt 20 ]; then
    echo "Memory: OK"
else
    echo "Memory: WARNING"
fi

DISK_USE=$(df -h / | awk '$NF=="/"{printf "%s", $5}' | sed 's/%//')
if [ "$DISK_USE" -lt 80 ]; then
    echo "Disk : OK"
else
    echo "Disk : WARNING"
fi

if systemctl is-active --quiet docker; then
    echo "Docker: OK"
else
    echo "Docker: WARNING"
fi

if ping -c 1 8.8.8.8 &> /dev/null; then
    echo "Network: OK"
else
    echo "Network: WARNING"
fi