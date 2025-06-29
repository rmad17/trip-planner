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
  dev = "docker://postgres/15/dev?search_path=public"
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
