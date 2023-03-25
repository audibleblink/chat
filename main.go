package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/abiosoft/readline"
	"github.com/joho/godotenv"
)

var (
	session_queries []gpt3.ChatCompletionRequestMessage
	client          gpt3.Client
	home            string

	interativeF = flag.Bool("i", false, "Start an interactive session")
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	err = godotenv.Load(fmt.Sprintf("%s/.chatgpt", home))
	if err != nil {
		log.Fatal(err)
	}
	client = gpt3.NewClient(os.Getenv("API_KEY"))

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Println("\nQUERY\tQuery to submit")
		flag.PrintDefaults()
	}

	flag.Parse()
}

func main() {

	if *interativeF {
		l, err := readline.NewEx(&readline.Config{
			Prompt:            "\033[31mChatGPT»\033[0m ",
			HistoryFile:       fmt.Sprintf("%s/.chatgpt_history", home),
			InterruptPrompt:   "^C",
			EOFPrompt:         "exit",
			HistorySearchFold: true,
		})
		if err != nil {
			log.Fatal(err)
		}
		defer l.Close()

		for {
			line, err := l.Readline()
			if err == readline.ErrInterrupt {
				if len(line) == 0 {
					break
				} else {
					continue
				}
			} else if err == io.EOF {
				break
			} else if len(line) == 0 {
				continue
			}

			line = strings.TrimSpace(line)
			reply, err := askChatGPT(line)
			if err != nil {
				log.Println(err)
				continue
			}
			fmt.Println(reply)
		}
		return
	}

	if len(flag.Args()) != 0 {
		query := strings.Join(flag.Args(), " ")
		answer, err := askChatGPT(query)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(answer)
		return
	}

	flag.Usage()
}

func askChatGPT(query string) (reply string, err error) {

	q := gpt3.ChatCompletionRequestMessage{
		Role:    "user",
		Content: query,
	}

	session_queries = append(session_queries, q)
	ctx := context.Background()
	resp, err := client.ChatCompletion(ctx, gpt3.ChatCompletionRequest{
		Temperature:      0.5,
		Stream:           false,
		MaxTokens:        500,
		TopP:             0,
		FrequencyPenalty: 0.0,
		PresencePenalty:  0.0,
		Messages:         session_queries,
	})
	if err != nil {
		return
	}

	chunks := make([]string, 0)
	for _, choice := range resp.Choices {
		chunks = append(chunks, choice.Message.Content)
	}
	reply = strings.Join(chunks, " ")
	return
}
