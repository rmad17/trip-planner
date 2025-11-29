data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "./cmd/atlas-loader",
  ]
}

env "local" {
  src = data.external_schema.gorm.url
  dev = "postgres://postgres:postgres@localhost:5432/triptest?search_path=public&sslmode=disable"
  migration {
    dir = "file://migrations"
  }
  url = getenv("DB_URL")
  
  # Add PostgreSQL specific settings
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
