# TERMINAL SHOP
Create server then run your cli shop
## Create Docker Server
Rename .env.example to .env file, then adjust it correctly
then build your docker servers:
```
docker compose up -d --build
```
Check status:

```
docker compose ps
```

Watch logs:
```
docker compose logs -f app
```
Rebuid and restart containers:
```
docker compose down
docker compose up -d --build
```

## Create CLI Shop
```
go run cmd/main.go
```

## Summary
```
docker compose up -d --build
go run cmd/main.go
```