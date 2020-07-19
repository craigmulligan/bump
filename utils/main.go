package utils

import (
	"fmt"
	"github.com/coreos/go-semver/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"log"
	"regexp"
	"strconv"
	"time"
)


func tagExists(tag string, r *git.Repository) bool {
	tagFoundErr := "tag was found"
	tags, err := r.Tags()
	if err != nil {
		log.Printf("get tags error: %s", err)
		return false
	}
	res := false
	err = tags.ForEach(func(t *plumbing.Reference) error {
		if t.Name().Short() == tag {
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

func SetTag(r *git.Repository, tag string) (bool, error) {
	if tagExists(tag, r) {
		return false, fmt.Errorf("tag %s already exists", tag)
	}
	h, err := r.Head()
	if err != nil {
		return false, err
	}

	cfg, err := r.ConfigScoped(config.SystemScope)
	if err != nil {
		return false, err
	}

	_, err = r.CreateTag(tag, h.Hash(), &git.CreateTagOptions{
		Tagger: &object.Signature{
			Name:  cfg.User.Name,
			Email: cfg.User.Email,
			When:  time.Now(),
		},
		Message: tag,
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func Push(r *git.Repository, tag string) error {
  return r.Push(&git.PushOptions{
    RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec("+refs/tags/" + tag +":refs/tags/" + tag),
		},
  })
}

func bumpNum(n []byte) []byte {
	i, _ := strconv.ParseInt(string(n), 10, 64)
	i++
	s := strconv.FormatInt(i, 10)
	return []byte(s)
}

func bumpPreRelease(v semver.Version) semver.Version {
	r, _ := regexp.Compile("(\\d+)")
	x := []byte(v.PreRelease)
	out := r.ReplaceAllFunc(x, bumpNum)
	s := semver.PreRelease(string(out))
	v.PreRelease = s
	return v
}

func GetTags(r git.Repository) ([]*semver.Version, error) {
  var tags []*semver.Version

	tagrefs, err := r.Tags()
	if err != nil {
		return nil, err
	}

	tagrefs.ForEach(func(t *plumbing.Reference) error {
		ref := t.Name()
		v, err := semver.NewVersion(ref.Short())
		if err != nil {
			return nil
		}
		tags = append(tags, v)
		return nil
	})

	if len(tags) == 0 {
		v, err := semver.NewVersion("0.0.0")
    if err != nil {
      return nil, err
    }
		tags = append(tags, v)
	}

  return tags, nil
}

func Latest(tags []*semver.Version) (*semver.Version, error) {
  // Gets the latest tag from a repo.
	semver.Sort(tags)
	latest := tags[len(tags)-1]
	return latest, nil
}

func IsCandidate(v semver.Version) bool {
  if (v.PreRelease != "") {
    return true
  }
  return false
}

type Level string

const (
	Major Level = "major"
	Minor       = "minor"
	Patch       = "patch"
  Candidate   = "rc"
  Promote     = "promote"
	Noop        = "noop"
)

func Bump(v semver.Version, level Level) (semver.Version, error) {
	switch level {
	case Major:
		v.BumpMajor()
	case Minor:
		v.BumpMinor()
	case Patch:
		v.BumpPatch()
  case Candidate:
    if IsCandidate(v) {
      v = bumpPreRelease(v)
    } else {
      v.PreRelease = semver.PreRelease("rc.0")
    }
  case Promote:
    v.PreRelease = ""
  case Noop:
	default:
		invalidLevelError := fmt.Errorf("%s is not a valid level", level)
    return v, invalidLevelError
	}

	return v, nil
}
