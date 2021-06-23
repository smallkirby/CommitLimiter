package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type GithubActor struct {
	Id    int64  `json:"id"`
	Login string `json:"lgoin"`
}

type GithubRepo struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type GithubAuthor struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type GithubCommit struct {
	Sha     string `json:"sha"`
	Message string `json:"message"`
	Url     string `json:"url"`
	Author  GithubAuthor
}

type GithubPush struct {
	Id      int64          `json:"push_id"`
	Size    int            `json:"size"`
	Ref     string         `json:"ref"`
	Commits []GithubCommit `json:"commits"`
}

type GithubEventPush struct {
	Id        string      `json:"id"`
	Type      string      `json:"type"`
	Actor     GithubActor `json:"actor"`
	Repo      GithubRepo  `json:"repo"`
	CreatedAt string      `json:"created_at"`
	Payload   GithubPush  `json:"payload"`
}

func fetchEventPushPage(username string, page int) ([]GithubEventPush, error) {
	// get events
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/users/%v/events?per_page=10&page=%v", username, page), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// parse as JSON
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var events []GithubEventPush
	if err := json.Unmarshal(body, &events); err != nil {
		log.Fatalln(err)
	}
	return events, nil
}

func FetchTodaysCommitAll(username string) (res_commits []GithubCommit, err error) {
	ty, tm, td := time.Now().Local().Date()
	need_continue := true
	for ix := 1; need_continue; ix++ {
		// fetch 10 events
		events, err := fetchEventPushPage(username, ix)
		if err != nil {
			return nil, err
		}
		// filter them
		for _, event := range events {
			_event_date, err := time.Parse(time.RFC3339Nano, event.CreatedAt)
			if err != nil {
				return nil, err
			}
			ey, em, ed := _event_date.Local().Date()
			if !(ty == ey && tm == em && td == ed) {
				need_continue = false
				break
			}
			if event.Type != "PushEvent" {
				continue
			}
			if event.Payload.Ref != "refs/heads/master" && event.Payload.Ref != "refs/heads/main" {
				continue
			}
			for _, commit := range event.Payload.Commits {
				res_commits = append(res_commits, commit)
			}
		}
	}
	return
}

func createConfig(username string, limit int) error {
	confdir := "/etc/smgithub"
	if _, err := os.Stat(confdir); os.IsNotExist(err) {
		if err := os.Mkdir(confdir, 0755); err != nil {
			return err
		}
		fmt.Printf("Created new config dir at %v\n", confdir)
	}
	conffile := "/etc/smgithub/setting.conf"
	if _, err := os.Stat(conffile); os.IsNotExist(err) {
		if _, err := os.Create(conffile); err != nil {
			return err
		}
		fmt.Printf("Created new config file at %v\n", conffile)
	}

	out := strings.Join([]string{
		fmt.Sprintf("USERNAME=%v", username),
		fmt.Sprintf("LIMIT=%v", limit),
	}, "\n")

	if err := ioutil.WriteFile(conffile, []byte(out), 0644); err != nil {
		return err
	}
	return nil
}

func readConfig() (string, int, error) {
	data, err := ioutil.ReadFile("/etc/smgithub/setting.conf")
	if err != nil {
		return "", 0, err
	}
	conf := strings.Split(string(data), "\n")
	if len(conf) != 2 {
		return "", 0, errors.New("Broken config file.")
	}

	username := strings.Split(conf[0], "=")[1]
	limit, _ := strconv.Atoi(strings.Split(conf[1], "=")[1])

	return username, limit, nil
}

func main() {
	// CLI interface
	flag_initialize := flag.Bool("init", false, "Init configuration.")
	flag_username := flag.String("username", "", "Github username.")
	flag_limit := flag.Int("limit", 3, "Permitted number of commits per day.")
	flag.Parse()
	if *flag_initialize {
		if *flag_username == "" {
			fmt.Println("Specify your Github username.")
			fmt.Printf("USAGE: %v --init --username <YOUR USERNAME> --limit <NUM>\n", os.Args[0])
			os.Exit(1)
		}
		createConfig(*flag_username, *flag_limit)
		fmt.Println("Successfully initialized.")
		os.Exit(0)
	}

	// read config
	username, limit, err := readConfig()
	if err != nil {
		log.Fatalln(err)
	}

	// fetch today's commits
	commits, err := FetchTodaysCommitAll(username)
	if err != nil {
		log.Fatalln(err)
	}

	// prohibit more commits
	if len(commits) >= limit {
		fmt.Println("Over threshold, prohibiting more commits...")
		if os.Geteuid() != 0 {
			fmt.Println("Need root permission.")
			os.Exit(1)
		}
		if err := DisalbleCommit(); err != nil {
			log.Fatalln(err)
		}
		fmt.Println("Prohibited github.com.")
	} else {
		fmt.Println("Allowing commits to github.com...")

		if os.Geteuid() != 0 {
			fmt.Println("Need root permission.")
			os.Exit(1)
		}

		if err := EnableCommit(); err != nil {
			log.Fatalln(err)
		}
		fmt.Println("Allowed github.com.")
		os.Exit(0)
	}
}

func EnableCommit() error {
	_hosts, err := ioutil.ReadFile("/etc/hosts")
	if err != nil {
		return err
	}
	hosts := strings.Split(string(_hosts), "\n")

	var new_hosts []string
	for _, host := range hosts {
		if strings.Contains(host, "smgithub") {
			if strings.Contains(host, "enabled") {
				new_hosts = append(new_hosts, "# 127.0.0.1 github.com # smgithub disabled")
				continue
			} else if strings.Contains(host, "disabled") {
				return nil
			} else {
				return errors.New("Broken field of 'smgithub' in /etc/hosts.")
			}
		}
		new_hosts = append(new_hosts, host)
	}
	out := strings.Join(new_hosts, "\n")
	if err := ioutil.WriteFile("/etc/hosts", []byte(out), 0000); err != nil {
		return err
	}

	return nil
}

func DisalbleCommit() error {
	_hosts, err := ioutil.ReadFile("/etc/hosts")
	if err != nil {
		return err
	}
	hosts := strings.Split(string(_hosts), "\n")

	var new_hosts []string
	status_changed := false
	for _, host := range hosts {
		if strings.Contains(host, "smgithub") {
			if strings.Contains(host, "enabled") {
				return nil
			} else if strings.Contains(host, "disabled") {
				new_hosts = append(new_hosts, "127.0.0.1 github.com # smgithub enabled")
				status_changed = true
				continue
			} else {
				return errors.New("Broken field of 'smgithub' in /etc/hosts.")
			}
		}
		new_hosts = append(new_hosts, host)
	}
	if !status_changed {
		new_hosts = append(new_hosts, "\n127.0.0.1 github.com # smgithub enabled")
	}
	out := strings.Join(new_hosts, "\n")
	if err := ioutil.WriteFile("/etc/hosts", []byte(out), 0000); err != nil {
		return err
	}

	return nil
}
