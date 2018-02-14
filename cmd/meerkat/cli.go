package meerkat

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ahmdrz/goinsta"
	"gopkg.in/yaml.v2"
)

func exists(path string) bool {
	_, err := os.Stat(path)
	return os.IsExist(err)
}

type Meerkat struct {
	Interval    int
	SleepTime   int
	Username    string
	Password    string
	TargetUsers []string
	OutputType  string

	TelegramToken string
	TelegramUser  int

	instagram     *goinsta.Instagram
	logger        *log.Logger
	lastTimeStamp int
	targetUsers   map[int64]User
	login         bool
	loggerFile    *os.File
}

type User struct {
	Username  string
	Bio       string
	Followers int
	Following int
	Posts     int
	Tags      int
}

func (m *Meerkat) parseArgs() error {
	if len(os.Args) > 1 {
		// meerkat init
		// meerkat init config.yaml
		if os.Args[1] == "init" {
			configFile := "meerkat.yaml"
			if len(os.Args) > 2 {
				configFile = os.Args[2]
			}
			file, err := os.OpenFile(configFile, os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			file.WriteString(configTemplate)
			file.Close()

			fmt.Println(configFile, "generated.")

			os.Exit(0)
		}
	}

	configFile := "meerkat.yaml"

	outputPtr := flag.String("output", "", "Log output file.")
	configPtr := flag.String("config", "", "Configuration file (YAML format)")

	flag.Parse()

	if *outputPtr == "" {
		m.logger = log.New(os.Stdout, "[meerkat] ", log.Ldate|log.Ltime)
	} else {
		file, err := os.OpenFile(*outputPtr, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return fmt.Errorf("output file [%s] does not exists", *outputPtr)
		}
		m.loggerFile = file
		m.logger = log.New(file, "[meerkat] ", log.Ldate|log.Ltime)
	}

	if *configPtr != "" {
		configFile = *configPtr
	}

	file, err := os.Open(configFile)
	if err != nil {
		return fmt.Errorf("config file [%s] does not exists", configFile)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(bytes, m)
	if err != nil {
		return err
	}

	return nil
}

func (m *Meerkat) Run(done chan bool) error {
	m.logger.Println("Logging in to the Instagram")

	m.instagram = goinsta.New(m.Username, m.Password)
	err := m.instagram.Login()
	if err != nil {
		return fmt.Errorf("Instagram error , %s", err.Error())
	}
	m.login = true

	m.logger.Println("Successfully logged in")

	select {
	case <-done:
		return fmt.Errorf("Signal on meerkat !")
	default:
		for _, username := range m.TargetUsers {
			m.logger.Printf("Getting %s information ", username)

			user, err := m.instagram.GetUserByUsername(username)
			if err != nil {
				return err
			}
			m.targetUsers[user.User.ID] = User{
				Username:  username,
				Followers: user.User.FollowerCount,
				Following: user.User.FollowingCount,
				Bio:       user.User.Biography,
				Posts:     user.User.MediaCount,
				Tags:      user.User.UserTagsCount,
			}

			m.logger.Printf("User %s-%d information has been retrived successfully.", username, user.User.ID)
		}
	}

	m.logger.Println("Starting watcher ...")

	var failure int = 0
	var exitErr error

	// TODO : make select statement better !
	tick := time.After(time.Duration(m.Interval) * time.Second)

	for failure != 3 {
		select {
		case <-done:
			return fmt.Errorf("Signal on meerkat !")
		case <-tick:
			tick = time.After(time.Duration(m.Interval) * time.Second)

			m.logger.Println("Sending request to get following activities")

			resp, err := m.instagram.GetFollowingRecentActivity()
			if err != nil {
				m.logger.Println("Error", err)
				failure++
				exitErr = err
				continue
			}

			// to find last time stamp
			maxTimeStamp := int(0)
			for _, story := range resp.Stories {
				unixTimeStamp := story.Args.Timestamp

				if unixTimeStamp <= m.lastTimeStamp {
					continue
				}

				for _, link := range story.Args.Links {
					if link.Type == "user" {
						userID, _ := strconv.ParseInt(link.ID, 10, 64)

						if user, ok := m.targetUsers[userID]; ok {
							unixTime := time.Unix(int64(unixTimeStamp), 0)

							message := fmt.Sprintf("[%s] [%s] %s\n", user.Username, unixTime.Format("15:04:05"), story.Args.Text)

							// TODO: parse to array of string and search over it.
							if strings.Contains(m.OutputType, "telegram") {
								m.sendToTelegram(m.TelegramUser, message)
							}
							if strings.Contains(m.OutputType, "logfile") {
								m.logger.Println(message)
							}
						}
					}
				}

				if unixTimeStamp > maxTimeStamp {
					maxTimeStamp = unixTimeStamp
				}
			}
			if maxTimeStamp != 0 {
				m.lastTimeStamp = maxTimeStamp
			}

			failure = 0

			for _, username := range m.TargetUsers {
				m.logger.Printf("Getting %s information ", username)

				user, err := m.instagram.GetUserByUsername(username)
				if err != nil {
					m.logger.Println("Error", err)
					failure++
					exitErr = err
					continue
				}
				tmpUser := m.targetUsers[user.User.ID]
				currentTime := time.Now().Format("15:04:05")

				message := fmt.Sprintf("[%s] [%s] : \n", username, currentTime)
				hasMessage := false

				if user.User.Biography != tmpUser.Bio {
					message += fmt.Sprintf("User %s biography changed to %s\n", username, user.User.Biography)
					hasMessage = true

					tmpUser.Bio = user.User.Biography
				}
				if user.User.FollowerCount != tmpUser.Followers {
					message += fmt.Sprintf("User %s followers changed from %d to %d\n", username, tmpUser.Followers, user.User.FollowerCount)
					hasMessage = true

					tmpUser.Followers = user.User.FollowerCount
				}
				if user.User.FollowingCount != tmpUser.Following {
					message += fmt.Sprintf("User %s following changed from %d to %d\n", username, tmpUser.Following, user.User.FollowingCount)
					hasMessage = true

					tmpUser.Following = user.User.FollowingCount
				}
				if user.User.FollowingCount != tmpUser.Following {
					message += fmt.Sprintf("User %s posts changed from %d to %d\n", username, tmpUser.Posts, user.User.MediaCount)
					hasMessage = true

					tmpUser.Posts = user.User.MediaCount
				}
				if user.User.UserTagsCount != tmpUser.Tags {
					message += fmt.Sprintf("User %s tags changed from %d to %d\n", username, tmpUser.Tags, user.User.UserTagsCount)
					hasMessage = true

					tmpUser.Tags = user.User.UserTagsCount
				}

				if hasMessage {
					m.targetUsers[user.User.ID] = tmpUser

					// TODO: parse to array of string and search over it.
					if strings.Contains(m.OutputType, "telegram") {
						m.sendToTelegram(m.TelegramUser, message)
					}
					if strings.Contains(m.OutputType, "logfile") {
						m.logger.Println(message)
					}
				}

				m.logger.Printf("User %s information has been updated successfully.", username)

				time.Sleep(time.Duration(m.SleepTime) * time.Second)
			}
		}
	}

	if failure == 3 {
		return exitErr
	}

	return nil
}

func (m *Meerkat) Logout() error {
	if m.login {
		return m.instagram.Logout()
	}
	if m.loggerFile != nil {
		return m.loggerFile.Close()
	}
	return nil
}

func New() (*Meerkat, error) {
	m := &Meerkat{}
	m.targetUsers = make(map[int64]User)
	m.loggerFile = nil

	if err := m.parseArgs(); err != nil {
		return nil, err
	}

	if len(m.TargetUsers) == 0 {
		return nil, fmt.Errorf("There is no targetusers in yaml config file")
	}

	if m.OutputType == "" {
		return nil, fmt.Errorf("Fill outputtype with ['logfile', 'telegram']")
	}

	if m.Interval < 10 {
		log.Println("Interval is low, try more than 10 seconds.")
	}

	if m.SleepTime < 10 {
		log.Println("SleepTime is low, try more than 10 seconds.")
	}

	return m, nil
}

func (m *Meerkat) sendToTelegram(to int, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%d&text=%s", m.TelegramToken, to, message)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var output struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}

	json.Unmarshal(bytes, &output)

	if !output.Ok {
		return fmt.Errorf("Telegram bot %s", output.Description)
	}

	return nil
}
