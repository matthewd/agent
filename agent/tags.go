package agent

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/buildkite/agent/logger"
	"github.com/buildkite/agent/retry"
	"github.com/denisbrodbeck/machineid"
)

type FetchTagsConfig struct {
	Tags                    []string
	TagsFromEC2MetaData     []string
	TagsFromEC2             bool
	TagsFromEC2Tags         bool
	TagsFromGCP             bool
	TagsFromGCPMetaData     []string
	TagsFromGCPLabels       bool
	TagsFromHost            bool
	WaitForEC2TagsTimeout   time.Duration
	WaitForGCPLabelsTimeout time.Duration
}

// FetchTags loads tags from a variety of sources
func FetchTags(l *logger.Logger, conf FetchTagsConfig) []string {
	tags := conf.Tags

	// Load tags from host
	if conf.TagsFromHost {
		var tags []string

		hostname, err := os.Hostname()
		if err != nil {
			l.Warn("Failed to find hostname: %v", err)
		}

		tags = append(tags,
			fmt.Sprintf("hostname=%s", hostname),
			fmt.Sprintf("os=%s", runtime.GOOS),
		)

		machineID, _ := machineid.ProtectedID("buildkite-agent")
		if machineID != "" {
			tags = append(tags, fmt.Sprintf("machine-id=%s", machineID))
		}

		return tags
	}

	// Attempt to add the EC2 tags
	if conf.TagsFromEC2 && awsSess != nil {
		l.Info("Fetching EC2 meta-data...")

		err := retry.Do(func(s *retry.Stats) error {
			ec2Tags, err := EC2MetaData{}.Get()
			if err != nil {
				l.Warn("%s (%s)", err, s)
			} else {
				l.Info("Successfully fetched EC2 meta-data")
				for tag, value := range ec2Tags {
					tags = append(tags, fmt.Sprintf("%s=%s", tag, value))
				}
				s.Break()
			}

			return err
		}, &retry.Config{Maximum: 5, Interval: 1 * time.Second, Jitter: true})

		// Don't blow up if we can't find them, just show a nasty error.
		if err != nil {
			l.Error(fmt.Sprintf("Failed to fetch EC2 meta-data: %s", err.Error()))
		}
	}

	// Attempt to add the EC2 tags
	if conf.TagsFromEC2Tags && awsSess != nil {
		l.Info("Fetching EC2 tags...")
		err := retry.Do(func(s *retry.Stats) error {
			ec2Tags, err := EC2Tags{}.Get()
			// EC2 tags are apparently "eventually consistent" and sometimes take several seconds
			// to be applied to instances. This error will cause retries.
			if err == nil && len(tags) == 0 {
				err = errors.New("EC2 tags are empty")
			}
			if err != nil {
				l.Warn("%s (%s)", err, s)
			} else {
				l.Info("Successfully fetched EC2 tags")
				for tag, value := range ec2Tags {
					tags = append(tags, fmt.Sprintf("%s=%s", tag, value))
				}
				s.Break()
			}
			return err
		}, &retry.Config{Maximum: 5, Interval: conf.WaitForEC2TagsTimeout / 5, Jitter: true})

		// Don't blow up if we can't find them, just show a nasty error.
		if err != nil {
			l.Error(fmt.Sprintf("Failed to find EC2 tags: %s", err.Error()))
		}
	}

	// Attempt to add the Google Cloud meta-data
	if conf.TagsFromGCP {
		gcpTags, err := GCPMetaData{}.Get()
		if err != nil {
			// Don't blow up if we can't find them, just show a nasty error.
			l.Error(fmt.Sprintf("Failed to fetch Google Cloud meta-data: %s", err.Error()))
		} else {
			for tag, value := range gcpTags {
				tags = append(tags, fmt.Sprintf("%s=%s", tag, value))
			}
		}
	}

	// Attempt to add the Google Cloud meta-data
	if len(conf.TagsFromGCPMetaData) > 0 {
		suffixes, err := parseTagValueSuffixePairs(conf.TagsFromGCPMetaData)
		if err != nil {
			// TODO: Would be good to log a nice message here, need to check this
			l.Error(fmt.Sprintf("TODO: %s", err.Error()))
		}

		gcpTags, err := GCPMetaData{}.GetSuffixes(suffixes)
		if err != nil {
			// Don't blow up if we can't find them, just show a nasty error.
			l.Error(fmt.Sprintf("Failed to fetch Google Cloud meta-data: %s", err.Error()))
		} else {
			for tag, value := range gcpTags {
				tags = append(tags, fmt.Sprintf("%s=%s", tag, value))
			}
		}

	}

	// Attempt to add the Google Compute instance labels
	if conf.TagsFromGCPLabels {
		l.Info("Fetching GCP instance labels...")
		err := retry.Do(func(s *retry.Stats) error {
			labels, err := GCPLabels{}.Get()
			if err == nil && len(labels) == 0 {
				err = errors.New("GCP instance labels are empty")
			}
			if err != nil {
				l.Warn("%s (%s)", err, s)
			} else {
				l.Info("Successfully fetched GCP instance labels")
				for label, value := range labels {
					tags = append(tags, fmt.Sprintf("%s=%s", label, value))
				}
				s.Break()
			}
			return err
		}, &retry.Config{Maximum: 5, Interval: conf.WaitForGCPLabelsTimeout / 5, Jitter: true})

		// Don't blow up if we can't find them, just show a nasty error.
		if err != nil {
			l.Error(fmt.Sprintf("Failed to find GCP instance labels: %s", err.Error()))
		}
	}

	return tags
}

// TODO: Seems maybe el crapola, dunno will ask lox
func parseTagValueSuffixePairs(suffixes []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, pair := range suffixes {
		x := strings.Split(pair, "=")
		// TODO: Should we just let people have stupid keys? Probably?
		key := strings.ToLower(strings.Trim(x[0], " "))

		uri, err := url.Parse(x[1])
		if err != nil {
			return result, err
		}

		meta_data_path := filepath.Clean(uri.Path)

		if filepath.IsAbs(meta_data_path) {
			meta_data_path, err = filepath.Rel("/", meta_data_path)
			if err != nil {
				return result, err
			}
		}

		result[key] = meta_data_path
	}

	return result, nil
}
