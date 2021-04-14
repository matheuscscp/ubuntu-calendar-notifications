# ubuntu-calendar-notifications

Ubuntu insistent desktop notifications for Google Calendar events.

## Setup and Usage

1. Create a project and a service account in Google Cloud Console.
2. Create a key for the service account and save the JSON credentials file somewhere in your computer.
3. If compatible, download the [latest release](https://github.com/matheuscscp/ubuntu-calendar-notifications/releases/latest) binary and skip to step 7.
4. Install go 1.16.
5. Clone this repo or download the [latest release](https://github.com/matheuscscp/ubuntu-calendar-notifications/releases/latest) source code.
6. Run `go build` in the root folder to generate the `ubuntu-calendar-notifications` binary.
7. Create the following script somewhere in your computer (replacing the placeholders!):

```bash
#!/bin/bash

CREDENTIALS_FILE=<path to JSON file with the google service account credentials>
GMAILS=<list of gmails and/or calendar IDs separated by comma>

kill -15 $(ps aux | grep ubuntu-calendar-notifications | awk '{print $2}') 2> /dev/null
CREDENTIALS_FILE=$CREDENTIALS_FILE GMAILS=$GMAILS <path to ubuntu-calendar-notifications binary> >> <path to log file> 2>&1 &
```
8. Grant permission for execution to the script file (`$ chmod +x <path to script file>`).
9. To ensure the app is always running, add a program to Ubuntu's Startup Applications Preferences pointing to the script file.
10. To acknowledge notifications for on-going events, the script file should be placed in folder used by the `PATH` environment variable (e.g. `/usr/local/bin`), so you can use it like a shell command (e.g. `calendar-ack`).
11. To start the app for the first time, run the shell command created on the previous step (`$ calendar-ack`).
12. For each gmail/calendar ID in the configuration, go to the Google Calendar Settings and grant access to the service account (you can find the service account email in Google Cloud Console). This step is not necessary for public calendars.
