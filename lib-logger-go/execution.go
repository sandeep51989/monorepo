package logger

import (
	"strconv"
	"time"
)

//ConfigureFromEnv will take a map of environmental variables and attempt to derive the internal configuration
func ConfigureFromEnv(envs map[string]string) {
	//get debug timer from environment
	if debugTimeString, ok := envs[EnvNameDebugTimer]; !ok {
		//use default if not found
		ConfigDebugTimer = DefaultDebugTimer
	} else {
		if debugTimeString != "" {
			//if the string is not empty, convert to integer, then to minutes
			if minutes, err := strconv.Atoi(debugTimeString); err != nil || minutes <= 0 {
				//use default time if there is an error or minutes is less or equal to 0
				ConfigDebugTimer = DefaultDebugTimer
			} else {
				ConfigDebugTimer = time.Duration(minutes) * time.Minute
			}
		} else {
			//if not found, use the default debug time
			ConfigDebugTimer = DefaultDebugTimer
		}
	}
}
