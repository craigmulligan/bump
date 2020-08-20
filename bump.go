package main

import (
	"github.com/hobochild/bump/utils"
	"flag"
	"fmt"
	"github.com/go-git/go-git/v5"
	"log"
	"os"
	"github.com/coreos/go-semver/semver"
  "bufio"
  "strings"
)

func prompt(msg string) (string) {
  reader := bufio.NewReader(os.Stdin)
  fmt.Print(msg, ": ")

  for {
    text, _ := reader.ReadString('\n')
    // convert CRLF to LF
    text = strings.Replace(text, "\n", "", -1)
    fmt.Print("\r")
    return text
  }
}

func defaultYes(input string) bool {
  return input == "y" || input == ""
}

func requestTag(latest semver.Version) (bool) {
  msg := fmt.Sprintf("Would you like to tag %s - [Y/n]", latest.String())
  input := strings.ToLower(prompt(msg))
  return defaultYes(input)
}

func requestPush(latest semver.Version) (bool) {
  msg := fmt.Sprintf("Would you like to push %s - [Y/n]", latest.String())
  input := strings.ToLower(prompt(msg))


  return defaultYes(input)
}

func requestPromote(latest semver.Version) (utils.Level) {
  msg := fmt.Sprintf("Current tag: %s - would you like to promote or bump [B/p]", latest.String())
  input := strings.ToLower(prompt(msg))
  if (input == "p") {
    level := utils.Level("promote")
    return level
  }
  return utils.Level("rc")
}


func requestCandidate(latest semver.Version) (utils.Level) {
  msg := fmt.Sprintf("Current tag: %s - would you release as a candidate first [Y/n]", latest.String())
  input := strings.ToLower(prompt(msg))
  if (defaultYes(input)) {
    level := utils.Level("rc")
    return level
  }
  return utils.Level("noop")
}

func requestBump(latest semver.Version) (utils.Level) {
  msg := fmt.Sprintf("Current tag: %s - how would you like to bump [major/minor/PATCH]", latest.String())
  input := strings.ToLower(prompt(msg))
  if input == "" {
    return utils.Level("patch")
  }
  return utils.Level(input)
}

// CheckIfError should be used to naively panics if an error is not nil.
func checkIfError(err error) {
	if err == nil {
		return
	}

  log.Fatalf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

func main() {
	bumpVersion := "0.0.0"
  level := utils.Level("noop")
	shouldFetch := flag.Bool("fetch", true, "Whether to fetch all tags from remote.")
	dir := flag.String("dir", ".git", "Path to the git directory.")
	showHelp := flag.Bool("h", false, "Print Usage.")
	showVersion := flag.Bool("v", false, "Print Version.")
	flag.Parse()

	if *showHelp {
		flag.Usage()
		return
	}

	if *showVersion {
		fmt.Println(bumpVersion)
		return
	}


	r, err := git.PlainOpen(*dir)
	checkIfError(err)

	if *shouldFetch {
		err = r.Fetch(&git.FetchOptions{RemoteName: "origin", Tags: git.AllTags, Progress: os.Stderr})

		if err != git.NoErrAlreadyUpToDate {
			checkIfError(err)
		}
	}

  tags, err := utils.GetTags(*r)
	checkIfError(err)

	latest, err := utils.Latest(tags)
	checkIfError(err)
  var bumped semver.Version

  if utils.IsCandidate(*latest) {
    level = requestPromote(*latest)
    bumped, err = utils.Bump(*latest, level)
  } else {
    // check what type of bump we want.
    level = requestBump(*latest)
    bumped, err = utils.Bump(*latest, level)
    checkIfError(err)
    
    // now check if should be a rc.
    level := requestCandidate(bumped)
    bumped, err = utils.Bump(bumped, level)
    checkIfError(err)
  }

  shouldTag := requestTag(bumped)
  checkIfError(err)

	if shouldTag {
		_, err = utils.SetTag(r, bumped.String())
		checkIfError(err)
    shouldPush := requestPush(bumped)
		fmt.Println(fmt.Sprintf("Tagged version: %s", bumped.String()))
    checkIfError(err)

    if (shouldPush) {
		  fmt.Println(fmt.Sprintf("Pushed version: %s", bumped.String()))
      err = utils.Push(r, bumped.String())
      checkIfError(err)
    }
	}

	os.Exit(0)
}
