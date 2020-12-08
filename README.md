# AYBUSH BOT

Aybush bot is a discord bot built with Golang and discordgo library.

# Getting Started
- Using docker

  Aybush bot can be easily deployed using docker. You can edit Dockerfile according to your preferences.

  For example:
  ```
  docker build -t aybush-bot:1.0.0 .
  docker run --name AybushBot -p 8000:8000 aybush-bot:1.0.0
  ```
  
- Using makefile

  I recommend you to run aybush bot using docker, it makes deployment very easy. If you don't like to use docker, you can run the app using make. However, if you are going to use make, you should be careful about environment variables which application could use.

  ```
  make all
  ```
  The command will build the Golang application and put the executable into what "executablePath" environment variable is set. The name of the executable is going to be what "executableName" environment variable is set.

# Configuration

TODO
