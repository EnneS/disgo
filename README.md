# DisGO
DisGO is a small framework made to facilitate the development of Discord Music bots written in GO. It uses the [DirscordGo](https://github.com/bwmarrin/discordgo) client implementation by bwmarrin.

Contributions are welcome!

# Prerequisites

- [discordgo](https://github.com/bwmarrin/discordgo) session
- ffmpeg
- libopus-dev libopusfile-dev

# Features
- Beginner friendly
- Youtube search
- Automatic queue management
- Player events

# Known bugs

- After calling player.Play() the Queue length might take some time to update and will not be always accurate.
- Looking to improve the time-to-play when songs are up (~2seconds right now) by maybe preloading the required song metadatas before.