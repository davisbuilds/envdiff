import os

database_url = os.environ["DATABASE_URL"]
redis_url = os.getenv("REDIS_URL")
debug = os.getenv("DEBUG", "false")

