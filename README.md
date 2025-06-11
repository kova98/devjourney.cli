# DevJourney CLI Tool

DevJourney CLI is a tool that allows uploading entries and content to devjourney.io 

## Features

- accept an entry in a markdown file
- parse it to extract:
    - text content
    - metadata
    - media content
- upload media content
- replace relative links with links of uploaded media
- authenticate using api key from ENV or parameter
- submit the finished content to devjourney

## Commands

```bash
# uploads markdown entry and all its media content
devjourney upload <file> [--api-key <key>]
```