package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	pb "github.com/demisto/pb-go"
)

var (
	appId, userKey, name, out, file, input, cmd *string
	debug                                       *bool
)

func init() {
	appId = flag.String("appId", "", "Application ID as received from pandoranbots.")
	userKey = flag.String("userKey", "", "User key as received from pandoranbots.")
	name = flag.String("name", "", "The bot name to use.")
	out = flag.String("out", "", "Output file. If not specified will write to standard output.")
	file = flag.String("file", "", "Input file for uploads or file name for downloads.")
	input = flag.String("input", "", "Input to talk.")
	cmd = flag.String("cmd", "", "The command to execute. Can be one of the following: list/createBot/deleteBot/listFiles/download/upload/downloadBot/deleteFile/verify/talk")
	debug = flag.Bool("debug", false, "Debug output")
}

func main() {
	flag.Parse()
	c, err := pb.New(pb.SetErrorLog(log.New(os.Stderr, "", log.Lshortfile)), pb.SetCredentials(*appId, *userKey))
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	if *debug {
		pb.SetTraceLog(log.New(os.Stderr, "TRACE: ", log.Lshortfile))(c)
	}
	if strings.ToLower(*cmd) != "list" && *name == "" {
		fmt.Println("You must specify the bot name")
		os.Exit(1)
	}
	switch strings.ToLower(*cmd) {
	case "list":
		res, err := c.List()
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(2)
		}
		for _, s := range res {
			fmt.Println(s)
		}
	case "createbot":
		err = c.CreateBot(*name)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Println("Bot successfully created.")
		}
	case "deletebot":
		err = c.DeleteBot(*name)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Println("Bot successfully deleted.")
		}
	case "upload":
		if *file == "" {
			fmt.Println("You must specify the file name to upload")
			os.Exit(1)
		}
		err = c.UploadFileFromPath(*name, *file)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Println("File successfully uploaded.")
		}
	case "download":
		if *out == "" {
			if *file == "" {
				fmt.Println("You must specify the file name to download")
				os.Exit(1)
			}
			err = c.GetFile(*name, *file, os.Stdout)
			if err != nil {
				fmt.Printf("%v\n", err)
			}
		} else {
			err = c.GetFileToPath(*name, *out)
			if err != nil {
				fmt.Printf("%v\n", err)
			} else {
				fmt.Println("File successfully downloaded.")
			}
		}
	case "deletefile":
		if *file == "" {
			fmt.Println("You must specify the file name to delete")
			os.Exit(1)
		}
		err = c.DeleteFile(*name, *file)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Println("File successfully deleted.")
		}
	case "listfiles":
		res, err := c.ListFiles(*name)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Printf("%v\n", res)
		}
	case "downloadbot":
		if *out == "" {
			err = c.DownloadFiles(*name, os.Stdout)
			if err != nil {
				fmt.Printf("%v\n", err)
			}
		} else {
			err = c.DownloadFilesToPath(*name, *out)
			if err != nil {
				fmt.Printf("%v\n", err)
			} else {
				fmt.Println("Bot files successfully downloaded.")
			}
		}
	case "verify":
		err = c.Verify(*name)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Println("Bot verified.")
		}
	case "talk":
		// While we are not quiting, let's talk
		if *input == "" {
			var sessionId int
			r := bufio.NewReader(os.Stdin)
			for line, err := r.ReadString('\n'); strings.ToLower(line) != "exit\n" && err == nil; line, err = r.ReadString('\n') {
				res, err := c.Talk(*name, line, "", sessionId, false)
				if err != nil {
					fmt.Printf("%v\n", err)
					os.Exit(2)
				}
				sessionId = res.SessionId
				for _, s := range res.Responses {
					fmt.Println(s)
				}
			}
		} else {
			res, err := c.Talk(*name, *input, "", 0, false)
			if err != nil {
				fmt.Printf("%v\n", err)
			} else {
				fmt.Printf("%v", res)
			}
		}
	default:
		fmt.Printf("Command [%s] was not recognized\n", *cmd)
	}
}
