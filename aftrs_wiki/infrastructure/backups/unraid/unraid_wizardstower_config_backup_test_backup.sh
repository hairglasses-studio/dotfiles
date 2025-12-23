#!/bin/bash

# Quick test of backup functionality
echo "🧪 Testing enhanced backup system..."

# Test backup script syntax
bash -n backup_unraid.sh && echo "✅ Basic backup script syntax OK" || echo "❌ Basic backup script syntax error"
bash -n backup_unraid_enhanced.sh && echo "✅ Enhanced backup script syntax OK" || echo "❌ Enhanced backup script syntax error"

# Test analysis script syntax
python3 -m py_compile analyze_unraid.py && echo "✅ Basic analysis script syntax OK" || echo "❌ Basic analysis script syntax error"
python3 -m py_compile analyze_unraid_enhanced.py && echo "✅ Enhanced analysis script syntax OK" || echo "❌ Enhanced analysis script syntax error"
python3 -m py_compile monitor_unraid.py && echo "✅ Monitoring script syntax OK" || echo "❌ Monitoring script syntax error"

# Test if we can create backup directory
mkdir -p test_backup && echo "✅ Can create backup directories" || echo "❌ Cannot create backup directories"
rmdir test_backup

# Test configuration file
if [ -f "config.yaml" ]; then
    echo "✅ Configuration file exists"
else
    echo "⚠️  Configuration file not found - will use defaults"
fi

echo "🧪 Enhanced test completed!"
