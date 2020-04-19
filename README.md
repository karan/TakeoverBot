# TakeoverBot

Download [fake local news sites](https://github.com/MassMove/AttackVectors), search for mentions of each on Twitter, and reply to them with an informative message.

A dumb "database" of CSV files is saved for recall.

```
$ cp template.env .env

# Add API keys

$ go run main.go
```

## TODO

- [ ] Using an actual DB for tracking tweets sent (instead of a local csv file)
- [ ] Smarter rate limit handling
- [ ] Discarding old tweets (say older than a week)
- [ ] Possibly running in multiple threads (go func())
- [ ] Variations of messages posted
- [ ] User opt-out
- [ ] Exclude tweets from accounts that belong to fake journos
- [ ] Saving aggregate data for analysis
- [ ] Linking to a form to collect feedback
