# TelegramGPTBot
Super simple telegram -> OpenAI API bot

## Build

`go build -o bot .`

## Run
`./bot -c config.cfg`
where config.cfg has structure
```ini
mode = debug
telegram_token = place_your_telegrambot_token_here
chatgpt_token = place_your_chatgptapi_token_here
bot_members = memberTelegramId1,memberTelegramId2,memberTelegramIdN
```

## Bot commands
/drop - drop chatting context