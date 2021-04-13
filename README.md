# ubuntu-calendar-notifications

Ubuntu desktop notifications for Google Calendar events.

## Setup and Usage

1. Create a project and a service account in Google Cloud Console
2. Save the JSON credentials file of the service account somewhere
3. Install go 1.16
4. Clone this repo
5. Run `go install` on the root
6. Create a shell command `calendar-ack` with the following script:

```bash
#!/bin/bash

CREDENTIALS_FILE=<path to JSON file with the google service account credentials>
GMAILS=<list of gmails separated by comma>

kill -15 $(ps aux | grep ubuntu-calendar-notifications | awk '{print $2}') 2> /dev/null
CREDENTIALS_FILE=$CREDENTIALS_FILE GMAILS=$GMAILS ubuntu-calendar-notifications >> <path to logs file> 2>&1 &
```
7. Add a program to Ubuntu's Startup Application Preferences that runs the command `calendar-ack`
8. (Optional) Go to the Calendar Settings of each gmail and grant access to the service account gmail to improve the notifications

Use the `calendar-ack` command to restart the process and stop notifications for the current event.
