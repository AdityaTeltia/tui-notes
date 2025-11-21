#!/bin/bash
# Quick script to associate Elastic IP via AWS CLI

echo "Associating Elastic IP 100.30.46.112 with EC2 instance..."

# Get instance ID
INSTANCE_ID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id)
echo "Instance ID: $INSTANCE_ID"

# Get allocation ID for the Elastic IP
ALLOCATION_ID=$(aws ec2 describe-addresses --public-ips 100.30.46.112 --query 'Addresses[0].AllocationId' --output text 2>/dev/null)

if [ -z "$ALLOCATION_ID" ]; then
    echo "Error: Could not find Elastic IP allocation. Please associate manually via AWS Console."
    exit 1
fi

echo "Allocation ID: $ALLOCATION_ID"

# Associate the Elastic IP
aws ec2 associate-address \
    --instance-id "$INSTANCE_ID" \
    --allocation-id "$ALLOCATION_ID" \
    --allow-reassociation

echo "Done! Your instance should now have IP: 100.30.46.112"
