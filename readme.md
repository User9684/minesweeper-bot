# Features
- Minesweeper, three difficulties
- Server-Specific leaderboard
- Global leaderboard

# Admin Commands
- Blacklist / Unblacklist user
- Configure automatically editing leaderboard messages

# Selfhosting Instructions (docker)
- Clone `app.example.env` and `db.example.env` and rename them to `app.env` and `db.env`
- Populate `app.env` and `db.env` accordingly 
- Run the command `docker-compose -f docker-compose.yml up`
- Voilà!

# Selfhosting Instructions (standalone)
- Create a MongoDB host, instructions can be found on the MongoDB website.
- Create a file named `.env` inside the `app` folder
- Populate accordingly to the example env file inside `minesweeper_docker`
- Run the command `go build .` and then run the built file.
- Voilà!
