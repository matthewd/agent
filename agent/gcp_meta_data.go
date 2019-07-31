package agent

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"

	"cloud.google.com/go/compute/metadata"
)

type GCPMetaData struct {
}

// TODO name this better
func (e GCPMetaData) GetSuffixes(s []string) (map[string]string, error) {
	// TODO Should this be additive on top of `Get()`?
	result := make(map[string]string)

	suffixes, err := parseMetaDataValueSuffixes(s)
	if err != nil {
		return result, err
	}

	for key, suffix := range suffixes {
		value, err := metadata.Get(suffix)
		if err != nil {
			return nil, err
		} else {
			result[key] = value
		}
	}

	return result, nil
}

func (e GCPMetaData) Get() (map[string]string, error) {
	result := make(map[string]string)

	instanceId, err := metadata.Get("instance/id")
	if err != nil {
		return result, err
	}
	result["gcp:instance-id"] = instanceId

	machineType, err := machineType()
	if err != nil {
		return result, err
	}
	result["gcp:machine-type"] = machineType

	preemptible, err := metadata.Get("instance/scheduling/preemptible")
	if err != nil {
		return result, err
	}
	result["gcp:preemptible"] = strings.ToLower(preemptible)

	projectId, err := metadata.ProjectID()
	if err != nil {
		return result, err
	}
	result["gcp:project-id"] = projectId

	zone, err := metadata.Zone()
	if err != nil {
		return result, err
	}
	result["gcp:zone"] = zone

	region, err := parseRegionFromZone(zone)
	if err != nil {
		return result, err
	}
	result["gcp:region"] = region

	return result, nil
}

func machineType() (string, error) {
	machType, err := metadata.Get("instance/machine-type")
	// machType is of the form "projects/<projNum>/machineTypes/<machType>".
	if err != nil {
		return "", err
	}
	index := strings.LastIndex(machType, "/")
	if index == -1 {
		return "", errors.New("cannot parse machine-type: " + machType)
	}
	return machType[index+1:], nil
}

func parseRegionFromZone(zone string) (string, error) {
	// zone is of the form "<region>-<letter>".
	index := strings.LastIndex(zone, "-")
	if index == -1 {
		return "", errors.New("cannot parse zone: " + zone)
	}
	return zone[:index], nil
}

// TODO: Seems quite crap, dunno will ask lox
func parseMetaDataValueSuffixes(suffixes []string) (map[string]string, error) {
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
