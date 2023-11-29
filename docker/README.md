### Step 1. Build Docker Image ###

```bash
docker build -f docker/Dockerfile --target packetrusher --tag packetrusher:latest .
```
### Step 2. Launch Tester ###
Make sure you have set up core-network already since we will reuse docker network)

```bash
docker-compose -f docker/docker-compose.yaml up -d
```
