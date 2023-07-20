# Inspiration
Giving credit where credit is due, this bot was pretty much entirely ripped off of [worldscutestvoid](https://github.com/cmsteffey)'s bot

# Features
- Minesweeper, three difficulties
- Custom Minesweeper game command
- Server-Specific leaderboard
- Global leaderboard
- Achievements

# Admin Commands
- Blacklist / Unblacklist user
- Configure automatically editing leaderboard messages

# Selfhosting Instructions (docker)
- Clone `app.example.env` and `db.example.env` and rename them to `app.env` and `db.env`
- Populate `app.env` and `db.env` accordingly 
- Run the start command
- Voilà!

# Start Command
If you do not already have a mongoDB host, run the command below
- `docker-compose -f docker-compose.yml up`

If you already have a mongoDB host, run the command below
- `docker-compose -f docker-compose.bot.yml up`

# Selfhosting Instructions (standalone)
- Create a MongoDB host, instructions can be found on the MongoDB website.
- Create a file named `.env` inside the `app` folder
- Populate accordingly to the example env file inside `minesweeper_docker`
- Run the command `go build .` and then run the built file.
- Voilà!
