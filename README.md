# DeerHacks Back-end API

## See Setup Video 

**[Here](https://drive.google.com/file/d/1s-L1dwfkkaSeSEja4EYnT1HKaZZkaRQx/view?usp=sharing)**

## Resources mentioned in video
- https://www.elephantsql.com/ (Choose PostgreSQL database)
- https://www.jetbrains.com/datagrip/ (Visualize data)
- https://discord.com/developers/applications (Discord Developer portal)

### Common commands
- `./CompileDaemon -command="./deerhacks-backend"` (Hot-reload Tool)
- `go build` (install dependencies and build project)
- `go run <path>` (Run a specific go file)

### .env format
```
PORT=8000
DB_URL = "host=<server name here> user=<username here> password=<password here> dbname=<same as username> port=5432 sslmode=disable"
SECRET = "youcantypeanythingyouwanthere"

CLIENT_ID = ""
CLIENT_SECRET = ""
BOT_TOKEN = ""
REDIRECT_URI = ""

GUILD_ID = "967161405017055342"
```