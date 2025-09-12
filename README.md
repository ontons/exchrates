# exchrates
Test assesment
* Has two command line arguments:
    * To run server `exchrates server`
    * To fetch data `exchrates fetch`
* Configuration stored in environment variables (example provided in example.env file)
* By default runs on port `:8080`
* Has two endpoints `/latest` and `/history?currency=usd`
* Logs stored in `/logs` folder
* Docker volumes stored in `/volumes` folder
## Run in Docker
1. Clone repository `git clone https://github.com/ontons/exchrates.git`
2. Navigate to newly created folder `cd exchrates`
3. Create .env file in repo root folder (copy from example.env and change if needed) `cp example.env .env`
4. Build and run containers: `sudo docker compose up -d --build`
5. Run fetch command to fetch latest rates `sudo docker compose exec app /app/exchrates fetch`
6. Navigate to `http://localhost:8080/latest` or `http://localhost:8080/history?currency=usd`
7. Stop containers: `sudo docker compose stop`
8. Remove containers and volumes `sudo docker compose down -v`
## Run in VSCode
1. Clone repository `git clone https://github.com/ontons/exchrates.git`
2. Create .env file in root folder (copy from example.env)
3. Start db container `sudo docker compose start mariadb`
4. Navigate to `app` folder
5. Install dependencies `go mod tidy`
6. Rund and debug with one of the VSCode launch configs for server or fetch