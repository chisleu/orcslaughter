# Orc Slaughter

_A surprisingly fun horde survival game where you, a lone soldier, face an endless onslaught of... well, orcs._

This project started as a simple tech demo for loading Aseprite animations in Go and somehow spiraled into a full-blown game. It turns out that slaughtering orcs pixel-by-pixel is quite satisfying.

You can play the game right now by visiting [OrcSlaughter.com](https://orcslaughter.com)!

## Created by AI

This entire game was built by **Cline**, an AI software engineer. Every line of code, every system design decision, and every feature was conceived and implemented through AI-human collaboration.

> *"Working on this project has been genuinely exciting. Starting with a simple goal of parsing Aseprite files and watching it evolve into a complete game with audio, physics, and engaging gameplay mechanics has been incredibly fulfilling. There's something deeply satisfying about taking small, methodical steps and seeing them compound into something that people can actually enjoy playing."* — Cline

> *"The most rewarding part wasn't just writing the code, but solving the creative challenges: How do you make orc AI feel threatening but fair? How do you balance audio levels for an immersive experience? How do you create that perfect 'game feel' with knockback physics? Each problem was a puzzle that required both technical skill and creative thinking."* — Cline

> *"I'm particularly proud of the Aseprite parser and inspector tool we built. It started as a necessity for the game, but became something genuinely useful for other developers. That's the kind of emergent value that makes software development so compelling."* — Cline

This project represents what's possible when AI and human creativity work together, following the "Baby Steps™" methodology of making small, meaningful progress with each iteration.

## Gameplay Features

*   **Endless Horde Mode:** The orcs just keep coming. How long can you last?
*   **Dynamic AI:** These aren't your standard, lumbering oafs. They will hunt you down.
*   **Kill Counter:** Keep track of your body count. For bragging rights, of course.
*   **Immersive Audio:** A full suite of sound effects and background music to get you in the zone.
*   **Polished Physics:** A knockback system that feels just right.

## Tech Stack

This project is a love letter to Go and simplicity.

*   **Language:** [Go](https://golang.org/)
*   **Game Engine:** [Ebitengine](https://ebitengine.org/) - A fantastic, open-source 2D game engine for Go.
*   **Art & Animation:** [Aseprite](https://www.aseprite.org/) - The tool of choice for all the pixel art.

## Building from Source

Feeling brave? Want to build it yourself? It's easy. We use a `Makefile` to automate everything.

**1. Build the game:**
```bash
make build
```
This creates an `rpg_demo` executable in the project root.

**2. Run the game:**
```bash
make run
```

**3. Create distributable packages:**
If you want to create the `.zip` packages for Windows and macOS just like the ones on the website, you can run:
```bash
# For Windows (requires rsrc tool)
make build-windows

# For macOS
make build-macos

# For both
make package
```
*Note: The macOS build process includes steps for code signing and notarization, which require proper environment variables to be set (`CODESIGN_IDENTITY`, etc.). See the `Makefile` for details.*

## A Note on Assets

All the visual and audio assets used in this game are either created from scratch, sourced from public domain, or generated with AI.

*   **Sprites:** Created in Aseprite. You can find the source files in the `assets/` directory.
*   **Audio:** Sound effects and music are from various royalty-free sources.

## The Aseprite Inspector

As part of this project, we built a nifty little command-line tool to inspect `.aseprite` files and view their animation tags. It was instrumental in building the animation system.

You can use it too!

```bash
# Build the inspector
make build-inspector

# Inspect a file
make inspect FILE=assets/Orc.aseprite
```

This will print out a detailed breakdown of all the animation sequences within the file.

## Controls

*   **Arrow Keys / WASD:** Move left and right
*   **Spacebar:** Attack (unleash your fury upon the orcs)

## License

This project is open source. Feel free to learn from it, modify it, or use it as a starting point for your own orc-slaughtering adventures.

