package cronetheus

import (
	"fmt"
	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"reflect"
	"strings"
	"sync"
	"time"
)

// CronSchedule struct defines the default running schedule configuration for the cron jobs (i.e */2 * * * *)
type CronSchedule struct {
	Second     string                 `yaml:"second"`
	Minute     string                 `yaml:"minute"`
	Hour       string                 `yaml:"hour"`
	DayOfMonth string                 `yaml:"day_of_month"`
	Month      string                 `yaml:"month"`
	DayOfWeek  string                 `yaml:"day_of_week"`
	XXX        map[string]interface{} `yaml:",inline"` // Catches all undefined fields and must be empty after parsing.
}

// JobConfig struct defines the fields needed to run/schedule a cron job
type JobConfig struct {
	CronID     string                 `yaml:"cron_id"`
	Descriptor string                 `yaml:"descriptor,omitempty"`
	User       string                 `yaml:"user"`
	Binary     string                 `yaml:"binary"`
	Args       string                 `yaml:"args,omitempty"`
	Schedule   CronSchedule           `yaml:"schedule,omitempty"`
	XXX        map[string]interface{} `yaml:",inline"` // Catches all undefined fields and must be empty after parsing.
}

// Config struct consists of JobConfigs
type Config struct {
	CronConfig []JobConfig            `yaml:"cron_config"`
	XXX        map[string]interface{} `yaml:",inline"` // Catches all undefined fields and must be empty after parsing.
	sync.Mutex
}

// DefaultCronSchedule is a CronSchedule with "*" on all fields
var DefaultCronSchedule = CronSchedule{
	Second:     "*",
	Minute:     "*",
	Hour:       "*",
	DayOfMonth: "*",
	Month:      "*",
	DayOfWeek:  "*",
}

// EmptyCronSchedule is a CronSchedule with empty fields
var EmptyCronSchedule = CronSchedule{
	Second:     "",
	Minute:     "",
	Hour:       "",
	DayOfMonth: "",
	Month:      "",
	DayOfWeek:  "",
}

// Equals compares two CronSchedules
func (cs *CronSchedule) Equals(other *CronSchedule) bool {
	return (cs.Second == other.Second && cs.Minute == other.Minute && cs.Hour == other.Hour && cs.DayOfMonth == other.DayOfMonth && cs.Month == other.Month && cs.DayOfWeek == other.DayOfWeek)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for JobConfig.
func (jc *JobConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain JobConfig
	if err := unmarshal((*plain)(jc)); err != nil {
		return err
	}

	// if both descriptor and schedule for a cron job is empty return error
	if jc.Descriptor == "" && jc.Schedule.Equals(&EmptyCronSchedule) {
		return fmt.Errorf("Both descriptor and schedule is not set")
	}

	// if given descriptor is not valid (check unix man page for cron) return error
	validDescriptors := []string{"@yearly", "@annually", "@monthly", "@weekly", "@daily", "@midnight", "@hourly"}

	if jc.Descriptor != "" && !SliceExists(validDescriptors, jc.Descriptor) {
		// if given descriptor is @every, a duration should be given as well
		// if given duration for @every descriptor is below a second return an error
		if strings.HasPrefix(jc.Descriptor, "@every") {
			glog.V(1).Infof("Cron job %s has @every descriptor", jc.CronID)
			if params := strings.Split(jc.Descriptor, " "); len(params) < 2 {
				return fmt.Errorf("A valid time duration should be specified after @every descriptor")
			}
			second, _ := time.ParseDuration("1s")
			timeParam, timeErr := time.ParseDuration(strings.Split(jc.Descriptor, " ")[1])
			if timeErr != nil {
				return timeErr
			}
			if second > timeParam {
				return fmt.Errorf("given time parameter should be greater than 1 seconds: %s", jc.Descriptor)
			}
		} else {
			return fmt.Errorf("descriptor %s is not valid", jc.Descriptor)
		}
	}

	// if required fields are empty, fill them with default values and log error
	if jc.CronID == "" {
		return fmt.Errorf("CronID field cannot be empty")
	}
	if jc.User == "" {
		return fmt.Errorf("User field cannot be empty")
	}
	if jc.Binary == "" {
		return fmt.Errorf("Binary field cannot be empty")
	}

	// check if there are extra unknown fields
	return checkOverflow(jc.XXX, "job_config")
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for CronSchedule.
func (cs *CronSchedule) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain CronSchedule
	if err := unmarshal((*plain)(cs)); err != nil {
		return err
	}
	// if given CronSchedule is empty set it to default CronSchedule
	if cs.Second == "" {
		glog.V(1).Infof("Second field is empty assigning a default value to it")
		cs.Second = DefaultCronSchedule.Second
	}
	if cs.Minute == "" {
		glog.V(1).Infof("Minute field is empty assigning a default value to it")
		cs.Minute = DefaultCronSchedule.Minute
	}
	if cs.Hour == "" {
		glog.V(1).Infof("Hour field is empty assigning a default value to it")
		cs.Hour = DefaultCronSchedule.Hour
	}
	if cs.DayOfMonth == "" {
		glog.V(1).Infof("DayOfMonth field is empty assigning a default value to it")
		cs.DayOfMonth = DefaultCronSchedule.DayOfMonth
	}
	if cs.Month == "" {
		glog.V(1).Infof("Month field is empty assigning a default value to it")
		cs.Month = DefaultCronSchedule.Month
	}
	if cs.DayOfWeek == "" {
		glog.V(1).Infof("DayOfWeek field is empty assigning a default value to it")
		cs.DayOfWeek = DefaultCronSchedule.DayOfWeek
	}

	// check if there are extra unknown fields
	return checkOverflow(cs.XXX, "cron_schedule")
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	// make sure cron_id's are unique, since they will be exposed as labels in metrics
	ids := make(map[string]int)
	for k := range c.CronConfig {
		if _, ok := ids[c.CronConfig[k].CronID]; ok {
			return fmt.Errorf("Error CronID's must be unique, found duplicate for: %q", c.CronConfig[k].CronID)
		}
		ids[c.CronConfig[k].CronID] = 0
	}
	// check if there are extra unknown fields
	return checkOverflow(c.XXX, "config")
}

// Init function loads the config file
func (c *Config) Init(filename string) error {
	c.Lock()
	defer c.Unlock()
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		glog.Errorf("Error reading config file: %q", filename)
		return err
	}
	err = yaml.Unmarshal([]byte(string(content)), c)
	if err != nil {
		return err
	}
	return nil
}

// checkOverflow func to check if there are extra unknown fields
func checkOverflow(m map[string]interface{}, ctx string) error {
	if len(m) > 0 {
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		return fmt.Errorf("unknown fields found in %s: %s", ctx, strings.Join(keys, ","))
	}
	return nil
}

// SliceExists func to check if given item exists in given slice
func SliceExists(slice interface{}, item interface{}) bool {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("SliceExists() given a non-slice type")
	}
	for i := 0; i < s.Len(); i++ {
		if s.Index(i).Interface() == item {
			return true
		}
	}
	return false
}

func (c Config) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("Error creating config string: %s", err)
	}
	return string(b)
}
