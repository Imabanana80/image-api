# Image API
This API allows authorized users to upload images, with the URL of that image returned.

# Usage
Create/Upload a new image by sending a POST to `https://img.potassium.sh/new`. (or your appropriate domain)

Include the `x-api-key` header as well as `content-type`.
Supported image formats include `image/jpeg`, `image/png`, `image/gif` and `image/webp`.

The body should contain the image binary file.

Example response:
```json
{
  "url": "/images/1f10a5e9-dc8f-4673-9d13-5601833aaa06.png",
  "filename": "1f10a5e9-dc8f-4673-9d13-5601833aaa06.png"
}
```

The image will be accessible at `https://img.potassium.sh/images/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.png`
