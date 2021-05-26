package main

import (
	"flag"
	"fmt"

	"github.com/chai2010/gettext-go/po"
)

func main() {
	var output string
	flag.StringVar(&output, "output", "", "")
	flag.Parse()
	args := flag.Args()

	var mimeHeader po.Header
	var messages []po.Message

	for _, arg := range args {
		file, err := po.LoadFile(arg)
		if err != nil {
			fmt.Println(err)
			return
		}

		if arg == output {
			mimeHeader = file.MimeHeader
		}

		if len(messages) == 0 {
			messages = file.Messages
		} else {
			for _, msg := range file.Messages {
				foundIdx := -1
				for idx, old := range messages {
					if msg.MsgId == old.MsgId {
						foundIdx = idx
						break
					}
				}
				if foundIdx < 0 {
					messages = append(messages, msg)
					continue
				}
				if messages[foundIdx].MsgStr == "" {
					messages[foundIdx].MsgStr = msg.MsgStr
				}
			}
		}
	}

	file := po.File{
		MimeHeader: mimeHeader,
		Messages:   messages,
	}

	if output == "" {
		fmt.Println(file.String())
		return
	}
	err := file.Save(output)
	if err != nil {
		fmt.Println(err)
		return
	}
}
