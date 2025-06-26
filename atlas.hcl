data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "./cmd/atlas-loader",
  ]
}

env "local" {
  src = data.external_schema.gorm.url
  dev = "docker://postgres/15/dev?search_path=public"
  migration {
    dir = "file://migrations"
  }
  url = getenv("DB_URL")  // Uses your existing DB_URL from .env
}

env "production" {
  src = data.external_schema.gorm.url
  dev = "docker://postgres/15/dev?search_path=public"
  migration {
    dir = "file://migrations"
  }
  url = getenv("DB_URL")
}
