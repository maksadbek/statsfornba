# System for NBA player's statistics

Pre-requisites:
- Code written entirely in Golang or Java
- Project delivered using Docker Compose / Minikube
-  Use a relational database for the task.

1. Log (save) NBA player statistics:
- Each player plays for a basketball team.
- Stat line per game is: Points, Rebounds, Assists, Steals, Blocks, Fouls, Turnovers, Minutes Played.
- The input should be validated as follows:
    - Points, rebounds, assists, steals, blocks, turnovers have to be a positive integer.
    - Fouls is an integer with max of 6.
    - Minutes played is a float between 0 and 48.0
    - This path is consumed by a non-human system

2. Calculate aggregate statistics:
a. Season Average per player
b. Season Average per team
This path is consumed by a human

Mandatory considerations:
- The system should be highly available and scalable: it should support batches of tens or hundreds of requests concurrently.
- The system should serve up to date data. once data is written, it should be available for statistics
- The system should have a solid (not necessarily SOLID from the SOLID principle) architecture
- The system should be maintainable and support frequent updates

---


### Solution

Use PostgreSQL as a database, and Kafka as a message queue.
All aggregated data is kept in 2 tables:
- player_avg keeps aggerated data for players for each season
- team_avg keeps aggerated data for teams for each season

Each log creates two inserts into both tables. Instead of keeping all the events, these tables keep only sum of all stats.
But also keeps game count. This schema allows to calculate the average value of metrics easily my dividing values to game count. 

To scale the writes to database, I've used Kafka as a message queue. All the incoming events are sent to Kafka topic.
The consumer application handles these events and inserts into database.


* Scalability of the system - writes can be easily scaled by sharding the Kafka topic into partitions and adding more consumer
applications. But the database write performance is a bottleneck. The workaround would be sharding the database as well. For example
the using CitusDB. Reads can be scaled by adding replicas to the database cluster.

* High availability. Availability is the ability of the system to operate w/o single point of failure. I've used only one instance of
Postgresql. But in live environment, we can use multiple standby replica instances for fault tolerance, and Patroni for managing failover.

* Serve up to date information. I assume "high consistency" is what you mean. If we want High availability, we have to give up
consistency. Because of the CAP theorem, we either choose being available or consistent during partitioning.
It's achievable, with a small setup, having __standby__ servers. But when we start scaling, writes will slower and slower since
we add more instances.

* I used Postgresql and Kafka because I'm more familiar with them. But also I believe they're scalable and perfectly covers requirements.

#### API reference

- `POST /upload`
   Receives a CSV file of stats and these stats are written to database.

   Example:

   1. create `stats.csv` file
   ```
   team,player,season,points,rebounds,assists,steals,blocks,fouls,turnovers,minutes played
   LAL,LeBron James,2023-2024,28,7,8,1,1,2,3,35
   GSW,Stephen Curry,2023-2024,32,5,6,2,0,3,4,37
   BKN,Kevin Durant,2023-2024,27,8,5,1,2,3,3,34
   MIL,Giannis Antetokounmpo,2023-2024,30,12,6,1,1,4,3,36
   DAL,Luka Dončić,2023-2024,29,9,9,1,0,2,5,38
   PHI,Joel Embiid,2023-2024,31,11,4,0,3,4,4,33
   DEN,Nikola Jokić,2023-2024,26,12,10,1,1,2,3,34
   BOS,Jayson Tatum,2023-2024,27,7,4,1,0,2,3,36
   ATL,Trae Young,2023-2024,24,3,10,1,0,3,5,35
   ```

   2. Upload CSV file using curl:
   ```bash
   curl -XPOST  -F "file=@stats.csv" localhost:8080/upload
   ```


- `GET /:team/:season`
   Query aggregated stats for the team, and for the specific season.

   Example:
   ```bash
   curl localhost:8080/NOP/2023-2024
   {"Points":24,"Rebounds":4.5,"Assists":4.5,"Steals":1,"Blocks":0.5,"Fouls":3,"Turnovers":2.5,"MinutesPlayed":51.75}
   ```

- `GET /:team/:season/:player`
   Query aggregated stats for the team, and for the specific season.
   ! Don't forget to url escape the player name if it contains empty space
   Example:
   ```bash
   curl localhost:8080/NOP/2023-2024/Brandon%20Ingram
   {"Points":25,"Rebounds":5,"Assists":4,"Steals":1,"Blocks":1,"Fouls":3,"Turnovers":2,"MinutesPlayed":51}
   ```



####  Example usage

1. Install docker-compose
2. Run `docker-compose up`, when it's finished you can use API available on `8080` port.
3. Upload csv file: `curl -XPOST  -F "file=@stats.csv" localhost:8080/upload`. Use sample data from `sample/stats.csv`.
4. Query:
  `curl localhost:8080/NOP/2023-2024`
  `curl localhost:8080/NOP/2023-2024/Brandon%20Ingram`
