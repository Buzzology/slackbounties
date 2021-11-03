# Configuration
Note that there is a sample configuration file in this directory: `sample.toml`

## Glossary
### DbConnection
The (MySQL compatible) database connection string that should be used for the bot.

### ReleaseBountyReaction
The reaction that should be added to a message in order to release the bounty. This will send the current bounty amount to the most recent person who has used the :TaskCompletedByMeReaction:.

### TaskCompletedByMeReaction
This reaction is used to signal that a task has been completed. The user who applies this emote to the message will be the one that receives the bounty if the message owner applies the :ReleaseBountyReaction:. If multiple people have used this emote the bounty will go to the most recent. 

### DailyDecay
This amount will be deducted from each user's balance when the daily tickover occurs. It will be applied before :DailyIncome: but will not drop a user's balance below zero.

### ApiConfig

#### Endpoint
This is the base url that the bot should use to interact with the slack api.

#### Token
This is the token that should be used by the bot to authenticate with the SlackApi.

### Boost Reactions
These are the emotes that are used to _boost_ the bounty on a particular message. You can add as many of these as you'd like so long as there's at least one.
