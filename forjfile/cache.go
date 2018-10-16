package forjfile

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"forjj/utils"
)

// Cache represents the application change status InMem Forjfile.
// Used to store and restore for process acceleration.
type Cache struct {
	cacheDir   string
	cacheName  string
	Data       *DeployForgeYaml
	RequestMD5 [16]byte
}

// SetCacheName define the cache file name to use
func (c *Cache) SetName(name string) {
	if c == nil {
		return
	}
	c.cacheName = name
}

// SetCacheDir define the Cache directory
func (c *Cache) SetCacheDir(cachePath string) error {
	if c == nil {
		return nil
	}
	if cachePath == "" {
		return nil
	}
	if v, err := utils.Abs(cachePath); err != nil {
		return err
	} else {
		c.cacheDir = v
	}
	return nil
}

// BuildMD5 build the MD5 sum of the data given.
// It returns the
func (c *Cache) BuildMD5(data interface{}) (result [16]byte, err error) {
	var jsonData []byte

	jsonData, err = json.Marshal(data) 
	if err != nil {
		return 
	}
	result = md5.Sum(jsonData)
	return
}

// SaveCache save the Forjfile in mem to the cache dir under the `name` given
// The request md5 is saved at the same time
func (c *Cache) SaveCache() error {
	if c == nil {
		return nil
	}
	if c.cacheDir == "" {
		return fmt.Errorf("Internal error: Missing properly call to SetCacheDir")
	}

	return nil
}

// LoadCache load the Forjfile in mem previously cached from the name given
func (c *Cache) LoadCache() error {
	return nil
}