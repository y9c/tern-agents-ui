// Command topencode is a chat with OpenCode drawn natively by Tern.
//
//	topencode [-model ID] [-mode MODE] [-effort LEVEL] [-resume ID | -continue] [-cwd DIR] [-log FILE [-debug]]
package main

import "github.com/saanuregh/tern-agents-ui/internal/chat"

func main() { chat.Main(chat.OpenCode()) }