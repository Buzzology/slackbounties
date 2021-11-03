## Configuration
Used to configure and customise the service. 

BalanceDecayPercentage: The percentage each user's balance will decay each night.

DailyIncome: The number of tokens each user will receive "for free".

BountyReactionCost: The amount that will be deducted from the user's balance for each `bounty` reaction they place on a message.

AwardAmount: The amount that will be awarded to the reviewer for each `bounty` emote on the message.

## SlackApp creation
- Copy manifest from existing app
- Make slash commands urls https (revert after creating app)
- Install to workplace
- Set logo
- Ensure that event subscription url is set
- Grab bot token for the app


# Create a free tier vm in gcp

## Install MySQL
IMPORTANT: Use legacy auth etc.
- https://www.digitalocean.com/community/tutorials/how-to-install-the-latest-mysql-on-debian-10

### Setting up
- Connect to mysql from the vm ssh window: `mysql -uroot -p`  
- Create a new user:
```
CREATE USER 'slackbounty'@'localhost' IDENTIFIED BY '^Slackbounty';

GRANT ALL PRIVILEGES ON *.* TO 'slackbounty'@'localhost' WITH GRANT OPTION;

CREATE USER 'slackbounty'@'%' IDENTIFIED BY '^Slackbounty';

GRANT ALL PRIVILEGES ON *.* TO 'slackbounty'@'%' WITH GRANT OPTION;

FLUSH PRIVILEGES;
```
- Change auth method for the new user:
`ALTER USER 'slackbounty'@'localhost' IDENTIFIED WITH mysql_native_password BY '^Slackbounty';`


## Install golang
https://www.digitalocean.com/community/tutorials/how-to-install-go-on-debian-10

## Unzip files
```
apt-get install unzip

cd /path/to/file
unzip file.zip
```

SSH Key: ~/.ssh/slackbounty
## Running
- SSH into the box
- Use screen: https://superuser.com/a/632219/124014
  - Run: `screen -dmSL slackbounty go run main.go --config="./config/remote.toml"`
  - Rejoin: `screen -x slackbounty`