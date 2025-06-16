
IMGID=$1
echo "Fetching Image for ID: $IMGID"
curl -L http://localhost:10022/serve?id=$IMGID --output downloaded_image.webp

