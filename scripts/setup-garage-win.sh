#!/bin/bash
set -e

source .env
sleep 20

CONTAINER="garaged-image-processor"

if ! docker ps | grep -q $CONTAINER; then
    echo "Container $CONTAINER is not running"
    exit 1
fi

NODE_ID=$(docker exec $CONTAINER //garage status | grep -oE '[a-f0-9]{16}' | head -n 1)

CURRENT_VERSION=$(docker exec $CONTAINER //garage layout show | grep "Current cluster layout version:" | grep -oE '[0-9]+' || echo "0")
NEW_VERSION=$((CURRENT_VERSION + 1))
docker exec  $CONTAINER //garage layout assign -z dc1 -c ${GARAGE_CAPACITY} $NODE_ID
docker exec $CONTAINER //garage layout apply --version $NEW_VERSION

docker exec $CONTAINER //garage key import ${S3_ACCESS_KEY} ${S3_SECRET_KEY} --yes || true

docker exec $CONTAINER //garage bucket create ${S3_BUCKET} || true

docker exec $CONTAINER //garage bucket allow ${S3_BUCKET} --read --write --owner --key ${S3_ACCESS_KEY}

echo "Garage configured successfully!"