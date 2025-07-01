package main

import (
	"fmt"
	"os"
	"strings"

	"rpg_demo/aseprite"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <aseprite-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s assets/Soldier.aseprite\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]

	// Load the Aseprite file
	aseFile, err := aseprite.LoadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading file: %v\n", err)
		os.Exit(1)
	}

	// Print file information
	fmt.Printf("Inspecting: %s\n", filename)
	fmt.Println(strings.Repeat("-", len(filename)+12))
	fmt.Printf("Dimensions:  %dx%d\n", aseFile.Header.Width, aseFile.Header.Height)
	fmt.Printf("Frames:      %d\n", aseFile.Header.Frames)
	fmt.Printf("Color Depth: %d bpp\n", aseFile.Header.ColorDepth)
	fmt.Printf("Speed:       %d ms (deprecated)\n", aseFile.Header.Speed)

	// Print animation tags
	if len(aseFile.Tags) > 0 {
		fmt.Println("\nAnimation Tags:")
		for _, tag := range aseFile.Tags {
			directionStr := getDirectionString(tag.Direction)
			repeatStr := getRepeatString(tag.Repeat)

			fmt.Printf("- \"%s\" (Frames: %d-%d, Direction: %s, Repeat: %s)\n",
				tag.Name, tag.FromFrame, tag.ToFrame, directionStr, repeatStr)
		}
	} else {
		fmt.Println("\nNo animation tags found.")
	}

	// Print frame durations if they vary
	fmt.Println("\nFrame Information:")
	for i, frame := range aseFile.Frames {
		if frame.Header.Duration > 0 {
			fmt.Printf("Frame %d: %d ms\n", i, frame.Header.Duration)
		}
	}

	// Summary for developers
	fmt.Println("\nDeveloper Summary:")
	fmt.Printf("- Total animation length: %d frames\n", len(aseFile.Frames))
	if len(aseFile.Tags) > 0 {
		fmt.Printf("- Animation sequences: %d\n", len(aseFile.Tags))
		fmt.Println("- Use tag names to reference specific animations in your game code")
	} else {
		fmt.Println("- No tagged sequences - consider adding animation tags in Aseprite")
	}
}

func getDirectionString(direction uint8) string {
	switch direction {
	case aseprite.DirectionForward:
		return "Forward"
	case aseprite.DirectionReverse:
		return "Reverse"
	case aseprite.DirectionPingPong:
		return "Ping-pong"
	case aseprite.DirectionPingPongRev:
		return "Ping-pong Reverse"
	default:
		return fmt.Sprintf("Unknown (%d)", direction)
	}
}

func getRepeatString(repeat uint16) string {
	switch repeat {
	case 0:
		return "Infinite"
	case 1:
		return "Once"
	default:
		return fmt.Sprintf("%d times", repeat)
	}
}
