package main

import "github.com/go-git/go-git/v5"
import "fmt"
import "os"
import "github.com/go-git/go-git/v5/plumbing"
import "github.com/go-git/go-git/v5/plumbing/object"
import "github.com/coreos/go-semver/semver"
import "regexp"
import "strconv"
import "flag"
import "log"

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	log.Fatalf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

func tagExists(tag string, r *git.Repository) bool {
	tagFoundErr := "tag was found"
	tags, err := r.TagObjects()
	if err != nil {
		log.Printf("get tags error: %s", err)
		return false
	}
	res := false
	err = tags.ForEach(func(t *object.Tag) error {
		if t.Name == tag {
			res = true
			return fmt.Errorf(tagFoundErr)
		}
		return nil
	})
	if err != nil && err.Error() != tagFoundErr {
		log.Printf("iterate tags error: %s", err)
		return false
	}
	return res
}

func setTag(r *git.Repository, tag string) (bool, error) {
	if tagExists(tag, r) {
		log.Printf("tag %s already exists", tag)
		return false, nil
	}
	h, err := r.Head()
	if err != nil {
		return false, err
	}

	_, err = r.CreateTag(tag, h.Hash(), &git.CreateTagOptions{
		Tagger: &object.Signature{
			Name: "Bump",
		},
		Message: tag,
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func bumpNum(n []byte) []byte {
	i, _ := strconv.ParseInt(string(n), 10, 64)
	i++
	s := strconv.FormatInt(i, 10)
	return []byte(s)
}

func bumpPreRelease(v *semver.Version) {
	// rc.13
	r, _ := regexp.Compile("(\\d+)")
	x := []byte(v.PreRelease)
	out := r.ReplaceAllFunc(x, bumpNum)
	s := semver.PreRelease(string(out))
	v.PreRelease = s
}

func usage() {
	fmt.Printf("Usage: %s [OPTIONS] [major|minor|patch]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	shouldTag := flag.Bool("tag", true, "Whether to commit the bumped tag.")
	isCandidate := flag.Bool("rc", false, "Whether to bump candidate version.")
	shouldFetch := flag.Bool("fetch", false, "Whether to fetch all tags from remote.")
	showHelp := flag.Bool("h", false, "Print Usage.")
	flag.Usage = usage

	flag.Parse()
	var level string

	if *showHelp {
		flag.Usage()
		return
	}

	if flag.NArg() > 0 {
		level = flag.Args()[0]
	} else {
		level = ""
	}

	var tags []*semver.Version
	r, err := git.PlainOpen(".git")
	CheckIfError(err)

	// fetch
	if *shouldFetch {
		err = r.Fetch(&git.FetchOptions{RemoteName: "origin", Tags: git.AllTags})
		CheckIfError(err)
	}

	tagrefs, err := r.Tags()
	CheckIfError(err)

	tagrefs.ForEach(func(t *plumbing.Reference) error {
		ref := t.Name()
		v, err := semver.NewVersion(ref.Short())
		CheckIfError(err)
		tags = append(tags, v)
		return nil
	})

	semver.Sort(tags)
	latest := tags[len(tags)-1]

	switch level {
	case "major":
		latest.BumpMajor()
	case "minor":
		latest.BumpMinor()
	case "patch":
		latest.BumpPatch()
	case "":
		break
	default:
		fmt.Printf("%s is not a valid Level.\n", level)
		os.Exit(1)
	}

	if *isCandidate {
		if latest.PreRelease == "" {
			latest.PreRelease = semver.PreRelease("rc.0")
		} else {
			bumpPreRelease(latest)
		}
	}

	log.Printf("Bumping to version: %s \n", latest.String())

	if *shouldTag {
		_, err := setTag(r, latest.String())
		CheckIfError(err)
		log.Println("Tagging")
	}

	// output the final version to stdout
	fmt.Println(latest.String())
	os.Exit(0)
}
