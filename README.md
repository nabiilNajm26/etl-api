# ETL API

A Go-based REST API that converts CSV files into PostgreSQL database tables automatically. Built for teams tired of manual data entry.

**Live Demo**: https://etl-api-production.up.railway.app

## What it does

1. **Upload a CSV file** → API creates PostgreSQL table with proper data types
2. **Get instant access** → REST endpoints to query your data with pagination  
3. **Stay secure** → JWT authentication keeps your data isolated

## Why I built this

Small businesses waste hours manually entering CSV data. Sales teams get spreadsheets, marketing gets campaign reports, HR gets survey results - everyone's copying data by hand.

This API eliminates that. Upload once, access forever.

## Quick test

```bash
# Health check
curl https://etl-api-production.up.railway.app/health

# Register
curl -X POST https://etl-api-production.up.railway.app/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@company.com","password":"password123"}'

# Login (copy the token)
curl -X POST https://etl-api-production.up.railway.app/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@company.com","password":"password123"}'

# Upload CSV
curl -X POST https://etl-api-production.up.railway.app/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@yourdata.csv" \
  -F "table_name=Sales Data"

# Get your data back
curl "https://etl-api-production.up.railway.app/data/TABLE_ID" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Tech stack

- **Go 1.23** - Fast HTTP server, clean code
- **PostgreSQL** - Dynamic table creation, connection pooling
- **JWT auth** - Secure user sessions
- **Railway** - One-click deployment with Docker

## API endpoints

| Method | Endpoint | What it does |
|--------|----------|--------------|
| `GET` | `/health` | Check if API is running |
| `POST` | `/auth/register` | Create account |
| `POST` | `/auth/login` | Get access token |
| `POST` | `/upload` | Upload CSV file |
| `GET` | `/tables` | List your uploaded tables |
| `GET` | `/data/{id}` | Get table data (paginated) |
| `DELETE` | `/tables/{id}` | Delete table |

## What makes it useful

**Smart data type detection** - Automatically figures out if columns are dates, numbers, or text

**Handles messy data** - Cleans column names, handles missing values, validates file formats

**Scales with you** - Tested with 10MB files, handles thousands of rows per second

**Security first** - bcrypt passwords, parameterized queries, user data isolation

## Local setup

```bash
git clone https://github.com/nabiilNajm26/etl-api.git
cd etl-api
go mod tidy

# Set environment variables
export DATABASE_URL="postgresql://user:pass@localhost:5432/etldb"  
export JWT_SECRET="your-secret-key"
export PORT="8080"

# Run it
go run .
```

## Common use cases

**Sales teams**: Upload monthly reports, get REST API access for dashboards

**Marketing**: Import campaign data, analyze performance with pagination  

**HR departments**: Process survey results, generate insights

**E-commerce**: Import product catalogs, manage inventory data

**Anyone with CSV files**: Stop copying data manually, start using APIs

---

Built by [Nabiil Najm](https://github.com/nabiilNajm26) • [Source](https://github.com/nabiilNajm26/etl-api)