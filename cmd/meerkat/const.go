package meerkat

const configTemplate = `
username: "#####"
password: "#####"

# interval
# in seconds.
interval: 15 

# sleeptime
# in seconds.
# after each request to get user's information , 
# we have to sleep , because instagram may ban our account.
sleeptime: 10

# output types: choose how you wants to know about users activity.
# types are : ["logfile", "telegram"]
# you can select multiple options using ',' seprator. ex. "telegram,logfile"
outputtype: "logfile"

# telegram bot token
# fill this field if you choose telegram in outputtype.
telegramtoken: "###"

# telegram id from user
# get it using @userinfobot on telegram
telegramuser: 0

targetusers: 
  - "###"
`
