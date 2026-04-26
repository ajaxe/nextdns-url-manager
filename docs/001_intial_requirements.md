# Custom Next DNS Client

## Description

As a user, I would like to be able to define custom application group so that I can enable/disable the allowed URLs in Next DNS using their API. The application should allow users to create and manage custom groups of URLs for different applications. The application should also provide a way to enable or disable the allowed URLs in Next DNS using their API.

The application should have a timer funtion which allows user to disable an application after certain amount of time.

## Requirements

### Cross-platform with Modern & Intuitive layout

Cross-platform Terminal UI based application with a modern and intuitive terminal layout. The user should be able to use arrow keys & keyboard for all operations.

### Interaction with Next DNS API

Will interact with Next DNS API, documentation - https://nextdns.github.io/api/. The `client_id` will be accepted as a command-line argument att the start of the application.

### Application configuration

The Terminal UI will enable the user to define an application which is a group of URLs. This configuration will be persisted in a local yaml configuration file.

Analyze optimal yaml configuration structure that will meet the requirements.

As part of the use, a user may manually update the yaml configuration file outside of the application. The application must merge the changes done via the application with the file contents.

### Timer for Allowed URLs

I would like an option to enable an application for certain duration will an intuitive format, some suggestions like

* `5s` - indicates five seconds
* `70s` - indicates senventy seconds which is equal to one minute and ten seconds.
* `1m5s` or `1m 5s` - indicates one minute and five seconds.
* `1h5m` or `1h 5m` - indicates one hour and five minutes

Analyze and suggest implementation components that will run in the background once the Terminal UI is exited to complete the timer functionality.

## Tech Stack and Contraints

* Terminal UI based on golang framework. Suggest modern looking TUI framework with minimal foot-print.

* Must build for windows and linux environments.

* Next DNS `client` credentials are only accepted as cli arguments.