
# DeerHacks API


[![DeerHacks Image](https://github.com/utmmcss/deerhacks/blob/c097731ac1a95f138462fbac6aa87ed0c7bfd191/public/backgrounds/collage.jpg?raw=true)](https://deerhacks.ca)

> DeerHacks Hackathon 2024 Backend API

[![Website Status](https://img.shields.io/website?down_color=red&down_message=offline&up_color=green&up_message=online&url=https%3A%2F%2Fdeerhacks.ca)](https://deerhacks.ca)

## Setup

1. Run `go build` to install dependencies
2. Add the required `.env` file with the schema specified below
3. Gather credentials from microservices and add it to the `.env` file

## Relevant URL'S

- [AWS S3 for saving resumes](https://aws.amazon.com/s3/pricing/?p=pm&c=s3&z=4)

- [AWS RDS for saving User Data](https://aws.amazon.com/rds/pricing/?p=ft&c=db)

- [Brevo for sending outbound emails](https://onboarding.brevo.com/account/register)

- [Discord Developer Portal](https://discord.com/developers/docs/intro)

  

## Getting Started

First, run the development server:

```bash
go  run  main.go
```

Send API requests to [http://localhost:8000](http://localhost:8000) (assuming port specified in .env file is 8000) with tools like Postman

### .env format

  

```bash
PORT=8000
DB_URL  =  "host=<server name here> user=<username here> password=<password here> dbname=<same as username> port=5432 sslmode=disable"
SECRET  =  "youcantypeanythingyouwanthere"

# From the Discord Developer Portal
CLIENT_ID  =  ""
CLIENT_SECRET  =  ""
BOT_TOKEN  =  ""
REDIRECT_URI  =  ""

# The discord server your discord bot will be in
GUILD_ID  =  "967161405017055342"

#Change this to "production" if public
APP_ENV  =  "development"
REGISTRATION_CUTOFF=1704085200  # (2024-01-01 00:00:00 EST)

# AWS IAM Credentials. Ensure full S3 access is given
AWS_ACCESS_KEY_ID  =  ""
AWS_SECRET_ACCESS_KEY  =  ""

# For sending emails
BREVO_API_KEY  =  ""
```