curl -X POST http://localhost:10022/carousel \
  -H "Content-Type: application/json" \
  -d '{
        "title": "Demo Image",
        "source_url": "https://github.com/yourorg",
        "image_url": "https://upload.wikimedia.org/wikipedia/commons/6/6e/Matterhorn_from_Domh%C3%BCtte_-_2.jpg"
      }'

