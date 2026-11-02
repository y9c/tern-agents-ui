// Command tpi is a chat with Pi drawn natively by Tern.
//
//	tpi [-model ID] [-mode MODE] [-effort LEVEL] [-resume ID | -continue] [-cwd DIR] [-log FILE [-debug]]
package main

import "github.com/saanuregh/tern-agents-ui/internal/chat"

func main() { chat.Main(chat.Pi()) }
