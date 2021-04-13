# ubuntu-calendar-notifications

Ubuntu insistent desktop notifications for Google Calendar events.

## Setup and Usage

1. Create a project and a service account in Google Cloud Console
2. Save the JSON credentials file of the service account somewhere in your computer
3. If compatible, download the [latest release](https://github.com/matheuscscp/ubuntu-calendar-notifications/releases/latest) binary and skip to step 7
4. Install go 1.16
5. Clone this repo or download the [latest release](https://github.com/matheuscscp/ubuntu-calendar-notifications/releases/latest) source code
6. Run `go build` in the root folder to generate the `ubuntu-calendar-notifications` binary
7. Create the following script replacing the placeholders:

```bash
#!/bin/bash

CREDENTIALS_FILE=<path to JSON file with the google service account credentials>
GMAILS=<list of gmails separated by comma>

kill -15 $(ps aux | grep ubuntu-calendar-notifications | awk '{print $2}') 2> /dev/null
CREDENTIALS_FILE=$CREDENTIALS_FILE GMAILS=$GMAILS <path to ubuntu-calendar-notifications binary> >> <path to log file> 2>&1 &
```
8. Add a program to Ubuntu's Startup Applications Preferences that runs the above script
9. To stop notifications for on-going events, create an easy shell command, like `calendar-ack`, that runs the above script
10. (Optional) Go to the Calendar Settings of each gmail and grant access to the service account email to improve the notifications (you can find the service account email in Google Cloud Console)
