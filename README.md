# Nutri Price Server

**Nutri Price Server** is the server-side application powering the [Nutri Price mobile app](https://github.com/SarunasBucius/nutri-price-app).  
It provides a GraphQL API, and in some cases RESTful API, for accessing product information, nutritional values, and recipes, enabling seamless meal planning, budgeting, and dietary awareness.

## 🚀 Features

- GraphQL API endpoints for products, nutrition values, purchases
- RESTful API endpoints for recipes
- Product and Recipe data management
- Integration with PostgreSQL database
- Database access with `pgx`
- Deployment-ready for Google Cloud

## 🛠️ Tech Stack

- **Go** (Golang)
- **gqlgen** (GraphQL server library for Go)
- **PostgreSQL** (Relational database)
- **pgx** (Go PostgreSQL driver and toolkit)
- **Google Cloud (GCP)** (deployment and hosting)

## 📦 Installation

### Prerequisites

- **Go** (v1.24 used during development)
- **Docker** (v27.3.1 used during development)
- **just** (optional; a command runner to simplify common tasks)

### Setup

1. Copy the example environment variables file and set your own values:
   ```bash
   cp .env.example .env
   ```
2. Ensure the required environment variables are set:

| Variable       | Description                                   |
|:---------------|:----------------------------------------------|
| `DATABASE_URL` | PostgreSQL database connection string         |
| `PORT`         | Port number for the API server               |

3. Run the Application

- Using `just` (recommended): `just start`

- *(You can check the `justfile` for other available commands.)*