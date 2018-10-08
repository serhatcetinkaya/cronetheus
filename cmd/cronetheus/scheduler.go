package main

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/robfig/cron"
	"github.com/serhatck/cronetheus"
	"io/ioutil"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

func Schedule(c *cronetheus.Config) (*cron.Cron, error) {
	cr := cron.New()
	// add a func to the cron.Cron object for every block in the config
	for _, v := range c.CronConfig {
		if v.Descriptor != "" {
			go func(descriptor string, cronId string, user string, binary string, args string) {
				cr.AddFunc(descriptor, func() { jobTemplate(cronId, user, binary, args) })
			}(v.Descriptor, v.CronID, v.User, v.Binary, v.Args)
			glog.V(1).Infof("Job added, CronID: %q", v.CronID)
		} else if formatSchedule(&v.Schedule) != "" {
			go func(schedule string, cronId string, user string, binary string, args string) {
				cr.AddFunc(schedule, func() { jobTemplate(cronId, user, binary, args) })
			}(formatSchedule(&v.Schedule), v.CronID, v.User, v.Binary, v.Args)
			glog.V(1).Infof("Job added, CronID: %q", v.CronID)
		} else {
			return nil, fmt.Errorf("Something is wrong with schedule/descriptor of the cron job")
		}
	}
	return cr, nil
}

// get userid and groupid from unix username
func getUID(uname string) (string, string, error) {
	user, err := user.Lookup(uname)
	if err != nil {
		return "", "", err
	}
	glog.Infof("Found user %q, UID: %s, GID, %s", uname, user.Uid, user.Gid)
	return user.Uid, user.Gid, nil
}

func jobTemplate(cron_id string, uname string, binary string, args string) {
	uid, gid, idErr := getUID(uname)
	if idErr != nil {
		glog.Errorf("Error getting userID: %q", idErr)
	}

	uidInt, _ := strconv.ParseUint(uid, 10, 32)
	gidInt, _ := strconv.ParseUint(gid, 10, 32)

	// check the executable path for given binary
	path, lookErr := exec.LookPath(binary)
	if lookErr != nil {
		glog.Errorf("Error searching for executable path for binary %s, %q", binary, lookErr)
	}
	glog.Infof("Found executable path %q for binary %q", path, binary)

	// prepare command with an error pipe, and SysProcAttr uid and gid
	fullCommand := fmt.Sprintf("%s %s", path, args)
	cmd := exec.Command("/bin/bash", "-c", fullCommand)
	stderrIn, _ := cmd.StderrPipe()
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uidInt), Gid: uint32(gidInt)}
	cmd.Start()

	cmdErrBytes, _ := ioutil.ReadAll(stderrIn)
	if len(cmdErrBytes) > 0 {
		glog.Errorf("Cron %s returned errors: %q", cron_id, string(cmdErrBytes))
		failedCronJobs.WithLabelValues(cron_id, uname).Inc()
	}

	cmdErr := cmd.Wait()
	if cmdErr != nil {
		glog.Errorf("Error running %s, %q", cron_id, cmdErr)
	}
}

func formatSchedule(cs *cronetheus.CronSchedule) string {
	return fmt.Sprintf("%s %s %s %s %s %s", cs.Second, cs.Minute, cs.Hour, cs.DayOfMonth, cs.Month, cs.DayOfWeek)
}
