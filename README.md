# ![](./sensum_logo.png "Sensum") SensumBot 

Telegram Bot for fetching the last messages from the famous app [Sensum](https://emeks.gitlab.io/sensum/) to a chat or a group

## Install

1. Set up the bot token and other variables in `config.json`

2. Set up the `$GOPATH` and such 

3. Build the executable

    ``` sh
    $ go build src/*
    ```

## Run

Simply run the executable

``` sh
$ ./bot
```

## To Do

- [x] Add a config file
- [x] Implement sensations votes counter and update them in the sent messages
- [x] Modularize components into different files
- [x] Add version release endpoint
- [x] Publish the repo

## Author

[Ariel Leandro Aguirre](mailto:ariedro@gmail.com)
