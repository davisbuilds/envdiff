const databaseUrl = process.env.DATABASE_URL;
const port = process.env.PORT || 3000;
const region = process.env["AWS_REGION"];
const logLevel = process.env.LOG_LEVEL ?? "info";
const debug = process.env['DEBUG'] || false;
